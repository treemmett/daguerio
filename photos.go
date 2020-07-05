package main

import (
	"database/sql"
	"errors"
	"fmt"
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
	"github.com/rwcarlsen/goexif/exif"
)

// Photo is a user uploaded picture
type Photo struct {
	ID            string     `json:"id"`
	Date          *time.Time `json:"date"`
	DateUploaded  *time.Time `json:"dateUploaded"`
	DominantColor string     `json:"dominantColor"`
	Latitude      *float64   `json:"latitude"`
	Longitude     *float64   `json:"longitude"`
	Height        int        `json:"height"`
	Size          int        `json:"size"`
	Width         int        `json:"width"`
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

	// get metadata
	file, err = os.Open(file.Name())
	if err != nil {
		return nil, errors.New("Failed to decode image\n" + err.Error())
	}
	defer file.Close()

	var time *time.Time = nil
	var lat *float64 = nil
	var long *float64 = nil
	meta, err := exif.Decode(file)
	if err == nil {
		timeActual, err := meta.DateTime()
		if err != nil {
			time = nil
		} else {
			time = &timeActual
		}

		latActual, longActual, err := meta.LatLong()
		if err != nil {
			lat = nil
			long = nil
		} else {
			lat = &latActual
			long = &longActual
		}
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
		"INSERT INTO photos (id, size, width, height, mime, \"dominantColor\", latitude, longitude, date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		id.String(),
		int(photo.Header.Size),
		img.Bounds().Dx(),
		img.Bounds().Dy(),
		photo.Header.Header.Get("Content-Type"),
		dominantColor,
		lat,
		long,
		time,
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

func deletePhoto(photoID string) (bool, error) {
	// get thumbnails of photo first
	thumbnails, err := getThumbnails(photoID)
	if err != nil {
		return false, err
	}

	for _, thumbnail := range thumbnails {
		err = removeThumbnail(thumbnail.ID)
		if err != nil {
			return false, err
		}
	}

	err = S3.RemoveObject(Config.S3Bucket, "photos/"+photoID)
	if err != nil {
		return false, err
	}

	result, err := DB.Exec("DELETE FROM photos WHERE id = $1", photoID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, fmt.Errorf("Failed to delete photo")
	}

	return true, nil
}

func getPhoto(ID string) (*Photo, error) {
	row := DB.QueryRow("SELECT id, date, \"dateUploaded\", \"dominantColor\", height, size, width, latitude, longitude FROM photos WHERE id = $1", ID)
	var p Photo
	err := row.Scan(
		&p.ID,
		&p.Date,
		&p.DateUploaded,
		&p.DominantColor,
		&p.Height,
		&p.Size,
		&p.Width,
		&p.Latitude,
		&p.Longitude,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Photo not found")
		}
		return nil, err
	}
	return &p, nil
}

func getPhotos() ([]*Photo, error) {
	rows, err := DB.Query("SELECT id, date, \"dateUploaded\", \"dominantColor\", height, size, width, latitude, longitude FROM photos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []*Photo
	for rows.Next() {
		var p Photo
		err = rows.Scan(
			&p.ID,
			&p.Date,
			&p.DateUploaded,
			&p.DominantColor,
			&p.Height,
			&p.Size,
			&p.Width,
			&p.Latitude,
			&p.Longitude,
		)
		photos = append(photos, &p)
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

func setPhotoDate(id string, date time.Time) (*Photo, error) {
	result, err := DB.Exec(
		"UPDATE photos SET date = $1 WHERE id = $2",
		date,
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Photo not found")
		}
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		return nil, fmt.Errorf("Failed to update photo")
	}

	return getPhoto(id)
}

func setPhotoLocation(id string, latitude float64, longitude float64) (*Photo, error) {
	result, err := DB.Exec(
		"UPDATE photos SET latitude = $1, longitude = $2 WHERE id = $3",
		latitude,
		longitude,
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Photo not found")
		}
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		return nil, fmt.Errorf("Failed to update photo")
	}

	return getPhoto(id)
}
