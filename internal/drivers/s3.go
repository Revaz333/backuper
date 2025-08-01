package drivers

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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
		o.UsePathStyle = false
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

func (d *S3) Cleanup(folderName string, maxImages int) error {

	var (
		objects []s3types.Object
		ctx     = context.TODO()
	)

	paginator := s3.NewListObjectsV2Paginator(d.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.Bucket),
		Prefix: aws.String(folderName),
	})

	for paginator.HasMorePages() {

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get page: %v", err)
		}

		objects = append(objects, page.Contents...)
	}

	if len(objects) <= maxImages {
		return nil
	}

	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(*objects[j].LastModified)
	})

	objectsToDelete := objects[5:]

	var identifiers []s3types.ObjectIdentifier

	for _, obj := range objectsToDelete {
		if strings.HasSuffix(*obj.Key, "/") {
			continue
		}

		identifiers = append(identifiers, s3types.ObjectIdentifier{Key: obj.Key})
	}

	_, err := d.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(d.Bucket),
		Delete: &types.Delete{
			Objects: identifiers,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to cleanup: %v", err)
	}

	return nil
}
