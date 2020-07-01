package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	initConfig()
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	http.Handle("/graphql", generateGQL())
	err = http.ListenAndServeTLS("0.0.0.0:"+strconv.Itoa(int(Config.Port)), os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), nil)
	fmt.Println(os.Getenv("SSL_CERT"))
	if err != nil {
		panic(err)
	}
}
