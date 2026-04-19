package api

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/auth"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// AttachmentResponse represents an attachment in API responses.
type AttachmentResponse struct {
	AttachmentIdentifier  string    `json:"attachment_identifier"`
	ElementIdentifier     string    `json:"element_identifier"`
	FileName      string    `json:"file_name"`
	FileSizeBytes int64     `json:"file_size_bytes"`
	ContentType   string    `json:"content_type"`
	UploadTime    time.Time `json:"upload_time"`
	UploadedByIdentifier  string    `json:"uploaded_by_id"`
}

// ListAttachmentsInput holds query parameters for listing attachments.
type ListAttachmentsInput struct {
	ElementIdentifier string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
}

// ListAttachmentsOutput returns a list of attachments.
type ListAttachmentsOutput struct {
	Body struct {
		Items []AttachmentResponse `json:"items"`
	}
}

// UploadAttachmentInput holds the request body for uploading an attachment.
type UploadAttachmentInput struct {
	ElementIdentifier   string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
	FileName    string `query:"file_name" required:"true" doc:"The name of the file"`
	ContentType string `header:"Content-Type" required:"true" doc:"The MIME type of the file"`
	RawBody     []byte `doc:"The file content"`
}

// UploadAttachmentOutput returns the uploaded attachment.
type UploadAttachmentOutput struct {
	Body AttachmentResponse
}

// GetAttachmentPresignedURLInput holds the path parameter for getting presigned URL.
type GetAttachmentPresignedURLInput struct {
	AttachmentIdentifier string `path:"attachment_identifier" format:"uuid" doc:"The attachment identifier"`
}

// GetAttachmentPresignedURLOutput returns the presigned URL.
type GetAttachmentPresignedURLOutput struct {
	Body struct {
		URL        string    `json:"url"`
		ExpiresAt  time.Time `json:"expires_at"`
	}
}

// DeleteAttachmentInput holds the path parameter for deleting an attachment.
type DeleteAttachmentInput struct {
	AttachmentIdentifier string `path:"attachment_identifier" format:"uuid" doc:"The attachment identifier"`
}

// DeleteAttachmentOutput is an empty response for successful deletion.
type DeleteAttachmentOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// RegisterAttachmentHandlers registers all attachment-related API handlers.
func RegisterAttachmentHandlers(api huma.API, attachmentService *service.AttachmentService) {
	// List attachments
	huma.Register(api, huma.Operation{
		OperationID: "listAttachments",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_identifier}/attachments",
		Summary:     "List attachments for an element",
		Tags:        []string{"attachments"},
	}, func(ctx context.Context, input *ListAttachmentsInput) (*ListAttachmentsOutput, error) {
		attachments, err := attachmentService.ListAttachmentsByElement(ctx, input.ElementIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list attachments", err)
		}

		output := &ListAttachmentsOutput{}
		output.Body.Items = make([]AttachmentResponse, len(attachments))

		for i, attachment := range attachments {
			output.Body.Items[i] = AttachmentResponse{
				AttachmentIdentifier:  attachment.AttachmentIdentifier,
				ElementIdentifier:     attachment.ElementIdentifier,
				FileName:      attachment.FileName,
				FileSizeBytes: attachment.FileSizeBytes,
				ContentType:   attachment.ContentType,
				UploadTime:    attachment.UploadTime,
				UploadedByIdentifier:  attachment.UploadedByIdentifier,
			}
		}

		return output, nil
	})

	// Upload attachment
	huma.Register(api, huma.Operation{
		OperationID: "uploadAttachment",
		Method:      http.MethodPost,
		Path:        "/api/v1/elements/{element_identifier}/attachments",
		Summary:     "Upload an attachment to an element",
		Tags:        []string{"attachments"},
	}, func(ctx context.Context, input *UploadAttachmentInput) (*UploadAttachmentOutput, error) {
		// Get authenticated user from context
		userIdentifier, ok := auth.GetUserIdentifierFromContext(ctx)
		if !ok {
			return nil, huma.NewError(http.StatusUnauthorized, "User not authenticated", nil)
		}

		// Upload attachment with file content
		attachment, err := attachmentService.UploadAttachment(
			ctx,
			input.ElementIdentifier,
			input.FileName,
			input.ContentType,
			int64(len(input.RawBody)),
			bytes.NewReader(input.RawBody),
			userIdentifier,
		)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to upload attachment", err)
		}

		return &UploadAttachmentOutput{
			Body: AttachmentResponse{
				AttachmentIdentifier:  attachment.AttachmentIdentifier,
				ElementIdentifier:     attachment.ElementIdentifier,
				FileName:      attachment.FileName,
				FileSizeBytes: attachment.FileSizeBytes,
				ContentType:   attachment.ContentType,
				UploadTime:    attachment.UploadTime,
				UploadedByIdentifier:  attachment.UploadedByIdentifier,
			},
		}, nil
	})

	// Get presigned URL for attachment
	huma.Register(api, huma.Operation{
		OperationID: "getAttachmentPresignedURL",
		Method:      http.MethodGet,
		Path:        "/api/v1/attachments/{attachment_identifier}",
		Summary:     "Get a presigned URL to download an attachment",
		Tags:        []string{"attachments"},
	}, func(ctx context.Context, input *GetAttachmentPresignedURLInput) (*GetAttachmentPresignedURLOutput, error) {
		const presignedURLExpiration = 15 * time.Minute
		url, err := attachmentService.GetPresignedURL(ctx, input.AttachmentIdentifier, presignedURLExpiration)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Failed to get presigned URL", err)
		}

		output := &GetAttachmentPresignedURLOutput{}
		output.Body.URL = url
		output.Body.ExpiresAt = time.Now().Add(presignedURLExpiration)

		return output, nil
	})

	// Delete attachment
	huma.Register(api, huma.Operation{
		OperationID: "deleteAttachment",
		Method:      http.MethodDelete,
		Path:        "/api/v1/attachments/{attachment_identifier}",
		Summary:     "Delete an attachment",
		Tags:        []string{"attachments"},
	}, func(ctx context.Context, input *DeleteAttachmentInput) (*DeleteAttachmentOutput, error) {
		err := attachmentService.DeleteAttachment(ctx, input.AttachmentIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete attachment", err)
		}

		return &DeleteAttachmentOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})
}
