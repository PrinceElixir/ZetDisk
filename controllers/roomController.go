package controllers

import (
    "context"
    "filestorage-backend/config"
    "filestorage-backend/models"
    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

func CreateRoom(c *fiber.Ctx) error {
    user := c.Locals("user").(string)
    userID, _ := primitive.ObjectIDFromHex(user)

    room := new(models.Room)
    if err := c.BodyParser(room); err != nil {
        return c.Status(fiber.StatusBadRequest).SendString(err.Error())
    }

   
    if room.Name == "" {
        return c.Status(fiber.StatusBadRequest).SendString("Room name is required")
    }

    
    userCollection := config.GetMongoCollection("users")
    var dbUser models.User
    err := userCollection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&dbUser)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch user")
    }

    roomCollection := config.GetMongoCollection("rooms")
    var roomCount int64
    roomCount, err = roomCollection.CountDocuments(context.TODO(), bson.M{"user_id": userID})
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to count rooms")
    }

   
    if dbUser.FreeVersion && roomCount >= 5 {
        return c.Status(fiber.StatusForbidden).SendString("Free version users can create a maximum of 5 rooms")
    }

   
    room.ID = primitive.NewObjectID()
    room.UserID = userID
    room.UsedMemory = 0
    room.CreatedAt = time.Now().Unix()
    room.UpdatedAt = time.Now().Unix()

    _, err = roomCollection.InsertOne(context.TODO(), room)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to create room")
    }

    return c.JSON(room)
}
