// Package schema initializes the GraphQL types and creates the query and mutation
// schema for use in the main app
package schema

import (
    "github.com/animal-crossing-exchange/ace-server/types"

    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "github.com/graphql-go/graphql"
)

// GenerateQuerySchema creates the Fields object containing queries. It also initializes
// the GraphQL types, so this function needs to be called before GenerateMutationSchema.
func GenerateQuerySchema(ctx context.Context, db mongo.Database) graphql.Fields {
    types.InitItemType(ctx, db)
    types.InitItemMarketRecordType(ctx, db)
    types.InitListingType(ctx, db)
    types.InitListingInquiry(ctx, db)
    types.InitTransactionType(ctx, db)
    types.InitUserType(ctx, db)

    GetItem := types.GetItem(ctx, *db.Collection("items"))
    GetItems := types.Items(ctx, *db.Collection("items"))

    return graphql.Fields {
        "item": &GetItem,
        "items": &GetItems,
    }
}

// GenerateMutationSchema creates the Fields object containing mutations. This function
// should be called after GenerateQuerySchema.
func GenerateMutationSchema(ctx context.Context, db mongo.Database) graphql.Fields {
    AddUser := types.AddUser(ctx, *db.Collection("users"))
    BanUser := types.BanUser(ctx, *db.Collection("users"))
    DeleteUser := types.DeleteUser(ctx, *db.Collection("users"))
    SetUserAdmin := types.SetUserAdmin(ctx, *db.Collection("users"))

    return graphql.Fields {
        "addUser": &AddUser,
        "banUser": &BanUser,
        "deleteUser": &DeleteUser,
        "setUserAdmin": &SetUserAdmin,
    }
}

