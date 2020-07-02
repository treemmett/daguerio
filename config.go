package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Configuration for application
type Configuration struct {
	Port        int64
	PsqlDb      string
	PsqlHost    string
	PsqlPass    string
	PsqlPort    int64
	PsqlUser    string
	S3AccessKey string
	S3Bucket    string
	S3Endpoint  string
	S3KeyID     string
	SSLCert     string
	SSLKey      string
}

// Config for application
var Config Configuration

func initConfig() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	var Port int64 = 4000
	if os.Getenv("PORT") != "" {
		port, err := strconv.ParseInt(os.Getenv("PORT"), 10, 0)
		if err == nil {
			Port = port
		}
	}

	PsqlDb := "photos"
	if os.Getenv("PSQL_DB") != "" {
		PsqlDb = os.Getenv("PSQL_DB")
	}

	PsqlHost := "localhost"
	if os.Getenv("PSQL_HOST") != "" {
		PsqlHost = os.Getenv("PSQL_HOST")
	}

	var PsqlPort int64 = 5432
	if os.Getenv("PSQL_PORT") != "" {
		port, err := strconv.ParseInt(os.Getenv("PSQL_PORT"), 10, 0)
		if err == nil {
			PsqlPort = port
		}
	}

	PsqlUser := "postgres"
	if os.Getenv("PSQL_USER") != "" {
		PsqlUser = os.Getenv("PSQL_USER")
	}

	Config = Configuration{
		Port:        Port,
		PsqlDb:      PsqlDb,
		PsqlHost:    PsqlHost,
		PsqlPass:    os.Getenv("PSQL_PASS"),
		PsqlPort:    PsqlPort,
		PsqlUser:    PsqlUser,
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3Endpoint:  os.Getenv("S3_ENDPOINT"),
		S3KeyID:     os.Getenv("S3_KEY_ID"),
		SSLCert:     os.Getenv("SSL_CERT"),
		SSLKey:      os.Getenv("SSL_KEY"),
	}
}
