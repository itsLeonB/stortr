package provider

import (
	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/service"
)

type Services struct {
	ExpenseBill service.ExpenseBillService
}

func ProvideServices(configs config.Google, repos *Repositories) *Services {
	validate := validator.New()

	return &Services{
		ExpenseBill: service.NewExpenseBillService(validate, repos.Storage, configs.BillBucketName),
	}
}
