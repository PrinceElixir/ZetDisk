package services

import (
    "context"
    "filestorage-backend/config"
    "filestorage-backend/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func init() {
    clientOptions := options.Client().ApplyURI(config.Get("MONGODB_URI"))
    var err error
    client, err = mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        panic(err)
    }
}

func GetMongoCollection(collectionName string) *mongo.Collection {
    return client.Database("filestorage").Collection(collectionName)
}

func FindUserByEmail(email string) (*models.User, error) {
    collection := GetMongoCollection("users")
    user := &models.User{}
    err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(user)
    return user, err
}

func CreateUser(user *models.User) error {
    collection := GetMongoCollection("users")
    _, err := collection.InsertOne(context.TODO(), user)
    return err
}
