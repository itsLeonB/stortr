package mapper

import (
	"github.com/itsLeonB/stortr/internal/dto"
	"github.com/itsLeonB/stortr/internal/entity"
)

func FileIdentifierFromDTO(fi dto.FileIdentifierDTO) entity.FileIdentifier {
	return entity.FileIdentifier{
		BucketName: fi.BucketName,
		ObjectKey:  fi.ObjectKey,
	}
}
