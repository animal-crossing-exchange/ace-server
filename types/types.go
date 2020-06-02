// Package types contains GraphQL types
package types

import (
    "context"
    "errors"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/graphql-go/graphql"
)

// TODO: add resolvers for IDs and created timestamps

// Some types have structs, but none of them are being used. I have left them in case we need
// them down the road.

// idResolver translates the MongoDB document's _id field to the id field of the GraphQL types.
func idResolver(p graphql.ResolveParams) (interface{}, error) {
    sourceObj := p.Source.(primitive.M)
    id := sourceObj["_id"].(primitive.ObjectID).String()[10:34]
    return id, nil
}

// timestampResolver gets the creation time of the document from its _id.
func timestampResolver(p graphql.ResolveParams) (interface{}, error) {
    sourceObj := p.Source.(primitive.M)
    id := sourceObj["_id"].(primitive.ObjectID)
    return id.Timestamp(), nil
}

// resolverGenerator provides a common implementation of a Resolve function on a GraphQL
// object. In the database, documents will store the ObjectID of the sub-document they reference
// under a key, but the GraphQL operation must return the document itself. This is performed
// by the type's Resolve function, which this function can generate since the logic is
// the same for any type. It takes a (probably Background) Context, a string representing the
// key to pull the sub-document from, and the MongoDB collection the sub-document is located in.
func resolverGenerator(ctx context.Context, objKey string, collection mongo.Collection) graphql.FieldResolveFn {
    return func (p graphql.ResolveParams) (interface{}, error) {
        sourceObj := p.Source.(primitive.M) // upper level document
        switch targetObj := sourceObj[objKey].(type) { // objKey could point to...
        case primitive.A: // an array of ObjectIDs
            targetObjArray := make([]bson.M, len(targetObj))
            // there is a method that works with Find to get a lot of documents at once
            // but this didn't seem to work
            for i, id := range targetObj {
                var obj bson.M
                timeout, cancel := context.WithTimeout(ctx, time.Second * 1)
                err := collection.FindOne(timeout, bson.M{"_id": id}).Decode(&obj)
                if err != nil {
                    cancel()
                    return nil, err
                }
                targetObjArray[i] = obj
                cancel()
            }
            return targetObjArray, nil
        case primitive.ObjectID: // or a singular ObjectID
            var obj bson.M
            timeout, cancel := context.WithTimeout(ctx, time.Second * 1)
            defer cancel()
            err := collection.FindOne(timeout, bson.M{"_id": targetObj}).Decode(&obj)
            if err != nil {
                return nil, err
            }
            return obj, nil
        case primitive.Null: // or it could be null
            // TODO: make sure this works when we have operations for types that can have null values
            return nil, nil
        default: // or objKey might not be part of the document
            return nil, errors.New(fmt.Sprintf("Invalid param for extraction from db.%s: %s", collection.Name(), objKey))
        }
    }
}

