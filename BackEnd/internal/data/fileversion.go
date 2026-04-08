package data

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// FileVersion uses UUID to classify different batch of comments
type FileVersion struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UUID      string             `bson:"uuid"`
	CreatedAt time.Time          `bson:"createdAt"`
}
