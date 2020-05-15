package model

import (
	"context"
	"testing"
	"time"

	"github.com/animal-crossing-exchange/ace-server/types"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const testDBName string = "interfacesTest"

type ListingTestSuite struct {
	suite.Suite
	client *mongo.Client
	ctx    context.Context
	li     ListingInterface
}

func (suite *ListingTestSuite) SetupSuite() {
	client, err := mongo.NewClient(options.Client().ApplyURI(DatabaseURI))
	suite.Require().Nil(err)
	suite.client = client

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	suite.Require().Nil(err)
	suite.ctx = ctx

	err = client.Ping(ctx, readpref.Primary())
	suite.Require().Nil(err)

	database := client.Database(testDBName)
	suite.li = NewListingInterface(database)
}

func (suite *ListingTestSuite) TearDownSuite() {
	suite.client.Disconnect(suite.ctx)
}

func (suite *ListingTestSuite) SetupTest() {
	_, err := suite.client.Database(testDBName).Collection(listingCollection).DeleteMany(suite.ctx, bson.D{})
	suite.Require().Nil(err)
}

func TestListingSuite(t *testing.T) {
	suite.Run(t, new(ListingTestSuite))
}

func (suite *ListingTestSuite) TestInsertRetrieveSingleListing() {
	l := types.Listing{FromUser: "Username#0001", Price: 9001, ItemType: "Rock"}
	err := suite.li.CreateListing(suite.ctx, l)
	suite.Require().Nil(err)

	ls, err := suite.li.GetListings(suite.ctx)
	suite.Require().Nil(err)
	suite.Assert().ElementsMatch(ls, []types.Listing{l})
}

func (suite *ListingTestSuite) TestInsertRetrieveListings() {
	l1 := types.Listing{FromUser: "Username#0001", Price: 9001, ItemType: "Rock"}
	l2 := types.Listing{FromUser: "OtherName#0001", Price: 8999, ItemType: "Rock"}
	err := suite.li.CreateListing(suite.ctx, l1)
	suite.Require().Nil(err)
	err = suite.li.CreateListing(suite.ctx, l2)
	suite.Require().Nil(err)

	ls, err := suite.li.GetListings(suite.ctx)
	suite.Require().Nil(err)
	suite.Assert().ElementsMatch(ls, []types.Listing{l1, l2})
}

func (suite *ListingTestSuite) TestGetEmptyListings() {
	ls, err := suite.li.GetListings(suite.ctx)
	suite.Require().Nil(err)
	suite.Assert().Empty(ls)
}
