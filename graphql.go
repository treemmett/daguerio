package main

import (
	"errors"
	"image/jpeg"
	"io"
	"os"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/esimov/stackblur-go"
	"github.com/graphql-go/graphql"
	handler "github.com/koblas/graphql-handler"
	"github.com/nfnt/resize"
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
	ID            string `json:"id"`
	DominantColor string `json:"dominantColor"`
	Height        int    `json:"height"`
	Size          int    `json:"size"`
	URL           string `json:"url"`
	Width         int    `json:"width"`
}

var photoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "photo",
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
		"url": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"width": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
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

				// save file to disk
				imageFile, err := os.Create("test.jpg")
				if err != nil {
					return nil, errors.New("Failed to save image\n" + err.Error())
				}

				_, err = io.Copy(imageFile, file.File)
				if err != nil {
					return nil, errors.New("Failed to save image\n" + err.Error())
				}

				// create thumbnails
				imageFile, err = os.Open("test.jpg")
				if err != nil {
					return nil, errors.New("Failed to decode image\n" + err.Error())
				}
				img, err := jpeg.Decode(imageFile)
				if err != nil {
					return nil, errors.New("Failed to decode image\n" + err.Error())
				}

				thumbnail := resize.Thumbnail(500, 500, img, resize.Bicubic)

				thumbnailFile, err := os.Create("thumbnail.jpg")
				if err != nil {
					return nil, errors.New("Failed to save thumbnail\n" + err.Error())
				}

				jpeg.Encode(thumbnailFile, thumbnail, nil)

				// blur thumbnail
				bluredThumbFile, err := os.Create("blurredThumb.jpg")
				if err != nil {
					return nil, errors.New("Failed to blur image\n" + err.Error())
				}

				err = jpeg.Encode(bluredThumbFile, stackblur.Process(thumbnail, 30), nil)
				if err != nil {
					return nil, errors.New("Failed to save blur\n" + err.Error())
				}

				// get dominant color
				colors, err := prominentcolor.Kmeans(img)
				if err != nil {
					return nil, errors.New("Unable to find dominant color\n" + err.Error())
				}
				dominantColor := strings.ToLower(colors[0].AsString())

				return photo{
					ID:            "string",
					DominantColor: dominantColor,
					Height:        img.Bounds().Dy(),
					URL:           "string",
					Width:         img.Bounds().Dx(),
					Size:          int(file.Header.Size),
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
