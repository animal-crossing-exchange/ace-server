package model

import (
	"context"

	"github.com/animal-crossing-exchange/ace-server/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const listingCollection string = "listings"

type ListingInterface interface {
	GetListings(ctx context.Context) ([]types.Listing, error)
	CreateListing(ctx context.Context, listing types.Listing) error
}

type ListingImplementation struct {
	coll *mongo.Collection
}

func NewListingInterface(db *mongo.Database) ListingInterface {
	return ListingImplementation{db.Collection(listingCollection)}
}

func (li ListingImplementation) GetListings(ctx context.Context) ([]types.Listing, error) {
	var listings []types.Listing
	cursor, err := li.coll.Find(ctx, bson.D{})
	if err != nil {
		return listings, err
	}

	if err = cursor.All(ctx, &listings); err != nil {
		return listings, err
	}

	return listings, nil
}

func (li ListingImplementation) CreateListing(ctx context.Context, listing types.Listing) error {
	_, err := li.coll.InsertOne(ctx, listing)
	return err
}
