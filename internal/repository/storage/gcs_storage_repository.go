package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/rotisserie/eris"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type gcsStorageRepository struct {
	client *storage.Client
}

func NewGCSStorageRepository(credentialsFile string) *gcsStorageRepository {
	if credentialsFile == "" {
		panic("credentials file cannot be empty")
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(credentialsFile)))
	if err != nil {
		panic(fmt.Sprintf("failed to create GCS client: %v", err))
	}

	return &gcsStorageRepository{client}
}

func (r *gcsStorageRepository) Upload(ctx context.Context, req *entity.StorageUploadRequest) error {
	bucket := r.client.Bucket(req.BucketName)
	obj := bucket.Object(req.ObjectKey)

	// Create a writer to upload the file
	writer := obj.NewWriter(ctx)
	writer.ContentType = req.ContentType
	writer.Metadata = map[string]string{
		"original_filename": req.Filename,
		"uploaded_at":       time.Now().Format(time.RFC3339),
	}

	// Set cache control for images
	writer.CacheControl = "public, max-age=3600" // 1 hour cache

	// Write the file data
	if _, err := io.Copy(writer, bytes.NewReader(req.Data)); err != nil {
		_ = writer.Close() // best-effort close on copy failure
		return eris.Wrap(err, "failed to upload file to GCS")
	}
	if err := writer.Close(); err != nil {
		return eris.Wrap(err, "failed to finalize upload to GCS")
	}

	// // Make the object publicly readable (optional)
	// if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
	// 	// Log warning but don't fail the operation
	// 	// In production, you might want to handle this differently based on your security requirements
	// 	fmt.Printf("Warning: failed to make object public: %v\n", err)
	// }

	// // Generate public URL
	// publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", req.BucketName, req.ObjectKey)

	return nil
}

func (r *gcsStorageRepository) Delete(ctx context.Context, bucketName, objectKey string) error {
	if bucketName == "" {
		return eris.New("missing bucket name")
	}

	bucket := r.client.Bucket(bucketName)
	obj := bucket.Object(objectKey)

	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			// Object doesn't exist, consider it already deleted
			return nil
		}
		return eris.Wrap(err, "failed to delete file from GCS")
	}

	return nil
}

func (r *gcsStorageRepository) GetSignedURL(ctx context.Context, bucketName, objectKey string, expiration time.Duration) (string, error) {
	if bucketName == "" {
		return "", eris.New("missing bucket name")
	}

	bucket := r.client.Bucket(bucketName)

	url, err := bucket.SignedURL(objectKey, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  http.MethodGet,
		Expires: time.Now().Add(expiration),
	})
	if err != nil {
		return "", eris.Wrap(err, "failed to generate signed URL")
	}

	return url, nil
}

func (r *gcsStorageRepository) GetAllObjectKeys(ctx context.Context, bucketName string) ([]string, error) {
	if bucketName == "" {
		return nil, eris.New("missing bucket name")
	}

	bucket := r.client.Bucket(bucketName)
	it := bucket.Objects(ctx, nil)
	objectKeys := make([]string, 0)

	for {
		attr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, eris.Wrap(err, "error listing objects in bucket")
		}
		objectKeys = append(objectKeys, attr.Name)
	}

	return objectKeys, nil
}

func (r *gcsStorageRepository) Close() error {
	return r.client.Close()
}
