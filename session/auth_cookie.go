package session

import (
	"fmt"
	"net/http"
	"time"
)

// NewCookieTokenStore creates a basic cookie-based token store that stores session tokens.
// The cookie is unsigned by default (just stores the session ID).
//
// Default options:
//   - Name: "auth_session"
//   - Secure: true
//   - HttpOnly: true
//   - SameSite: Lax mode
//   - Path: "/"
//
// Example usage:
//
//	// With default options
//	tokenStore, _ := NewCookieTokenStore()
//
//	// With custom options
//	tokenStore, _ := NewCookieTokenStore(
//	    WithCookieName("my_session"),
//	    WithCookieSecure(false), // For development
//	)
func NewCookieTokenStore(opts ...func(*CookieOptions)) (TokenStore, error) {
	cookieOptions := DefaultCookieOptions("auth_session", opts...)
	cookie, err := NewCookie[string](cookieOptions)
	if err != nil {
		return nil, err
	}
	return cookieTokenStore{cookie}, nil
}

// cookieTokenStore implements TokenStore using unsigned cookies.
// Stores only the session token.
type cookieTokenStore struct{ Cookie[string] }

// GetToken retrieves the session token from the cookie.
func (c cookieTokenStore) GetToken(r *http.Request) (sessionToken string, err error) {
	return c.Get(r)
}

// SetToken stores the session token in a cookie with the given expiration time.
func (c cookieTokenStore) SetToken(w http.ResponseWriter, sessionToken string, expiresAt time.Time) error {
	return c.Set(w, sessionToken, func(c *http.Cookie) {
		c.Expires = expiresAt.UTC()
	})
}

// DeleteToken removes the session cookie from the response.
func (c cookieTokenStore) DeleteToken(w http.ResponseWriter) {
	c.Delete(w)
}

// cachedSession stores session and user data in an encrypted cookie.
// This enables optimistic authentication without session validation until the data becomes stale.
type cachedSession[S Session, U any] struct {
	Session     S         `json:"session"`     // Full session data
	User        U         `json:"user"`        // Full user data
	Token       string    `json:"token"`       // Original session token for session store fallback
	StaleAtTime time.Time `json:"staleAtTime"` // When to trigger session store validation
}

// NewTrustedCookieTokenStore creates a cookie-based token store with session/user caching.
// This implements TrustedTokenStore to avoid session validation until the cache becomes stale.
//
// The cookie MUST be signed or encrypted for security (default is signed). Session and user data
// are stored in the cookie and trusted for the stale duration. After the stale duration elapses,
// the system falls back to session store validation using the stored session token.
//
// Default options:
//   - Secret: secret
//   - Name: "auth_session"
//   - Secure: true
//   - HttpOnly: true
//   - SameSite: Lax mode
//   - Path: "/"
//
// Example usage:
//
//	// With encrypted cookies
//	tokenStore, _ := NewTrustedCookieTokenStore[MySession, MyUser](
//	    Secret(secret),
//	    time.Minute * 15, // Trust for 15 minutes before session store check
//	    WithCookieEncrypted(true),
//	    WithCookieName("my_session"),
//	    WithCookieSecure(false), // For development
//	)
func NewTrustedCookieTokenStore[S Session, U any](secret [][]byte, staleDuration time.Duration, opts ...func(*CookieOptions)) (TrustedTokenStore[S, U], error) {
	opts = append(opts, WithCookieSecret(secret))
	cookieOptions := DefaultCookieOptions("auth_session", opts...)

	if len(cookieOptions.Secret) == 0 || !cookieOptions.Encrypted {
		return nil, fmt.Errorf("cookie must be signed or encrypted")
	}

	cookie, err := NewRawCookie[cachedSession[S, U]](cookieOptions)
	if err != nil {
		return nil, err
	}
	return &trustedCookieTokenStore[S, U]{cookie, staleDuration}, nil
}

// trustedCookieTokenStore implements TrustedTokenStore with encrypted cookie caching.
// Stores session and user data in the cookie to avoid session validation until stale.
type trustedCookieTokenStore[S Session, U any] struct {
	cookie        RawCookie[cachedSession[S, U]]
	staleDuration time.Duration
}

// GetToken retrieves the encrypted cookie value (not the session token inside).
// The cookie contains the full cached session data.
func (c *trustedCookieTokenStore[S, U]) GetToken(r *http.Request) (sessionToken string, err error) {
	cookie, err := r.Cookie(c.cookie.Name())
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetToken stores the encrypted cookie with cached session/user data and expiration.
// The sessionToken parameter is the encrypted cookie value from ToTrustedToken.
func (c *trustedCookieTokenStore[S, U]) SetToken(w http.ResponseWriter, sessionToken string, expiresAt time.Time) error {
	cookie := c.cookie.SetCookie(sessionToken, func(cookie *http.Cookie) {
		cookie.Expires = expiresAt.UTC()
	})
	http.SetCookie(w, cookie)
	return nil
}

// DeleteToken removes the cached session cookie from the response.
func (c *trustedCookieTokenStore[S, U]) DeleteToken(w http.ResponseWriter) {
	http.SetCookie(w, c.cookie.RemoveCookie())
}

// FromTrustedToken validates and extracts cached session/user data from the encrypted cookie.
// Returns StaleSessionError if the cache is stale (triggers session store fallback).
// Returns ErrExpiredSession if the session itself has expired.
func (c *trustedCookieTokenStore[S, U]) FromTrustedToken(trustedToken string) (session S, user U, err error) {
	// Parse and decrypt cookie data
	data, err := c.cookie.Parse(trustedToken)
	if err != nil {
		return session, user, err
	}

	// Check if session expired
	_, _, expiresAt := data.Session.GetSession()
	if time.Now().After(expiresAt) {
		return session, user, ErrExpiredSession
	}

	// Check if cache is stale - trigger session store validation
	if time.Now().After(data.StaleAtTime) {
		return session, user, &StaleSessionError{SessionToken: data.Token}
	}

	// Cache is fresh - return cached data
	return data.Session, data.User, nil
}

// ToTrustedToken creates an encrypted cookie containing session, user, and metadata.
// The cookie will be trusted for staleDuration before requiring session store validation.
func (c *trustedCookieTokenStore[S, U]) ToTrustedToken(sessionToken string, session S, user U) (trustedToken string, err error) {
	return c.cookie.Data(cachedSession[S, U]{
		Session:     session,
		User:        user,
		Token:       sessionToken,
		StaleAtTime: time.Now().Add(c.staleDuration),
	})
}
