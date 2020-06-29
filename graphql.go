package main

import (
	"io"
	"os"

	"github.com/graphql-go/graphql"
	handler "github.com/koblas/graphql-handler"
)

var uploadScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name: "Upload",
	ParseValue: func(value interface{}) interface{} {
		if v, ok := value.(*handler.MultipartFile); ok {
			return v
		}
		return nil
	},
})

type photo struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Size int    `json:"size"`
}

var photoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "photo",
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

var photos []photo = []photo{
	{
		ID:  "a",
		URL: "a.com",
	},
	{
		ID:  "b",
		URL: "b.com",
	},
}

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"photos": &graphql.Field{
			Args: graphql.FieldConfigArgument{
				"photo": &graphql.ArgumentConfig{
					Type: uploadScalar,
				},
			},
			Type: graphql.NewList(photoType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return photos, nil
			},
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"addPhoto": &graphql.Field{
			Type: photoType,
			Args: graphql.FieldConfigArgument{
				"photo": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(uploadScalar),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				file := params.Args["photo"].(*handler.MultipartFile)
				defer file.File.Close()

				f, _ := os.Create("test.jpg")
				io.Copy(f, file.File)

				return photo{
					ID:   "string",
					URL:  "string",
					Size: int(file.Header.Size),
				}, nil
			},
		},
	},
})

func generateGQL() *handler.Handler {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})

	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	return h
}
