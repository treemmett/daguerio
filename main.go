package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/graphql-go/graphql"
	"github.com/joho/godotenv"
)

type Photo struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

var photos []Photo = []Photo{
	{
		ID:  "a",
		URL: "a.com",
	},
	{
		ID:  "b",
		URL: "b.com",
	},
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	photoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Photo",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"size": &graphql.Field{
				Type: graphql.Int,
			},
			"url": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	rootQuery := graphql.ObjectConfig(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"photos": &graphql.Field{
				Type: graphql.NewList(photoType),
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return photos, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
	})
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: r.URL.Query().Get("query"),
		})
		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	err = http.ListenAndServeTLS("0.0.0.0:4000", os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), nil)
	fmt.Println(os.Getenv("SSL_CERT"))
	if err != nil {
		panic(err)
	}
}
