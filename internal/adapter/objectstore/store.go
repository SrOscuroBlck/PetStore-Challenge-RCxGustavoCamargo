package objectstore

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"roboticCrewChallenge/internal/platform/id"
)

const (
	presignTTL      = 15 * time.Minute
	objectKeyPrefix = "pets/"
)

type PictureStore struct {
	client *minio.Client
	bucket string
}

func New(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*PictureStore, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	return &PictureStore{client: client, bucket: bucket}, nil
}

func (s *PictureStore) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		switch minio.ToErrorResponse(err).Code {
		case minio.BucketAlreadyOwnedByYou, minio.BucketAlreadyExists:
			return nil
		default:
			return fmt.Errorf("create bucket %q: %w", s.bucket, err)
		}
	}
	return nil
}

func (s *PictureStore) Upload(ctx context.Context, body io.Reader, size int64, contentType string) (string, error) {
	objectID, err := id.New()
	if err != nil {
		return "", fmt.Errorf("generate object key: %w", err)
	}
	key := objectKeyPrefix + objectID.String()

	if _, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return "", fmt.Errorf("put object %q: %w", key, err)
	}
	return key, nil
}

func (s *PictureStore) PresignedURL(ctx context.Context, objectKey string) (string, error) {
	signed, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, presignTTL, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign get %q: %w", objectKey, err)
	}
	return signed.String(), nil
}
