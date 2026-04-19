package domain

import "time"

// Attachment references a file stored in the file storage backend (S3/SeaweedFS).
type Attachment struct {
	AttachmentIdentifier string
	ElementIdentifier    string
	FileStorageKey       string
	FileName             string
	FileSizeBytes        int64
	ContentType          string
	UploadTime           time.Time
	UploadedByIdentifier string
}
