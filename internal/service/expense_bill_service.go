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

type ExpenseBillService interface {
	Upload(ctx context.Context, req *dto.UploadBillRequest) (string, error)
	GetURL(ctx context.Context, objectKey string) (string, error)
	Delete(ctx context.Context, objectKey string) error
}

type expenseBillServiceImpl struct {
	validate    *validator.Validate
	storageRepo repository.StorageRepository
	bucketName  string
}

func NewExpenseBillService(
	validate *validator.Validate,
	storageRepo repository.StorageRepository,
	bucketName string,
) ExpenseBillService {
	return &expenseBillServiceImpl{
		validate,
		storageRepo,
		bucketName,
	}
}

func (ebs *expenseBillServiceImpl) Upload(ctx context.Context, req *dto.UploadBillRequest) (string, error) {
	if err := ebs.validateUploadRequest(req); err != nil {
		return "", err
	}

	objectKey := ebs.generateObjectKey(req.Filename)

	storageReq := entity.StorageUploadRequest{
		Data:        req.ImageData,
		ContentType: req.ContentType,
		Filename:    req.Filename,
		BucketName:  ebs.bucketName,
		ObjectKey:   objectKey,
	}

	if err := ebs.storageRepo.Upload(ctx, &storageReq); err != nil {
		return "", err
	}

	return objectKey, nil
}

func (ebs *expenseBillServiceImpl) GetURL(ctx context.Context, objectKey string) (string, error) {
	return ebs.storageRepo.GetSignedURL(ctx, ebs.bucketName, objectKey, appconstant.SignedURLDuration)
}

func (ebs *expenseBillServiceImpl) Delete(ctx context.Context, objectKey string) error {
	return ebs.storageRepo.Delete(ctx, ebs.bucketName, objectKey)
}

func (ebs *expenseBillServiceImpl) validateUploadRequest(req *dto.UploadBillRequest) error {
	if req == nil {
		return ungerr.BadRequestError("request is nil")
	}
	if len(req.ImageData) == 0 {
		return ungerr.BadRequestError("image data is required")
	}
	if len(req.ImageData) > appconstant.MaxFileSize {
		return ungerr.BadRequestError(appconstant.ErrFileTooLarge)
	}
	if err := ebs.validate.Struct(req); err != nil {
		return eris.Wrap(err, appconstant.ErrStructValidation)
	}
	return nil
}

func (ebs *expenseBillServiceImpl) generateObjectKey(filename string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().Format("2006/01/02")
	return fmt.Sprintf("bills/%s/%s%s", timestamp, uuid.NewString(), ext)
}
