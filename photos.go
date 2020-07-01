package main

import (
	"errors"
	"image/jpeg"
	"io"
	"os"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/esimov/stackblur-go"
	handler "github.com/koblas/graphql-handler"
	"github.com/nfnt/resize"
)

// Photo is a user uploaded picture
type Photo struct {
	ID            string `json:"id"`
	DominantColor string `json:"dominantColor"`
	Height        int    `json:"height"`
	Size          int    `json:"size"`
	URL           string `json:"url"`
	Width         int    `json:"width"`
}

func addPhoto(photo *handler.MultipartFile) (*Photo, error) {
	// save file to disk
	imageFile, err := os.Create("test.jpg")
	if err != nil {
		return nil, errors.New("Failed to save image\n" + err.Error())
	}

	_, err = io.Copy(imageFile, photo.File)
	if err != nil {
		return nil, errors.New("Failed to save image\n" + err.Error())
	}

	// create thumbnails
	imageFile, err = os.Open("test.jpg")
	if err != nil {
		return nil, errors.New("Failed to decode image\n" + err.Error())
	}
	img, err := jpeg.Decode(imageFile)
	if err != nil {
		return nil, errors.New("Failed to decode image\n" + err.Error())
	}

	thumbnail := resize.Thumbnail(500, 500, img, resize.Bicubic)

	thumbnailFile, err := os.Create("thumbnail.jpg")
	if err != nil {
		return nil, errors.New("Failed to save thumbnail\n" + err.Error())
	}

	jpeg.Encode(thumbnailFile, thumbnail, nil)

	// blur thumbnail
	bluredThumbFile, err := os.Create("blurredThumb.jpg")
	if err != nil {
		return nil, errors.New("Failed to blur image\n" + err.Error())
	}

	err = jpeg.Encode(bluredThumbFile, stackblur.Process(thumbnail, 30), nil)
	if err != nil {
		return nil, errors.New("Failed to save blur\n" + err.Error())
	}

	// get dominant color
	colors, err := prominentcolor.Kmeans(img)
	if err != nil {
		return nil, errors.New("Unable to find dominant color\n" + err.Error())
	}
	dominantColor := strings.ToLower(colors[0].AsString())

	return &Photo{
		ID:            "string",
		DominantColor: dominantColor,
		Height:        img.Bounds().Dy(),
		URL:           "string",
		Width:         img.Bounds().Dx(),
		Size:          int(photo.Header.Size),
	}, nil
}

func getPhotos() ([]Photo, error) {
	return []Photo{
		{
			ID:  "a",
			URL: "a.com",
		},
		{
			ID:  "b",
			URL: "b.com",
		},
	}, nil
}
