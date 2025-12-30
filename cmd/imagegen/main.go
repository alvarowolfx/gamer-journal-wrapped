package main

import (
	"flag"
	"fmt"
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
)

var (
	mysqlDSN     = "root:@/gaming_journal?parseTime=true"
	serperAPIKey string
	startYear    int
	endYear      int
	outFolder    = "./out/"
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

	flag.IntVar(&startYear, "start", 2021, "start year to render gamer wrapped")
	flag.IntVar(&endYear, "end", 2025, "end year to render gamer wrapped")
	flag.StringVar(&outFolder, "out", "./out/", "output folder")
	flag.Parse()

	for year := startYear; year <= endYear; year++ {
		yearStr := fmt.Sprintf("%d", year)
		fmt.Println("Rendering wrapped for", yearStr)

		mostPlayedConsoles := []imagegen.MostPlayedByPlaytime{}
		err = db.Select(&mostPlayedConsoles, imagegen.QueryMostPlayedConsoles, yearStr)
		if err != nil {
			log.Fatalf("failed to query most played console: %v", err)
		}
		renderAndSaveAllMostPlayedWrapped("Most played consoles in "+yearStr, mostPlayedConsoles)

		mostPlayedPlatform := []imagegen.MostPlayedByPlaytime{}
		err = db.Select(&mostPlayedPlatform, imagegen.QueryMostPlayedPlatforms, yearStr)
		if err != nil {
			log.Fatalf("failed to query most played platform: %v", err)
		}
		renderAndSaveNMostPlayedWrapped("Most played platform in "+yearStr, mostPlayedPlatform, 9)

		mostPlayedGames := []imagegen.MostPlayedGame{}
		err = db.Select(&mostPlayedGames, imagegen.QueryMostPlayedGames, yearStr)
		if err != nil {
			log.Fatalf("failed to query most played games: %v", err)
		}
		renderAndSaveNMostPlayedWrapped("Most played games in "+yearStr, mostPlayedGames, 8)

		mostPlayedGameSerie := []imagegen.MostPlayedByPlaytime{}
		err = db.Select(&mostPlayedGameSerie, imagegen.QueryMostPlayedSeries, yearStr)
		if err != nil {
			log.Fatalf("failed to query most played game serie: %v", err)
		}
		renderAndSaveNMostPlayedWrapped("Most played game serie in "+yearStr, mostPlayedGameSerie, 8)

		gamesByStatus := []imagegen.MostPlayedByNumGames{}
		err = db.Select(&gamesByStatus, imagegen.QueryGamesByStatus, yearStr)
		if err != nil {
			log.Fatalf("failed to query most played games: %v", err)
		}
		renderAndSaveAllMostPlayedWrapped("Games beaten in "+yearStr, gamesByStatus)

		busiestMonth := []imagegen.MostPlayedByPlaytime{}
		err = db.Select(&busiestMonth, imagegen.QueryBusiestMonths, yearStr)
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
}

func renderAndSaveNMostPlayedWrapped[T imagegen.BarChartItem](title string, data []T, n int) {
	for _, orientation := range []imagegen.Orientation{imagegen.Vertical, imagegen.Horizontal} {
		imagegen.RenderMostPlayedWrapped(title, toBarChartItems(data), n, orientation, serperAPIKey).
			SavePNG(fmt.Sprintf("%s/%s_%s.png", outFolder, orientation.String(), util.ToSnakecase(title)))
	}
}

func renderAndSaveAllMostPlayedWrapped[T imagegen.BarChartItem](title string, data []T) {
	for _, orientation := range []imagegen.Orientation{imagegen.Vertical, imagegen.Horizontal} {
		imagegen.RenderMostPlayedWrapped(title, toBarChartItems(data), len(data), orientation, serperAPIKey).
			SavePNG(fmt.Sprintf("%s/%s_%s.png", outFolder, orientation.String(), util.ToSnakecase(title)))
	}
}

func toBarChartItems[T imagegen.BarChartItem](arr []T) []imagegen.BarChartItem {
	narr := make([]imagegen.BarChartItem, len(arr))
	for i, d := range arr {
		narr[i] = d
	}
	return narr
}
