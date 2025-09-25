package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/rotisserie/eris"
)

type Task struct {
	ID        uuid.UUID   `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Type      string      `json:"type"`
	Payload   []byte      `json:"payload"`
	message   TaskMessage `json:"-"`
}

func (t Task) IsZero() bool {
	return t.ID == uuid.Nil
}

func NewTask(tm TaskMessage) (Task, error) {
	if tm == nil {
		return Task{}, eris.New("task message is nil")
	}

	payload, err := json.Marshal(tm)
	if err != nil {
		return Task{}, eris.Wrap(err, "error marshaling payload")
	}

	return Task{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Source:    config.AppName,
		Type:      tm.Type(),
		Payload:   payload,
		message:   tm,
	}, nil
}

type TaskMessage interface {
	Type() string
}

type OrphanedBillCleanupTask struct {
	BillObjectKeys []string `json:"billObjectKeys"`
}

func (obc OrphanedBillCleanupTask) Type() string {
	return "orphaned-bill-cleanup"
}
