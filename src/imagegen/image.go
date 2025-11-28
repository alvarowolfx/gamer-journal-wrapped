package imagegen

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/util"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
	"github.com/nfnt/resize"
)

func ResizeAndCropImage(w, h uint, img image.Image) image.Image {
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	topCrop, _ := analyzer.FindBestCrop(img, int(w), int(h))

	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	croppedImg := img.(SubImager).SubImage(topCrop)
	return ResizeImage(w, h, croppedImg)
}

func ResizeImage(w, h uint, img image.Image) image.Image {
	icon := resize.Resize(w, h, img, resize.NearestNeighbor)
	return icon
}

func AutoResizeImage(h uint, img image.Image) image.Image {
	if img.Bounds().Max.Y > img.Bounds().Max.X {
		return ResizeImage(0, h, img)
	}
	return ResizeImage(h, 0, img)
}

func DownloadImageFromUrl(name, folder, url string) (image.Image, error) {
	fmt.Println("Downloading image for game:", name)
	fmt.Println("from url:", url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("failed to download image:", err)
		return nil, err
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer([]byte{})
	iconReader := io.TeeReader(response.Body, buf)

	icon, _, err := image.Decode(iconReader)
	if err != nil {
		fmt.Println("failed to decode image:", err)
		return nil, err
	}

	werr := os.WriteFile(fmt.Sprintf("%s/%s.png", folder, util.ToSnakecase(name)), buf.Bytes(), 0644)
	if werr != nil {
		fmt.Println("failed to save box art:", werr)
	}

	return icon, werr
}
