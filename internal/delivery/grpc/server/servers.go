package server

import (
	"github.com/itsLeonB/stortr-protos/gen/go/imageupload/v1"
	"github.com/itsLeonB/stortr/internal/provider"
	"github.com/rotisserie/eris"
	"google.golang.org/grpc"
)

type Servers struct {
	Image imageupload.ImageUploadServiceServer
}

func ProvideServers(services *provider.Services) *Servers {
	return &Servers{
		Image: newImageUploadServer(services.Image),
	}
}

func (s *Servers) Register(grpcServer *grpc.Server) error {
	if s.Image == nil {
		return eris.New("expense bill server is nil")
	}

	imageupload.RegisterImageUploadServiceServer(grpcServer, s.Image)

	return nil
}
