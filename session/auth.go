// Package session implements session based authentication type-safe cookies, and flash messages.
package session

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrExpiredSession is returned when a session is expired and has been successfully deleted.
	ErrExpiredSession = errors.New("expired session")

	// ErrInvalidSessionSecret is returned when a session secret does not match the stored secret.
	ErrInvalidSessionSecret = errors.New("invalid session secret")
)

// StaleSessionError indicates that the trusted token is stale and requires
// fallback to session store validation.
type StaleSessionError struct{ SessionToken string }

func (e *StaleSessionError) Error() string { return "stale session" }

type (
	// Session represents a user session with an ID, secret hash and expiration time.
	//
	// The secret hash must not be exposed.
	// If the session type is returned to the client, ensure that the secret hash field is always omitted.
	Session interface {
		GetSession() (sessionId string, sessionSecretHash []byte, expiresAt time.Time)
	}

	// SessionStore provides storage operations for sessions and users.
	SessionStore[S Session, U any] interface {
		// GetSessionAndUser retrieves a session and its associated user by session ID.
		GetSessionAndUser(sessionId string) (S, U, error)

		// CreateSession creates a new session for a user with the given session data.
		CreateSession(sessionData Session, user U) (S, error)

		// UpdateSession updates a session's expiration time (used for session refresh).
		UpdateSession(sessionId string, expiresAt time.Time) (S, error)

		// DeleteSession deletes a session by ID.
		DeleteSession(sessionId string) error

		// DeleteUserSessions deletes all sessions for user.
		DeleteUserSessions(user U) error
	}

	// TokenStore provides token storage operations for HTTP requests/responses.
	// Default implementation uses unsigned cookies.
	TokenStore interface {
		// GetToken retrieves the session token from the request.
		GetToken(r *http.Request) (sessionToken string, err error)

		// SetToken stores the session token in the response.
		SetToken(w http.ResponseWriter, sessionToken string, expiresAt time.Time) error

		// DeleteToken removes the session token from the response.
		DeleteToken(w http.ResponseWriter)
	}

	// TrustedTokenStore validates session data optimistically to avoid checking the session store.
	// Use this for short-lived signed cookies or JWT tokens.
	TrustedTokenStore[S Session, U any] interface {
		TokenStore

		// FromTrustedToken attempts to validate and extract session data from a token.
		// The sessionToken parameter comes from TokenStore.GetToken().
		// Returns a *StaleSessionError to fallback to validating the session token with the session store.
		FromTrustedToken(trustedToken string) (session S, user U, err error)

		// ToTrustedToken transforms session data into a trusted token string.
		// The returned token will be passed to TokenStore.SetToken().
		ToTrustedToken(sessionToken string, session S, user U) (trustedToken string, err error)
	}
)

// Auth provides user authentication and session management operations.
//
// Example usage:
//
//	// Create auth instance
//	auth, _ := NewAuth(sessionStore)
//
//	// Login user
//	session, _ := auth.LoginRequest(w, user)
//
//	// Authenticate user on protected routes
//	session, user, _ := auth.AuthenticateRequest(w, r)
//
//	// Logout user
//	session, user, _ := auth.LogoutRequest(w, r)
type Auth[S Session, U any] interface {
	// Sessions

	// CreateSession creates a new session for the given user.
	// Returns the session token (<session id>.<session secret>) and the created session.
	CreateSession(user U) (sessionToken string, session S, err error)

	// InvalidateSession deletes a session by its ID.
	InvalidateSession(sessionId string) error

	// InvalidateSession deletes all sessions for user.
	InvalidateAllSessions(user U) error

	// Session Tokens

	// GetSessionToken retrieves the session token from the request.
	GetSessionToken(r *http.Request) (sessionToken string, err error)

	// SetSessionToken stores the session token in the response.
	// Uses TrustedTokenStore if available, otherwise uses regular TokenStore.
	SetSessionToken(w http.ResponseWriter, sessionToken string, session S, user U) error

	// DeleteSessionToken removes the session token from the response.
	DeleteSessionToken(w http.ResponseWriter)

	// ValidateSessionToken validates a session token and returns session data.
	// Automatically refreshes sessions that are past their refresh threshold.
	// Returns optimistic=true if a trusted token store was used and the session token
	// was validated without checking the session store.
	// The token store should always be updated with the returned session if optimistic=false.
	ValidateSessionToken(sessionToken string) (session S, user U, optimistic bool, err error)

	// Request Helpers

	// LoginRequest creates a new session and sets the session token in the response.
	// Use this after successful user login or registration.
	LoginRequest(w http.ResponseWriter, user U) (session S, err error)

	// AuthenticateRequest validates the session token from the request and returns session data.
	// Automatically refreshes sessions that are past their refresh threshold.
	// Updates the token store automatically with the returned session when optimistic=false.
	AuthenticateRequest(w http.ResponseWriter, r *http.Request) (session S, user U, err error)

	// LogoutRequest validates the request, invalidates the session, and deletes the session token.
	// Use this for logout operations.
	LogoutRequest(w http.ResponseWriter, r *http.Request) (session S, user U, err error)
}

