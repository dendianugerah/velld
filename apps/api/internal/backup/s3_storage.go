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
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return objectKey, nil
}

// UploadStream streams an object of unknown length straight to S3. The caller
// writes to the returned pipe and must close it; minio switches to a multipart
// upload on its own because the size is -1. This is what lets a dump be
// uploaded as it is produced, without ever landing on local disk.
func (s *S3Storage) UploadStream(ctx context.Context, objectKey string, produce func(io.Writer) error) error {
	reader, writer := io.Pipe()

	produceErr := make(chan error, 1)
	go func() {
		err := produce(writer)
		writer.CloseWithError(err)
		produceErr <- err
	}()

	_, uploadErr := s.client.PutObject(ctx, s.bucket, objectKey, reader, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
		PartSize:    16 * 1024 * 1024,
		NumThreads:  4,
	})

	// Unblock the producer if the upload failed first.
	reader.CloseWithError(uploadErr)

	if err := <-produceErr; err != nil {
		return fmt.Errorf("produce stream: %w", err)
	}
	if uploadErr != nil {
		return fmt.Errorf("upload stream to S3: %w", uploadErr)
	}

	return nil
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

// ObjectKeyFor returns the full object key for a file placed under subfolder.
func (s *S3Storage) ObjectKeyFor(subfolder, fileName string) string {
	return s.getObjectKeyWithPath(fileName, subfolder)
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
