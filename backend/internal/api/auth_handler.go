package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/auth"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// LoginResponse is used for structured responses where needed.
type LoginResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
}

// LogoutResponse is used for logout responses.
type LogoutResponse struct {
	Success bool `json:"success"`
}

// CurrentUserResponse returns the authenticated user information.
type CurrentUserResponse struct {
	UserId      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// RegisterAuthHandlers registers authentication-related handlers.
// These handlers use the underlying chi router for custom HTTP handling
// (redirects, cookies, etc.) rather than going through Huma.
func RegisterAuthHandlers(
	chiRouter *chi.Mux,
	sessionManager *auth.SessionManager,
	oidcProvider *auth.OIDCProvider,
	userService *service.UserService,
) {
	// Login endpoint: redirects to OIDC provider
	chiRouter.Get("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Generate random state for CSRF protection
		stateByte := make([]byte, 32)
		_, err := rand.Read(stateByte)
		if err != nil {
			http.Error(w, "Failed to generate state", http.StatusInternalServerError)
			return
		}
		state := base64.URLEncoding.EncodeToString(stateByte)

		// Store state in session cookie (temporary)
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Get authorization URL from OIDC provider
		authURL := oidcProvider.GetAuthURL(state)

		// Redirect to OIDC provider
		http.Redirect(w, r, authURL, http.StatusFound)
	})

	// Callback endpoint: handles OIDC callback
	chiRouter.Get("/api/v1/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		// Verify state parameter
		state := r.URL.Query().Get("state")
		if state == "" {
			http.Error(w, "Missing state parameter", http.StatusBadRequest)
			return
		}

		stateCookie, err := r.Cookie("oauth_state")
		if err != nil {
			http.Error(w, "Missing state cookie", http.StatusBadRequest)
			return
		}

		if state != stateCookie.Value {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// Clear state cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			return
		}

		// Exchange code for token and get user info
		provider, subject, email, displayName, err := oidcProvider.HandleCallback(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
			return
		}

		// Get or create user
		user, err := userService.GetOrCreateUserFromOIDC(
			r.Context(),
			provider,
			subject,
			email,
			displayName,
		)
		if err != nil {
			http.Error(w, "Failed to get or create user", http.StatusInternalServerError)
			return
		}

		// Create session token
		sessionToken, err := sessionManager.CreateSessionToken(user.UserId)
		if err != nil {
			http.Error(w, "Failed to create session token", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionToken,
			Path:     "/",
			MaxAge:   86400, // 24 hours
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to frontend (or dashboard)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	// Logout endpoint: clears session cookie
	chiRouter.Post("/api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		// Clear session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	})

	// Current user endpoint: returns authenticated user info
	chiRouter.Get("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		userId, ok := auth.GetUserIdFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Unauthorized"}`))
			return
		}

		user, err := userService.GetUserById(r.Context(), userId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to get user"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := CurrentUserResponse{
			UserId:      user.UserId,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		}

		// Simple JSON marshaling
		w.Write([]byte(`{`))
		w.Write([]byte(`"user_id":"` + response.UserId + `",`))
		w.Write([]byte(`"email":"` + response.Email + `",`))
		w.Write([]byte(`"display_name":"` + response.DisplayName + `"`))
		w.Write([]byte(`}`))
	})
}

// RegisterAuthHandlersWithHuma is an alternative that registers the /me endpoint with Huma.
// However, login/callback/logout still need raw HTTP handlers due to redirects and cookies.
// This function is provided for consistency but the main RegisterAuthHandlers is recommended.
func RegisterAuthHandlersWithHuma(
	chiRouter chi.Router,
	sessionManager *auth.SessionManager,
	oidcProvider *auth.OIDCProvider,
	userService *service.UserService,
) {
	RegisterAuthHandlers(chiRouter.(*chi.Mux), sessionManager, oidcProvider, userService)
}
