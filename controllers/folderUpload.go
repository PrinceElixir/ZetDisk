package controllers

import (
    "filestorage-backend/services"
    "github.com/gofiber/fiber/v2"

)

func InitiateFolderUpload(c *fiber.Ctx) error {
    folderName := c.Query("folderName")

    uploadID, key, err := services.InitiateMultipartUpload(folderName + ".zip")
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "uploadID": uploadID,
        "key":      key,
    })
}

func CompleteFolderUpload(c *fiber.Ctx) error {
    uploadID := c.Query("uploadID")
    folderPath := c.Query("folderPath")

    err := services.UploadFolder(folderPath, uploadID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Folder upload completed successfully",
    })
}
