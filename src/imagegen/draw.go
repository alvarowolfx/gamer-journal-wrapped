package imagegen

import (
	"image"
	"log"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

const (
	W               = 1080
	H               = 1920
	Ratio           = 2
	BarColor        = "#B4F8C8"
	BackgroundColor = "#000000"
	TitleFontSize   = 48
	BoldFontSize    = 20
	RegularFontSize = 14
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

func RenderMostPlayedWrapped(title string, data []BarChartItem) SaveableDrawing {
	dc := gg.NewContext(W/Ratio, H/Ratio)
	dc.SetHexColor(BackgroundColor)
	dc.DrawRectangle(0, 0, W/Ratio, H/Ratio)
	dc.Fill()
	margin := 20.0
	marginTop := float64(H) / 9.5
	barHeight := (float64(H) / float64(len(data))) / 6
	if barHeight > 40 {
		barHeight = 40
	}
	maxMetric := -1
	for _, d := range data {
		m := d.GetMetric()
		if m > maxMetric {
			maxMetric = m
		}
	}
	maxMetric = int(float64(maxMetric) * 1.25)

	iconRatio := 2.0
	textSize := float64(W/Ratio - 8*margin)

	dc.SetHexColor(BarColor)
	dc.SetFontFace(titleFont)
	dc.DrawStringWrapped(title, (W/Ratio)/2, 5*margin, 0.5, 0.5, textSize, 1, gg.AlignCenter)
	dc.Fill()

	for i, d := range data {
		fullbarSize := float64(W/Ratio - 4*margin)
		icon := d.RenderIcon(uint(barHeight*iconRatio) - 4)

		if icon != nil {
			fullbarSize = float64(W/Ratio - 6*margin)
		}

		size := (float64(d.GetMetric()) / float64(maxMetric)) * fullbarSize
		x := float64(1.5 * margin)
		y := marginTop + (float64(i) * (barHeight + 2.5*margin))

		if icon == nil {
			y = marginTop + (float64(i) * (barHeight + 2*margin))
		}

		if icon != nil {
			dc.SetHexColor("#FFF")
			dc.DrawRectangle(x, y, barHeight*iconRatio, barHeight*iconRatio)
			dc.Fill()

			dc.DrawImageAnchored(icon, int(x+(barHeight*(iconRatio/2))), int(y+barHeight), 0.5, 0.5)

			x += 8 + barHeight*iconRatio
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
