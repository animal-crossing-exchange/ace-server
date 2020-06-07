package types

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
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
                Resolve: idResolver,
            },
            "created": &graphql.Field {
                Type: graphql.String,
                Resolve: timestampResolver,
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

func CreateInquiry(ctx context.Context, db mongo.Database) graphql.Field {
    inquiriesCollection := db.Collection("inquiries")
    listingsCollection := db.Collection("listings")
    usersCollection := db.Collection("users")

    return graphql.Field {
        Type: ListingInquiryType,
        Description: "Create a listing inquiry",
        Args: graphql.FieldConfigArgument {
            "listingID": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "userID": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "note": &graphql.ArgumentConfig {
                Type: graphql.String,
                DefaultValue: nil,
            },
        },
        Resolve: func (p graphql.ResolveParams) (interface{}, error) {
            listingID, prs := p.Args["listingID"]
            if !prs {
                return nil, errors.New("No listing ID given for inquiry creation")
            }
            listingObjID, err := primitive.ObjectIDFromHex(listingID.(string))
            if err != nil {
                return nil, err
            }
            userID, prs := p.Args["userID"]
            if !prs {
                return nil, errors.New("No user ID given for inquiry creation")
            }
            userObjID, err := primitive.ObjectIDFromHex(userID.(string))
            if err != nil {
                return nil, err
            }
            note := p.Args["note"]

            timeout, cancel := context.WithTimeout(ctx, time.Second)
            defer cancel()

            var listing bson.M
            err = listingsCollection.FindOne(timeout, bson.M{"_id": listingObjID}).Decode(&listing)
            if err != nil {
                return nil, err
            }

            var user bson.M
            err = usersCollection.FindOne(timeout, bson.M{"_id": userObjID}).Decode(&user)
            if err != nil {
                return nil, err
            }

            // check if user is trying to make an inquiry to themself
            if listing["seller"] == userObjID {
                return nil, errors.New("User cannot create inquiry towards their own listing")
            }
            // check if user is trying to make multipe inquiries
            for _, iID := range user["inquiries"].(bson.A) {
                var i bson.M
                err = inquiriesCollection.FindOne(timeout, bson.M{"_id": iID}).Decode(&i)
                if err != nil {
                    return nil, err
                }
                if i["buyer"] == userObjID {
                    return nil, errors.New("User cannot make multiple inquiries towards same listing")
                }
            }

            res, err := inquiriesCollection.InsertOne(timeout, bson.M{
                "note": note,
                "accepted": nil,
                "declined": nil,
                "buyer": userObjID,
                "listing": listingObjID,
            })
            if err != nil {
                return nil, err
            }
            var inquiry bson.M
            err = inquiriesCollection.FindOne(timeout, bson.M{"_id": res.InsertedID}).Decode(&inquiry)
            if err != nil {
                return nil, err
            }

            err = addToBsonArray(timeout, listingObjID, *listingsCollection, "inquiries", res.InsertedID)
            if err != nil {
                inquiriesCollection.DeleteOne(timeout, bson.M{"_id": res.InsertedID})
                return nil, err
            }

            err = addToBsonArray(timeout, userObjID, *usersCollection, "inquiries", res.InsertedID)
            if err != nil {
                inquiriesCollection.DeleteOne(timeout, bson.M{"_id": res.InsertedID})
                pullFromBsonArray(timeout, listingObjID, *listingsCollection, "inquiries", res.InsertedID)
                return nil, err
            }

            return inquiry, nil
        },
    }
}

