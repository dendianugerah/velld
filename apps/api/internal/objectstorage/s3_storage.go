package objectstorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	IsR2       bool
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
	isR2   bool
}

func NewS3Storage(ctx context.Context, config S3Config, createBucket bool) (*S3Storage, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	}
	if config.IsR2 {
		opts.BucketLookup = minio.BucketLookupPath
	}
	client, err := minio.New(config.Endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	if createBucket && !config.IsR2 {
		exists, err := client.BucketExists(ctx, config.Bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to check bucket existence: %w", err)
		}
		if !exists {
			if err := client.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{Region: config.Region}); err != nil {
				return nil, fmt.Errorf("failed to create bucket: %w", err)
			}
		}
	}

	return &S3Storage{
		client: client,
		bucket: config.Bucket,
		prefix: config.PathPrefix,
		isR2:   config.IsR2,
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
		ContentType:          "application/octet-stream",
		DisableContentSha256: s.isR2,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return objectKey, nil
}

// UploadFileWithPath uploads a file to S3 with a custom subfolder path
func (s *S3Storage) UploadFileWithPath(ctx context.Context, localPath string, subfolder string) (string, error) {
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
	objectKey := s.getObjectKeyWithPath(fileName, subfolder)

	_, err = s.client.PutObject(ctx, s.bucket, objectKey, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType:          "application/octet-stream",
		DisableContentSha256: s.isR2,
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

func (s *S3Storage) TestConnection(ctx context.Context) (result error) {
	key := s.getObjectKey(".velld-connection-test-" + uuid.NewString())
	probe := []byte("velld object storage connection test")
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(probe), int64(len(probe)), minio.PutObjectOptions{
		ContentType:          "application/octet-stream",
		DisableContentSha256: s.isR2,
	})
	if err != nil {
		return fmt.Errorf("could not write test object: %w", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := s.DeleteFile(cleanupCtx, key); err != nil {
			result = errors.Join(result, fmt.Errorf("could not remove test object: %w", err))
		}
	}()

	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("could not open test object: %w", err)
	}
	defer object.Close()
	actual, err := io.ReadAll(io.LimitReader(object, int64(len(probe))+1))
	if err != nil {
		return fmt.Errorf("could not read test object: %w", err)
	}
	if !bytes.Equal(actual, probe) {
		return fmt.Errorf("test object did not match the data written")
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

func (s *S3Storage) getObjectKeyWithPath(fileName string, subfolder string) string {
	// Build path: prefix/subfolder/fileName
	var parts []string

	if s.prefix != "" {
		parts = append(parts, strings.TrimSuffix(s.prefix, "/"))
	}

	if subfolder != "" {
		parts = append(parts, strings.Trim(subfolder, "/"))
	}

	parts = append(parts, strings.TrimPrefix(fileName, "/"))

	return strings.Join(parts, "/")
}

// MoveFile moves/renames an object in S3 (copy then delete)
func (s *S3Storage) MoveFile(ctx context.Context, oldKey, newKey string) error {
	// Copy to new location
	src := minio.CopySrcOptions{
		Bucket: s.bucket,
		Object: oldKey,
	}
	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: newKey,
	}

	_, err := s.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	// Delete old object
	err = s.client.RemoveObject(ctx, s.bucket, oldKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete old object: %w", err)
	}

	return nil
}
