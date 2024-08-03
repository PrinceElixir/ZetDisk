package routes

import (
    "filestorage-backend/controllers"
    "filestorage-backend/middleware"
    "github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
    api := app.Group("/api")

    api.Post("/register", controllers.Register)
    api.Post("/login", controllers.Login)
    api.Post("/password/reset/request", controllers.RequestPasswordReset)
    api.Post("/password/reset/verify", controllers.VerifyOTP)
    api.Post("/password/reset", controllers.ResetPassword)

    api.Get("/auth/google", controllers.GoogleLogin)
    api.Get("/auth/google/callback", controllers.GoogleCallback)

    protected := api.Group("/protected", middleware.JWTProtected())

    protected.Get("/", controllers.Protected)
    protected.Post("/rooms", controllers.CreateRoom)
    // protected.Post("/contents", controllers.UploadContent)
    protected.Post("/upload/initiate", controllers.InitiateUpload)
    protected.Get("/upload/url", controllers.GetUploadURL)
    protected.Post("/upload/complete", controllers.CompleteUpload)
}
