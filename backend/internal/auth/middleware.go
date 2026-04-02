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

// ContextKeyUserId is the key used to store the authenticated user ID in the request context
const ContextKeyUserId contextKey = "userId"

// AuthMiddleware handles authentication via API keys and session cookies
type AuthMiddleware struct {
	sessionManager *SessionManager
	apiKeyStore    domain.APIKeyStore
	userService    *service.UserService
}

// NewAuthMiddleware creates a new AuthMiddleware with the given dependencies
func NewAuthMiddleware(sessionManager *SessionManager, apiKeyStore domain.APIKeyStore, userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
		apiKeyStore:    apiKeyStore,
		userService:    userService,
	}
}

// Authenticate is HTTP middleware that validates authentication via API key or session cookie
// Sets userId in the request context if authenticated, otherwise returns 401
func (authMiddleware *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		// Try API key authentication first
		authHeader := request.Header.Get("Authorization")
		if authHeader != "" {
			bearerToken := extractBearerToken(authHeader)
			if bearerToken != "" {
				userId, err := authMiddleware.authenticateWithAPIKey(request.Context(), bearerToken)
				if err == nil {
					contextWithUserId := context.WithValue(request.Context(), ContextKeyUserId, userId)
					next.ServeHTTP(responseWriter, request.WithContext(contextWithUserId))
					return
				}
			}
		}

		// Try session cookie authentication
		sessionCookie, err := request.Cookie("session")
		if err == nil {
			userId, err := authMiddleware.sessionManager.ValidateSessionToken(sessionCookie.Value)
			if err == nil {
				contextWithUserId := context.WithValue(request.Context(), ContextKeyUserId, userId)
				next.ServeHTTP(responseWriter, request.WithContext(contextWithUserId))
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

	return apiKeyEntity.UserId, nil
}

// extractBearerToken extracts the token from a "Bearer <token>" authorization header
func extractBearerToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// GetUserIdFromContext retrieves the authenticated user ID from the request context
// Returns the user ID and a boolean indicating if it was found
func GetUserIdFromContext(ctx context.Context) (string, bool) {
	userId, ok := ctx.Value(ContextKeyUserId).(string)
	return userId, ok
}