// NewAuth creates a new Auth instance with the given session store and options.
//
// Defaults:
//   - IsDev: false
//   - Duration: 30 days
//   - TokenStore: Unsigned cookie token store (stores session ID)
//   - RefreshThreshold: Half-time (refresh when 50% of duration has passed)
//   - GenerateSessionToken: Base32 encoding of 20 random bytes
//
// Example:
//
//	auth, err := NewAuth(sessionStore,
//	    WithDev(true),
//	    WithDuration(time.Hour * 24 * 7),
//	    WithTokenStore(customTokenStore),
//	)
func NewAuth[S Session, U any](sessionStore SessionStore[S, U], opts ...func(*AuthOptions)) (Auth[S, U], error) {
	options := AuthOptions{
		RefreshThreshold: halfTimeRefreshThreshold,
	}
	// Configure auth options
	for _, opt := range opts {
		opt(&options)
	}
	// Set missing required options
	if options.Duration == 0 {
		options.Duration = time.Hour * 24 * 30 // 30 days
	}
	if options.GenerateSessionToken == nil {
		options.GenerateSessionToken = generateSessionToken
	}
	if options.TokenStore == nil {
		cookieStore, err := NewCookieTokenStore(func(co *CookieOptions) {
			co.Secure = !options.IsDev
		})
		if err != nil {
			return nil, err
		}
		options.TokenStore = cookieStore
	}
	trustedTokenStore, _ := options.TokenStore.(TrustedTokenStore[S, U])

	if sessionStore == nil {
		return nil, fmt.Errorf("session storage is nil")
	}

	return &auth[S, U]{
		options:           options,
		sessionStore:      sessionStore,
		tokenStore:        options.TokenStore,
		trustedTokenStore: trustedTokenStore,
	}, nil
}

type auth[S Session, U any] struct {
	options           AuthOptions
	sessionStore      SessionStore[S, U]
	trustedTokenStore TrustedTokenStore[S, U]
	tokenStore        TokenStore
}

// CreateSession creates a new session.
// The session token is <session id>.<session secret>.
func (a *auth[S, U]) CreateSession(user U) (string, S, error) {
	var session S

	// Generate session id and secret
	sessionId, sessionSecret, err := a.options.GenerateSessionToken()
	if err != nil {
		return "", session, err
	}
	// Hash session secret
	sessionSecretHash := hashSessionSecret(sessionSecret)

	sessionToken := fmt.Sprintf("%s.%s", sessionId, sessionSecret)
	sessionData := sessionData{
		id:         sessionId,
		secretHash: sessionSecretHash,
		expiresAt:  time.Now().Add(a.options.Duration),
	}

	session, err = a.sessionStore.CreateSession(sessionData, user)
	return sessionToken, session, err
}

// InvalidateSession deletes a session.
func (a *auth[S, U]) InvalidateSession(sessionId string) error {
	return a.sessionStore.DeleteSession(sessionId)
}

// InvalidateSession deletes all sessions for user.
func (a *auth[S, U]) InvalidateAllSessions(user U) error {
	return a.sessionStore.DeleteUserSessions(user)
}

// GetSessionToken retrieves the session token from the request.
func (a *auth[S, U]) GetSessionToken(r *http.Request) (string, error) {
	return a.tokenStore.GetToken(r)
}

// SetSessionToken stores the session token in the response.
func (a *auth[S, U]) SetSessionToken(w http.ResponseWriter, sessionToken string, session S, user U) error {
	if a.trustedTokenStore != nil {
		// Transform the session token
		transformedToken, err := a.trustedTokenStore.ToTrustedToken(sessionToken, session, user)
		if err != nil {
			return err
		}
		sessionToken = transformedToken
	}
	_, _, expiresAt := session.GetSession()
	return a.tokenStore.SetToken(w, sessionToken, expiresAt)
}

// DeleteSessionToken removes the session token from the response.
func (a *auth[S, U]) DeleteSessionToken(w http.ResponseWriter) {
	a.tokenStore.DeleteToken(w)
}

