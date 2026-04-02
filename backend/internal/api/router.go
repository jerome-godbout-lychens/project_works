package api

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/danielgtaylor/huma/v2"
	humachi "github.com/danielgtaylor/huma/v2/adapters/humachi"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/auth"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// NewRouter creates and configures the HTTP router with all API handlers.
func NewRouter(
	projectService *service.ProjectService,
	elementService *service.ElementService,
	versionService *service.VersionService,
	linkService *service.LinkService,
	attachmentService *service.AttachmentService,
	customFieldService *service.CustomFieldService,
	phaseService *service.PhaseService,
	userService *service.UserService,
	groupService *service.GroupService,
	apiKeyStore domain.APIKeyStore,
	authMiddleware *auth.AuthMiddleware,
	sessionManager *auth.SessionManager,
	oidcProvider *auth.OIDCProvider,
) http.Handler {
	router := chi.NewRouter()

	// Apply authentication middleware to all routes except /auth endpoints
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth middleware for login and callback endpoints
			if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/callback" {
				next.ServeHTTP(w, r)
				return
			}
			authMiddleware.Authenticate(next).ServeHTTP(w, r)
		})
	})

	// Create Huma API with Chi adapter
	humConfig := huma.DefaultConfig("Project Works API", "1.0.0")
	humConfig.Servers = []*huma.Server{
		{
			URL:         "http://localhost:8080",
			Description: "Development server",
		},
	}
	humaAPI := humachi.New(router, humConfig)

	// Register all handler groups
	RegisterProjectHandlers(humaAPI, projectService)
	RegisterElementHandlers(humaAPI, elementService)
	RegisterVersionHandlers(humaAPI, versionService)
	RegisterLinkHandlers(humaAPI, linkService)
	RegisterAttachmentHandlers(humaAPI, attachmentService)
	RegisterCustomFieldHandlers(humaAPI, customFieldService)
	RegisterPhaseHandlers(humaAPI, phaseService)
	RegisterUserHandlers(humaAPI, userService)
	RegisterGroupHandlers(humaAPI, groupService)
	RegisterAccessHandlers(humaAPI, groupService)
	RegisterAPIKeyHandlers(humaAPI, apiKeyStore)
	RegisterAuthHandlers(router, sessionManager, oidcProvider, userService)

	return router
}
