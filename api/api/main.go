package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/animal-crossing-exchange/ace-server/model"
	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var database *mongo.Database
var li *model.ListingInterface

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI(model.DatabaseURI))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database(model.DatabaseName)
	li := model.NewListingInterface(database)

	http.Handle("/graphql", gqlHandler(li))
	http.ListenAndServe(":3000", nil)
}

type reqBody struct {
	Query string `json:"query"`
}

func gqlHandler(li model.ListingInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "No query data", 400)
			return
		}

		var rBody reqBody
		err := json.NewDecoder(r.Body).Decode(&rBody)
		if err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
			return
		}

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		schema, err := gqlSchema(ctx, li)
		if err != nil {
			http.Error(w, "Internal server error", 500)
			log.Printf("error constructing schema: %s", err)
			return
		}

		params := graphql.Params{Schema: schema, RequestString: rBody.Query}
		run := graphql.Do(params)
		if len(run.Errors) > 0 {
			log.Printf("failed to execute graphql operation, errors: %+v", run.Errors)
		}
		rJSON, _ := json.Marshal(r)
		fmt.Fprintf(w, "%s", rJSON)

	})
}

func gqlSchema(ctx context.Context, li model.ListingInterface) (graphql.Schema, error) {
	fields := graphql.Fields{
		"listings": &graphql.Field{
			Type:        graphql.NewList(ListingType),
			Description: "Get all listings",
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return li.GetListings(ctx)
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	return graphql.NewSchema(schemaConfig)
}

var ListingType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Listing",
		Fields: graphql.Fields{
			"fromUser": &graphql.Field{
				Type: graphql.String,
			},
			"price": &graphql.Field{
				Type: graphql.Int,
			},
			"itemType": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)
