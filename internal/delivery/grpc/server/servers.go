package server

import (
	"github.com/itsLeonB/stortr-protos/gen/go/expensebill/v1"
	"github.com/itsLeonB/stortr/internal/provider"
	"github.com/rotisserie/eris"
	"google.golang.org/grpc"
)

type Servers struct {
	ExpenseBill expensebill.ExpenseBillServiceServer
}

func ProvideServers(services *provider.Services) *Servers {
	return &Servers{
		ExpenseBill: newExpenseBillServer(services.ExpenseBill),
	}
}

func (s *Servers) Register(grpcServer *grpc.Server) error {
	if s.ExpenseBill == nil {
		return eris.New("expense bill server is nil")
	}

	expensebill.RegisterExpenseBillServiceServer(grpcServer, s.ExpenseBill)

	return nil
}
