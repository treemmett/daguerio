package main

import (
	"github.com/graphql-go/graphql"
	handler "github.com/koblas/graphql-handler"
)

func generateGQL() *handler.Handler {
	uploadScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name: "Upload",
		ParseValue: func(value interface{}) interface{} {
			if v, ok := value.(*handler.MultipartFile); ok {
				return v
			}
			return nil
		},
	})

	thumbnailTypeEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "ThumbnailType",
		Values: graphql.EnumValueConfigMap{
			"NORMAL": &graphql.EnumValueConfig{
				Value: "NORMAL",
			},
			"BLUR": &graphql.EnumValueConfig{
				Value: "BLUR",
			},
		},
	})

	thumbnailType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Thumbnail",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"height": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"size": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"type": &graphql.Field{
				Type: graphql.NewNonNull(thumbnailTypeEnum),
			},
			"width": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	})

	photoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Photo",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"dominantColor": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"height": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"size": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"thumbnails": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(thumbnailType)),
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return getThumbnails(params.Source.(Photo).ID)
				},
			},
			"url": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"width": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
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
					return getPhotos()
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
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

					return addPhoto(file)
				},
			},
		},
	})

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
