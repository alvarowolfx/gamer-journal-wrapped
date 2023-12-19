package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/imagegen"
	"github.com/alvarowolfx/gamer-journal-wrapped/src/util"
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	AssetsFolder = "./assets/"
	OutFolder    = "./out/"
)

var (
	mysqlDSN     = "root:@/gaming_journal?parseTime=true"
	serperAPIKey string
	year         int
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to read .env: %v \n", err)
	}

	serperAPIKey = os.Getenv("SERPER_API_KEY")
	imagegen.LoadFonts()

	db, err := sqlx.Connect("mysql", mysqlDSN)
	if err != nil {
		log.Fatal(err)
	}

	flag.IntVar(&year, "year", 2023, "year to render gamer wrapped")
	flag.Parse()

	yearStr := fmt.Sprintf("%d", year)

	fmt.Println("Rendering wrapped for", yearStr)

	mostPlayedConsoles := []MostPlayedByPlaytime{}
	err = db.Select(&mostPlayedConsoles, `
	select c.name as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
	from playthroughs p
		inner join consoles c on JSON_CONTAINS(p.console, CONCAT('"', c.record_id, '"'))
	where p.year_start_date = ?
	group by c.name
	order by playtime desc`, yearStr)
	if err != nil {
		log.Fatalf("failed to query most played console: %v", err)
	}
	renderAndSaveAllMostPlayedWrapped("Most played consoles in "+yearStr, mostPlayedConsoles)

	mostPlayedPlatform := []MostPlayedByPlaytime{}
	err = db.Select(&mostPlayedPlatform, `
	select pt.name as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
	from playthroughs p	
		inner join games g on JSON_CONTAINS(p.games, CONCAT('"', g.record_id, '"'))
		inner join platforms pt on JSON_CONTAINS(g.platforms, CONCAT('"', pt.record_id, '"'))
	where p.year_start_date = ? 
	group by pt.name
	order by playtime desc;`, yearStr)
	if err != nil {
		log.Fatalf("failed to query most played platform: %v", err)
	}
	renderAndSaveNMostPlayedWrapped("Most played platform in "+yearStr, mostPlayedPlatform, 9)

	mostPlayedGames := []MostPlayedGame{}
	err = db.Select(&mostPlayedGames, `
	select g.name as title, pt.name as platform, c.name as console, p.playtime/(60*60) as playtime
	from playthroughs p	
		inner join games g on JSON_CONTAINS(p.games, CONCAT('"', g.record_id, '"'))
		inner join consoles c on JSON_CONTAINS(p.console, CONCAT('"', c.record_id, '"'))
		inner join platforms pt on JSON_CONTAINS(g.platforms, CONCAT('"', pt.record_id, '"'))
	where p.year_start_date = ?
	order by playtime desc;`, yearStr)
	if err != nil {
		log.Fatalf("failed to query most played games: %v", err)
	}
	renderAndSaveNMostPlayedWrapped("Most played games in "+yearStr, mostPlayedGames, 8)

	mostPlayedGameSerie := []MostPlayedByPlaytime{}
	err = db.Select(&mostPlayedGameSerie, `
	select s.name as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
	from playthroughs p	
		inner join games g on JSON_CONTAINS(p.games, CONCAT('"', g.record_id, '"'))
		inner join serie s on JSON_CONTAINS(g.serie, CONCAT('"', s.record_id, '"'))
	where p.year_start_date = ?
	group by s.name
	order by playtime desc;`, yearStr)
	if err != nil {
		log.Fatalf("failed to query most played game serie: %v", err)
	}
	renderAndSaveNMostPlayedWrapped("Most played game serie in "+yearStr, mostPlayedGameSerie, 8)

	gamesByStatus := []MostPlayedByNumGames{}
	err = db.Select(&gamesByStatus, `
	select p.status as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
	from playthroughs p	
	where p.year_start_date = ?
		and p.status not in ('Playing')
	group by p.status
	order by count desc;`, yearStr)
	if err != nil {
		log.Fatalf("failed to query most played games: %v", err)
	}
	renderAndSaveAllMostPlayedWrapped("Games beaten in "+yearStr, gamesByStatus)

	busiestMonth := []MostPlayedByPlaytime{}
	err = db.Select(&busiestMonth, `
	select EXTRACT(MONTH from p.start_date) as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
		from playthroughs p	
	where p.year_start_date = ?
	group by title
	order by title asc;
	`, yearStr)
	if err != nil {
		log.Fatalf("failed to query busiest month: %v", err)
	}
	busiestMonthData := make([]imagegen.BarChartItem, len(busiestMonth))
	for i, d := range busiestMonth {
		monthNum, _ := strconv.ParseInt(d.Title, 10, 64)
		d.Title = time.Month(int(monthNum)).String()
		d.NoIcon = true
		busiestMonthData[i] = d
	}
	renderAndSaveAllMostPlayedWrapped("Busiest months in "+yearStr, busiestMonthData)
}

