package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/airtablesql"
	"github.com/alvarowolfx/gamer-journal-wrapped/src/imagegen"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mehanizm/airtable"
	"golang.org/x/sync/errgroup"
)

var (
	db           *sqlx.DB
	serperAPIKey string
)

type StatsResponse struct {
	Year               int                             `json:"year"`
	MostPlayedConsoles []imagegen.MostPlayedByPlaytime `json:"most_played_consoles"`
	MostPlayedPlatform []imagegen.MostPlayedByPlaytime `json:"most_played_platforms"`
	MostPlayedGames    []imagegen.MostPlayedGame       `json:"most_played_games"`
	MostPlayedSeries   []imagegen.MostPlayedByPlaytime `json:"most_played_series"`
	GamesByStatus      []imagegen.MostPlayedByNumGames `json:"games_by_status"`
	BusiestMonths      []imagegen.MostPlayedByPlaytime `json:"busiest_months"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to read .env: %v \n", err)
	}

	serperAPIKey = os.Getenv("SERPER_API_KEY")
	airtableAPIKey := os.Getenv("AIRTABLE_API_KEY")
	imagegen.LoadFonts()

	client := airtable.NewClient(airtableAPIKey)
	provider, err := airtablesql.NewProvider(client)
	if err != nil {
		log.Fatalf("failed to init airtable sql provider: %v", err)
	}

	engine := sqle.NewDefault(provider)

	sqlPort := 3307
	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("localhost:%d", sqlPort),
	}
	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		log.Fatalf("failed to create mysql server: %v", err)
	}

	go func() {
		if err = s.Start(); err != nil {
			log.Fatalf("failed to start mysql server: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	mysqlDSN := fmt.Sprintf("root:@tcp(localhost:%d)/gaming_journal?parseTime=true", sqlPort)
	db, err = sqlx.Connect("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("failed to connect to internal mysql: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/api/stats", handleGetStats)
	e.GET("/api/charts/:type", handleGetChart)

	e.Logger.Fatal(e.Start(":8080"))
}

func handleGetStats(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = fmt.Sprintf("%d", time.Now().Year())
	}
	year, _ := strconv.Atoi(yearStr)

	stats := StatsResponse{Year: year}

	var g errgroup.Group

	g.Go(func() error {
		return db.Select(&stats.MostPlayedConsoles, imagegen.QueryMostPlayedConsoles, yearStr)
	})

	g.Go(func() error {
		return db.Select(&stats.MostPlayedPlatform, imagegen.QueryMostPlayedPlatforms, yearStr)
	})

	g.Go(func() error {
		return db.Select(&stats.MostPlayedGames, imagegen.QueryMostPlayedGames, yearStr)
	})

	g.Go(func() error {
		return db.Select(&stats.MostPlayedSeries, imagegen.QueryMostPlayedSeries, yearStr)
	})

	g.Go(func() error {
		return db.Select(&stats.GamesByStatus, imagegen.QueryGamesByStatus, yearStr)
	})

	var busiestMonth []imagegen.MostPlayedByPlaytime
	g.Go(func() error {
		return db.Select(&busiestMonth, imagegen.QueryBusiestMonths, yearStr)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	stats.BusiestMonths = make([]imagegen.MostPlayedByPlaytime, len(busiestMonth))
	for i, d := range busiestMonth {
		monthNum, _ := strconv.ParseInt(d.Title, 10, 64)
		d.Title = time.Month(int(monthNum)).String()
		d.NoIcon = true
		stats.BusiestMonths[i] = d
	}

	return c.JSON(http.StatusOK, stats)
}

func handleGetChart(c echo.Context) error {
	chartType := c.Param("type")
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = fmt.Sprintf("%d", time.Now().Year())
	}

	orientationStr := c.QueryParam("orientation")
	orientation := imagegen.Vertical
	if orientationStr == "horizontal" {
		orientation = imagegen.Horizontal
	}

	var title string
	var data []imagegen.BarChartItem
	var limit int

	switch chartType {
	case "consoles":
		title = "Most played consoles in " + yearStr
		var rows []imagegen.MostPlayedByPlaytime
		err := db.Select(&rows, imagegen.QueryMostPlayedConsoles, yearStr)
		if err != nil {
			return err
		}
		data = toBarChartItems(rows)
		limit = len(data)
	case "platforms":
		title = "Most played platform in " + yearStr
		var rows []imagegen.MostPlayedByPlaytime
		err := db.Select(&rows, imagegen.QueryMostPlayedPlatforms, yearStr)
		if err != nil {
			return err
		}
		data = toBarChartItems(rows)
		limit = 9
	case "games":
		title = "Most played games in " + yearStr
		var rows []imagegen.MostPlayedGame
		err := db.Select(&rows, imagegen.QueryMostPlayedGames, yearStr)
		if err != nil {
			return err
		}
		data = toBarChartItems(rows)
		limit = 8
	case "series":
		title = "Most played game serie in " + yearStr
		var rows []imagegen.MostPlayedByPlaytime
		err := db.Select(&rows, imagegen.QueryMostPlayedSeries, yearStr)
		if err != nil {
			return err
		}
		data = toBarChartItems(rows)
		limit = 8
	case "status":
		title = "Games beaten in " + yearStr
		var rows []imagegen.MostPlayedByNumGames
		err := db.Select(&rows, imagegen.QueryGamesByStatus, yearStr)
		if err != nil {
			return err
		}
		data = toBarChartItems(rows)
		limit = len(data)
	case "months":
		title = "Busiest months in " + yearStr
		var rows []imagegen.MostPlayedByPlaytime
		err := db.Select(&rows, imagegen.QueryBusiestMonths, yearStr)
		if err != nil {
			return err
		}
		data = make([]imagegen.BarChartItem, len(rows))
		for i, d := range rows {
			monthNum, _ := strconv.ParseInt(d.Title, 10, 64)
			d.Title = time.Month(int(monthNum)).String()
			d.NoIcon = true
			data[i] = d
		}
		limit = len(data)
	default:
		return c.String(http.StatusBadRequest, "Invalid chart type")
	}

	drawing := imagegen.RenderMostPlayedWrapped(title, data, limit, orientation)

	c.Response().Header().Set(echo.HeaderContentType, "image/png")
	return drawing.EncodePNG(c.Response().Writer)
}

func toBarChartItems[T imagegen.BarChartItem](arr []T) []imagegen.BarChartItem {
	narr := make([]imagegen.BarChartItem, len(arr))
	for i, d := range arr {
		narr[i] = d
	}
	return narr
}
