package easy_mvc

import "mime/multipart"

type MultipartFile struct {
	multipart.File
	*multipart.FileHeader
	Error error
}
