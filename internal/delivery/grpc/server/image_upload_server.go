package server

import (
	"context"
	"io"

	"github.com/itsLeonB/stortr-protos/gen/go/genericupload/v1"
	"github.com/itsLeonB/stortr-protos/gen/go/imageupload/v1"
	"github.com/itsLeonB/stortr/internal/appconstant"
	"github.com/itsLeonB/stortr/internal/delivery/grpc/mapper"
	"github.com/itsLeonB/stortr/internal/dto"
	"github.com/itsLeonB/stortr/internal/service"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/types/known/emptypb"
)

type imageUploadServer struct {
	imageupload.UnimplementedImageUploadServiceServer
	imageSvc service.ImageService
}

func newImageUploadServer(imageSvc service.ImageService) imageupload.ImageUploadServiceServer {
	return &imageUploadServer{
		imageSvc: imageSvc,
	}
}

func (ebs *imageUploadServer) UploadStream(stream imageupload.ImageUploadService_UploadStreamServer) error {
	metadata, imageData, err := receiveStreamData(stream)
	if err != nil {
		return err
	}

	fileID, err := mapper.FromFileIdentifierProto(metadata.GetFileIdentifier())
	if err != nil {
		return err
	}

	// Create domain request
	request := &dto.ImageUploadRequest{
		ImageData:         imageData,
		ContentType:       metadata.ContentType,
		FileSize:          metadata.FileSize,
		FileIdentifierDTO: fileID,
	}

	// Call service layer
	uri, err := ebs.imageSvc.Upload(stream.Context(), request)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&genericupload.UploadStreamResponse{Uri: uri})
}

func (ebs *imageUploadServer) GetUrl(ctx context.Context, req *genericupload.GetUrlRequest) (*genericupload.GetUrlResponse, error) {
	if req == nil {
		return nil, ungerr.BadRequestError("request is nil")
	}

	fileID, err := mapper.FromFileIdentifierProto(req.GetFileIdentifier())
	if err != nil {
		return nil, err
	}

	url, err := ebs.imageSvc.GetURL(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &genericupload.GetUrlResponse{
		Url: url,
	}, nil
}

func (ebs *imageUploadServer) Delete(ctx context.Context, req *genericupload.DeleteRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, ungerr.BadRequestError("request is nil")
	}

	fileID, err := mapper.FromFileIdentifierProto(req.GetFileIdentifier())
	if err != nil {
		return nil, err
	}

	if err := ebs.imageSvc.Delete(ctx, fileID); err != nil {
		return nil, err
	}

	return nil, nil
}

func receiveStreamData(stream imageupload.ImageUploadService_UploadStreamServer) (*genericupload.Metadata, []byte, error) {
	var metadata *genericupload.Metadata
	var imageData []byte

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, eris.Wrap(err, "failed to receive stream data")
		}

		switch data := req.Data.(type) {
		case *genericupload.UploadStreamRequest_Metadata:
			if metadata != nil {
				return nil, nil, ungerr.BadRequestError("metadata already received")
			}
			metadata = data.Metadata
			fileSize := metadata.GetFileSize()
			if fileSize <= 0 {
				return nil, nil, ungerr.BadRequestError("file size must be greater than zero")
			}
			if fileSize > appconstant.MaxFileSize {
				return nil, nil, ungerr.UnprocessableEntityError("file too large")
			}
			imageData = make([]byte, 0, fileSize)

		case *genericupload.UploadStreamRequest_Chunk:
			if metadata == nil {
				return nil, nil, ungerr.BadRequestError("metadata must be sent first")
			}

			nextSize := int64(len(imageData)) + int64(len(data.Chunk))
			if metadata.GetFileSize() > 0 && nextSize > metadata.GetFileSize() {
				return nil, nil, ungerr.BadRequestError("stream exceeds declared file size")
			}

			imageData = append(imageData, data.Chunk...)
		}
	}

	if metadata == nil {
		return nil, nil, ungerr.BadRequestError("no metadata received")
	}

	// Validate file size
	if int64(len(imageData)) != metadata.GetFileSize() {
		return nil, nil, ungerr.BadRequestError("actual file size doesn't match expected size")
	}

	return metadata, imageData, nil
}
