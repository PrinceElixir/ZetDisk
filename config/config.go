package config

import (
    "context"
    "log"

    "github.com/spf13/viper"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func LoadConfig() {
    viper.SetConfigFile(".env")
    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error loading .env file: %s", err)
    }
}

func Get(key string) string {
    return viper.GetString(key)
}

func GetMongoCollection(collectionName string) *mongo.Collection {
    clientOptions := options.Client().ApplyURI(Get("MONGODB_URI"))
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatalf("Error connecting to MongoDB: %s", err)
    }
    return client.Database("filestorage").Collection(collectionName)
}
