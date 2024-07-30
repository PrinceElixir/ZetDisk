package services

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func InitiateMultipartUpload(fileName string) (string, string, error) {
    svc := s3.New(session.Must(session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })))

    input := &s3.CreateMultipartUploadInput{
        Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
        Key:    aws.String(fileName),
    }

    result, err := svc.CreateMultipartUpload(input)
    if err != nil {
        return "", "", err
    }

    return *result.UploadId, *result.Key, nil
}

func UploadFileParts(file multipart.File, fileName, uploadID string) ([]*s3.CompletedPart, error) {
    svc := s3.New(session.Must(session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })))

    var parts []*s3.CompletedPart
    buffer := make([]byte, 5*1024*1024) 

    partNumber := int64(1)
    for {
        bytesRead, err := file.Read(buffer)
        if err != nil && err != io.EOF {
            return nil, err
        }

        if bytesRead == 0 {
            break
        }

        uploadInput := &s3.UploadPartInput{
            Body:       bytes.NewReader(buffer[:bytesRead]),
            Bucket:     aws.String(os.Getenv("S3_BUCKET_NAME")),
            Key:        aws.String(fileName),
            PartNumber: aws.Int64(partNumber),
            UploadId:   aws.String(uploadID),
        }

        uploadResult, err := svc.UploadPart(uploadInput)
        if err != nil {
            return nil, err
        }

        parts = append(parts, &s3.CompletedPart{
            ETag:       uploadResult.ETag,
            PartNumber: aws.Int64(partNumber),
        })

        partNumber++
    }

    return parts, nil
}

func CompleteMultipartUpload(uploadID, fileName string, parts []*s3.CompletedPart) error {
    svc := s3.New(session.Must(session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })))

    input := &s3.CompleteMultipartUploadInput{
        Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
        Key:    aws.String(fileName),
        MultipartUpload: &s3.CompletedMultipartUpload{
            Parts: parts,
        },
        UploadId: aws.String(uploadID),
    }

    _, err := svc.CompleteMultipartUpload(input)
    return err
}

func GetPresignedURL(uploadID, fileName string, partNumber int64) (string, error) {
    svc := s3.New(session.Must(session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })))

    req, _ := svc.UploadPartRequest(&s3.UploadPartInput{
        Bucket:     aws.String(os.Getenv("S3_BUCKET_NAME")),
        Key:        aws.String(fileName),
        PartNumber: aws.Int64(partNumber),
        UploadId:   aws.String(uploadID),
    })

    url, err := req.Presign(15 * time.Minute)
    if err != nil {
        return "", err
    }

    return url, nil
}