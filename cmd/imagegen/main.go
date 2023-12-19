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
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	AssetsFolder = "./assets/"
	OutFolder    = "./out/"
)

var (
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

	db, err := sqlx.Connect("mysql", "root:@/Gaming Journal?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	flag.IntVar(&year, "year", 2023, "year to render gamer wrapped")
	flag.Parse()

	yearStr := fmt.Sprintf("%d", year)

	fmt.Println("Rendering wrapped for", yearStr)

	mostPlayedConsolesData := []MostPlayedByPlaytime{}
	err = db.Select(&mostPlayedConsolesData, `
	select c.name as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
	from playthroughs p
		inner join consoles c on JSON_CONTAINS(p.console, CONCAT('"', c.record_id, '"'))
	where p.year_start_date = ?
	group by c.name
	order by playtime desc`, yearStr)
	if err != nil {
		log.Fatalf("failed to query most played console: %v", err)
	}
	mostPlayedConsoles := make([]imagegen.BarChartItem, len(mostPlayedConsolesData))
	for i, d := range mostPlayedConsolesData {
		mostPlayedConsoles[i] = d
	}
	imagegen.RenderMostPlayedWrapped("Most played consoles in "+yearStr, mostPlayedConsoles, OutFolder)

	mostPlayedPlatformData := []MostPlayedByPlaytime{}
	err = db.Select(&mostPlayedPlatformData, `
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
	mostPlayedPlatform := make([]imagegen.BarChartItem, len(mostPlayedPlatformData))
	for i, d := range mostPlayedPlatformData {
		mostPlayedPlatform[i] = d
	}
	if len(mostPlayedPlatform) > 9 {
		mostPlayedPlatform = mostPlayedPlatform[0:9]
	}
	imagegen.RenderMostPlayedWrapped("Most played platform in "+yearStr, mostPlayedPlatform, OutFolder)

	mostPlayedGamesData := []MostPlayedGame{}
	err = db.Select(&mostPlayedGamesData, `
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
	mostPlayedGames := make([]imagegen.BarChartItem, len(mostPlayedGamesData))
	for i, d := range mostPlayedGamesData {
		mostPlayedGames[i] = d
	}
	imagegen.RenderMostPlayedWrapped("Most played games in "+yearStr, mostPlayedGames[:8], OutFolder)

	gamesByStatusData := []MostPlayedByNumGames{}
	err = db.Select(&gamesByStatusData, `
	select p.status as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
	from playthroughs p	
	where p.year_start_date = ?
		and p.status not in ('Playing')
	group by p.status
	order by count desc;`, yearStr)
	if err != nil {
		log.Fatalf("failed to query most played games: %v", err)
	}
	gamesByStatus := make([]imagegen.BarChartItem, len(gamesByStatusData))
	for i, d := range gamesByStatusData {
		gamesByStatus[i] = d
	}

	imagegen.RenderMostPlayedWrapped("Games beaten in "+yearStr, gamesByStatus, OutFolder)

	busiestMonthData := []MostPlayedByPlaytime{}
	err = db.Select(&busiestMonthData, `
	select EXTRACT(MONTH from p.start_date) as title, sum(p.playtime)/(60*60) as playtime, count(*) as count
	from playthroughs p	
	where p.year_start_date = ?
	group by title
	order by title asc;
	`, yearStr)
	if err != nil {
		log.Fatalf("failed to query busiest month: %v", err)
	}
	busiestMonth := make([]imagegen.BarChartItem, len(busiestMonthData))
	for i, d := range busiestMonthData {
		monthNum, _ := strconv.ParseInt(d.Title, 10, 64)
		d.Title = time.Month(int(monthNum)).String()
		d.NoIcon = true
		busiestMonth[i] = d
	}

	imagegen.RenderMostPlayedWrapped("Busiest months in "+yearStr, busiestMonth, OutFolder)
}

type MostPlayedByPlaytime struct {
	Title    string
	Playtime int
	Count    int
	NoIcon   bool
}

func (mp MostPlayedByPlaytime) GetTitle() string {
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
	newIcon := imagegen.ResizeImage(height, 0, icon)
	return newIcon
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

	newIcon := imagegen.ResizeAndCropImage(0, height, icon)
	return newIcon
}

func loadIconForName(name string, isBoxArt bool) image.Image {
	var icon image.Image
	iconContent, _ := os.Open(fmt.Sprintf("%s/%s.png", AssetsFolder, imagegen.ToSnakecase(name)))
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
