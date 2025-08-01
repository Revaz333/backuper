package drivers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type (
	S3 struct {
		Client *s3.Client
		Bucket string
	}

	S3Config struct {
		Region    string
		AccessKey string
		SecretKey string
		Endpoint  string
		Bucket    string
	}
)

func NewS3() DriverInterface {

	return &S3{}
}

func (d *S3) Init(cfgData any) error {

	cfg, ok := cfgData.(S3Config)
	if !ok {
		return fmt.Errorf("config type must be - `S3Config`")
	}

	baseCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("failed to load s3 config: %v", err)
	}

	s3Client := s3.NewFromConfig(baseCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(cfg.Endpoint)
	})

	d.Client = s3Client
	d.Bucket = cfg.Bucket

	return nil
}

func (d *S3) Ping() error {

	_, err := d.Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: &d.Bucket,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to s3 bucket, go error: %v", err)
	}

	return nil
}

func (d *S3) CheckFolderExist(folderName string) (bool, error) {

	resp, err := d.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:  aws.String(d.Bucket),
		Prefix:  aws.String(folderName),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list objects: %v", err)
	}

	if len(resp.Contents) == 0 {
		return false, nil
	}

	return true, nil
}

func (d *S3) CreateFolder(folderName string) error {

	_, err := d.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(folderName),
		Body:   strings.NewReader(""),
	})
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}

	return nil
}

func (d *S3) Upload(folderName, filePath string) error {

	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	contentHash := calculateSHA256(file)

	_, err = d.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:         aws.String(d.Bucket),
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
