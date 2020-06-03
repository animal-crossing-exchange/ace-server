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
            "created": &graphql.Field {
                Type: graphql.String,
                Resolve: timestampResolver,
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

// CreateListing creates a new listing in the database, and also updates the listings
// field of the associated item and user.
func CreateListing(ctx context.Context, db mongo.Database) graphql.Field {
    itemsCollection := db.Collection("items")
    listingsCollection := db.Collection("listings")
    usersCollection := db.Collection("users")

    return graphql.Field {
        Type: ListingType,
        Description: "Create a new listing",
        Args: graphql.FieldConfigArgument {
            "itemID": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "userID": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "price": &graphql.ArgumentConfig {
                Type: graphql.Int,
            },
        },
        Resolve: func (p graphql.ResolveParams) (interface{}, error) {
            itemID, prs := p.Args["itemID"]
            if !prs {
                return nil, errors.New("Item ID not given for listing creation")
            }
            itemObjID, err := primitive.ObjectIDFromHex(itemID.(string))
            if err != nil {
                return nil, err
            }
            userID, prs := p.Args["userID"]
            if !prs {
                return nil, errors.New("User ID not given for listing creation")
            }
            userObjID, err := primitive.ObjectIDFromHex(userID.(string))
            if err != nil {
                return nil, err
            }
            price, prs := p.Args["price"].(int)
            if !prs {
                return nil, errors.New("Price not given for listing creation")
            }
            if price < 0 || price > 100000000 {
                return nil, errors.New("Price must be between 0 and 100 mil")
            }

            timeout, cancel := context.WithTimeout(ctx, time.Second)
            defer cancel()

            var item bson.M
            err = itemsCollection.FindOne(timeout, bson.M{"_id": itemObjID}).Decode(&item)
            if err != nil {
                return nil, err
            }

            var user bson.M
            err = usersCollection.FindOne(timeout, bson.M{"_id": userObjID}).Decode(&user)
            if err != nil {
                return nil, err
            }

            res, err := listingsCollection.InsertOne(timeout, bson.M{
                "price": price,
                "deleted": false,
                "accepted": nil,
                "seller": userObjID,
                "buyer": nil,
                "item": itemObjID,
                "inquiries": bson.A{},
            })
            if err != nil {
                return nil, err
            }
            var listing bson.M
            err = listingsCollection.FindOne(timeout, bson.M{"_id": res.InsertedID}).Decode(&listing)
            if err != nil {
                return nil, err
            }

            err = addToBsonArray(timeout, itemObjID, *itemsCollection, "listings", res.InsertedID)
            if err != nil {
                listingsCollection.DeleteOne(timeout, bson.M{"_id": res.InsertedID})
                return nil, err
            }

            err = addToBsonArray(timeout, userObjID, *usersCollection, "listings", res.InsertedID)
            if err != nil {
                listingsCollection.DeleteOne(timeout, bson.M{"_id": res.InsertedID})
                pullFromBsonArray(ctx, itemObjID, *itemsCollection, "listings", res.InsertedID)
                return nil, err
            }

            return listing, nil
        },
    }
}

