package types

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "github.com/graphql-go/graphql"
)

type TransactionStruct struct {
    id string
    state string
    price int
    buyerReportedComplete string
    sellerReportedComplete string
    reportedFailed string
    note string
    listing *ListingStruct
    buyer *UserStruct
    seller *UserStruct
    goesFirst *UserStruct
    unhappyUser *UserStruct
}

// TransactionType corresponds to the "transactions" collection
var TransactionType = graphql.NewObject(
    graphql.ObjectConfig {
        Name: "Transaction",
        Fields: graphql.Fields {
            "id": &graphql.Field {
                Type: graphql.ID,
                Resolve: idResolver,
            },
            "state": &graphql.Field {
                Type: graphql.String, // TODO TransactionState scalar
            },
            "price": &graphql.Field {
                Type: graphql.Int,
            },
            "buyerReportedComplete": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
            "sellerReportedComplete": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
            "reportedFailed": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
            "note": &graphql.Field {
                Type: graphql.String,
            },
        },
    },
)

func InitTransactionType(ctx context.Context, db mongo.Database) {
    TransactionType.AddFieldConfig("listing", &graphql.Field {
        Type: ListingType,
        Resolve: resolverGenerator(ctx, "listing", *db.Collection("listings")),
    })
    TransactionType.AddFieldConfig("buyer", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "buyer", *db.Collection("users")),
    })
    TransactionType.AddFieldConfig("seller", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "seller", *db.Collection("users")),
    })
    TransactionType.AddFieldConfig("goesFirst", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "goesFirst", *db.Collection("users")),
    })
    TransactionType.AddFieldConfig("unhappyUser", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "unhappyUser", *db.Collection("users")),
    })
}

