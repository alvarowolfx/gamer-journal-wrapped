package imagegen

import (
	"image"
	"log"

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

var (
	titleFont   font.Face
	boldFont    font.Face
	regularFont font.Face
)

type BarChartItem interface {
	GetTitle() string
	GetMetric() int
	RenderMetric() string
	RenderIcon(height uint) image.Image
}

type SaveableDrawing interface {
	SavePNG(path string) error
}

func RenderMostPlayedWrapped(title string, data []BarChartItem, n int) SaveableDrawing {
	dc := gg.NewContext(W, H)
	dc.SetHexColor(BackgroundColor)
	dc.DrawRectangle(0, 0, W, H)
	dc.Fill()
	margin := 20.0 * Ratio
	if n > 10 {
		margin = 18.0 * Ratio
	}
	marginTop := float64(H) / 5
	barHeight := Ratio * ((float64(H) / float64(n)) / 3)
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

	textSize := float64(W - 8*margin)

	dc.SetHexColor(BarColor)
	dc.SetFontFace(titleFont)
	dc.DrawStringWrapped(title, W/2, 5*margin, 0.5, 0.5, textSize, 1, gg.AlignCenter)
	dc.Fill()

	for i, d := range data {
		if i > n-1 {
			break
		}
		fullbarSize := float64(W - 4*margin)
		icon := d.RenderIcon(uint(barHeight*2) - 4)

		if icon != nil {
			fullbarSize = float64(W - 6*margin)
		}

		size := (float64(d.GetMetric()) / float64(maxMetric)) * fullbarSize
		x := float64(1.5 * margin)
		y := marginTop + (float64(i) * (barHeight + 2.5*margin))

		if icon == nil {
			y = marginTop + (float64(i) * (barHeight + 2*margin))
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

	return dc
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
