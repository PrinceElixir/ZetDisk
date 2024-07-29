package main

import (
    "filestorage-backend/config"
    "filestorage-backend/routes"
    "github.com/gofiber/fiber/v2"
    "github.com/joho/godotenv"
    "log"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    config.LoadConfig()

    app := fiber.New()

    routes.Setup(app)

    log.Fatal(app.Listen(":3000"))
}
