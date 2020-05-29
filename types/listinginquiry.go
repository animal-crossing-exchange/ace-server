package types

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "github.com/graphql-go/graphql"
)

type ListingInquiryStruct struct {
    id string
    note string
    deleted bool
    accepted string
    declined string
    buyer *UserStruct
    listing *ListingStruct
}

// ListingInquiryType corresponds to the "inquiries" collection
var ListingInquiryType = graphql.NewObject(
    graphql.ObjectConfig {
        Name: "ListingInquiry",
        Fields: graphql.Fields {
            "id": &graphql.Field {
                Type: graphql.ID,
            },
            "note": &graphql.Field {
                Type: graphql.String,
            },
            "deleted": &graphql.Field {
                Type: graphql.Boolean,
            },
            "accepted": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
            "declined": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
        },
    },
)

func InitListingInquiry(ctx context.Context, db mongo.Database) {
    ListingInquiryType.AddFieldConfig("buyer", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "buyer", *db.Collection("users")),
    })
    ListingInquiryType.AddFieldConfig("listing", &graphql.Field {
        Type: ListingType,
        Resolve: resolverGenerator(ctx, "listing", *db.Collection("listings")),
    })
}

