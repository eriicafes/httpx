package session

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewCookieTokenStore(t *testing.T) {
	tokenStore, err := NewCookieTokenStore()
	if err != nil {
		t.Fatalf("failed to create token store: %v", err)
	}

	// Test setting and getting token
	w := httptest.NewRecorder()
	testToken := "session123.secret456"
	expiresAt := time.Now().Add(time.Hour)

	err = tokenStore.SetToken(w, testToken, expiresAt)
	if err != nil {
		t.Fatalf("failed to set token: %v", err)
	}

	// Get token from request
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	retrievedToken, err := tokenStore.GetToken(r)
	if err != nil {
		t.Fatalf("failed to get token: %v", err)
	}

	if retrievedToken != testToken {
		t.Errorf("expected token %q, got %q", testToken, retrievedToken)
	}
}

func TestCookieTokenStore_CustomOptions(t *testing.T) {
	tokenStore, err := NewCookieTokenStore(
		WithCookieName("custom_session"),
		WithCookieSecure(false),
		WithCookiePath("/api"),
	)
	if err != nil {
		t.Fatalf("failed to create token store: %v", err)
	}

	w := httptest.NewRecorder()
	tokenStore.SetToken(w, "token", time.Now().Add(time.Hour))

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set")
	}

	cookie := cookies[0]
	if cookie.Name != "custom_session" {
		t.Errorf("expected cookie name 'custom_session', got %q", cookie.Name)
	}
	if cookie.Secure {
		t.Error("expected Secure to be false")
	}
	if cookie.Path != "/api" {
		t.Errorf("expected Path '/api', got %q", cookie.Path)
	}
}

func TestCookieTokenStore_DeleteToken(t *testing.T) {
	tokenStore, err := NewCookieTokenStore()
	if err != nil {
		t.Fatalf("failed to create token store: %v", err)
	}

	w := httptest.NewRecorder()
	tokenStore.DeleteToken(w)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no deletion cookie set")
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("expected MaxAge -1 for deletion, got %d", cookies[0].MaxAge)
	}
}

func TestCookieTokenStore_Expiration(t *testing.T) {
	tokenStore, err := NewCookieTokenStore()
	if err != nil {
		t.Fatalf("failed to create token store: %v", err)
	}

	w := httptest.NewRecorder()
	expiresAt := time.Now().Add(time.Hour * 24)

	tokenStore.SetToken(w, "token", expiresAt)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set")
	}

	// Check that expiration is set correctly (within 1 minute tolerance)
	if cookies[0].Expires.Before(expiresAt.Add(-time.Minute)) ||
		cookies[0].Expires.After(expiresAt.Add(time.Minute)) {
		t.Errorf("expected Expires around %v, got %v", expiresAt, cookies[0].Expires)
	}
}

