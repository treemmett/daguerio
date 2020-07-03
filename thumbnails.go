package main

import (
	"errors"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"time"

	"github.com/esimov/stackblur-go"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
	"github.com/nfnt/resize"
)

// Thumbnail for an upload photo
type Thumbnail struct {
	ID      string `json:"id"`
	Height  int    `json:"height"`
	PhotoID string `json:"photoId"`
	Size    int    `json:"size"`
	Type    string `json:"type"`
	Width   int    `json:"width"`
}

func createThumbnails(img image.Image, photoID string) error {
	thumbnail := resize.Thumbnail(500, 500, img, resize.Bicubic)

	for i := 0; i < 2; i++ {
		var img image.Image
		var thumbType string

		if i == 0 {
			img = thumbnail
			thumbType = "NORMAL"
		} else {
			img = stackblur.Process(thumbnail, 30)
			thumbType = "BLUR"
		}

		file, err := ioutil.TempFile(os.TempDir(), "thumbnail-")
		if err != nil {
			return err
		}
		defer file.Close()
		defer os.Remove(file.Name())

		err = jpeg.Encode(file, img, nil)
		if err != nil {
			return err
		}

		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}

		stats, err := file.Stat()
		if err != nil {
			return err
		}

		file, err = os.Open(file.Name())
		if err != nil {
			return err
		}
		_, err = S3.PutObject(Config.S3Bucket, "thumbnails/"+id.String(), file, stats.Size(), minio.PutObjectOptions{
			ContentType: "image/jpeg",
		})
		if err != nil {
			return err
		}

		_, err = DB.Exec(
			"INSERT INTO thumbnails (id, size, width, height, mime, type, \"photoId\") VALUES ($1, $2, $3, $4, $5, $6, $7)",
			id.String(),
			stats.Size(),
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			"image/jpeg",
			thumbType,
			photoID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getThumbnails(photoID string) ([]*Thumbnail, error) {
	rows, err := DB.Query(
		"SELECT id, height, \"photoId\", size, type, width FROM thumbnails WHERE \"photoId\" = $1",
		photoID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var thumbnails []*Thumbnail

	for rows.Next() {
		var t Thumbnail
		err = rows.Scan(
			&t.ID,
			&t.Height,
			&t.PhotoID,
			&t.Size,
			&t.Type,
			&t.Width,
		)
		if err != nil {
			return nil, err
		}
		thumbnails = append(thumbnails, &t)
	}

	return thumbnails, nil
}

func getThumbnailURL(id string) (string, error) {
	url, err := S3.PresignedGetObject(Config.S3Bucket, "thumbnails/"+id, time.Second*60*60, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func removeThumbnail(id string) error {
	rows, err := DB.Exec("DELETE FROM thumbnails WHERE id = $1", id)
	if err != nil {
		return err
	}
	count, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("Thumbnail not found")
	}

	return nil
}
