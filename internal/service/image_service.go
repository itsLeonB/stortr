package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/stortr/internal/appconstant"
	"github.com/itsLeonB/stortr/internal/dto"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/itsLeonB/stortr/internal/mapper"
	"github.com/itsLeonB/stortr/internal/repository"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
)

type ImageService interface {
	Upload(ctx context.Context, req *dto.ImageUploadRequest) (string, error)
	GetURL(ctx context.Context, fileID dto.FileIdentifierDTO) (string, error)
	Delete(ctx context.Context, fileID dto.FileIdentifierDTO) error
}

type imageServiceImpl struct {
	validate    *validator.Validate
	storageRepo repository.StorageRepository
}

func NewImageService(
	validate *validator.Validate,
	storageRepo repository.StorageRepository,
) ImageService {
	return &imageServiceImpl{
		validate,
		storageRepo,
	}
}

func (ubs *imageServiceImpl) Upload(ctx context.Context, req *dto.ImageUploadRequest) (string, error) {
	if err := ubs.validateUploadRequest(req); err != nil {
		return "", err
	}

	storageReq := entity.StorageUploadRequest{
		Data:           req.ImageData,
		ContentType:    req.ContentType,
		FileIdentifier: mapper.FileIdentifierFromDTO(req.FileIdentifierDTO),
	}

	if err := ubs.storageRepo.Upload(ctx, &storageReq); err != nil {
		return "", err
	}

	return ubs.storageRepo.ToURI(storageReq.FileIdentifier), nil
}

func (ubs *imageServiceImpl) GetURL(ctx context.Context, fileID dto.FileIdentifierDTO) (string, error) {
	return ubs.storageRepo.GetSignedURL(ctx, mapper.FileIdentifierFromDTO(fileID), appconstant.SignedURLDuration)
}

func (ubs *imageServiceImpl) Delete(ctx context.Context, fileID dto.FileIdentifierDTO) error {
	return ubs.storageRepo.Delete(ctx, mapper.FileIdentifierFromDTO(fileID))
}

func (ubs *imageServiceImpl) validateUploadRequest(req *dto.ImageUploadRequest) error {
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
