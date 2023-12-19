package imagegen

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"

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
	resizedImg := img.(SubImager).SubImage(topCrop)
	icon := resize.Resize(w, h, resizedImg, resize.NearestNeighbor)
	return icon
}

func ResizeImage(w, h uint, img image.Image) image.Image {
	icon := resize.Resize(w, h, img, resize.NearestNeighbor)
	return icon
}

func DownloadImageFromUrl(name, folder, url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer([]byte{})
	iconReader := io.TeeReader(response.Body, buf)

	icon, _, err := image.Decode(iconReader)

	werr := os.WriteFile(fmt.Sprintf("%s/%s.png", folder, ToSnakecase(name)), buf.Bytes(), 0644)
	if werr != nil {
		fmt.Println("failed to save box art:", werr)
	}

	return icon, err
}
