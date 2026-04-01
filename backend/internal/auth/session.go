package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionManager handles JWT session token creation and validation
type SessionManager struct {
	secretKey []byte
}

// NewSessionManager creates a new SessionManager with the given secret key
func NewSessionManager(secretKey string) *SessionManager {
	return &SessionManager{
		secretKey: []byte(secretKey),
	}
}

// CreateSessionToken creates a JWT token with the provided userId
// Token expires in 24 hours from now
func (sessionManager *SessionManager) CreateSessionToken(userId string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"sub": userId,
		"iat": now.Unix(),
		"exp": expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(sessionManager.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateSessionToken parses and validates a JWT token string
// Returns the userId (sub claim) if valid, otherwise returns an error
func (sessionManager *SessionManager) ValidateSessionToken(tokenString string) (string, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return sessionManager.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("token is invalid")
	}

	userId, ok := claims["sub"].(string)
	if !ok || userId == "" {
		return "", errors.New("missing or invalid userId claim")
	}

	return userId, nil
}
