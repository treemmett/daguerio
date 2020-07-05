package main

import (
	"time"

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
			"url": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return getThumbnailURL(params.Source.(*Thumbnail).ID)
				},
			},
			"width": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	})

	coordinatesType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Coordinates",
		Fields: graphql.Fields{
			"longitude": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
			"latitude": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
		},
	})

	photoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Photo",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"date": &graphql.Field{
				Type: graphql.DateTime,
			},
			"dateUploaded": &graphql.Field{
				Type: graphql.DateTime,
			},
			"dominantColor": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"location": &graphql.Field{
				Type: coordinatesType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					type Coordinates struct {
						Latitude  float64
						Longitude float64
					}

					photo := params.Source.(*Photo)

					if photo.Latitude == nil || photo.Longitude == nil {
						return nil, nil
					}

					return Coordinates{
						Latitude:  *photo.Latitude,
						Longitude: *photo.Longitude,
					}, nil
				},
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
					return getThumbnails(params.Source.(*Photo).ID)
				},
			},
			"url": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return getPhotoURL(params.Source.(*Photo).ID)
				},
			},
			"width": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"photo": &graphql.Field{
				Type: photoType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return getPhoto(params.Args["id"].(string))
				},
			},
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
			"deletePhoto": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return deletePhoto(params.Args["id"].(string))
				},
			},
			"setPhotoDate": &graphql.Field{
				Type: photoType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"date": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.DateTime),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return setPhotoDate(params.Args["id"].(string), params.Args["date"].(time.Time))
				},
			},
			"setPhotoLocation": &graphql.Field{
				Type: photoType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"latitude": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Float),
					},
					"longitude": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Float),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return setPhotoLocation(params.Args["id"].(string), params.Args["latitude"].(float64), params.Args["longitude"].(float64))
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
