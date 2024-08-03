package controllers

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "filestorage-backend/config"
    "filestorage-backend/models"
    "filestorage-backend/utils"
    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)
   

func generateOTP() string {
    bytes := make([]byte, 4)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

func RequestPasswordReset(c *fiber.Ctx) error {
    type request struct {
        Email string `json:"email"`
    }

    var req request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
    }

    collection := config.GetMongoCollection("users")
    var user models.User
    err := collection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
    }

    otp := generateOTP()
    if err := utils.SetOTP(req.Email, otp); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store OTP"})
    }

    if err := utils.SendOTPEmail(req.Email, otp); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP email"})
    }

    return c.JSON(fiber.Map{"message": "OTP sent to your email"})
}

func VerifyOTP(c *fiber.Ctx) error {
    type request struct {
        Email string `json:"email"`
        OTP   string `json:"otp"`
    }

    var req request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
    }

    storedOTP, err := utils.GetOTP(req.Email)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or expired OTP"})
    }

    if storedOTP != req.OTP {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid OTP"})
    }

    return c.JSON(fiber.Map{"message": "OTP verified"})
}

func ResetPassword(c *fiber.Ctx) error {
    type request struct {
        Email    string `json:"email"`
        OTP      string `json:"otp"`
        Password string `json:"password"`
    }

    var req request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
    }

    storedOTP, err := utils.GetOTP(req.Email)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or expired OTP"})
    }

    if storedOTP != req.OTP {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid OTP"})
    }

    hashedPassword, err := hashPassword(req.Password)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
    }

    collection := config.GetMongoCollection("users")
    _, err = collection.UpdateOne(context.TODO(), bson.M{"email": req.Email}, bson.M{"$set": bson.M{"password": hashedPassword}})
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update password"})
    }

    return c.JSON(fiber.Map{"message": "Password reset successful"})
}
