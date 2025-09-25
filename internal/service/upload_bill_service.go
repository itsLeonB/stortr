package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
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
	CleanupOrphaned(ctx context.Context) error
}

type uploadBillServiceImpl struct {
	validate    *validator.Validate
	storageRepo repository.StorageRepository
	bucketName  string
	taskQueue   repository.TaskQueue
	logger      ezutil.Logger
}

func NewUploadBillService(
	validate *validator.Validate,
	storageRepo repository.StorageRepository,
	bucketName string,
	taskQueue repository.TaskQueue,
	logger ezutil.Logger,
) UploadBillService {
	return &uploadBillServiceImpl{
		validate,
		storageRepo,
		bucketName,
		taskQueue,
		logger,
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

func (ubs *uploadBillServiceImpl) CleanupOrphaned(ctx context.Context) error {
	taskMsg := entity.OrphanedBillCleanupTask{}
	tasks, err := ubs.taskQueue.GetAllPending(ctx, taskMsg.Type())
	if err != nil {
		return err
	}

	if len(tasks) < 1 {
		return eris.New("empty task queue")
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Timestamp.After(tasks[j].Timestamp)
	})

	latestTask := tasks[0]

	if err = json.Unmarshal(latestTask.Payload, &taskMsg); err != nil {
		return eris.Wrap(err, "error unmarshaling task message")
	}

	ubs.logger.Infof("obtained object keys from task payload:\n%s", strings.Join(taskMsg.BillObjectKeys, "\n"))

	validObjectKeys := make(map[string]struct{}, len(taskMsg.BillObjectKeys))
	for _, key := range taskMsg.BillObjectKeys {
		validObjectKeys[key] = struct{}{}
	}

	allObjectKeys, err := ubs.storageRepo.GetAllObjectKeys(ctx, ubs.bucketName)
	if err != nil {
		return err
	}

	ubs.logger.Infof(
		"obtained object keys from bucket: %s,\n%s",
		ubs.bucketName,
		strings.Join(allObjectKeys, "\n"),
	)

	hasDeleted := false
	for _, key := range allObjectKeys {
		if _, exists := validObjectKeys[key]; !exists {
			hasDeleted = true
			ubs.logger.Infof("deleting orphaned bill: %s", key)
			if e := ubs.storageRepo.Delete(ctx, ubs.bucketName, key); e != nil {
				ubs.logger.Errorf("error deleting object: %s from bucket: %s: %v", key, ubs.bucketName, e)
			}
		}
	}

	if !hasDeleted {
		ubs.logger.Info("no orphaned bill to delete")
	}

	return ubs.taskQueue.DeleteAll(ctx, taskMsg.Type())
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
