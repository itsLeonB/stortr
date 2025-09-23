package server

import (
	"github.com/itsLeonB/stortr-protos/gen/go/uploadbill/v1"
	"github.com/itsLeonB/stortr/internal/provider"
	"github.com/rotisserie/eris"
	"google.golang.org/grpc"
)

type Servers struct {
	UploadBill uploadbill.UploadBillServiceServer
}

func ProvideServers(services *provider.Services) *Servers {
	return &Servers{
		UploadBill: newUploadBillServer(services.UploadBill),
	}
}

func (s *Servers) Register(grpcServer *grpc.Server) error {
	if s.UploadBill == nil {
		return eris.New("expense bill server is nil")
	}

	uploadbill.RegisterUploadBillServiceServer(grpcServer, s.UploadBill)

	return nil
}
