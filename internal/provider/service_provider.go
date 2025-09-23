package provider

import (
	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/service"
)

type Services struct {
	UploadBill service.UploadBillService
}

func ProvideServices(configs config.Google, repos *Repositories) *Services {
	validate := validator.New()

	return &Services{
		UploadBill: service.NewUploadBillService(validate, repos.Storage, configs.BillBucketName),
	}
}
