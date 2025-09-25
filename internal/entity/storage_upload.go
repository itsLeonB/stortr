package entity

type StorageUploadRequest struct {
	Data        []byte
	ContentType string
	FileIdentifier
}

type FileIdentifier struct {
	BucketName string
	ObjectKey  string
}
