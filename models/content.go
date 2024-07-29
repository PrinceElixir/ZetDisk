package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Content struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    RoomID    primitive.ObjectID `bson:"room_id" json:"room_id"`
    Type      string             `json:"type" bson:"type"` // "this type can be either video, audio or picture you can adjust this and make changes in the model, queuing is left, this mode; might have changes after that"
    URL       string             `json:"url" bson:"url"`
    CreatedAt int64              `json:"created_at" bson:"created_at"`
}
