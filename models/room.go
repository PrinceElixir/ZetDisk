package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Room struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
    Name          string             `json:"name" bson:"name"`
    Anonymous     bool               `json:"anonymous" bson:"anonymous"`
    Private       bool               `json:"private" bson:"private"`
    BannerPicture string             `json:"banner_picture,omitempty" bson:"banner_picture,omitempty"`
    DisplayPicture string            `json:"display_picture,omitempty" bson:"display_picture,omitempty"`
    UsedMemory    int64              `json:"used_memory" bson:"used_memory"` // in bytes
    CreatedAt     int64              `json:"created_at" bson:"created_at"`
    UpdatedAt     int64              `json:"updated_at" bson:"updated_at"`
    Contents      []Content          `json:"contents" bson:"contents"`
}
