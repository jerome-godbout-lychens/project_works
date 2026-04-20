package service

import (
	"context"
	"sync"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// ─── fakeCacheStore ───────────────────────────────────────────────────────

type fakeCacheStore struct {
	data             map[string]interface{}
	invalidatedKeys  []string
	prefixInvalidated []string
	mu               sync.Mutex
}

func newFakeCacheStore() *fakeCacheStore {
	return &fakeCacheStore{data: make(map[string]interface{})}
}

func (c *fakeCacheStore) Get(_ context.Context, cacheKey string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.data[cacheKey]
	return value, ok
}

func (c *fakeCacheStore) Set(_ context.Context, cacheKey string, value interface{}, _ time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[cacheKey] = value
}

func (c *fakeCacheStore) Invalidate(_ context.Context, cacheKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, cacheKey)
	c.invalidatedKeys = append(c.invalidatedKeys, cacheKey)
}

func (c *fakeCacheStore) InvalidateByPrefix(_ context.Context, prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prefixInvalidated = append(c.prefixInvalidated, prefix)
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// ─── fakeElementStore ─────────────────────────────────────────────────────

type fakeElementStore struct {
	elements map[string]*domain.Element
}

func newFakeElementStore() *fakeElementStore {
	return &fakeElementStore{elements: make(map[string]*domain.Element)}
}

func (s *fakeElementStore) GetElementByIdentifier(_ context.Context, elementIdentifier string) (*domain.Element, error) {
	if element, ok := s.elements[elementIdentifier]; ok {
		// Return a copy so callers mutating fields do not affect the store.
		copied := *element
		return &copied, nil
	}
	return nil, domain.ErrElementNotFound
}

func (s *fakeElementStore) ListElementsByProject(_ context.Context, _ string, _ domain.ElementFilter) ([]domain.Element, error) {
	return nil, nil
}

func (s *fakeElementStore) CreateElement(_ context.Context, element *domain.Element) error {
	copied := *element
	s.elements[element.ElementIdentifier] = &copied
	return nil
}

func (s *fakeElementStore) UpdateElement(_ context.Context, element *domain.Element) error {
	if _, ok := s.elements[element.ElementIdentifier]; !ok {
		return domain.ErrElementNotFound
	}
	copied := *element
	s.elements[element.ElementIdentifier] = &copied
	return nil
}

func (s *fakeElementStore) DeleteElement(_ context.Context, elementIdentifier string) error {
	if _, ok := s.elements[elementIdentifier]; !ok {
		return domain.ErrElementNotFound
	}
	delete(s.elements, elementIdentifier)
	return nil
}

func (s *fakeElementStore) SearchElements(_ context.Context, _ string, _ string, _ int, _ int) ([]domain.Element, error) {
	return nil, nil
}

// ─── fakeElementLinkStore ─────────────────────────────────────────────────

type fakeElementLinkStore struct {
	links map[string]*domain.ElementLink
}

func newFakeElementLinkStore() *fakeElementLinkStore {
	return &fakeElementLinkStore{links: make(map[string]*domain.ElementLink)}
}

func (s *fakeElementLinkStore) CreateLink(_ context.Context, link *domain.ElementLink) error {
	copied := *link
	s.links[link.LinkIdentifier] = &copied
	return nil
}

func (s *fakeElementLinkStore) GetLinkByIdentifier(_ context.Context, linkIdentifier string) (*domain.ElementLink, error) {
	if link, ok := s.links[linkIdentifier]; ok {
		copied := *link
		return &copied, nil
	}
	return nil, domain.ErrLinkNotFound
}

func (s *fakeElementLinkStore) UpdateLinkType(_ context.Context, linkIdentifier string, linkType domain.LinkType) error {
	link, ok := s.links[linkIdentifier]
	if !ok {
		return domain.ErrLinkNotFound
	}
	link.LinkType = linkType
	return nil
}

func (s *fakeElementLinkStore) DeleteLink(_ context.Context, linkIdentifier string) error {
	if _, ok := s.links[linkIdentifier]; !ok {
		return domain.ErrLinkNotFound
	}
	delete(s.links, linkIdentifier)
	return nil
}

func (s *fakeElementLinkStore) ListLinksByElement(
	_ context.Context,
	elementIdentifier string,
	direction domain.LinkDirection,
) ([]domain.ElementLink, error) {
	var out []domain.ElementLink
	for _, link := range s.links {
		matchesOutgoing := link.SourceElementIdentifier == elementIdentifier
		matchesIncoming := link.DestinationElementIdentifier == elementIdentifier
		switch direction {
		case domain.LinkDirectionOutgoing:
			if matchesOutgoing {
				out = append(out, *link)
			}
		case domain.LinkDirectionIncoming:
			if matchesIncoming {
				out = append(out, *link)
			}
		case domain.LinkDirectionBoth:
			if matchesOutgoing || matchesIncoming {
				out = append(out, *link)
			}
		}
	}
	return out, nil
}

// ─── fakeCustomFieldValueStore ────────────────────────────────────────────

type fakeCustomFieldValueStore struct {
	valuesByElement map[string]map[string]domain.CustomFieldValue
}

func newFakeCustomFieldValueStore() *fakeCustomFieldValueStore {
	return &fakeCustomFieldValueStore{valuesByElement: make(map[string]map[string]domain.CustomFieldValue)}
}

func (s *fakeCustomFieldValueStore) SetFieldValue(_ context.Context, elementIdentifier string, fieldDefinitionIdentifier string, value interface{}) error {
	if _, ok := s.valuesByElement[elementIdentifier]; !ok {
		s.valuesByElement[elementIdentifier] = make(map[string]domain.CustomFieldValue)
	}
	s.valuesByElement[elementIdentifier][fieldDefinitionIdentifier] = domain.CustomFieldValue{
		FieldDefinitionIdentifier: fieldDefinitionIdentifier,
		FieldValue:                value,
	}
	return nil
}

func (s *fakeCustomFieldValueStore) GetFieldValues(_ context.Context, elementIdentifier string) ([]domain.CustomFieldValue, error) {
	entries := s.valuesByElement[elementIdentifier]
	values := make([]domain.CustomFieldValue, 0, len(entries))
	for _, value := range entries {
		values = append(values, value)
	}
	return values, nil
}

func (s *fakeCustomFieldValueStore) DeleteFieldValue(_ context.Context, elementIdentifier string, fieldDefinitionIdentifier string) error {
	if entries, ok := s.valuesByElement[elementIdentifier]; ok {
		delete(entries, fieldDefinitionIdentifier)
	}
	return nil
}

// ─── fakeAttachmentStore ──────────────────────────────────────────────────

type fakeAttachmentStore struct {
	attachments map[string]*domain.Attachment
}

func newFakeAttachmentStore() *fakeAttachmentStore {
	return &fakeAttachmentStore{attachments: make(map[string]*domain.Attachment)}
}

func (s *fakeAttachmentStore) CreateAttachment(_ context.Context, attachment *domain.Attachment) error {
	copied := *attachment
	s.attachments[attachment.AttachmentIdentifier] = &copied
	return nil
}

func (s *fakeAttachmentStore) GetAttachmentByIdentifier(_ context.Context, attachmentIdentifier string) (*domain.Attachment, error) {
	if attachment, ok := s.attachments[attachmentIdentifier]; ok {
		copied := *attachment
		return &copied, nil
	}
	return nil, domain.ErrAttachmentNotFound
}

func (s *fakeAttachmentStore) ListAttachmentsByElement(_ context.Context, elementIdentifier string) ([]domain.Attachment, error) {
	var out []domain.Attachment
	for _, attachment := range s.attachments {
		if attachment.ElementIdentifier == elementIdentifier {
			out = append(out, *attachment)
		}
	}
	return out, nil
}

func (s *fakeAttachmentStore) DeleteAttachment(_ context.Context, attachmentIdentifier string) error {
	if _, ok := s.attachments[attachmentIdentifier]; !ok {
		return domain.ErrAttachmentNotFound
	}
	delete(s.attachments, attachmentIdentifier)
	return nil
}

// ─── fakePendingChangeStore ───────────────────────────────────────────────

type fakePendingChangeStore struct {
	changesByElement map[string]map[string]interface{}
	upsertCalls      int
	deleteCalls      int
}

func newFakePendingChangeStore() *fakePendingChangeStore {
	return &fakePendingChangeStore{changesByElement: make(map[string]map[string]interface{})}
}

func (s *fakePendingChangeStore) UpsertPendingChange(_ context.Context, elementIdentifier string, snapshotBeforeEdits map[string]interface{}) error {
	s.upsertCalls++
	// UpsertPendingChange semantics: snapshot is only set on first insert. For test purposes
	// we record the first snapshot seen for each element and keep it stable afterwards.
	if _, exists := s.changesByElement[elementIdentifier]; !exists {
		s.changesByElement[elementIdentifier] = snapshotBeforeEdits
	}
	return nil
}

func (s *fakePendingChangeStore) GetStalePendingChanges(_ context.Context, _ time.Duration) ([]domain.PendingChange, error) {
	return nil, nil
}

func (s *fakePendingChangeStore) DeletePendingChange(_ context.Context, elementIdentifier string) error {
	s.deleteCalls++
	delete(s.changesByElement, elementIdentifier)
	return nil
}

// ─── fakeProjectStore ─────────────────────────────────────────────────────

type fakeProjectStore struct {
	projects map[string]*domain.Project
}

func newFakeProjectStore() *fakeProjectStore {
	return &fakeProjectStore{projects: make(map[string]*domain.Project)}
}

func (s *fakeProjectStore) GetProjectByIdentifier(_ context.Context, projectIdentifier string) (*domain.Project, error) {
	if project, ok := s.projects[projectIdentifier]; ok {
		copied := *project
		return &copied, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (s *fakeProjectStore) ListProjects(_ context.Context, folderPathPrefix string) ([]domain.Project, error) {
	var out []domain.Project
	for _, project := range s.projects {
		if folderPathPrefix == "" || (len(project.FolderPath) >= len(folderPathPrefix) && project.FolderPath[:len(folderPathPrefix)] == folderPathPrefix) {
			out = append(out, *project)
		}
	}
	return out, nil
}

func (s *fakeProjectStore) ListFolderPaths(_ context.Context) ([]string, error) {
	seen := make(map[string]struct{})
	for _, project := range s.projects {
		seen[project.FolderPath] = struct{}{}
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	return paths, nil
}

func (s *fakeProjectStore) CreateProject(_ context.Context, project *domain.Project) error {
	copied := *project
	s.projects[project.ProjectIdentifier] = &copied
	return nil
}

func (s *fakeProjectStore) UpdateProject(_ context.Context, project *domain.Project) error {
	if _, ok := s.projects[project.ProjectIdentifier]; !ok {
		return domain.ErrProjectNotFound
	}
	copied := *project
	s.projects[project.ProjectIdentifier] = &copied
	return nil
}

func (s *fakeProjectStore) DeleteProject(_ context.Context, projectIdentifier string) error {
	if _, ok := s.projects[projectIdentifier]; !ok {
		return domain.ErrProjectNotFound
	}
	delete(s.projects, projectIdentifier)
	return nil
}

// ─── fakePhaseStore ───────────────────────────────────────────────────────

type fakePhaseStore struct {
	phases map[string]*domain.Phase
}

func newFakePhaseStore() *fakePhaseStore {
	return &fakePhaseStore{phases: make(map[string]*domain.Phase)}
}

func (s *fakePhaseStore) CreatePhase(_ context.Context, phase *domain.Phase) error {
	copied := *phase
	s.phases[phase.PhaseIdentifier] = &copied
	return nil
}

func (s *fakePhaseStore) ListPhasesByProject(_ context.Context, projectIdentifier string) ([]domain.Phase, error) {
	var out []domain.Phase
	for _, phase := range s.phases {
		if phase.ProjectIdentifier == projectIdentifier {
			out = append(out, *phase)
		}
	}
	return out, nil
}

func (s *fakePhaseStore) UpdatePhase(_ context.Context, phase *domain.Phase) error {
	if _, ok := s.phases[phase.PhaseIdentifier]; !ok {
		return domain.ErrPhaseNotFound
	}
	copied := *phase
	s.phases[phase.PhaseIdentifier] = &copied
	return nil
}

func (s *fakePhaseStore) DeletePhase(_ context.Context, phaseIdentifier string) error {
	if _, ok := s.phases[phaseIdentifier]; !ok {
		return domain.ErrPhaseNotFound
	}
	delete(s.phases, phaseIdentifier)
	return nil
}

// ─── fakeUserStore ────────────────────────────────────────────────────────

type fakeUserStore struct {
	usersByIdentifier       map[string]*domain.User
	usersByExternalIdentity map[string]*domain.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		usersByIdentifier:       make(map[string]*domain.User),
		usersByExternalIdentity: make(map[string]*domain.User),
	}
}

func externalIdentityKey(provider, subject string) string {
	return provider + "|" + subject
}

func (s *fakeUserStore) GetUserByIdentifier(_ context.Context, userIdentifier string) (*domain.User, error) {
	if user, ok := s.usersByIdentifier[userIdentifier]; ok {
		copied := *user
		return &copied, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *fakeUserStore) GetUserByExternalIdentity(_ context.Context, provider string, subject string) (*domain.User, error) {
	if user, ok := s.usersByExternalIdentity[externalIdentityKey(provider, subject)]; ok {
		copied := *user
		return &copied, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *fakeUserStore) CreateUser(_ context.Context, user *domain.User) error {
	copied := *user
	s.usersByIdentifier[user.UserIdentifier] = &copied
	s.usersByExternalIdentity[externalIdentityKey(user.ExternalIdentityProvider, user.ExternalIdentitySubject)] = &copied
	return nil
}

func (s *fakeUserStore) UpdateUser(_ context.Context, user *domain.User) error {
	copied := *user
	s.usersByIdentifier[user.UserIdentifier] = &copied
	s.usersByExternalIdentity[externalIdentityKey(user.ExternalIdentityProvider, user.ExternalIdentitySubject)] = &copied
	return nil
}

func (s *fakeUserStore) ListUsers(_ context.Context, _ int, _ int) ([]domain.User, error) {
	users := make([]domain.User, 0, len(s.usersByIdentifier))
	for _, user := range s.usersByIdentifier {
		users = append(users, *user)
	}
	return users, nil
}
