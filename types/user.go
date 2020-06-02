package types

import (
    "context"
    "errors"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/graphql-go/graphql"
)

type UserStruct struct {
    id string
    discordId int
    lastLogin string
    reputation int
    admin bool
    banned string
    banNote string
    inquiries []*ListingInquiryStruct
    transactions []*TransactionStruct
}

// UserType corresponds to the "users" collection
var UserType = graphql.NewObject(
    graphql.ObjectConfig {
        Name: "User",
        Fields: graphql.Fields {
            "id": &graphql.Field {
                Type: graphql.ID,
                Resolve: idResolver,
            },
            "discordID": &graphql.Field {
                Type: graphql.Int,
            },
            "created": &graphql.Field {
                Type: graphql.String,
                Resolve: timestampResolver,
            },
            "lastLogin": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
            "reputation": &graphql.Field {
                Type: graphql.Int,
            },
            "admin": &graphql.Field {
                Type: graphql.Boolean,
            },
            "banned": &graphql.Field {
                Type: graphql.String, // TODO Date scalar
            },
            "banNote": &graphql.Field {
                Type: graphql.String,
            },
        },
    },
)

func InitUserType(ctx context.Context, db mongo.Database) {
    UserType.AddFieldConfig("inquiries", &graphql.Field {
        Type: graphql.NewList(ListingInquiryType),
        Resolve: resolverGenerator(ctx, "inquiries", *db.Collection("inquiries")),
    })
    UserType.AddFieldConfig("transactions", &graphql.Field {
        Type: graphql.NewList(TransactionType),
        Resolve: resolverGenerator(ctx, "transactions", *db.Collection("transactions")),
    })
}

// AddUser creates a new user from a Discord ID. Before adding the user to the DB,
// it checks to make sure the user doesn't already exist.
func AddUser(ctx context.Context, usersCollection mongo.Collection) graphql.Field {
    return graphql.Field {
        Type: UserType,
        Description: "Create a new user",
        Args: graphql.FieldConfigArgument {
            "discordID": &graphql.ArgumentConfig {
                Type: graphql.Int,
            },
        },
        Resolve: func (p graphql.ResolveParams) (interface{}, error) {
            discordID := p.Args["discordID"]
            var user bson.M
            timeout, cancel := context.WithTimeout(ctx, time.Second)
            defer cancel()
            err := usersCollection.FindOne(timeout, bson.M{"discordID": discordID}).Decode(&user)
            if err != nil && err.Error() != "mongo: no documents in result" {
                return nil, err
            } else if err == nil {
                return nil, errors.New(fmt.Sprintf("User with Discord ID already in DB: %d", discordID))
            }
            _, err = usersCollection.InsertOne(timeout, bson.M{
                "discordID": discordID,
                "lastLogin": primitive.NewDateTimeFromTime(time.Now()),
                "reputation": 0,
                "admin": false,
                "banned": nil,
                "banNote": nil,
                "transactions": bson.A{},
                "listings": bson.A{},
                "inquiries": bson.A{},
            })
            if err != nil {
                return nil, err
            }
            err = usersCollection.FindOne(timeout, bson.M{"discordID": discordID}).Decode(&user)
            if err != nil {
                return nil, err
            }
            return user, nil
        },
    }
}

// SetUserAdmin sets the admin boolean on a user in the database.
func SetUserAdmin(ctx context.Context, usersCollection mongo.Collection) graphql.Field {
    return graphql.Field {
        Type: UserType,
        Description: "Update a user's admin status",
        Args: graphql.FieldConfigArgument {
            "id": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "isAdmin": &graphql.ArgumentConfig {
                Type: graphql.Boolean,
            },
        },
        Resolve: func (p graphql.ResolveParams) (interface{}, error) {
            id, prs := p.Args["id"]
            if !prs {
                return nil, errors.New("No user ID given for admin update")
            }
            isAdmin, prs := p.Args["isAdmin"]
            if !prs {
                return nil, errors.New("No user ID given for admin update")
            }
            objID, err := primitive.ObjectIDFromHex(id.(string))
            if err != nil {
                return nil, err
            }
            filter := bson.D{{"_id", objID}}
            update := bson.D{{"$set", bson.D{{"admin", isAdmin}}}}
            opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
            timeout, cancel := context.WithTimeout(ctx, time.Second * 3)
            defer cancel()
            var updatedUser bson.M
            err = usersCollection.FindOneAndUpdate(timeout, filter, update, opts).Decode(&updatedUser)
            if err != nil {
                return nil, err
            }
            return updatedUser, nil
        },
    }
}

// DeleteUser deletes a user from the database.
func DeleteUser(ctx context.Context, usersCollection mongo.Collection) graphql.Field {
    return graphql.Field {
        Type: UserType,
        Description: "Delete a user",
        Args: graphql.FieldConfigArgument {
            "id": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
        },
        Resolve: func (p graphql.ResolveParams) (interface{}, error) {
            id, prs := p.Args["id"]
            if !prs {
                return nil, errors.New("No user ID given for user deletion")
            }
            objID, err := primitive.ObjectIDFromHex(id.(string))
            if err != nil {
                return nil, err
            }
            filter := bson.D{{"_id", objID}}
            opts := options.FindOneAndDelete()
            timeout, cancel := context.WithTimeout(ctx, time.Second * 3)
            defer cancel()
            var deletedUser bson.M
            err = usersCollection.FindOneAndDelete(timeout, filter, opts).Decode(&deletedUser)
            if err != nil {
                return nil, err
            }
            return deletedUser, nil
        },
    }
}

// BanUser sets the ban date on a user's record in the database. A note can optionally
// be provided. If the user is already banned, the ban date and note will simply be updated.
func BanUser(ctx context.Context, usersCollection mongo.Collection) graphql.Field {
    return graphql.Field {
        Type: UserType,
        Description: "Ban a user",
        Args: graphql.FieldConfigArgument {
            "id": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "note": &graphql.ArgumentConfig {
                Type: graphql.String,
                DefaultValue: nil,
            },
        },
        Resolve: func (p graphql.ResolveParams) (interface{}, error) {
            id, prs := p.Args["id"]
            if !prs {
                return nil, errors.New("No user ID given for user ban")
            }
            objID, err := primitive.ObjectIDFromHex(id.(string))
            if err != nil {
                return nil, err
            }
            note := p.Args["note"]
            date := time.Now().String()
            filter := bson.M{"_id": objID}
            update := bson.D{{"$set", bson.M{"banned": date, "banNote": note}}}
            opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
            timeout, cancel := context.WithTimeout(ctx, time.Second)
            defer cancel()
            var bannedUser bson.M
            err = usersCollection.FindOneAndUpdate(timeout, filter, update, opts).Decode(&bannedUser)
            if err != nil {
                return nil, err
            }
            return bannedUser, nil
        },
    }
}

