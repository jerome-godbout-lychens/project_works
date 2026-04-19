package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// ContextKeyUserIdentifier is the key used to store the authenticated user ID in the request context
const ContextKeyUserIdentifier contextKey = "userIdentifier"

// ContextKeyIsSuperAdmin flags whether the authenticated user is the super admin
const ContextKeyIsSuperAdmin contextKey = "isSuperAdmin"

// AuthMiddleware handles authentication via API keys and session cookies
type AuthMiddleware struct {
	sessionManager   *SessionManager
	apiKeyStore      domain.APIKeyStore
	userService      *service.UserService
	superAdminUserIdentifier string // empty when no super admin is configured
}

// NewAuthMiddleware creates a new AuthMiddleware with the given dependencies.
// superAdminUserIdentifier may be empty when no super admin is configured.
func NewAuthMiddleware(sessionManager *SessionManager, apiKeyStore domain.APIKeyStore, userService *service.UserService, superAdminUserIdentifier string) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager:   sessionManager,
		apiKeyStore:      apiKeyStore,
		userService:      userService,
		superAdminUserIdentifier: superAdminUserIdentifier,
	}
}

// Authenticate is HTTP middleware that validates authentication via API key or session cookie
// Sets userIdentifier in the request context if authenticated, otherwise returns 401
func (authMiddleware *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		// Try API key authentication first
		authHeader := request.Header.Get("Authorization")
		if authHeader != "" {
			bearerToken := extractBearerToken(authHeader)
			if bearerToken != "" {
				userIdentifier, err := authMiddleware.authenticateWithAPIKey(request.Context(), bearerToken)
				if err == nil {
					requestContext := context.WithValue(request.Context(), ContextKeyUserIdentifier, userIdentifier)
					requestContext = context.WithValue(requestContext, ContextKeyIsSuperAdmin, authMiddleware.superAdminUserIdentifier != "" && userIdentifier == authMiddleware.superAdminUserIdentifier)
					next.ServeHTTP(responseWriter, request.WithContext(requestContext))
					return
				}
			}
		}

		// Try session cookie authentication
		sessionCookie, err := request.Cookie("session")
		if err == nil {
			userIdentifier, err := authMiddleware.sessionManager.ValidateSessionToken(sessionCookie.Value)
			if err == nil {
				requestContext := context.WithValue(request.Context(), ContextKeyUserIdentifier, userIdentifier)
				requestContext = context.WithValue(requestContext, ContextKeyIsSuperAdmin, authMiddleware.superAdminUserIdentifier != "" && userIdentifier == authMiddleware.superAdminUserIdentifier)
				next.ServeHTTP(responseWriter, request.WithContext(requestContext))
				return
			}
		}

		// No valid authentication found
		responseWriter.WriteHeader(http.StatusUnauthorized)
		responseWriter.Write([]byte("Unauthorized"))
	})
}

// authenticateWithAPIKey validates an API key and returns the associated user ID
func (authMiddleware *AuthMiddleware) authenticateWithAPIKey(ctx context.Context, rawKey string) (string, error) {
	hashedKey := HashAPIKey(rawKey)
	apiKeyEntity, err := authMiddleware.apiKeyStore.GetAPIKeyByHash(ctx, hashedKey)
	if err != nil {
		return "", err
	}

	return apiKeyEntity.UserIdentifier, nil
}

// extractBearerToken extracts the token from a "Bearer <token>" authorization header
func extractBearerToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// GetUserIdentifierFromContext retrieves the authenticated user ID from the request context.
// Returns the user ID and a boolean indicating if it was found.
func GetUserIdentifierFromContext(ctx context.Context) (string, bool) {
	userIdentifier, ok := ctx.Value(ContextKeyUserIdentifier).(string)
	return userIdentifier, ok
}

// IsSuperAdmin returns true when the authenticated user is the configured super admin.
func IsSuperAdmin(ctx context.Context) bool {
	flag, _ := ctx.Value(ContextKeyIsSuperAdmin).(bool)
	return flag
}
