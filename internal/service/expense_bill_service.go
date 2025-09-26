package service

import (
	"context"
	"sort"
	"strings"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/meq"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/itsLeonB/stortr/internal/repository"
	"github.com/rotisserie/eris"
)

type ExpenseBillService interface {
	CleanupOrphaned(ctx context.Context) error
}

type expenseBillServiceImpl struct {
	storageRepo repository.StorageRepository
	taskQueue   meq.TaskQueue[entity.OrphanedBillCleanupTask]
	logger      ezutil.Logger
}

func NewExpenseBillService(
	storageRepo repository.StorageRepository,
	taskQueue meq.TaskQueue[entity.OrphanedBillCleanupTask],
	logger ezutil.Logger,
) ExpenseBillService {
	return &expenseBillServiceImpl{
		storageRepo,
		taskQueue,
		logger,
	}
}

func (ubs *expenseBillServiceImpl) CleanupOrphaned(ctx context.Context) error {
	bucketName, validObjectKeys, err := ubs.getBucketNameAndObjectKeys(ctx)
	if err != nil {
		return err
	}

	allObjectKeys, err := ubs.storageRepo.GetAllObjectKeys(ctx, bucketName)
	if err != nil {
		return err
	}

	ubs.logger.Infof("obtained object keys from bucket: %s,\n%s", bucketName, strings.Join(allObjectKeys, "\n"))

	hasDeleted := false
	for _, key := range allObjectKeys {
		if _, exists := validObjectKeys[key]; !exists {
			hasDeleted = true
			ubs.logger.Infof("deleting orphaned bill: %s", key)
			if e := ubs.storageRepo.Delete(ctx, entity.FileIdentifier{
				BucketName: bucketName,
				ObjectKey:  key,
			}); e != nil {
				ubs.logger.Errorf("error deleting object: %s from bucket: %s: %v", key, bucketName, e)
			}
		}
	}

	if !hasDeleted {
		ubs.logger.Info("no orphaned bill to delete")
	}

	return ubs.taskQueue.DeleteAll(ctx)
}

func (ubs *expenseBillServiceImpl) getBucketNameAndObjectKeys(ctx context.Context) (string, map[string]struct{}, error) {
	tasks, err := ubs.taskQueue.GetAllPending(ctx)
	if err != nil {
		return "", nil, err
	}

	if len(tasks) < 1 {
		return "", nil, eris.New("empty task queue")
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Timestamp.After(tasks[j].Timestamp)
	})

	latestTask := tasks[0]
	objectKeys := latestTask.Message.BillObjectKeys
	bucketName := latestTask.Message.BucketName

	ubs.logger.Infof("obtained object keys from task payload:\n%s", strings.Join(objectKeys, "\n"))

	validObjectKeys := make(map[string]struct{}, len(objectKeys))
	for _, key := range objectKeys {
		validObjectKeys[key] = struct{}{}
	}

	return bucketName, validObjectKeys, nil
}
