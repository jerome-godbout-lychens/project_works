package domain

import "time"

// Attachment references a file stored in the file storage backend (S3/SeaweedFS).
type Attachment struct {
	AttachmentId string
	ElementId    string
	FileStorageKey       string
	FileName             string
	FileSizeBytes        int64
	ContentType          string
	UploadTime           time.Time
	UploadedById string
}
