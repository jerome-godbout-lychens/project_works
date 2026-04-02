package domain

import "context"

// AttachmentStore handles attachment metadata persistence.
type AttachmentStore interface {
	CreateAttachment(context context.Context, attachment *Attachment) error
	GetAttachmentById(context context.Context, attachmentId string) (*Attachment, error)
	ListAttachmentsByElement(context context.Context, elementId string) ([]Attachment, error)
	DeleteAttachment(context context.Context, attachmentId string) error
}
