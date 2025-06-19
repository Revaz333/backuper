package pkg

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	Client *s3.Client
	Bucket string
}

func NewStorage(
	region string,
	accessKey string,
	secretKey string,
	endpoint string,
	bucket string,
) (*Storage, error) {

	baseCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return &Storage{}, fmt.Errorf("failed to s3 load config: %v", err)
	}

	s3Client := s3.NewFromConfig(baseCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &Storage{s3Client, bucket}, nil
}

func (s *Storage) CheckFolderExist(folderName string) (bool, error) {

	resp, err := s.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.Bucket),
		Prefix:  aws.String(folderName),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		log.Fatalf("Failed to list objects: %v", err)
	}

	if len(resp.Contents) == 0 {
		return false, nil
	}

	return true, nil
}

func (s *Storage) CreateFolder(folderName string) error {

	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(folderName),
		Body:   strings.NewReader(""),
	})
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}

	return nil
}

func (s *Storage) Upload(folderName, filePath string) error {

	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	contentHash := calculateSHA256(file)

	_, err = s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:         aws.String(s.Bucket),
		Key:            aws.String(folderName),
		Body:           bytes.NewReader(file),
		ContentType:    aws.String(detectMimeType(filePath)),
		ChecksumSHA256: aws.String(contentHash),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

func calculateSHA256(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func detectMimeType(path string) string {

	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
