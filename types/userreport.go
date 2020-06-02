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

// UserReportType corresponds to the "reports" collection
var UserReportType = graphql.NewObject(
    graphql.ObjectConfig {
        Name: "UserReport",
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
        },
    },
)

func InitUserReportType(ctx context.Context, db mongo.Database) {
    UserReportType.AddFieldConfig("reporter", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "reporter", *db.Collection("users")),
    })
    UserReportType.AddFieldConfig("scumbag", &graphql.Field {
        Type: UserType,
        Resolve: resolverGenerator(ctx, "scumbag", *db.Collection("users")),
    })
}

// ReportUser creates a new report from a reporter, a problematic user, and a note. If
// a report with the same users has already been created, an error is returned. For now,
// the function does not check to see if the users actually exist, only that the IDs are
// valid ObjectIDs.
func ReportUser(ctx context.Context, reportsCollection mongo.Collection) graphql.Field {
    return graphql.Field {
        Type: UserReportType,
        Description: "Report a user",
        Args: graphql.FieldConfigArgument {
            "reporterID": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "scumbagID": &graphql.ArgumentConfig {
                Type: graphql.ID,
            },
            "note": &graphql.ArgumentConfig {
                Type: graphql.String,
            },
        },
        Resolve: func (p graphql.ResolveParams) (interface{}, error) {
            reporterID, prs := p.Args["reporterID"]
            if !prs {
                return nil, errors.New("Reporter ID not given for user report")
            }
            rObjID, err := primitive.ObjectIDFromHex(reporterID.(string))
            if err != nil {
                return nil, err
            }
            scumbagID, prs := p.Args["scumbagID"]
            if !prs {
                return nil, errors.New("Scumbag ID not given for user report")
            }
            sObjID, err := primitive.ObjectIDFromHex(scumbagID.(string))
            if err != nil {
                return nil, err
            }
            note, prs := p.Args["note"]
            if !prs {
                return nil, errors.New("Note not given for user report")
            }
            timeout, cancel := context.WithTimeout(ctx, time.Second)
            defer cancel()
            var userreport bson.M
            err = reportsCollection.FindOne(timeout, bson.M{"reporter": rObjID, "scumbag": sObjID}).Decode(&userreport)
            if err != nil && err.Error() != "mongo: no documents in result" {
                return nil, err
            } else if err == nil {
                return nil, errors.New(fmt.Sprintf("Report already created"))
            }
            res, err := reportsCollection.InsertOne(timeout, bson.M{
                "reporter": rObjID,
                "scumbag": sObjID,
                "note": note,
            })
            if err != nil {
                return nil, err
            }
            err = reportsCollection.FindOne(timeout, bson.M{"_id": res.InsertedID}).Decode(&userreport)
            if err != nil {
                return nil, err
            }
            return userreport, nil
        },
    }
}

