package ttypes

import (
    "github.com/animal-crossing-exchange/ace-server/thelpers"

    "context"
    "os"
    "testing"

    "go.mongodb.org/mongo-driver/mongo"
)

var ctx context.Context
var db mongo.Database

func TestMain(m *testing.M) {
    ctx := context.Background()
    client := thelpers.SetupDB()
    defer client.Disconnect(ctx)
    db = *client.Database("acex_test")
    thelpers.ClearDB(ctx, db)

    end := make(chan bool)
    go thelpers.StartServer("acex_test", end)
    <-end

    res := m.Run()
    thelpers.ClearDB(ctx, db)

    end <- true
    <-end
    os.Exit(res)
}

func TestAddUser(t *testing.T) {
    query := `
    mutation {
        addUser(discordID: 1337) {
            discordID
            lastLogin
            reputation
            admin
            banned
            banNote
            inquiries {
                id
            }
            listings {
                id
            }
            transactions {
                id
            }
        }
    }`

    result := thelpers.ExecQuery(query)
    data := result["data"].(map[string]interface{})["addUser"].(map[string]interface{})

    item, prs := data["discordID"]
    if !prs {
        t.Error("AddUser: discordID not in result")
    } else if item.(float64) != 1337 {
        t.Errorf("AddUser: Wrong discordID, expected %d, got %f", 1337, item.(float64))
    }

    item, prs = data["lastLogin"]
    if !prs {
        t.Error("AddUser: lastLogin not in result")
    }

    item, prs = data["reputation"]
    if !prs {
        t.Error("AddUser: reputation not in result")
    } else if item.(float64) != 0 {
        t.Errorf("AddUser: Wrong reputation, expected %d, got %f", 0, item.(float64))
    }

    item, prs = data["admin"]
    if !prs {
        t.Error("AddUser: admin not in result")
    } else if item.(bool) != false {
        t.Errorf("AddUser: Wrong admin, expected %t, got %t", false, item.(bool))
    }

    item, prs = data["banned"]
    if !prs {
        t.Error("AddUser: banned not in result")
    } else if item != nil {
        t.Error("AddUser: banned not nil")
    }

    item, prs = data["banNote"]
    if !prs {
        t.Error("AddUser: banNote not in result")
    } else if item != nil {
        t.Error("AddUser: banNote not nil")
    }

    item, prs = data["inquiries"]
    if !prs {
        t.Error("AddUser: inquiries not in result")
    } else if len(item.([]interface{})) != 0 {
        t.Error("AddUser: inquiries not empty")
    }

    item, prs = data["listings"]
    if !prs {
        t.Error("AddUser: listings not in result")
    } else if len(item.([]interface{})) != 0 {
        t.Error("AddUser: listings not empty")
    }

    item, prs = data["transactions"]
    if !prs {
        t.Error("AddUser: transactions not in result")
    } else if len(item.([]interface{})) != 0 {
        t.Error("AddUser: transactions not empty")
    }
}

