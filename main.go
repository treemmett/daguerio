package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	initConfig()
	initS3()
	connectToSQL()
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	http.Handle("/graphql", generateGQL())
	err = http.ListenAndServeTLS("0.0.0.0:"+strconv.Itoa(int(Config.Port)), os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), nil)
	if err != nil {
		panic(err)
	}
}
