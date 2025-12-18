package goss

import "io"

type UploadReq struct {
	FileName    string
	IsPrivate   bool
	StorageType StorageType
	Reader      io.Reader
	Overwrite   bool
	ContentType string
}
