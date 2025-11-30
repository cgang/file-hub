package stor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cgang/file-hub/pkg/config"
)

var (
	s3Client *s3.Client // Shared S3 client instance
)

func newS3Client(cfg *config.S3Config) *s3.Client {
	opts := &s3.Options{
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		BaseEndpoint: aws.String(cfg.Endpoint),
	}

	return s3.New(*opts)
}

type s3Storage struct {
	bucket string
}

func hashPrefix(s string) string {
	b := sha256.Sum224([]byte(s))
	return hex.EncodeToString(b[:2])
}

// getS3Key converts a file path to an S3 object key
func (s *s3Storage) getS3Key(repo, name string) string {
	name = path.Clean(name)
	// Use a hash prefix to avoid too many objects in a single S3 prefix
	prefix := hashPrefix(name)
	return path.Join(prefix, repo, name)
}

func (s *s3Storage) PutFile(ctx context.Context, repo, name string, data io.Reader) error {
	key := s.getS3Key(repo, name)

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   data,
	}

	_, err := s3Client.PutObject(ctx, input)
	return err
}

// DeleteFile deletes a file or directory from S3
func (s *s3Storage) DeleteFile(ctx context.Context, repo, name string) error {
	key := s.getS3Key(repo, name)

	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}

// OpenFile opens a file for reading
func (s *s3Storage) OpenFile(ctx context.Context, repo, name string) (io.ReadCloser, error) {
	key := s.getS3Key(repo, name)

	output, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

func (s *s3Storage) CopyFile(ctx context.Context, repo, srcName, destName string) error {
	srcKey := s.getS3Key(repo, srcName)
	destKey := s.getS3Key(repo, destName)

	_, err := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(path.Join(s.bucket, srcKey)),
		Key:        aws.String(destKey),
	})
	return err
}

func (s *s3Storage) Scan(ctx context.Context, repo string, visit func(*FileMeta) error) error {
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(s.getS3Key(repo, "")),
		Delimiter: aws.String("/"),
	}

	for {
		output, err := s3Client.ListObjectsV2(ctx, input)
		if err != nil {
			return err
		}

		for _, dir := range output.CommonPrefixes {
			// TODO get correct last modified time for directory
			meta := newDirMeta(aws.ToString(dir.Prefix), time.Now())
			if err := visit(meta); err != nil {
				return err
			}
		}

		for _, obj := range output.Contents {
			meta := newFileMeta(aws.ToString(obj.Key), aws.ToTime(obj.LastModified))
			meta.Size = aws.ToInt64(obj.Size)
			// TODO support content type
			meta.ContentType = getContentType(meta.Name) // temporary workaround

			if err := visit(meta); err != nil {
				return err
			}
		}

		if aws.ToBool(output.IsTruncated) {
			input.ContinuationToken = output.ContinuationToken
		} else {
			return nil
		}
	}
}
