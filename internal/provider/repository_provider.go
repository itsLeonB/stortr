package provider

import (
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/repository"
)

type Repositories struct {
	Storage   repository.StorageRepository
	TaskQueue repository.TaskQueue
}

func ProvideRepositories(configs config.Config, dbs *DBs, logger ezutil.Logger) *Repositories {
	if dbs == nil {
		panic("dbs cannot be nil")
	}

	return &Repositories{
		Storage:   repository.NewStorageRepository(configs),
		TaskQueue: repository.NewTaskQueue(logger, dbs.AsynqClient, dbs.AsynqInspector),
	}
}

func (r *Repositories) Shutdown() error {
	return r.Storage.Close()
}
