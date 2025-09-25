package provider

import (
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/meq"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/entity"
	"github.com/itsLeonB/stortr/internal/repository"
)

type Repositories struct {
	Storage                      repository.StorageRepository
	OrphanedBillCleanupTaskQueue meq.TaskQueue[entity.OrphanedBillCleanupTask]
}

func ProvideRepositories(configs config.Config, dbs meq.DB, logger ezutil.Logger) *Repositories {
	if dbs == nil {
		panic("dbs cannot be nil")
	}

	return &Repositories{
		Storage:                      repository.NewGCSStorageRepository([]byte(configs.ServiceAccount)),
		OrphanedBillCleanupTaskQueue: meq.NewTaskQueue[entity.OrphanedBillCleanupTask](logger, dbs),
	}
}

func (r *Repositories) Shutdown() error {
	return r.Storage.Close()
}
