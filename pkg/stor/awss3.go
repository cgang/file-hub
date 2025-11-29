package stor

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cgang/file-hub/pkg/config"
)

type s3Storage struct {
	client *s3.Client
}

func NewS3Storage(cfg *config.S3Config) *s3Storage {
	opts := &s3.Options{
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		BaseEndpoint: aws.String(cfg.Endpoint),
	}

	client := s3.New(*opts)

	return &s3Storage{
		client: client,
	}
}
