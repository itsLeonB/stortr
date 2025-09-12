package dto

type UploadBillRequest struct {
	ImageData   []byte `validate:"required"`
	ContentType string `validate:"required,oneof=image/jpeg image/png image/jpg image/webp"`
	Filename    string `validate:"required,min=4"`
	FileSize    int64  `validate:"required,min=1"`
}
