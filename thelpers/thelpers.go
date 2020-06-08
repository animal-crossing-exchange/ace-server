package thelpers

import (
    "github.com/animal-crossing-exchange/ace-server/schema"

    "context"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "strings"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/graphql-go/graphql"
)

var db mongo.Database

func StartServer(dbName string, end chan bool) {
    ctx := context.Background()

    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        panic(err)
    }
    pingTimeout, cancel := context.WithTimeout(ctx, time.Second)
    defer cancel()
    err = client.Ping(pingTimeout, nil)
    if err != nil {
        panic(err)
    }
    defer client.Disconnect(ctx)

    rootQuery := graphql.ObjectConfig{ Name: "RootQuery", Fields: schema.GenerateQuerySchema(ctx, *client.Database(dbName)) }
    rootMutation := graphql.ObjectConfig{ Name: "RootMutation", Fields: schema.GenerateMutationSchema(ctx, *client.Database(dbName)) }
    schemaConfig := graphql.SchemaConfig{ Query: graphql.NewObject(rootQuery), Mutation: graphql.NewObject(rootMutation) }
    schema, err := graphql.NewSchema(schemaConfig)
    if err != nil {
        panic(err)
    }

    http.HandleFunc("/test/graphql", func(w http.ResponseWriter, r *http.Request) {
        query := r.URL.Query().Get("query")
        params := graphql.Params{ Schema: schema, RequestString: query }
        result := graphql.Do(params)
        /*
        if len(result.Errors) > 0 {
            for _, err := range result.Errors {
                log.Print(err)
            }
        }
        */
        json.NewEncoder(w).Encode(result)
    })

    server := &http.Server {
        Addr: ":8081",
    }

    go server.ListenAndServe()
    end <- true
    <-end
    server.Close()
    end <- true
}

func SetupDB() *mongo.Client {
    ctx := context.Background()
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        panic(err)
    }
    pingTimeout, cancel := context.WithTimeout(ctx, time.Second)
    defer cancel()
    err = client.Ping(pingTimeout, nil)
    if err != nil {
        panic(err)
    }
    return client
}

func ClearDB(ctx context.Context, db mongo.Database) {
    _, err := db.Collection("items").DeleteMany(ctx, bson.M{}, nil)
    if err != nil {
        panic(err)
    }
    _, err = db.Collection("records").DeleteMany(ctx, bson.M{}, nil)
    if err != nil {
        panic(err)
    }
    _, err = db.Collection("listings").DeleteMany(ctx, bson.M{}, nil)
    if err != nil {
        panic(err)
    }
    _, err = db.Collection("inquiries").DeleteMany(ctx, bson.M{}, nil)
    if err != nil {
        panic(err)
    }
    _, err = db.Collection("transactions").DeleteMany(ctx, bson.M{}, nil)
    if err != nil {
        panic(err)
    }
    _, err = db.Collection("users").DeleteMany(ctx, bson.M{}, nil)
    if err != nil {
        panic(err)
    }
    _, err = db.Collection("reports").DeleteMany(ctx, bson.M{}, nil)
    if err != nil {
        panic(err)
    }
}

func ExecQuery(query string) map[string]interface{} {
    query = strings.ReplaceAll(query, "\n", "")
    query = strings.ReplaceAll(query, " ", "+")
    resp, err := http.Get("http://localhost:8081/test/graphql?query=" + query)
    if err != nil {
        panic(err)
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    var j map[string]interface{}
    err = json.Unmarshal(body, &j)
    if err != nil {
        panic(err)
    }
    return j
}

