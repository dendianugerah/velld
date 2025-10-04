package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	Endpoint   string
	Region     string
	Bucket     string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	PathPrefix string
}

type S3Storage struct {
	client *minio.Client
	bucket string
	prefix string
}

func NewS3Storage(config S3Config) (*S3Storage, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &S3Storage{
		client: client,
		bucket: config.Bucket,
		prefix: config.PathPrefix,
	}, nil
}

func (s *S3Storage) UploadFile(ctx context.Context, localPath string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	fileName := filepath.Base(localPath)
	objectKey := s.getObjectKey(fileName)

	_, err = s.client.PutObject(ctx, s.bucket, objectKey, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return objectKey, nil
}

func (s *S3Storage) DownloadFile(ctx context.Context, objectKey, localPath string) error {
	object, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer object.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, object)
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}

	return nil
}

func (s *S3Storage) DeleteFile(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}
	return nil
}

func (s *S3Storage) ListFiles(ctx context.Context) ([]string, error) {
	var files []string

	opts := minio.ListObjectsOptions{
		Prefix:    s.prefix,
		Recursive: true,
	}

	for object := range s.client.ListObjects(ctx, s.bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

func (s *S3Storage) GetFileSize(ctx context.Context, objectKey string) (int64, error) {
	info, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to stat object: %w", err)
	}
	return info.Size, nil
}

func (s *S3Storage) TestConnection(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket does not exist: %s", s.bucket)
	}
	return nil
}

func (s *S3Storage) getObjectKey(fileName string) string {
	if s.prefix == "" {
		return fileName
	}
	
	// Ensure prefix doesn't end with / and fileName doesn't start with /
	prefix := strings.TrimSuffix(s.prefix, "/")
	fileName = strings.TrimPrefix(fileName, "/")
	return fmt.Sprintf("%s/%s", prefix, fileName)
}
