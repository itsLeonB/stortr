package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/itsLeonB/stortr/internal/appconstant"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/rotisserie/eris"
)

type localStorageRepository struct {
	basePath string
}

func NewLocalStorageRepository(basePath string) *localStorageRepository {
	return &localStorageRepository{basePath: basePath}
}

func (r *localStorageRepository) Upload(ctx context.Context, req *entity.StorageUploadRequest) error {
	bucketPath := filepath.Join(r.basePath, req.BucketName)
	if err := os.MkdirAll(bucketPath, 0755); err != nil {
		return eris.Wrap(err, appconstant.ErrProcessFile)
	}
	objectPath := filepath.Join(bucketPath, req.ObjectKey)
	if err := os.WriteFile(objectPath, req.Data, 0644); err != nil {
		return eris.Wrap(err, appconstant.ErrProcessFile)
	}
	return nil
}

func (r *localStorageRepository) Delete(ctx context.Context, bucketName, objectKey string) error {
	objectPath := filepath.Join(r.basePath, bucketName, objectKey)
	if err := os.Remove(objectPath); err != nil {
		return eris.Wrap(err, appconstant.ErrProcessFile)
	}
	return nil
}

func (r *localStorageRepository) GetSignedURL(ctx context.Context, bucketName, objectKey string, expiration time.Duration) (string, error) {
	objectPath := filepath.Join(r.basePath, bucketName, objectKey)
	if _, err := os.Stat(objectPath); err != nil {
		return "", eris.New("object not found")
	}
	// For local storage, just return the file path as a "signed URL"
	return fmt.Sprintf("file://%s", objectPath), nil
}

func (r *localStorageRepository) Close() error {
	// No resources to close for local storage
	return nil
}
