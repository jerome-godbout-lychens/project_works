package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// VersionableSnapshot is the canonical serializable form of an element's full state.
// It includes core fields, custom fields, supervisors, clients, links, and attachment metadata.
// This struct is serialized to deterministic JSON for diffing and SHA computation.
type VersionableSnapshot struct {
	ElementType   domain.ElementType   `json:"element_type"`
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	InterestLevel *int                 `json:"interest_level,omitempty"`
	AssigneeId    *string              `json:"assignee_id,omitempty"`
	TaskStatus    *domain.TaskStatus   `json:"task_status,omitempty"`
	TaskProgress  *int                 `json:"task_progress,omitempty"`
	ParentFeatureId  *string           `json:"parent_feature_id,omitempty"`
	StartPhaseId     *string           `json:"start_phase_id,omitempty"`
	DeliveryPhaseId  *string           `json:"delivery_phase_id,omitempty"`
	CustomFieldValues []SnapshotCustomField `json:"custom_field_values"`
	Supervisors       []string              `json:"supervisors"`
	Clients           []SnapshotContact     `json:"clients"`
	Links             []SnapshotLink        `json:"links"`
	Attachments       []SnapshotAttachment  `json:"attachments"`
}

type SnapshotCustomField struct {
	FieldDefinitionId string      `json:"field_definition_id"`
	FieldName         string      `json:"field_name"`
	FieldValue        interface{} `json:"field_value"`
}

type SnapshotContact struct {
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	ContactNotes string `json:"contact_notes"`
}

type SnapshotLink struct {
	SourceElementId      string          `json:"source_element_id"`
	DestinationElementId string          `json:"destination_element_id"`
	LinkType             domain.LinkType `json:"link_type"`
}

type SnapshotAttachment struct {
	AttachmentId string `json:"attachment_id"`
	FileName     string `json:"file_name"`
	ContentType  string `json:"content_type"`
	FileSizeBytes int64 `json:"file_size_bytes"`
}

// BuildSnapshot assembles a VersionableSnapshot from the element and its related data.
func BuildSnapshot(
	element *domain.Element,
	links []domain.ElementLink,
	attachments []domain.Attachment,
) *VersionableSnapshot {
	snapshot := &VersionableSnapshot{
		ElementType:     element.ElementType,
		Title:           element.Title,
		Description:     element.Description,
		InterestLevel:   element.InterestLevel,
		AssigneeId:      element.AssigneeId,
		TaskStatus:      element.TaskStatus,
		TaskProgress:    element.TaskProgress,
		ParentFeatureId: element.ParentFeatureId,
		StartPhaseId:    element.StartPhaseId,
		DeliveryPhaseId: element.DeliveryPhaseId,
	}

	// Custom fields — sorted by definition ID for deterministic output
	snapshot.CustomFieldValues = make([]SnapshotCustomField, len(element.CustomFieldValues))
	for i, customField := range element.CustomFieldValues {
		snapshot.CustomFieldValues[i] = SnapshotCustomField{
			FieldDefinitionId: customField.FieldDefinitionId,
			FieldName:         customField.FieldName,
			FieldValue:        customField.FieldValue,
		}
	}
	sort.Slice(snapshot.CustomFieldValues, func(i, j int) bool {
		return snapshot.CustomFieldValues[i].FieldDefinitionId < snapshot.CustomFieldValues[j].FieldDefinitionId
	})

	// Supervisors — sorted for determinism
	snapshot.Supervisors = make([]string, len(element.Supervisors))
	copy(snapshot.Supervisors, element.Supervisors)
	sort.Strings(snapshot.Supervisors)

	// Clients — sorted by name
	snapshot.Clients = make([]SnapshotContact, len(element.Clients))
	for i, client := range element.Clients {
		snapshot.Clients[i] = SnapshotContact{
			ContactName:  client.ContactName,
			ContactEmail: client.ContactEmail,
			ContactNotes: client.ContactNotes,
		}
	}
	sort.Slice(snapshot.Clients, func(i, j int) bool {
		return snapshot.Clients[i].ContactName < snapshot.Clients[j].ContactName
	})

	// Links — sorted by destination + type
	snapshot.Links = make([]SnapshotLink, len(links))
	for i, link := range links {
		snapshot.Links[i] = SnapshotLink{
			SourceElementId:      link.SourceElementId,
			DestinationElementId: link.DestinationElementId,
			LinkType:             link.LinkType,
		}
	}
	sort.Slice(snapshot.Links, func(i, j int) bool {
		if snapshot.Links[i].DestinationElementId != snapshot.Links[j].DestinationElementId {
			return snapshot.Links[i].DestinationElementId < snapshot.Links[j].DestinationElementId
		}
		return string(snapshot.Links[i].LinkType) < string(snapshot.Links[j].LinkType)
	})

	// Attachments — sorted by ID
	snapshot.Attachments = make([]SnapshotAttachment, len(attachments))
	for i, attachment := range attachments {
		snapshot.Attachments[i] = SnapshotAttachment{
			AttachmentId:  attachment.AttachmentId,
			FileName:      attachment.FileName,
			ContentType:   attachment.ContentType,
			FileSizeBytes: attachment.FileSizeBytes,
		}
	}
	sort.Slice(snapshot.Attachments, func(i, j int) bool {
		return snapshot.Attachments[i].AttachmentId < snapshot.Attachments[j].AttachmentId
	})

	return snapshot
}

// SerializeSnapshot converts a snapshot to deterministic JSON bytes.
func SerializeSnapshot(snapshot *VersionableSnapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

// DeserializeSnapshot parses JSON bytes back into a VersionableSnapshot.
func DeserializeSnapshot(data []byte) (*VersionableSnapshot, error) {
	var snapshot VersionableSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// ComputeContentSha returns the SHA-256 hex digest of the serialized snapshot.
func ComputeContentSha(snapshotJson []byte) string {
	hash := sha256.Sum256(snapshotJson)
	return fmt.Sprintf("%x", hash)
}
