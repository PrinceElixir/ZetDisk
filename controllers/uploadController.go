package controllers

import (
    "context"
    "filestorage-backend/config"
    "filestorage-backend/models"
    "filestorage-backend/services"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "os"
    "strconv"
    "time"
)

func InitiateUpload(c *fiber.Ctx) error {
    fileName := c.Query("fileName")

    uploadID, key, err := services.InitiateMultipartUpload(fileName)
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

func GetUploadURL(c *fiber.Ctx) error {
    uploadID := c.Query("uploadID")
    fileName := c.Query("fileName")
    partNumber, err := strconv.ParseInt(c.Query("partNumber"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid part number",
        })
    }

    url, err := services.GetPresignedURL(uploadID, fileName, partNumber)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "url": url,
    })
}

func CompleteUpload(c *fiber.Ctx) error {
    uploadID := c.Query("uploadID")
    fileName := c.Query("fileName")
    var parts []*s3.CompletedPart
    if err := c.BodyParser(&parts); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    err := services.CompleteMultipartUpload(uploadID, fileName, parts)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    // here we save metadata to mongoDb database
    contentCollection := config.GetMongoCollection("contents")
    contentID := primitive.NewObjectID()

    roomID, err := primitive.ObjectIDFromHex(fileName)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid room ID",
        })
    }

    content := models.Content{
        ID:        contentID,
        RoomID:    roomID, 
        Type:      "",
        URL:       "https://" + os.Getenv("S3_BUCKET_NAME") + ".s3.amazonaws.com/" + fileName,
        CreatedAt: time.Now().Unix(),
    }

    _, err = contentCollection.InsertOne(context.TODO(), content)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to save content metadata",
        })
    }

    return c.JSON(content)
}


// func UploadContent(c *fiber.Ctx) error {
   
//     roomID := c.FormValue("room_id")
//     fileType := c.FormValue("type") 
//     file, err := c.FormFile("file")

//     if err != nil {
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//             "error": "File not provided",
//         })
//     }

//     roomIDHex, err := primitive.ObjectIDFromHex(roomID)
//     if err != nil {
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//             "error": "Invalid room ID",
//         })
//     }

//     fileType = file.Header.Get("Content-Type")
//     if fileType == "" {
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//             "error": "File type not provided",
//         })
//     }

//     fileHeader, err := file.Open()
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "error": "Unable to open file",
//         })
//     }
//     defer fileHeader.Close()

    
//     fileName := primitive.NewObjectID().Hex() + "-" + file.Filename

//     uploadID, key, err := services.InitiateMultipartUpload(fileName)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "error": "Failed to initiate upload",
//         })
//     }

    
//     parts, err := services.UploadFileParts(fileHeader, fileName, uploadID)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "error": "Failed to upload file parts",
//         })
//     }

    
//     err = services.CompleteMultipartUpload(uploadID, fileName, parts)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "error": "Failed to complete upload",
//         })
//     }

    
//     contentCollection := config.GetMongoCollection("contents")
//     contentID := primitive.NewObjectID()

//     content := models.Content{
//         ID:        contentID,
//         RoomID:    roomIDHex,
//         Type:      fileType,
//         URL:       "https://" + os.Getenv("S3_BUCKET_NAME") + ".s3.amazonaws.com/" + key,
//         CreatedAt: time.Now().Unix(),
//     }

//     _, err = contentCollection.InsertOne(context.TODO(), content)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//             "error": "Failed to save content metadata",
//         })
//     }

//     return c.JSON(content)
// }
