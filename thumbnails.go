package main

import (
	"errors"
	"image"
	"image/jpeg"
	"os"

	"github.com/esimov/stackblur-go"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

func createThumbnails(img image.Image, photoID string) error {

	thumbnail := resize.Thumbnail(500, 500, img, resize.Bicubic)
	thumbnailFile, err := os.Create("thumbnail.jpg")
	if err != nil {
		return err
	}
	err = jpeg.Encode(thumbnailFile, thumbnail, nil)
	if err != nil {
		return err
	}

	thumbnailID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	thumbnailStat, err := thumbnailFile.Stat()
	if err != nil {
		return err
	}

	_, err = DB.Query(
		"INSERT INTO thumbnails (id, size, width, height, mime, type, \"photoId\") VALUES ($1, $2, $3, $4, $5, $6, $7)",
		thumbnailID.String(),
		thumbnailStat.Size(),
		thumbnail.Bounds().Dx(),
		thumbnail.Bounds().Dy(),
		"image/jpeg",
		"NORMAL",
		photoID,
	)

	if err != nil {
		return err
	}

	// create blurred thumbnail
	blurredThumbnailFile, err := os.Create("blurred_thumbnail.jpg")
	if err != nil {
		removeThumbnail(thumbnailID.String())
		return err
	}

	err = jpeg.Encode(blurredThumbnailFile, stackblur.Process(thumbnail, 30), nil)
	if err != nil {
		removeThumbnail(thumbnailID.String())
		return err
	}

	blurredID, err := uuid.NewRandom()
	if err != nil {
		removeThumbnail(thumbnailID.String())
		return err
	}

	blurredStat, err := blurredThumbnailFile.Stat()
	if err != nil {
		removeThumbnail(thumbnailID.String())
		return err
	}

	_, err = DB.Query(
		"INSERT INTO thumbnails (id, size, width, height, mime, type, \"photoId\") VALUES ($1, $2, $3, $4, $5, $6, $7)",
		blurredID.String(),
		blurredStat.Size(),
		thumbnail.Bounds().Dx(),
		thumbnail.Bounds().Dy(),
		"image/jpeg",
		"BLUR",
		photoID,
	)
	if err != nil {
		removeThumbnail(thumbnailID.String())
		return err
	}

	return nil
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
