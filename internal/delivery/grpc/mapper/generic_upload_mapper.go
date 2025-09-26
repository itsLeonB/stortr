package mapper

import (
	"github.com/itsLeonB/stortr-protos/gen/go/genericupload/v1"
	"github.com/itsLeonB/stortr/internal/dto"
	"github.com/rotisserie/eris"
)

func FromFileIdentifierProto(pb *genericupload.FileIdentifier) (dto.FileIdentifierDTO, error) {
	if pb == nil {
		return dto.FileIdentifierDTO{}, eris.New("file identifier proto is nil")
	}
	return dto.FileIdentifierDTO{
		BucketName: pb.GetBucketName(),
		ObjectKey:  pb.GetObjectKey(),
	}, nil
}