// ValidateSessionToken validates a session token and returns session data.
// Automatically refreshes sessions that are past their refresh threshold.
// Returns optimistic=true if a trusted token store was used and the session token
// was validated without checking the session store.
// The token store should always be updated with the returned session if optimistic=false.
func (a *auth[S, U]) ValidateSessionToken(sessionToken string) (S, U, bool, error) {
	if a.trustedTokenStore != nil {
		// Optimistically get trusted session and user from token
		session, user, err := a.trustedTokenStore.FromTrustedToken(sessionToken)
		if err == nil {
			// Token is fresh, no fallback needed
			return session, user, true, nil
		}

		// Check if error is a fallback error
		var staleErr *StaleSessionError
		if errors.As(err, &staleErr) {
			// Trusted token is stale, fallback to session store validation
			sessionToken = staleErr.SessionToken
		} else {
			// Validation failed completely
			return session, user, false, err
		}
	}

	var session S
	var user U

	sessionId, sessionSecret, _ := strings.Cut(sessionToken, ".")
	sessionSecretHash := hashSessionSecret(sessionSecret)

	// Get session and user from storage
	session, user, err := a.sessionStore.GetSessionAndUser(sessionId)
	if err != nil {
		return session, user, false, err
	}

	_, storedSessionSecretHash, expiresAt := session.GetSession()
	// Validate session secret
	if subtle.ConstantTimeCompare(sessionSecretHash, storedSessionSecretHash) == 0 {
		return session, user, false, ErrInvalidSessionSecret
	}
	// Check if session is expired
	if time.Now().After(expiresAt) {
		if err := a.sessionStore.DeleteSession(sessionId); err != nil {
			return session, user, false, err
		}
		return session, user, false, ErrExpiredSession
	}

	// Refresh session
	if a.options.RefreshThreshold != nil {
		newExpiresAt, needsRefresh := a.options.RefreshThreshold(expiresAt, a.options.Duration)
		if needsRefresh {
			session, err = a.sessionStore.UpdateSession(sessionId, newExpiresAt)
			if err != nil {
				return session, user, false, err
			}
		}
	}
	// Always refresh token store
	return session, user, false, nil
}

// LoginRequest creates a new session and sets the session token.
// Call this after successful user login or registration.
// Returns the created session or an error.
func (a *auth[S, U]) LoginRequest(w http.ResponseWriter, user U) (S, error) {
	// Create session
	sessionToken, session, err := a.CreateSession(user)
	if err != nil {
		return session, err
	}

	// Set request session
	err = a.SetSessionToken(w, sessionToken, session, user)
	return session, err
}

// AuthenticateRequest validates the session token from the request.
// Refreshes sessions past their refresh threshold.
// Deletes invalid or expired session tokens from the response.
// Returns the session, user, and any error encountered.
func (a *auth[S, U]) AuthenticateRequest(w http.ResponseWriter, r *http.Request) (S, U, error) {
	var session S
	var user U

	// Get session token from request
	sessionToken, err := a.GetSessionToken(r)
	if err != nil {
		return session, user, err
	}

	// Validate session token
	session, user, optimistic, err := a.ValidateSessionToken(sessionToken)
	if err != nil {
		// delete invalid session
		a.DeleteSessionToken(w)
		return session, user, err
	}

	// Set request session if optimistic=false
	if !optimistic {
		err = a.SetSessionToken(w, sessionToken, session, user)
		if err != nil {
			return session, user, err
		}
	}

	return session, user, nil
}

// LogoutRequest validates the request, invalidates the session, and deletes the session token.
func (a *auth[S, U]) LogoutRequest(w http.ResponseWriter, r *http.Request) (S, U, error) {
	var session S
	var user U

	// Get session token from request
	sessionToken, err := a.GetSessionToken(r)
	if err != nil {
		return session, user, err
	}

	// Validate session token
	session, user, _, err = a.ValidateSessionToken(sessionToken)
	if err != nil {
		// Delete session token even if validation fails
		a.DeleteSessionToken(w)
		return session, user, err
	}

	// Invalidate the session
	sessionId, _, _ := session.GetSession()
	if err := a.InvalidateSession(sessionId); err != nil {
		return session, user, err
	}

	// Delete session token
	a.DeleteSessionToken(w)

	return session, user, nil
}

// hashSessionSecret hashes a session token with SHA-256.
// Session tokens are hashed before storage to prevent token leakage from data breaches.
func hashSessionSecret(s string) []byte {
	hash := sha256.Sum256([]byte(s))
	return hash[:]
}

// generateSessionToken generates a secure random string.
// Uses base32 encoding (lowercase, no padding) of 20 random bytes.
func generateSessionToken() (string, string, error) {
	// generate random bytes for session id
	sessionIdBytes := make([]byte, 20)
	if _, err := rand.Read(sessionIdBytes); err != nil {
		return "", "", err
	}
	// generate random bytes for session secret
	sessionSecretBytes := make([]byte, 20)
	if _, err := rand.Read(sessionSecretBytes); err != nil {
		return "", "", err
	}

	// encode with base32 no padding lowercase
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	sessionId := encoding.EncodeToString(sessionIdBytes)
	sessionSecret := encoding.EncodeToString(sessionSecretBytes)

	return strings.ToLower(sessionId), strings.ToLower(sessionSecret), nil
}

// halfTimeRefreshThreshold is the default refresh threshold function.
// Refreshes sessions when they have passed 50% of their total duration.
// For example, a 30-day session will be refreshed after 15 days.
func halfTimeRefreshThreshold(expiresAt time.Time, duration time.Duration) (newExpiresAt time.Time, shouldRefresh bool) {
	if time.Until(expiresAt) < duration/2 {
		return time.Now().Add(duration), true
	}
	return
}

type sessionData struct {
	id         string
	secretHash []byte
	expiresAt  time.Time
}

func (s sessionData) GetSession() (sessionId string, sessionSecretHash []byte, expiresAt time.Time) {
	return s.id, s.secretHash, s.expiresAt
}
