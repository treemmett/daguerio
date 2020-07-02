package main

import (
	"log"

	"github.com/minio/minio-go"
)

// S3 is a client to access the S3 bucket
var S3 *minio.Client

func initS3() {
	client, err := minio.New(Config.S3Endpoint, Config.S3KeyID, Config.S3AccessKey, true)
	if err != nil {
		log.Fatal(err)
	}

	S3 = client
}
