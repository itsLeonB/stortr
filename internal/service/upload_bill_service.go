package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/itsLeonB/stortr/internal/appconstant"
	"github.com/itsLeonB/stortr/internal/dto"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/itsLeonB/stortr/internal/repository"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
)

type UploadBillService interface {
	Upload(ctx context.Context, req *dto.UploadBillRequest) (string, error)
	GetURL(ctx context.Context, objectKey string) (string, error)
	Delete(ctx context.Context, objectKey string) error
}

type uploadBillServiceImpl struct {
	validate    *validator.Validate
	storageRepo repository.StorageRepository
	bucketName  string
}

func NewUploadBillService(
	validate *validator.Validate,
	storageRepo repository.StorageRepository,
	bucketName string,
) UploadBillService {
	return &uploadBillServiceImpl{
		validate,
		storageRepo,
		bucketName,
	}
}

func (ubs *uploadBillServiceImpl) Upload(ctx context.Context, req *dto.UploadBillRequest) (string, error) {
	if err := ubs.validateUploadRequest(req); err != nil {
		return "", err
	}

	objectKey := ubs.generateObjectKey(req.Filename)

	storageReq := entity.StorageUploadRequest{
		Data:        req.ImageData,
		ContentType: req.ContentType,
		Filename:    req.Filename,
		BucketName:  ubs.bucketName,
		ObjectKey:   objectKey,
	}

	if err := ubs.storageRepo.Upload(ctx, &storageReq); err != nil {
		return "", err
	}

	return objectKey, nil
}

func (ubs *uploadBillServiceImpl) GetURL(ctx context.Context, objectKey string) (string, error) {
	return ubs.storageRepo.GetSignedURL(ctx, ubs.bucketName, objectKey, appconstant.SignedURLDuration)
}

func (ubs *uploadBillServiceImpl) Delete(ctx context.Context, objectKey string) error {
	return ubs.storageRepo.Delete(ctx, ubs.bucketName, objectKey)
}

func (ubs *uploadBillServiceImpl) validateUploadRequest(req *dto.UploadBillRequest) error {
	if req == nil {
		return ungerr.BadRequestError("request is nil")
	}
	if len(req.ImageData) == 0 {
		return ungerr.BadRequestError("image data is required")
	}
	if len(req.ImageData) > appconstant.MaxFileSize {
		return ungerr.BadRequestError(appconstant.ErrFileTooLarge)
	}
	if err := ubs.validate.Struct(req); err != nil {
		return eris.Wrap(err, appconstant.ErrStructValidation)
	}
	return nil
}

func (ubs *uploadBillServiceImpl) generateObjectKey(filename string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().Format("2006/01/02")
	return fmt.Sprintf("bills/%s/%s%s", timestamp, uuid.NewString(), ext)
}