func TestNewTrustedCookieTokenStore(t *testing.T) {
	secret := generateTestSecret()
	staleDuration := time.Minute * 15

	tokenStore, err := NewTrustedCookieTokenStore[SessionData, UserData](
		Secret(secret),
		staleDuration,
		WithCookieEncrypted(true),
	)
	if err != nil {
		t.Fatalf("failed to create trusted token store: %v", err)
	}

	// Create test session and user
	session := SessionData{
		Id:        "session123",
		UserId:    "user@example.com",
		Role:      "admin",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	user := UserData{
		Email: "user@example.com",
		Name:  "Test User",
		Role:  "admin",
	}

	sessionToken := "session123.secret456"

	// Convert to trusted token
	trustedToken, err := tokenStore.ToTrustedToken(sessionToken, session, user)
	if err != nil {
		t.Fatalf("failed to create trusted token: %v", err)
	}

	// Verify from trusted token
	retrievedSession, retrievedUser, err := tokenStore.FromTrustedToken(trustedToken)
	if err != nil {
		t.Fatalf("failed to get from trusted token: %v", err)
	}

	if retrievedSession.Id != session.Id {
		t.Errorf("expected session ID %q, got %q", session.Id, retrievedSession.Id)
	}
	if retrievedUser.Email != user.Email {
		t.Errorf("expected user email %q, got %q", user.Email, retrievedUser.Email)
	}
}

func TestTrustedCookieTokenStore_StaleToken(t *testing.T) {
	secret := generateTestSecret()
	staleDuration := -time.Hour // Negative duration to create already-stale token

	tokenStore, err := NewTrustedCookieTokenStore[SessionData, UserData](
		Secret(secret),
		staleDuration,
		WithCookieEncrypted(true),
	)
	if err != nil {
		t.Fatalf("failed to create trusted token store: %v", err)
	}

	session := SessionData{
		Id:        "session123",
		UserId:    "user@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	user := UserData{
		Email: "user@example.com",
		Name:  "Test User",
	}

	sessionToken := "session123.secret456"
	trustedToken, _ := tokenStore.ToTrustedToken(sessionToken, session, user)

	// Token should be immediately stale since staleDuration is negative
	_, _, err = tokenStore.FromTrustedToken(trustedToken)
	if err == nil {
		t.Error("expected StaleSessionError for stale token")
	}

	var staleErr *StaleSessionError
	if !errors.As(err, &staleErr) {
		t.Errorf("expected StaleSessionError, got %T: %v", err, err)
	}

	if staleErr.SessionToken != sessionToken {
		t.Errorf("expected session token %q in error, got %q", sessionToken, staleErr.SessionToken)
	}
}

func TestTrustedCookieTokenStore_ExpiredSession(t *testing.T) {
	secret := generateTestSecret()

	tokenStore, err := NewTrustedCookieTokenStore[SessionData, UserData](
		Secret(secret),
		time.Hour, // Long stale duration
		WithCookieEncrypted(true),
	)
	if err != nil {
		t.Fatalf("failed to create trusted token store: %v", err)
	}

	// Create session that's already expired
	session := SessionData{
		Id:        "session123",
		UserId:    "user@example.com",
		ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
	}
	user := UserData{
		Email: "user@example.com",
		Name:  "Test User",
	}

	sessionToken := "session123.secret456"
	trustedToken, _ := tokenStore.ToTrustedToken(sessionToken, session, user)

	// Should return ErrExpiredSession
	_, _, err = tokenStore.FromTrustedToken(trustedToken)
	if err != ErrExpiredSession {
		t.Errorf("expected ErrExpiredSession, got %v", err)
	}
}

func TestTrustedCookieTokenStore_SetGetToken(t *testing.T) {
	secret := generateTestSecret()

	tokenStore, err := NewTrustedCookieTokenStore[SessionData, UserData](
		Secret(secret),
		time.Minute*15,
		WithCookieEncrypted(true),
	)
	if err != nil {
		t.Fatalf("failed to create trusted token store: %v", err)
	}

	w := httptest.NewRecorder()
	expiresAt := time.Now().Add(time.Hour)

	// Set encrypted token
	err = tokenStore.SetToken(w, "encrypted_token_value", expiresAt)
	if err != nil {
		t.Fatalf("failed to set token: %v", err)
	}

	// Get token
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	retrievedToken, err := tokenStore.GetToken(r)
	if err != nil {
		t.Fatalf("failed to get token: %v", err)
	}

	if retrievedToken != "encrypted_token_value" {
		t.Errorf("expected token 'encrypted_token_value', got %q", retrievedToken)
	}
}

func TestTrustedCookieTokenStore_RequiresEncryption(t *testing.T) {
	secret := generateTestSecret()

	// Should fail without secret
	_, err := NewTrustedCookieTokenStore[SessionData, UserData](
		nil,
		time.Minute*15,
		WithCookieEncrypted(true),
	)
	if err == nil {
		t.Error("expected error for trusted cookie without secret")
	}

	// Should fail without encryption
	_, err = NewTrustedCookieTokenStore[SessionData, UserData](
		Secret(secret),
		time.Minute*15,
		WithCookieEncrypted(false),
	)
	if err == nil {
		t.Error("expected error for non-encrypted trusted cookie")
	}
}

func TestTrustedCookieTokenStore_DeleteToken(t *testing.T) {
	secret := generateTestSecret()

	tokenStore, err := NewTrustedCookieTokenStore[SessionData, UserData](
		Secret(secret),
		time.Minute*15,
		WithCookieEncrypted(true),
	)
	if err != nil {
		t.Fatalf("failed to create trusted token store: %v", err)
	}

	w := httptest.NewRecorder()
	tokenStore.DeleteToken(w)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no deletion cookie set")
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("expected MaxAge -1 for deletion, got %d", cookies[0].MaxAge)
	}
}
