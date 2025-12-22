package session

import (
	"net/http"
	"time"
)

// DefaultCookieOptions creates a CookieOptions with secure defaults.
// The name parameter is required. Additional options can be provided to override defaults.
//
// Default values:
//   - Secure: true
//   - HttpOnly: true
//   - SameSite: Lax mode
//   - Path: "/"
//
// Cookie signing/encryption:
//   - No Secret: Unsigned cookie (not recommended for sensitive data)
//   - Secret provided: Signed cookie (HMAC-SHA256, prevents tampering)
//   - Secret + Encrypted=true: Encrypted cookie (AES-256-GCM, hides data)
//
// Example usage:
//
//	// Basic unsigned cookie
//	opts := DefaultCookieOptions("session_id")
//
//	// Signed cookie
//	opts := DefaultCookieOptions("session_id",
//	    WithCookieSecret(Secret(key)),
//	)
//
//	// Encrypted cookie
//	opts := DefaultCookieOptions("session_id",
//	    WithCookieSecret(Secret(key)),
//	    WithEncrypted(true),
//	)
func DefaultCookieOptions(name string, opts ...func(*CookieOptions)) CookieOptions {
	options := CookieOptions{
		Name:     name,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

type CookieOptions struct {
	// Secret is a slice of keys for signing and encrypting cookies.
	// Use Secret(key) for a single key or Secret(key1, key2, ...) for key rotation.
	// The first key is used for signing/encrypting, all keys are tried for verification/decryption.
	//
	// Cookie security based on Secret:
	//   - No Secret (nil or empty): Unsigned cookie
	//   - Secret provided: Signed cookie (HMAC-SHA256)
	//   - Secret + Encrypted=true: Encrypted cookie (AES-256-GCM)
	Secret [][]byte

	// Encrypted enables AES-256-GCM encryption.
	// Requires Secret to be set with keys of exactly 32 bytes each.
	Encrypted bool

	// Name configures the cookie name.
	Name string

	// Duration configures the session duration.
	// This is used to set the Expires time when a cookie is set.
	Duration time.Duration

	// Other cookie options

	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

// WithCookieName sets the cookie name.
func WithCookieName(name string) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.Name = name
	}
}

// WithCookieSecret sets the cookie secret for signing/encryption.
// Use Secret(key) for a single key or Secret(key1, key2, ...) for key rotation.
//
// Cookie behavior based on Secret:
//   - No Secret: Unsigned cookie (not recommended for sensitive data)
//   - Secret provided: Signed cookie (HMAC-SHA256, prevents tampering)
//   - Secret + WithEncrypted(true): Encrypted cookie (AES-256-GCM, hides data)
func WithCookieSecret(secret [][]byte) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.Secret = secret
	}
}

// WithCookieEncrypted enables AES-256-GCM encryption for cookies.
// Requires a secret key of exactly 32 bytes.
// Default: false (not encrypted).
func WithCookieEncrypted(encrypted bool) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.Encrypted = encrypted
	}
}

// WithCookieDuration sets the cookie duration.
// If > 0, creates a persistent cookie. If 0, creates a session cookie.
// Default: 0 (session cookie).
func WithCookieDuration(duration time.Duration) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.Duration = duration
	}
}

// WithCookiePath sets the cookie path.
// Default: "/" (all paths).
func WithCookiePath(path string) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.Path = path
	}
}

// WithCookieDomain sets the cookie domain.
// Default: "" (current domain only).
func WithCookieDomain(domain string) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.Domain = domain
	}
}

// WithCookieSecure sets the Secure flag on the cookie.
// When true, cookie is only sent over HTTPS.
// Default: true (secure).
func WithCookieSecure(secure bool) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.Secure = secure
	}
}

// WithCookieHttpOnly sets the HttpOnly flag on the cookie.
// When true, cookie is not accessible via JavaScript.
// Default: true (http only).
func WithCookieHttpOnly(httpOnly bool) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.HttpOnly = httpOnly
	}
}

// WithCookieSameSite sets the SameSite attribute on the cookie.
// Options: http.SameSiteDefaultMode, http.SameSiteLaxMode, http.SameSiteStrictMode, http.SameSiteNoneMode
// Default: http.SameSiteLaxMode
func WithCookieSameSite(sameSite http.SameSite) func(*CookieOptions) {
	return func(o *CookieOptions) {
		o.SameSite = sameSite
	}
}
