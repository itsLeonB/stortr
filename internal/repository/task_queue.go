package repository

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/rotisserie/eris"
)

type TaskQueue interface {
	Enqueue(ctx context.Context, task entity.Task) error
	GetAllPending(ctx context.Context, taskType string) ([]entity.Task, error)
	DeleteAll(ctx context.Context, taskType string) error
}

type asynqTaskQueue struct {
	logger    ezutil.Logger
	client    *asynq.Client
	inspector *asynq.Inspector
}

func NewTaskQueue(
	logger ezutil.Logger,
	client *asynq.Client,
	inspector *asynq.Inspector,
) TaskQueue {
	return &asynqTaskQueue{
		logger,
		client,
		inspector,
	}
}

func (tq *asynqTaskQueue) Enqueue(ctx context.Context, task entity.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return eris.Wrap(err, "error marshaling task")
	}

	asynqTask := asynq.NewTask(task.Type, payload)

	info, err := tq.client.EnqueueContext(ctx, asynqTask, asynq.Queue(task.Type))
	if err != nil {
		return eris.Wrap(err, "error enqueuing task")
	}

	tq.logger.Infof("enqueued task: BatchID=%s, Queue=%s", info.ID, info.Queue)

	return nil
}

func (tq *asynqTaskQueue) GetAllPending(ctx context.Context, taskType string) ([]entity.Task, error) {
	pendingTasks, err := tq.inspector.ListPendingTasks(taskType, asynq.PageSize(1000))
	if err != nil {
		return nil, eris.Wrap(err, "error listing pending tasks")
	}

	return ezutil.MapSliceWithError(pendingTasks, mapToTask)
}

func (tq *asynqTaskQueue) DeleteAll(ctx context.Context, taskType string) error {
	if err := tq.inspector.DeleteQueue(taskType, true); err != nil {
		return eris.Wrap(err, "error deleting queue")
	}
	return nil
}

func mapToTask(taskInfo *asynq.TaskInfo) (entity.Task, error) {
	if taskInfo == nil {
		return entity.Task{}, eris.New("task info is nil")
	}

	var payload entity.Task
	if err := json.Unmarshal(taskInfo.Payload, &payload); err != nil {
		return entity.Task{}, eris.Wrap(err, "error unmarshal to task")
	}

	return payload, nil
}
