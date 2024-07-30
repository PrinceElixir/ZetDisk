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

func AddContent(c *fiber.Ctx) error {
    user := c.Locals("user").(string)
    userID, _ := primitive.ObjectIDFromHex(user)

    content := new(models.Content)
    if err := c.BodyParser(content); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    if content.RoomID.IsZero() || content.Type == "" || content.URL == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "RoomID, Type, and URL are required",
        })
    }

    roomCollection := config.GetMongoCollection("rooms")
    var room models.Room
    err := roomCollection.FindOne(context.TODO(), bson.M{"_id": content.RoomID, "user_id": userID}).Decode(&room)
    if err != nil {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Room not found or you do not have permission",
        })
    }

    content.ID = primitive.NewObjectID()
    content.CreatedAt = time.Now().Unix()

    update := bson.M{
        "$push": bson.M{"contents": content},
        "$set":  bson.M{"updated_at": time.Now().Unix()},
    }
    _, err = roomCollection.UpdateOne(context.TODO(), bson.M{"_id": content.RoomID}, update)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to add content to room",
        })
    }

    return c.JSON(content)
}
