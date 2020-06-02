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

type ItemStruct struct {
    ID string `bson:"_id"`
    Name string `bson:"name"`
    Variations []string `bson:"variations"`
    Category string `bson:"category"`
    InGamePrice int `bson:"inGamePrice"`
    CurrentAvg int `bson:"currentAvg"`
    CurrentMedian int `bson:"currentMedian"`
    Records []*ItemMarketRecordStruct `bson:"records"`
    Listings []*ListingStruct `bson:"listings"`
}

// ItemType corresponds to the "items" collection
var ItemType = graphql.NewObject(
    graphql.ObjectConfig {
        Name: "Item",
        Fields: graphql.Fields {
            "id": &graphql.Field {
                Type: graphql.ID,
                Resolve: idResolver,
            },
            "name": &graphql.Field {
                Type: graphql.String,
            },
            "variations": &graphql.Field {
                Type: graphql.NewList(graphql.String),
            },
            "category": &graphql.Field {
                Type: graphql.String,
            },
            "inGamePrice": &graphql.Field {
                Type: graphql.Int,
            },
            "currentAvg": &graphql.Field {
                Type: graphql.Int,
            },
            "currentMedian": &graphql.Field {
                Type: graphql.Int,
            },
        },
    },
)

func InitItemType(ctx context.Context, db mongo.Database) {
    ItemType.AddFieldConfig("records", &graphql.Field {
        Type: graphql.NewList(ItemMarketRecordType),
        Resolve: resolverGenerator(ctx, "records", *db.Collection("records")),
    })
    ItemType.AddFieldConfig("listings", &graphql.Field {
        Type: graphql.NewList(ListingType),
        Resolve: resolverGenerator(ctx, "listings", *db.Collection("listings")),
    })
}

// GetItem is a query for getting an item by either ID or name.
func GetItem(ctx context.Context, itemsCollection mongo.Collection) graphql.Field {
    return graphql.Field {
        Type: ItemType,
        Description: "Get an Item by name",
        Args: graphql.FieldConfigArgument {
            "id": &graphql.ArgumentConfig {
                Type: graphql.ID,
                DefaultValue: nil,
            },
            "name": &graphql.ArgumentConfig {
                Type: graphql.String,
            },
        },
        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
            var result bson.M
            timeout, cancel := context.WithTimeout(ctx, time.Second * 3)
            defer cancel()
            var err error
            if id, prs := p.Args["id"]; prs {
                objID, err := primitive.ObjectIDFromHex(id.(string))
                if err != nil {
                    return nil, err
                }
                err = itemsCollection.FindOne(timeout, bson.M{"_id": objID}).Decode(&result)
            } else if name, prs := p.Args["name"]; prs {
                err = itemsCollection.FindOne(timeout, bson.M{"name": name}).Decode(&result)
            } else {
                return nil, errors.New("No arguments given, need Item 'name' or 'ID'")
            }
            if err != nil {
                return nil, err
            }
            return result, nil
        },
    }
}

// Items is a query for getting all items. This is a very expensive query so we
// might need to limit it somehow.
func Items(ctx context.Context, itemsCollection mongo.Collection) graphql.Field {
    return graphql.Field {
        Type: graphql.NewList(ItemType),
        Description: "Get all Items",
        Resolve: func(params graphql.ResolveParams) (interface{}, error) {
            timeout, cancel := context.WithTimeout(ctx, time.Second * 3)
            defer cancel()
            cursor, err := itemsCollection.Find(timeout, bson.D{})
            if err != nil {
                return nil, err
            }
            items := make([]bson.M, 0)
            for cursor.Next(timeout) {
                var item bson.M
                if err = cursor.Decode(&item); err != nil {
                    return nil, err
                }
                items = append(items, item)
            }
            return items, nil
        },
    }
}

