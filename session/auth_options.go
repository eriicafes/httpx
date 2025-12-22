package session

import "time"

type AuthOptions struct {
	IsDev bool

	// Duration configures the session duration. default is 30 days.
	Duration time.Duration

	// TokenStore configures the token store.
	// TokenStore may implement [TrustedTokenStore].
	TokenStore TokenStore

	// RefreshThreshold configures how often a session is refreshed. default is half-time refresh threshold.
	// Set to nil to disabled refresh.
	RefreshThreshold func(expiresAt time.Time, duration time.Duration) (newExpiresAt time.Time, shouldRefresh bool)

	// GenerateSessionToken configures the session token generator function.
	// default is base32 encoding (lowercase, no padding) of 20 random bytes for session id and secret.
	GenerateSessionToken func() (sessionId, sessionSecret string, err error)
}

// WithDev enables development mode.
// In dev mode, Secure flag is disabled for cookies.
func WithDev(isDev bool) func(*AuthOptions) {
	return func(o *AuthOptions) {
		o.IsDev = isDev
	}
}

// WithDuration sets the session duration.
// Default: 30 days if not specified.
func WithDuration(duration time.Duration) func(*AuthOptions) {
	return func(o *AuthOptions) {
		o.Duration = duration
	}
}

// WithTokenStore sets the token store.
// TokenStore may implement [TrustedTokenStore].
func WithTokenStore(store TokenStore) func(*AuthOptions) {
	return func(o *AuthOptions) {
		o.TokenStore = store
	}
}

// WithRefreshThreshold sets the session refresh threshold function.
// Set to nil to disable automatic session refresh.
// Default: half-time refresh threshold.
func WithRefreshThreshold(fn func(expiresAt time.Time, duration time.Duration) (newExpiresAt time.Time, shouldRefresh bool)) func(*AuthOptions) {
	return func(o *AuthOptions) {
		o.RefreshThreshold = fn
	}
}

// WithGenerateSessionToken sets the session token generator function.
// Default: base32 encoding (lowercase, no padding) of 20 random bytes for session id and secret.
func WithGenerateSessionToken(fn func() (sessionId, sessionSecret string, err error)) func(*AuthOptions) {
	return func(o *AuthOptions) {
		o.GenerateSessionToken = fn
	}
}
