package entity

type StorageUploadRequest struct {
	Data        []byte
	ContentType string
	Filename    string
	BucketName  string
	ObjectKey   string
}
