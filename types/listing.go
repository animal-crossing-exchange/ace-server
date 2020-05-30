package types

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "github.com/graphql-go/graphql"
)

type ListingStruct struct {
    id string
    price int
    deleted bool
    accepted string
    seller *UserStruct
    buyer *UserStruct
    item *ItemStruct
    inquiries []*ListingInquiryStruct
}

// ListingType corresponds to the "listings" collection
var ListingType = graphql.NewObject(
    graphql.ObjectConfig {
        Name: "Listing",
        Fields: graphql.Fields {
            "id": &graphql.Field {
                Type: graphql.ID,
                Resolve: idResolver,
            },
            "price": &graphql.Field {
                Type: graphql.Int,
            },
            "deleted": &graphql.Field {
                Type: graphql.Boolean,
            },
            "accepted": &graphql.Field {
                Type: graphql.String, //TODO change this to a custom Date scalar
            },
        },
    },
)

func InitListingType(ctx context.Context, db mongo.Database) {
    ListingType.AddFieldConfig("seller", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "seller", *db.Collection("users")),
    })
    ListingType.AddFieldConfig("buyer", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "buyer", *db.Collection("users")),
    })
    ListingType.AddFieldConfig("item", &graphql.Field {
        Type: ItemType,
        Resolve: resolverGenerator(ctx, "item", *db.Collection("items")),
    })
    ListingType.AddFieldConfig("inquiries", &graphql.Field {
        Type: graphql.NewList(ListingInquiryType),
        Resolve: resolverGenerator(ctx, "inquiries", *db.Collection("inquiries")),
    })
}

