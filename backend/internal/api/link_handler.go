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
	LinkIdentifier               string    `json:"link_identifier"`
	SourceElementIdentifier      string    `json:"source_element_identifier"`
	DestinationElementIdentifier string    `json:"destination_element_identifier"`
	LinkType             string    `json:"link_type"`
	CreationTime         time.Time `json:"creation_time"`
}

// ListElementLinksInput holds query parameters for listing element links.
type ListElementLinksInput struct {
	ElementIdentifier string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
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
	ElementIdentifier string `path:"element_identifier" format:"uuid" doc:"The source element identifier"`
	Body      struct {
		DestinationElementIdentifier string `json:"destination_element_identifier" required:"true" doc:"The destination element identifier"`
		LinkType             string `json:"link_type" required:"true" doc:"The type of link (related, child, implement)"`
	}
}

// CreateElementLinkOutput returns the created link.
type CreateElementLinkOutput struct {
	Body LinkResponse
}

// UpdateLinkInput holds the request body for updating a link's type.
type UpdateLinkInput struct {
	LinkIdentifier string `path:"link_identifier" format:"uuid" doc:"The link identifier"`
	Body   struct {
		LinkType string `json:"link_type" required:"true" doc:"New link type (related, child, implement)"`
	}
}

// UpdateLinkOutput returns the updated link.
type UpdateLinkOutput struct {
	Body LinkResponse
}

// DeleteLinkInput holds the path parameter for deleting a link.
type DeleteLinkInput struct {
	LinkIdentifier string `path:"link_identifier" format:"uuid" doc:"The link identifier"`
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
		Path:        "/api/v1/elements/{element_identifier}/links",
		Summary:     "List links for an element",
		Description: "Retrieve all links associated with an element, optionally filtered by direction.",
		Tags:        []string{"links"},
	}, func(ctx context.Context, input *ListElementLinksInput) (*ListElementLinksOutput, error) {
		links, err := linkService.ListLinksByElement(ctx, input.ElementIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list links", err)
		}

		output := &ListElementLinksOutput{}
		output.Body.Items = make([]LinkResponse, 0)

		for _, link := range links {
			// Filter by direction if specified
			if input.Direction == "outgoing" && link.SourceElementIdentifier != input.ElementIdentifier {
				continue
			}
			if input.Direction == "incoming" && link.DestinationElementIdentifier != input.ElementIdentifier {
				continue
			}

			output.Body.Items = append(output.Body.Items, LinkResponse{
				LinkIdentifier:               link.LinkIdentifier,
				SourceElementIdentifier:      link.SourceElementIdentifier,
				DestinationElementIdentifier: link.DestinationElementIdentifier,
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
		Path:        "/api/v1/elements/{element_identifier}/links",
		Summary:     "Create a new link",
		Tags:        []string{"links"},
	}, func(ctx context.Context, input *CreateElementLinkInput) (*CreateElementLinkOutput, error) {
		// Validate link type
		linkType := domain.LinkType(input.Body.LinkType)
		if !linkType.IsValid() {
			return nil, huma.NewError(http.StatusBadRequest, "Invalid link type", nil)
		}

		link := &domain.ElementLink{
			LinkIdentifier:               uuid.New().String(),
			SourceElementIdentifier:      input.ElementIdentifier,
			DestinationElementIdentifier: input.Body.DestinationElementIdentifier,
			LinkType:             linkType,
			CreationTime:         time.Now().UTC(),
		}

		err := linkService.CreateLink(ctx, link)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to create link", err)
		}

		return &CreateElementLinkOutput{
			Body: LinkResponse{
				LinkIdentifier:               link.LinkIdentifier,
				SourceElementIdentifier:      link.SourceElementIdentifier,
				DestinationElementIdentifier: link.DestinationElementIdentifier,
				LinkType:             string(link.LinkType),
				CreationTime:         link.CreationTime,
			},
		}, nil
	})

	// Update link type
	huma.Register(api, huma.Operation{
		OperationID: "updateElementLink",
		Method:      http.MethodPatch,
		Path:        "/api/v1/links/{link_identifier}",
		Summary:     "Update a link's type",
		Description: "Change the type of an existing link between two elements.",
		Tags:        []string{"links"},
	}, func(ctx context.Context, input *UpdateLinkInput) (*UpdateLinkOutput, error) {
		newLinkType := domain.LinkType(input.Body.LinkType)
		if !newLinkType.IsValid() {
			return nil, huma.NewError(http.StatusBadRequest, "Invalid link type", nil)
		}

		if err := linkService.UpdateLinkType(ctx, input.LinkIdentifier, newLinkType); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to update link", err)
		}

		// Fetch updated link to return in response
		link, err := linkService.GetLinkByIdentifier(ctx, input.LinkIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to fetch updated link", err)
		}

		return &UpdateLinkOutput{
			Body: LinkResponse{
				LinkIdentifier:               link.LinkIdentifier,
				SourceElementIdentifier:      link.SourceElementIdentifier,
				DestinationElementIdentifier: link.DestinationElementIdentifier,
				LinkType:             string(link.LinkType),
				CreationTime:         link.CreationTime,
			},
		}, nil
	})

	// Delete link
	huma.Register(api, huma.Operation{
		OperationID: "deleteElementLink",
		Method:      http.MethodDelete,
		Path:        "/api/v1/links/{link_identifier}",
		Summary:     "Delete a link",
		Tags:        []string{"links"},
	}, func(ctx context.Context, input *DeleteLinkInput) (*DeleteLinkOutput, error) {
		err := linkService.DeleteLink(ctx, input.LinkIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete link", err)
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
