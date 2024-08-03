package controllers

import (
    "archive/zip"
    "context"
    "filestorage-backend/config"
    "filestorage-backend/models"
    "filestorage-backend/services"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "io"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
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
    folderName := c.Query("folderName")
    var parts []*s3.CompletedPart
    if err := c.BodyParser(&parts); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    err := services.CompleteMultipartUpload(uploadID, folderName+".zip", parts)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

  
    zipFilePath := "/tmp/" + folderName + ".zip"
    err = services.DownloadFileFromS3(zipFilePath, folderName+".zip")
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to download zip file from S3",
        })
    }

    
    err = unzipFolder(zipFilePath, "/tmp/"+folderName)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to unzip folder",
        })
    }

   
    contentCollection := config.GetMongoCollection("contents")
    roomID, err := primitive.ObjectIDFromHex(folderName)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid room ID",
        })
    }

    err = filepath.Walk("/tmp/"+folderName, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            contentType := determineContentType(path)
            content := models.Content{
                ID:        primitive.NewObjectID(),
                RoomID:    roomID,
                Type:      contentType,
                URL:       "https://" + os.Getenv("S3_BUCKET_NAME") + ".s3.amazonaws.com/" + folderName + "/" + info.Name(),
                CreatedAt: time.Now().Unix(),
            }
            _, err = contentCollection.InsertOne(context.TODO(), content)
            if err != nil {
                return err
            }
        }
        return nil
    })

    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to save content metadata",
        })
    }

    return c.JSON(fiber.Map{
        "message": "Folder upload and extraction completed successfully",
    })
}

func unzipFolder(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        fPath := filepath.Join(dest, f.Name)
        if !strings.HasPrefix(fPath, filepath.Clean(dest)+string(os.PathSeparator)) {
            return fmt.Errorf("illegal file path: %s", fPath)
        }
        if f.FileInfo().IsDir() {
            os.MkdirAll(fPath, os.ModePerm)
        } else {
            os.MkdirAll(filepath.Dir(fPath), os.ModePerm)
            outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            rc, err := f.Open()
            if err != nil {
                return err
            }
            _, err = io.Copy(outFile, rc)
            outFile.Close()
            rc.Close()
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func determineContentType(filePath string) string {
    switch {
    case strings.HasSuffix(filePath, ".mp4"), strings.HasSuffix(filePath, ".mkv"):
        return "video"
    case strings.HasSuffix(filePath, ".mp3"), strings.HasSuffix(filePath, ".wav"):
        return "audio"
    case strings.HasSuffix(filePath, ".jpg"), strings.HasSuffix(filePath, ".png"):
        return "picture"
    default:
        return "unknown"
    }
}
