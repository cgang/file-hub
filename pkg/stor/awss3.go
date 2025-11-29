package stor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/model"
)

type s3Storage struct {
	client   *s3.Client
	bucket   string
	userRoot string
}

func newS3Storage(user *model.User, cfg *config.S3Config, bucket, prefix string) *s3Storage {
	opts := &s3.Options{
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		BaseEndpoint: aws.String(cfg.Endpoint),
	}

	client := s3.New(*opts)

	return &s3Storage{
		client:   client,
		bucket:   bucket,
		userRoot: prefix,
	}
}

// getS3Key converts a file path to an S3 object key
func (s *s3Storage) getS3Key(path string) string {
	cleanPath := strings.TrimPrefix(filepath.Clean(path), "/")
	if cleanPath == "." || cleanPath == "" {
		return s.userRoot
	}
	// Ensure we don't duplicate the user root
	if strings.HasPrefix(cleanPath, s.userRoot) {
		return cleanPath
	}
	return filepath.Join(s.userRoot, cleanPath)
}

// getParentKey gets the parent directory key for a given path
func (s *s3Storage) getParentKey(path string) string {
	cleanPath := strings.TrimPrefix(filepath.Clean(path), "/")
	// Ensure we don not duplicate the user root
	if after, ok := strings.CutPrefix(cleanPath, s.userRoot); ok {
		cleanPath = after
	}
	parent := filepath.Dir(cleanPath)
	if parent == "." || parent == "/" {
		return s.userRoot
	}
	return filepath.Join(s.userRoot, parent) + "/"
}

// DeleteFile deletes a file or directory from S3
func (s *s3Storage) DeleteFile(ctx context.Context, path string) error {
	key := s.getS3Key(path)

	// First, try to delete as a file
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	// Also try to delete as a directory (with trailing slash)
	dirKey := key
	if !strings.HasSuffix(key, "/") {
		dirKey = key + "/"
	}
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(dirKey),
	})
	// Ignore error for directory deletion as it might not exist

	return nil
}

// CreateDir creates a directory in S3 by creating an empty object with a trailing slash
func (s *s3Storage) CreateDir(ctx context.Context, path string) error {
	key := s.getS3Key(path)
	if !strings.HasSuffix(key, "/") {
		key = key + "/"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte{}),
	})
	return err
}

// CopyFile copies a file or directory in S3
func (s *s3Storage) CopyFile(ctx context.Context, src, dst string) error {
	srcKey := s.getS3Key(src)
	dstKey := s.getS3Key(dst)

	// Copy the object
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(fmt.Sprintf("/%s/%s", s.bucket, srcKey)),
		Key:        aws.String(dstKey),
	})
	return err
}

// MoveFile moves a file or directory in S3 by copying and then deleting the source
func (s *s3Storage) MoveFile(ctx context.Context, src, dst string) error {
	// Copy the file first
	if err := s.CopyFile(ctx, src, dst); err != nil {
		return err
	}

	// Then delete the source
	return s.DeleteFile(ctx, src)
}

// OpenFile opens a file for reading
func (s *s3Storage) OpenFile(ctx context.Context, path string) (io.ReadCloser, error) {
	key := s.getS3Key(path)

	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

// WriteToFile writes content to a file in S3
func (s *s3Storage) WriteToFile(ctx context.Context, path string, content io.Reader) error {
	key := s.getS3Key(path)

	// Ensure parent directory exists
	parentKey := s.getParentKey(path)
	if parentKey != s.userRoot {
		_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(parentKey),
			Body:   bytes.NewReader([]byte{}),
		})
		if err != nil {
			return err
		}
	}

	// Upload the file
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   content,
	})
	return err
}
