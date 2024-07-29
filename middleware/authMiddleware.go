package middleware

import (
    "filestorage-backend/utils"
    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v4"
)

func JWTProtected() fiber.Handler {
    return func(c *fiber.Ctx) error {
        tokenString := c.Get("Authorization")
        if tokenString == "" {
            return c.Status(fiber.StatusUnauthorized).SendString("Missing or malformed JWT")
        }

        claims := &utils.JWTClaims{}
        token, err := jwt.ParseWithClaims(tokenString[len("Bearer "):], claims, func(token *jwt.Token) (interface{}, error) {
            return []byte(utils.JWTSecret), nil
        })

        if err != nil || !token.Valid {
            return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired JWT")
        }

        c.Locals("user", claims.UserID)
        return c.Next()
    }
}
