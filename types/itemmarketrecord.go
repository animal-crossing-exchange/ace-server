package types

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "github.com/graphql-go/graphql"
)

type ItemMarketRecordStruct struct {
    id string
    date string
    avg int
    median int
    high int
    low int
    numListings int
}

// ItemMarketRecordType corresponds to the "records" collection
var ItemMarketRecordType = graphql.NewObject(
    graphql.ObjectConfig {
        Name: "ItemMarketRecord",
        Fields: graphql.Fields {
            "id": &graphql.Field {
                Type: graphql.ID,
                Resolve: idResolver,
            },
            "date": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
            "avg": &graphql.Field {
                Type: graphql.Int,
            },
            "median": &graphql.Field {
                Type: graphql.Int,
            },
            "high": &graphql.Field {
                Type: graphql.Int,
            },
            "low": &graphql.Field {
                Type: graphql.Int,
            },
            "numListings": &graphql.Field {
                Type: graphql.Int,
            },
        },
    },
)

func InitItemMarketRecordType(ctx context.Context, db mongo.Database) {
    ItemMarketRecordType.AddFieldConfig("item", &graphql.Field {
        Type: ItemType,
        Resolve: resolverGenerator(ctx, "item", *db.Collection("items")),
    })
}

