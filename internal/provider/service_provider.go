package provider

import (
	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/service"
)

type Services struct {
	Image       service.ImageService
	ExpenseBill service.ExpenseBillService
}

func ProvideServices(configs config.Google, repos *Repositories, logger ezutil.Logger) *Services {
	if repos == nil {
		panic("repos cannot be nil")
	}

	validate := validator.New()

	return &Services{
		Image: service.NewImageService(
			validate,
			repos.Storage,
		),
		ExpenseBill: service.NewExpenseBillService(
			repos.Storage,
			repos.OrphanedBillCleanupTaskQueue,
			logger,
		),
	}
}
