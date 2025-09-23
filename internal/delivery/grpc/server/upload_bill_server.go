package server

import (
	"context"
	"io"

	"github.com/itsLeonB/stortr-protos/gen/go/uploadbill/v1"
	"github.com/itsLeonB/stortr/internal/appconstant"
	"github.com/itsLeonB/stortr/internal/dto"
	"github.com/itsLeonB/stortr/internal/service"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/types/known/emptypb"
)

type uploadBillServer struct {
	uploadbill.UnimplementedUploadBillServiceServer
	uploadbillSvc service.UploadBillService
}

func newUploadBillServer(uploadbillSvc service.UploadBillService) uploadbill.UploadBillServiceServer {
	return &uploadBillServer{
		uploadbillSvc: uploadbillSvc,
	}
}

func (ebs *uploadBillServer) UploadStream(stream uploadbill.UploadBillService_UploadStreamServer) error {
	metadata, imageData, err := receiveStreamData(stream)
	if err != nil {
		return err
	}

	// Create domain request
	request := &dto.UploadBillRequest{
		ImageData:   imageData,
		ContentType: metadata.ContentType,
		Filename:    metadata.Filename,
		FileSize:    metadata.FileSize,
	}

	// Call service layer
	objectKey, err := ebs.uploadbillSvc.Upload(stream.Context(), request)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&uploadbill.UploadStreamResponse{ObjectKey: objectKey})
}

func (ebs *uploadBillServer) GetUrl(ctx context.Context, req *uploadbill.GetUrlRequest) (*uploadbill.GetUrlResponse, error) {
	if req == nil {
		return nil, ungerr.BadRequestError("request is nil")
	}

	if req.GetObjectKey() == "" {
		return nil, ungerr.BadRequestError("object key is empty")
	}

	url, err := ebs.uploadbillSvc.GetURL(ctx, req.GetObjectKey())
	if err != nil {
		return nil, err
	}

	return &uploadbill.GetUrlResponse{
		Url: url,
	}, nil
}

func (ebs *uploadBillServer) Delete(ctx context.Context, req *uploadbill.DeleteRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, ungerr.BadRequestError("request is nil")
	}

	if req.GetObjectKey() == "" {
		return nil, ungerr.BadRequestError("object key is empty")
	}

	if err := ebs.uploadbillSvc.Delete(ctx, req.GetObjectKey()); err != nil {
		return nil, err
	}

	return nil, nil
}

func receiveStreamData(stream uploadbill.UploadBillService_UploadStreamServer) (*uploadbill.BillMetadata, []byte, error) {
	var metadata *uploadbill.BillMetadata
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
		case *uploadbill.UploadStreamRequest_BillMetadata:
			if metadata != nil {
				return nil, nil, ungerr.BadRequestError("metadata already received")
			}
			metadata = data.BillMetadata
			fileSize := metadata.GetFileSize()
			if fileSize <= 0 {
				return nil, nil, ungerr.BadRequestError("file size must be greater than zero")
			}
			if fileSize > appconstant.MaxFileSize {
				return nil, nil, ungerr.UnprocessableEntityError("file too large")
			}
			imageData = make([]byte, 0, fileSize)

		case *uploadbill.UploadStreamRequest_Chunk:
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
