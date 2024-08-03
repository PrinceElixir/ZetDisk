package services

import (
	"archive/zip"
	"bytes"

	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

func ZipFolder(srcFolder string) (string, error) {
	zipFile, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(srcFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcFolder, path)
		if err != nil {
			return err
		}

		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		_, err = io.Copy(zipFileWriter, fileToZip)
		return err
	})

	if err != nil {
		return "", err
	}

	return zipFile.Name(), nil
}

func UploadFolder(folderPath, uploadID string) error {
	zipFilePath, err := ZipFolder(folderPath)
	if err != nil {
		return err
	}
	defer os.Remove(zipFilePath)

	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	
	if err != nil {
		return err
	}

	parts, err := UploadFileParts(zipFile, filepath.Base(zipFilePath), uploadID)
	if err != nil {
		return err
	}

	return CompleteMultipartUpload(uploadID, filepath.Base(zipFilePath), parts)
}

func DownloadFileFromS3(filePath, key string) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))
	downloader := s3manager.NewDownloader(sess)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}
