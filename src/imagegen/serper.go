package imagegen

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type serperImagesResponse struct {
	Images []serperImageResponse `json:"images"`
}

type serperImageResponse struct {
	Title    string `json:"title"`
	ImageURL string `json:"imageUrl"`
}

func FindBoxArtUrl(title string, isBoxArt bool, serperAPIKey string) (string, error) {
	url := "https://google.serper.dev/images"
	method := "POST"

	if isBoxArt {
		title = fmt.Sprintf("%s box art", title)
	}
	payload := strings.NewReader(fmt.Sprintf(`{"q":"%s"}`, title))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-API-KEY", serperAPIKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var serperImagesResponse serperImagesResponse
	err = json.Unmarshal(body, &serperImagesResponse)
	if err != nil {
		return "", err
	}
	for _, img := range serperImagesResponse.Images {
		return img.ImageURL, nil
	}
	return "", fmt.Errorf("image not found for search: %s", title)
}
