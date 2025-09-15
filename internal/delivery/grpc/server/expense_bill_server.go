package server

import (
	"context"
	"io"

	"github.com/itsLeonB/stortr-protos/gen/go/expensebill/v1"
	"github.com/itsLeonB/stortr/internal/appconstant"
	"github.com/itsLeonB/stortr/internal/dto"
	"github.com/itsLeonB/stortr/internal/service"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ExpenseBillServer struct {
	expensebill.UnimplementedExpenseBillServiceServer
	expenseBillSvc service.ExpenseBillService
}

func newExpenseBillServer(expenseBillSvc service.ExpenseBillService) expensebill.ExpenseBillServiceServer {
	return &ExpenseBillServer{
		expenseBillSvc: expenseBillSvc,
	}
}

func (ebs *ExpenseBillServer) UploadStream(stream expensebill.ExpenseBillService_UploadStreamServer) error {
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
	objectKey, err := ebs.expenseBillSvc.Upload(stream.Context(), request)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&expensebill.UploadStreamResponse{ObjectKey: objectKey})
}

func (ebs *ExpenseBillServer) GetUrl(ctx context.Context, req *expensebill.GetUrlRequest) (*expensebill.GetUrlResponse, error) {
	if req == nil {
		return nil, ungerr.BadRequestError("request is nil")
	}

	if req.GetObjectKey() == "" {
		return nil, ungerr.BadRequestError("object key is empty")
	}

	url, err := ebs.expenseBillSvc.GetURL(ctx, req.GetObjectKey())
	if err != nil {
		return nil, err
	}

	return &expensebill.GetUrlResponse{
		Url: url,
	}, nil
}

func (ebs *ExpenseBillServer) Delete(ctx context.Context, req *expensebill.DeleteRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, ungerr.BadRequestError("request is nil")
	}

	if req.GetObjectKey() == "" {
		return nil, ungerr.BadRequestError("object key is empty")
	}

	if err := ebs.expenseBillSvc.Delete(ctx, req.GetObjectKey()); err != nil {
		return nil, err
	}

	return nil, nil
}

func receiveStreamData(stream expensebill.ExpenseBillService_UploadStreamServer) (*expensebill.BillMetadata, []byte, error) {
	var metadata *expensebill.BillMetadata
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
		case *expensebill.UploadStreamRequest_BillMetadata:
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

		case *expensebill.UploadStreamRequest_Chunk:
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
