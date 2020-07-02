package main

import (
	"errors"
	"image/jpeg"
	"io"
	"os"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/google/uuid"
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
	defer imageFile.Close()

	_, err = io.Copy(imageFile, photo.File)
	if err != nil {
		return nil, errors.New("Failed to save image\n" + err.Error())
	}

	imageFile, err = os.Open("test.jpg")
	if err != nil {
		return nil, errors.New("Failed to decode image\n" + err.Error())
	}
	defer imageFile.Close()
	img, err := jpeg.Decode(imageFile)
	if err != nil {
		return nil, errors.New("Failed to decode image\n" + err.Error())
	}

	// get dominant color
	colors, err := prominentcolor.Kmeans(img)
	if err != nil {
		return nil, errors.New("Unable to find dominant color\n" + err.Error())
	}
	dominantColor := strings.ToLower(colors[0].AsString())

	// save image to database
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.New("ID generation failed\n" + err.Error())
	}
	_, err = DB.Exec(
		"INSERT INTO photos (id, size, width, height, mime, \"dominantColor\") VALUES ($1, $2, $3, $4, $5, $6)",
		id.String(),
		int(photo.Header.Size),
		img.Bounds().Dx(),
		img.Bounds().Dy(),
		photo.Header.Header.Get("Content-Type"),
		dominantColor,
	)
	if err != nil {
		return nil, errors.New("Failed to save image to db\n" + err.Error())
	}

	// create thumbnails
	thumbnail := resize.Thumbnail(500, 500, img, resize.Bicubic)

	thumbnailFile, err := os.Create("thumbnail.jpg")
	if err != nil {
		return nil, errors.New("Failed to save thumbnail\n" + err.Error())
	}
	defer thumbnailFile.Close()

	jpeg.Encode(thumbnailFile, thumbnail, nil)

	// blur thumbnailS
	err = createThumbnails(img, id.String())
	if err != nil {
		return nil, errors.New("Failed to create thumbnails\n" + err.Error())
	}

	return &Photo{
		ID:            id.String(),
		DominantColor: dominantColor,
		Height:        img.Bounds().Dy(),
		URL:           "string",
		Width:         img.Bounds().Dx(),
		Size:          int(photo.Header.Size),
	}, nil
}

func getPhotos() ([]Photo, error) {
	rows, err := DB.Query("SELECT id, \"dominantColor\", height, size, width FROM photos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var p Photo
		err = rows.Scan(
			&p.ID,
			&p.DominantColor,
			&p.Height,
			&p.Size,
			&p.Width,
		)
		photos = append(photos, p)
	}

	return photos, nil
}
