package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/platform/id"
)

const objectKeyPrefix = "pets/"

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

// Get streams the stored object back. The caller owns PictureContent.Body and
// must close it. A missing object maps to domain.ErrPictureNotFound so the
// picture path can answer 404 without leaking storage detail.
func (s *PictureStore) Get(ctx context.Context, objectKey string) (domain.PictureContent, error) {
	// Only pet pictures live under this prefix; refusing anything else keeps the
	// read path from reaching unrelated objects even if the bucket ever holds them.
	if !strings.HasPrefix(objectKey, objectKeyPrefix) {
		return domain.PictureContent{}, domain.ErrPictureNotFound
	}
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return domain.PictureContent{}, fmt.Errorf("get object %q: %w", objectKey, err)
	}
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		if minio.ToErrorResponse(err).Code == minio.NoSuchKey {
			return domain.PictureContent{}, fmt.Errorf("stat object %q: %w", objectKey, errors.Join(domain.ErrPictureNotFound, err))
		}
		return domain.PictureContent{}, fmt.Errorf("stat object %q: %w", objectKey, err)
	}
	return domain.PictureContent{Body: obj, ContentType: info.ContentType, Size: info.Size}, nil
}
