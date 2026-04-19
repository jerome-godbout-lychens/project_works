package domain

import "context"

// AttachmentStore handles attachment metadata persistence.
type AttachmentStore interface {
	CreateAttachment(context context.Context, attachment *Attachment) error
	GetAttachmentByIdentifier(context context.Context, attachmentIdentifier string) (*Attachment, error)
	ListAttachmentsByElement(context context.Context, elementIdentifier string) ([]Attachment, error)
	DeleteAttachment(context context.Context, attachmentIdentifier string) error
}
