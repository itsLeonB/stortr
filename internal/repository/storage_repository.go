package repository

import (
	"context"
	"time"

	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/itsLeonB/stortr/internal/repository/storage"
)

type StorageRepository interface {
	Upload(ctx context.Context, req *entity.StorageUploadRequest) error
	Delete(ctx context.Context, bucketName, objectKey string) error
	GetSignedURL(ctx context.Context, bucketName, objectKey string, expiration time.Duration) (string, error)
	Close() error
}

func NewStorageRepository(configs config.Config) StorageRepository {
	switch configs.Env {
	case "local":
		return storage.NewLocalStorageRepository("./data")
	case "dev":
		return storage.NewGCSStorageRepository(configs.ServiceAccount)
	case "prod":
		return storage.NewGCSStorageRepository(configs.ServiceAccount)
	default:
		panic("unsupported environment: " + configs.Env)
	}
}
