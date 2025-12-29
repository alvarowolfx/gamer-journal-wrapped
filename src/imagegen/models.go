package imagegen

import (
	"fmt"
	"image"
	"math"
	"os"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/util"
)

const (
	AssetsFolder = "./assets/"
)

type MostPlayedByPlaytime struct {
	Title    string  `db:"title" json:"title"`
	Playtime float64 `db:"playtime" json:"playtime"`
	Count    int     `db:"count" json:"count"`
	NoIcon   bool    `json:"-"`
}

func (mp MostPlayedByPlaytime) GetTitle() string {
	if mp.Count == 1 {
		return fmt.Sprintf("%s (%d game)", mp.Title, mp.Count)
	}
	return fmt.Sprintf("%s (%d games)", mp.Title, mp.Count)
}

func (mp MostPlayedByPlaytime) GetMetric() int {
	return int(math.Round(mp.Playtime))
}

func (mp MostPlayedByPlaytime) RenderMetric() string {
	return fmt.Sprintf("%dh", mp.GetMetric())
}

func (mp MostPlayedByPlaytime) RenderIcon(height uint) image.Image {
	if mp.NoIcon {
		return nil
	}
	icon, err := LoadIconForName(mp.Title, false, "")
	if err != nil {
		fmt.Println("failed to load icon for game: ", mp.Title)
		return nil
	}
	return AutoResizeImage(height, icon)
}

type MostPlayedByNumGames struct {
	Title    string  `db:"title" json:"title"`
	Playtime float64 `db:"playtime" json:"playtime"`
	Count    int     `db:"count" json:"count"`
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
	Title    string  `db:"title" json:"title"`
	Platform string  `db:"platform" json:"platform"`
	Console  string  `db:"console" json:"console"`
	Playtime float64 `db:"playtime" json:"playtime"`
}

func (mpg MostPlayedGame) GetTitle() string {
	return fmt.Sprintf("%s (%s) on %s", mpg.Title, mpg.Platform, mpg.Console)
}

func (mpg MostPlayedGame) GetMetric() int {
	return int(math.Round(mpg.Playtime))
}

func (mpg MostPlayedGame) RenderMetric() string {
	return fmt.Sprintf("%dh", mpg.GetMetric())
}

func (mpg MostPlayedGame) RenderIcon(height uint) image.Image {
	icon, err := LoadIconForName(mpg.Title, true, "")
	if err != nil {
		fmt.Println("failed to load icon for game: ", mpg.Title)
		return nil
	}
	return AutoResizeImage(height, icon)
}

func LoadIconForName(name string, isBoxArt bool, serperAPIKey string) (image.Image, error) {
	var icon image.Image
	iconContent, _ := os.Open(fmt.Sprintf("%s/%s.png", AssetsFolder, util.ToSnakecase(name)))
	if iconContent == nil {
		if serperAPIKey == "" {
			return nil, fmt.Errorf("serper api key not provided")
		}
		boxArtURL, err := FindBoxArtUrl(name, isBoxArt, serperAPIKey)
		if err != nil {
			fmt.Println("failed to find image for game: ", name)
			return nil, err
		}

		icon, err = DownloadImageFromUrl(name, AssetsFolder, boxArtURL)
		if err != nil {
			fmt.Println("failed to download image for game: ", name)
			return nil, err
		}
		return icon, nil
	}
	icon, _, err := image.Decode(iconContent)
	return icon, err
}
