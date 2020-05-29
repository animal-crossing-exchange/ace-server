package main

import (
    "api-test/schema"

    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "net/http"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/graphql-go/graphql"
)

func main() {
    ctx := context.Background()

    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    pingTimeout, _ := context.WithTimeout(ctx, time.Second)
    err = client.Ping(pingTimeout, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    rootQuery := graphql.ObjectConfig{ Name: "RootQuery", Fields: schema.GenerateQuerySchema(ctx, *client.Database("acex")) }
    rootMutation := graphql.ObjectConfig{ Name: "RootMutation", Fields: schema.GenerateMutationSchema(ctx, *client.Database("acex")) }
    schemaConfig := graphql.SchemaConfig{ Query: graphql.NewObject(rootQuery), Mutation: graphql.NewObject(rootMutation) }
    schema, err := graphql.NewSchema(schemaConfig)
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("hit")
        query := r.URL.Query().Get("query")
        params := graphql.Params{ Schema: schema, RequestString: query }
        result := graphql.Do(params)
        if len(result.Errors) > 0 {
            for _, err := range result.Errors {
                log.Print(err)
            }
        }
        json.NewEncoder(w).Encode(result)
    })

    fmt.Println("API started")
    http.ListenAndServe(":8080", nil)

}

