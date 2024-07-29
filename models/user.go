package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Name        string             `json:"name" bson:"name"`
    Email       string             `json:"email" bson:"email"`
    Password    string             `json:"password,omitempty" bson:"password"`
    DisplayName string             `json:"display_name,omitempty" bson:"display_name,omitempty"`
    Picture     string             `json:"picture,omitempty" bson:"picture,omitempty"`
    FreeVersion bool               `json:"free_version" bson:"free_version"`
}
