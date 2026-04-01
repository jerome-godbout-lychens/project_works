package s3

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// S3FileStorage implements domain.FileStorage using the AWS SDK v2 S3 client.
// Works with both AWS S3 and S3-compatible services (SeaweedFS, MinIO).
type S3FileStorage struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
}

// NewS3FileStorage creates a new S3-backed file storage.
// If endpoint is empty, it connects to native AWS S3.
// If endpoint is set (e.g. "http://localhost:8333"), it connects to that S3-compatible service.
func NewS3FileStorage(ctx context.Context, endpoint string, bucketName string, region string) (domain.FileStorage, error) {
	optFunctions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}

	configuration, err := awsconfig.LoadDefaultConfig(ctx, optFunctions...)
	if err != nil {
		return nil, err
	}

	clientOptions := []func(*s3.Options){}
	if endpoint != "" {
		clientOptions = append(clientOptions, func(options *s3.Options) {
			options.BaseEndpoint = aws.String(endpoint)
			options.UsePathStyle = true // required for SeaweedFS / MinIO
		})
	}

	client := s3.NewFromConfig(configuration, clientOptions...)
	presigner := s3.NewPresignClient(client)

	return &S3FileStorage{
		client:     client,
		presigner:  presigner,
		bucketName: bucketName,
	}, nil
}

func (storage *S3FileStorage) PutFile(ctx context.Context, storageKey string, data io.Reader, contentType string) error {
	_, err := storage.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(storage.bucketName),
		Key:         aws.String(storageKey),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	return err
}

func (storage *S3FileStorage) GetFile(ctx context.Context, storageKey string) (io.ReadCloser, error) {
	output, err := storage.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(storage.bucketName),
		Key:    aws.String(storageKey),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

func (storage *S3FileStorage) DeleteFile(ctx context.Context, storageKey string) error {
	_, err := storage.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(storage.bucketName),
		Key:    aws.String(storageKey),
	})
	return err
}

func (storage *S3FileStorage) GeneratePresignedURL(ctx context.Context, storageKey string, expiration time.Duration) (string, error) {
	presignedRequest, err := storage.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(storage.bucketName),
		Key:    aws.String(storageKey),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", err
	}
	return presignedRequest.URL, nil
}
