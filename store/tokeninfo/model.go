package tokeninfo

import (
	"context"

	"github.com/lebensborned/medobs-test/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionName for token info
const CollectionName = "tokeninfo"

var ctx = context.TODO()

// Model for token info
type Model struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GUID         string             `json:"guid" bson:"guid"`
	RefreshToken string             `json:"refresh_token" bson:"refresh_token"`
}

// Save or update the model in DB
func (m *Model) Save(store store.Storage) error {

	f := bson.M{"guid": m.GUID}
	if !m.ID.IsZero() {
		f = bson.M{"_id": m.ID}
	}

	result, err := store.Database().Collection(CollectionName).UpdateOne(ctx, f, bson.M{"$set": m}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		m.ID = result.UpsertedID.(primitive.ObjectID)
	}
	return nil
}

// FindByGUID returns document by user GUID
func FindByGUID(store store.Storage, id string) (*Model, error) {
	var usr Model
	if err := store.Database().Collection(CollectionName).FindOne(ctx, bson.M{"guid": id}).Decode(&usr); err != nil {
		return nil, err
	}

	return &usr, nil
}
