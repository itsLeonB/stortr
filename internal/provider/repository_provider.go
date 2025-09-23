package provider

import (
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/repository"
)

type Repositories struct {
	Storage repository.StorageRepository
}

func ProvideRepositories(configs config.Config) *Repositories {
	return &Repositories{
		Storage: repository.NewStorageRepository(configs),
	}
}

func (r *Repositories) Shutdown() error {
	return r.Storage.Close()
}