func renderAndSaveNMostPlayedWrapped[T imagegen.BarChartItem](title string, data []T, n int) {
	imagegen.RenderMostPlayedWrapped(title, toBarChartItems(data), n).
		SavePNG(fmt.Sprintf("%s/%s.png", OutFolder, util.ToSnakecase(title)))
}

func renderAndSaveAllMostPlayedWrapped[T imagegen.BarChartItem](title string, data []T) {
	imagegen.RenderMostPlayedWrapped(title, toBarChartItems(data), len(data)).
		SavePNG(fmt.Sprintf("%s/%s.png", OutFolder, util.ToSnakecase(title)))
}

func toBarChartItems[T imagegen.BarChartItem](arr []T) []imagegen.BarChartItem {
	narr := make([]imagegen.BarChartItem, len(arr))
	for i, d := range arr {
		narr[i] = d
	}
	return narr
}

type MostPlayedByPlaytime struct {
	Title    string
	Playtime int
	Count    int
	NoIcon   bool
}

func (mp MostPlayedByPlaytime) GetTitle() string {
	if mp.Count == 1 {
		return fmt.Sprintf("%s (%d game)", mp.Title, mp.Count)
	}
	return fmt.Sprintf("%s (%d games)", mp.Title, mp.Count)
}

func (mp MostPlayedByPlaytime) GetMetric() int {
	return mp.Playtime
}

func (mp MostPlayedByPlaytime) RenderMetric() string {
	return fmt.Sprintf("%dh", mp.Playtime)
}

func (mp MostPlayedByPlaytime) RenderIcon(height uint) image.Image {
	if mp.NoIcon {
		return nil
	}
	icon := loadIconForName(mp.Title, false)
	if icon == nil {
		return nil
	}
	return imagegen.AutoResizeImage(height, icon)
}

type MostPlayedByNumGames struct {
	Title    string
	Playtime int
	Count    int
}

func (mp MostPlayedByNumGames) GetTitle() string {
	return mp.Title
}

func (mp MostPlayedByNumGames) GetMetric() int {
	return mp.Count
}

func (mp MostPlayedByNumGames) RenderMetric() string {
	return fmt.Sprintf("%d", mp.Count)
}

func (mp MostPlayedByNumGames) RenderIcon(height uint) image.Image {
	return nil
}

type MostPlayedGame struct {
	Title    string
	Platform string
	Console  string
	Playtime int
}

func (mpg MostPlayedGame) GetTitle() string {
	return fmt.Sprintf("%s (%s) on %s", mpg.Title, mpg.Platform, mpg.Console)
}

func (mpg MostPlayedGame) GetMetric() int {
	return mpg.Playtime
}

func (mpg MostPlayedGame) RenderMetric() string {
	return fmt.Sprintf("%dh", mpg.Playtime)
}

func (mpg MostPlayedGame) RenderIcon(height uint) image.Image {
	icon := loadIconForName(mpg.Title, true)
	if icon == nil {
		return nil
	}
	return imagegen.AutoResizeImage(height, icon)
}

func loadIconForName(name string, isBoxArt bool) image.Image {
	var icon image.Image
	iconContent, _ := os.Open(fmt.Sprintf("%s/%s.png", AssetsFolder, util.ToSnakecase(name)))
	if iconContent == nil {
		boxArtURL, err := imagegen.FindBoxArtUrl(name, isBoxArt, serperAPIKey)
		if err != nil {
			fmt.Println("failed to find image for game: ", name)
			return nil
		}

		icon, err := imagegen.DownloadImageFromUrl(name, AssetsFolder, boxArtURL)
		if err != nil {
			fmt.Println("failed to download image for game: ", name)
			return nil
		}
		return icon
	}
	icon, _, _ = image.Decode(iconContent)
	return icon
}
