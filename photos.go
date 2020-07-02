package main

import (
	"errors"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/google/uuid"
	handler "github.com/koblas/graphql-handler"
	"github.com/minio/minio-go"
)

// Photo is a user uploaded picture
type Photo struct {
	ID            string `json:"id"`
	DominantColor string `json:"dominantColor"`
	Height        int    `json:"height"`
	Size          int    `json:"size"`
	Width         int    `json:"width"`
}

func addPhoto(photo *handler.MultipartFile) (*Photo, error) {
	// save file to temp
	file, err := ioutil.TempFile(os.TempDir(), "photo-")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	defer os.Remove(file.Name())

	_, err = io.Copy(file, photo.File)
	if err != nil {
		return nil, errors.New("Failed to save image\n" + err.Error())
	}

	file, err = os.Open(file.Name())
	if err != nil {
		return nil, errors.New("Failed to decode image\n" + err.Error())
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
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
	err = createThumbnails(img, id.String())
	if err != nil {
		return nil, errors.New("Failed to create thumbnails\n" + err.Error())
	}

	// add image to s3
	file, err = os.Open(file.Name())
	if err != nil {
		return nil, errors.New("Failed to decode image\n" + err.Error())
	}
	defer file.Close()
	_, err = S3.PutObject(Config.S3Bucket, "photos/"+id.String(), file, photo.Header.Size, minio.PutObjectOptions{
		ContentType: photo.Header.Header.Get("Content-Type"),
	})
	if err != nil {
		return nil, errors.New("Failed to save image in store\n" + err.Error())
	}

	return &Photo{
		ID:            id.String(),
		DominantColor: dominantColor,
		Height:        img.Bounds().Dy(),
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

func getPhotoURL(id string) (string, error) {
	url, err := S3.PresignedGetObject(Config.S3Bucket, "photos/"+id, time.Second*60*60, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
