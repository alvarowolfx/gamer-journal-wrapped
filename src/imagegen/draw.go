package imagegen

import (
	"image"
	"log"
	"sync"

	"io"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

const (
	Ratio = 2
	W     = 540 * Ratio
	H     = 960 * Ratio
	//BarColor        = "#B4F8C8"
	//BackgroundColor = "#000000"
	IconColor       = "#FFF"
	BarColor        = "#FCA311"
	BackgroundColor = "#14213D"
	TitleFontSize   = 48 * Ratio
	BoldFontSize    = 20 * Ratio
	RegularFontSize = 14 * Ratio
)

type Orientation int

const (
	Horizontal Orientation = 0
	Vertical   Orientation = iota
)

var (
	// Lock to prevent concurrent rendering, which was
	// causing some glitches on text rendering
	lock = sync.Mutex{}
)

func (o Orientation) String() string {
	switch o {
	case Horizontal:
		return "horizontal"
	case Vertical:
		return "vertical"
	default:
		return "unknown"
	}
}

var (
	titleFont   font.Face
	boldFont    font.Face
	regularFont font.Face
)

type BarChartItem interface {
	GetTitle() string
	GetMetric() int
	RenderMetric() string
	RenderIcon(height uint, serperAPIKey string) image.Image
}

type SaveableDrawing interface {
	SavePNG(path string) error
	EncodePNG(w io.Writer) error
}

type drawing struct {
	*gg.Context
}

func (d *drawing) EncodePNG(w io.Writer) error {
	return d.Context.EncodePNG(w)
}

func RenderMostPlayedWrapped(title string, data []BarChartItem, n int, orientation Orientation, serperAPIKey string) SaveableDrawing {
	width := W
	height := H
	if orientation != Vertical {
		width, height = height, width
	}
	lock.Lock()
	defer lock.Unlock()
	dc := gg.NewContext(width, height)
	dc.SetHexColor(BackgroundColor)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Fill()
	margin := 20.0 * Ratio
	if n > 10 {
		margin = 18.0 * Ratio
	}
	marginTop := 4 * margin
	if orientation != Vertical {
		marginTop = 1.5 * margin
	}
	// ((2 * barHeight) + (margin)) * n = 3*height/4
	// ((2 * barHeight) + (margin)) = 3*height/4n
	// (2 * barHeight) = (3*height/4n) - (margin)
	// barHeight = ((3*height/4n) - (margin)) / 2
	barHeight := ((float64(3*height/4) - float64(margin)) / float64(n)) / 2
	if barHeight > 40*Ratio {
		barHeight = 40 * Ratio
	}
	maxMetric := -1
	total := 0
	for _, d := range data {
		m := d.GetMetric()
		if m > maxMetric {
			maxMetric = m
		}
		total += m
	}
	maxMetric = int(float64(maxMetric) * 1.25)

	textSize := float64(width) - 8*margin
	if orientation != Vertical {
		textSize = float64(width) - 2*margin
	}

	dc.SetHexColor(BarColor)
	dc.SetFontFace(titleFont)
	dc.DrawStringWrapped(title, float64(width)/2, marginTop, 0.5, 0.5, textSize, 1, gg.AlignCenter)
	dc.Fill()

	if orientation == Vertical {
		marginTop += 5 * margin
	} else {
		marginTop += 2.5 * margin
	}

	for i, d := range data {
		if i > n-1 {
			break
		}
		fullbarSize := float64(width) - 4*margin
		icon := d.RenderIcon(uint(barHeight*2)-4, serperAPIKey)

		if icon != nil {
			fullbarSize = float64(width) - 6*margin
		}

		size := (float64(d.GetMetric()) / float64(maxMetric)) * fullbarSize
		x := float64(1.5 * margin)
		y := marginTop + (float64(i) * ((barHeight * 2) + margin/2))

		if icon == nil {
			y = marginTop + (float64(i) * ((barHeight * 2) + margin/2))
		}

		if icon != nil {
			dc.SetHexColor(IconColor)
			dc.DrawRectangle(x, y, barHeight*2, barHeight*2)
			dc.Fill()

			dc.DrawImageAnchored(icon, int(x+barHeight), int(y+barHeight), 0.5, 0.5)

			x += 8 + barHeight*2
		}

		dc.SetHexColor(BarColor)
		dc.SetFontFace(regularFont)
		dc.DrawStringWrapped(d.GetTitle(), x+textSize/2, y+1.4*barHeight, 0.5, 0.5, textSize, 1, gg.AlignLeft)

		dc.SetFontFace(boldFont)
		dc.DrawString(d.RenderMetric(), x+size+(1.5*margin), y+(barHeight/2)+(BoldFontSize/4))
		dc.Fill()

		dc.DrawRectangle(x, y, size+margin, barHeight)
		dc.Fill()
	}

	return &drawing{dc}
}

func LoadFonts() {
	var err error
	titleFont, err = gg.LoadFontFace("./fonts/SpaceMono-Bold.ttf", TitleFontSize)
	if err != nil {
		log.Fatalf("failed to load title font")
	}
	boldFont, err = gg.LoadFontFace("./fonts/SpaceMono-Bold.ttf", BoldFontSize)
	if err != nil {
		log.Fatalf("failed to load bold font")
	}
	regularFont, err = gg.LoadFontFace("./fonts/SpaceMono-Regular.ttf", RegularFontSize)
	if err != nil {
		log.Fatalf("failed to load regular font")
	}
}
