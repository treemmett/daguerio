package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	http.Handle("/graphql", generateGQL())
	err = http.ListenAndServeTLS("0.0.0.0:4000", os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), nil)
	fmt.Println(os.Getenv("SSL_CERT"))
	if err != nil {
		panic(err)
	}
}
