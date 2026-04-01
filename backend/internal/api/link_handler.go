package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// LinkResponse represents an element link in API responses.
type LinkResponse struct {
	LinkId               string    `json:"link_id"`
	SourceElementId      string    `json:"source_element_id"`
	DestinationElementId string    `json:"destination_element_id"`
	LinkType             string    `json:"link_type"`
	CreationTime         time.Time `json:"creation_time"`
}

// ListElementLinksInput holds query parameters for listing element links.
type ListElementLinksInput struct {
	ElementId string `path:"element_id" format:"uuid" doc:"The element identifier"`
	Direction string `query:"direction" default:"both" doc:"Direction of links: outgoing, incoming, or both"`
}

// ListElementLinksOutput returns a list of element links.
type ListElementLinksOutput struct {
	Body struct {
		Items []LinkResponse `json:"items"`
	}
}

// CreateElementLinkInput holds the request body for creating a link.
type CreateElementLinkInput struct {
	ElementId            string `path:"element_id" format:"uuid" doc:"The source element identifier"`
	DestinationElementId string `json:"destination_element_id" required:"true" doc:"The destination element identifier"`
	LinkType             string `json:"link_type" required:"true" doc:"The type of link (related, child, implement)"`
}

// CreateElementLinkOutput returns the created link.
type CreateElementLinkOutput struct {
	Body LinkResponse
}

// DeleteLinkInput holds the path parameter for deleting a link.
type DeleteLinkInput struct {
	LinkId string `path:"link_id" format:"uuid" doc:"The link identifier"`
}

// DeleteLinkOutput is an empty response for successful deletion.
type DeleteLinkOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// RegisterLinkHandlers registers all link-related API handlers.
func RegisterLinkHandlers(api huma.API, linkService *service.LinkService) {
	// List element links
	huma.Register(api, huma.Operation{
		OperationID: "listElementLinks",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_id}/links",
		Summary:     "List links for an element",
		Description: "Retrieve all links associated with an element, optionally filtered by direction.",
		Tags:        []string{"links"},
	}, func(ctx context.Context, input *ListElementLinksInput) (*ListElementLinksOutput, error) {
		links, err := linkService.ListLinksByElement(ctx, input.ElementId)
		if err != nil {
			return nil, huma.Error(http.StatusInternalServerError, "Failed to list links", err)
		}

		output := &ListElementLinksOutput{}
		output.Body.Items = make([]LinkResponse, 0)

		for _, link := range links {
			// Filter by direction if specified
			if input.Direction == "outgoing" && link.SourceElementId != input.ElementId {
				continue
			}
			if input.Direction == "incoming" && link.DestinationElementId != input.ElementId {
				continue
			}

			output.Body.Items = append(output.Body.Items, LinkResponse{
				LinkId:               link.LinkId,
				SourceElementId:      link.SourceElementId,
				DestinationElementId: link.DestinationElementId,
				LinkType:             string(link.LinkType),
				CreationTime:         link.CreationTime,
			})
		}

		return output, nil
	})

	// Create element link
	huma.Register(api, huma.Operation{
		OperationID: "createElementLink",
		Method:      http.MethodPost,
		Path:        "/api/v1/elements/{element_id}/links",
		Summary:     "Create a new link",
		Tags:        []string{"links"},
	}, func(ctx context.Context, input *CreateElementLinkInput) (*CreateElementLinkOutput, error) {
		// Validate link type
		linkType := domain.LinkType(input.LinkType)
		if !linkType.IsValid() {
			return nil, huma.Error(http.StatusBadRequest, "Invalid link type", nil)
		}

		link := &domain.ElementLink{
			LinkId:               uuid.New().String(),
			SourceElementId:      input.ElementId,
			DestinationElementId: input.DestinationElementId,
			LinkType:             linkType,
			CreationTime:         time.Now().UTC(),
		}

		err := linkService.CreateLink(ctx, link)
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to create link", err)
		}

		return &CreateElementLinkOutput{
			Body: LinkResponse{
				LinkId:               link.LinkId,
				SourceElementId:      link.SourceElementId,
				DestinationElementId: link.DestinationElementId,
				LinkType:             string(link.LinkType),
				CreationTime:         link.CreationTime,
			},
		}, nil
	})

	// Delete link
	huma.Register(api, huma.Operation{
		OperationID: "deleteElementLink",
		Method:      http.MethodDelete,
		Path:        "/api/v1/links/{link_id}",
		Summary:     "Delete a link",
		Tags:        []string{"links"},
	}, func(ctx context.Context, input *DeleteLinkInput) (*DeleteLinkOutput, error) {
		err := linkService.DeleteLink(ctx, input.LinkId)
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to delete link", err)
		}

		return &DeleteLinkOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})
}
