package session

import (
	"bytes"
	"encoding/hex"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type UserData struct {
	Email          string `json:"email"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	HashedPassword string `json:"-"`
}

type SessionData struct {
	Id         string    `json:"id"`
	SecretHash string    `json:"-"`
	UserId     string    `json:"userId"`
	Role       string    `json:"role"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

func (s SessionData) GetSession() (string, []byte, time.Time) {
	secretHash, _ := hex.DecodeString(s.SecretHash)
	return s.Id, secretHash, s.ExpiresAt
}

// Mock session store for testing
type mockSessionStore struct {
	sessions map[string]SessionData
	users    map[string]UserData
}

func newMockSessionStore() *mockSessionStore {
	return &mockSessionStore{
		sessions: make(map[string]SessionData),
		users:    make(map[string]UserData),
	}
}

func (m *mockSessionStore) GetSessionAndUser(sessionId string) (SessionData, UserData, error) {
	session, ok := m.sessions[sessionId]
	if !ok {
		return SessionData{}, UserData{}, errors.New("session not found")
	}

	user, ok := m.users[session.UserId]
	if !ok {
		return SessionData{}, UserData{}, errors.New("user not found")
	}

	return session, user, nil
}

func (m *mockSessionStore) CreateSession(sessionData Session, user UserData) (SessionData, error) {
	sessionId, sessionSecretHash, expiresAt := sessionData.GetSession()

	session := SessionData{
		Id:         sessionId,
		SecretHash: hex.EncodeToString(sessionSecretHash),
		UserId:     user.Email,
		Role:       user.Role,
		ExpiresAt:  expiresAt,
	}
	m.sessions[sessionId] = session
	return session, nil
}

func (m *mockSessionStore) UpdateSession(sessionId string, expiresAt time.Time) (SessionData, error) {
	session, ok := m.sessions[sessionId]
	if !ok {
		return SessionData{}, errors.New("session not found")
	}

	session.ExpiresAt = expiresAt
	m.sessions[sessionId] = session
	return session, nil
}

func (m *mockSessionStore) DeleteSession(sessionId string) error {
	delete(m.sessions, sessionId)
	return nil
}

func (m *mockSessionStore) DeleteUserSessions(user UserData) error {
	for id, session := range m.sessions {
		if session.UserId == user.Email {
			delete(m.sessions, id)
		}
	}
	return nil
}

func TestAuth_CreateSession(t *testing.T) {
	sessionStore := newMockSessionStore()
	duration := time.Hour * 24 * 7 // 7 days
	auth, _ := NewAuth(sessionStore, WithDuration(duration))

	user := UserData{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "user",
	}

	sessionToken, session, err := auth.CreateSession(user)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Session token should be in format: sessionId.sessionSecret
	sessionId, sessionSecretHash, expiresAt := session.GetSession()

	if sessionId == "" {
		t.Error("expected session ID to be non-empty")
	}
	if len(sessionSecretHash) == 0 {
		t.Error("expected session secret hash to be non-empty")
	}

	// Parse session token
	parts := strings.Split(sessionToken, ".")
	if len(parts) != 2 {
		t.Fatalf("expected session token format 'sessionId.sessionSecret', got %q", sessionToken)
	}
	tokenSessionId := parts[0]
	tokenSessionSecret := parts[1]

	// Session token should contain correct session ID
	if tokenSessionId != sessionId {
		t.Errorf("expected session token to contain session ID %q, got %q", sessionId, tokenSessionId)
	}

	// Session secret in token should hash to stored hash
	computedHash := hashSessionSecret(tokenSessionSecret)
	if !bytes.Equal(computedHash, sessionSecretHash) {
		t.Error("session secret hash mismatch: hashing the secret from token should equal stored hash")
	}

	// Session should exist in database
	stored, ok := sessionStore.sessions[sessionId]
	if !ok {
		t.Fatal("session not found in store")
	}

	// Verify stored session has the expected properties from Session interface
	storedSessionId, storedSecretHash, storedExpiresAt := stored.GetSession()

	if storedSessionId != sessionId {
		t.Errorf("expected stored session ID %q, got %q", sessionId, storedSessionId)
	}

	if !bytes.Equal(storedSecretHash, sessionSecretHash) {
		t.Error("stored secret hash does not match session secret hash")
	}

	// Verify stored expiration matches
	if !storedExpiresAt.Equal(expiresAt) {
		t.Errorf("expected stored expiration %v, got %v", expiresAt, storedExpiresAt)
	}

	// ExpiresAt should match configured duration
	expectedExpiry := time.Now().Add(duration)
	if expiresAt.Before(expectedExpiry.Add(-time.Second)) || expiresAt.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("expected expiration around %v (now + %v), got %v", expectedExpiry, duration, expiresAt)
	}
}

func TestAuth_InvalidateSession(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test"}
	_, session, _ := auth.CreateSession(user)

	sessionId, _, _ := session.GetSession()

	// Verify session exists
	if _, ok := sessionStore.sessions[sessionId]; !ok {
		t.Fatal("session should exist before invalidation")
	}

	// Invalidate session
	err := auth.InvalidateSession(sessionId)
	if err != nil {
		t.Fatalf("failed to invalidate session: %v", err)
	}

	// Verify session is deleted
	if _, ok := sessionStore.sessions[sessionId]; ok {
		t.Error("session should be deleted after invalidation")
	}
}

func TestAuth_InvalidateAllSessions(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test"}

	// Create multiple sessions for the user
	auth.CreateSession(user)
	auth.CreateSession(user)
	auth.CreateSession(user)

	// Verify sessions exist
	count := 0
	for _, session := range sessionStore.sessions {
		if session.UserId == user.Email {
			count++
		}
	}
	if count != 3 {
		t.Fatalf("expected 3 sessions, got %d", count)
	}

	// Invalidate all sessions
	err := auth.InvalidateAllSessions(user)
	if err != nil {
		t.Fatalf("failed to invalidate all sessions: %v", err)
	}

	// Verify all sessions are deleted
	for _, session := range sessionStore.sessions {
		if session.UserId == user.Email {
			t.Error("found session after invalidating all sessions")
		}
	}
}

func TestAuth_ValidateSessionToken(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test"}
	sessionStore.users[user.Email] = user // Add user to store
	sessionToken, createdSession, _ := auth.CreateSession(user)

	// Validate the session token
	session, retrievedUser, optimistic, err := auth.ValidateSessionToken(sessionToken)
	if err != nil {
		t.Fatalf("failed to validate session token: %v", err)
	}

	if optimistic {
		t.Error("expected optimistic=false for regular token store")
	}

	sessionId, _, _ := session.GetSession()
	createdSessionId, _, _ := createdSession.GetSession()

	if sessionId != createdSessionId {
		t.Errorf("expected session ID %q, got %q", createdSessionId, sessionId)
	}

	if retrievedUser != user {
		t.Errorf("expected user %+v, got %+v", user, retrievedUser)
	}
}

func TestAuth_ValidateSessionToken_ExpiredSession(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test"}
	sessionStore.users[user.Email] = user
	sessionToken, session, _ := auth.CreateSession(user)

	// Manually expire the session by updating expiresAt to past
	sessionId, _, _ := session.GetSession()
	storedSession := sessionStore.sessions[sessionId]
	storedSession.ExpiresAt = time.Now().Add(-time.Hour) // Expired 1 hour ago
	sessionStore.sessions[sessionId] = storedSession

	// Validate expired session token
	_, _, _, err := auth.ValidateSessionToken(sessionToken)
	if err != ErrExpiredSession {
		t.Errorf("expected ErrExpiredSession, got %v", err)
	}

	// Verify session was deleted from store
	if _, ok := sessionStore.sessions[sessionId]; ok {
		t.Error("expected expired session to be deleted from store")
	}
}

func TestAuth_ValidateSessionToken_InvalidSecret(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test"}
	sessionStore.users[user.Email] = user
	_, session, _ := auth.CreateSession(user)

	sessionId, _, _ := session.GetSession()

	// Create invalid session token with wrong secret
	invalidToken := sessionId + ".wrongsecret123"

	// Validate with invalid secret
	_, _, _, err := auth.ValidateSessionToken(invalidToken)
	if err != ErrInvalidSessionSecret {
		t.Errorf("expected ErrInvalidSessionSecret, got %v", err)
	}
}

func TestAuth_LoginRequest(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test"}

	w := httptest.NewRecorder()
	session, err := auth.LoginRequest(w, user)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	sessionId, _, _ := session.GetSession()
	if sessionId == "" {
		t.Error("expected session ID")
	}

	// Check that cookie was set
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set")
	}

	if cookies[0].Value == "" {
		t.Error("expected cookie value")
	}
}

func TestAuth_AuthenticateRequest(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test", Role: "admin"}
	sessionStore.users[user.Email] = user

	// Login to get session cookie
	w := httptest.NewRecorder()
	auth.LoginRequest(w, user)

	// Create request with session cookie
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	// Authenticate request
	w2 := httptest.NewRecorder()
	session, authenticatedUser, err := auth.AuthenticateRequest(w2, r)
	if err != nil {
		t.Fatalf("failed to authenticate request: %v", err)
	}

	sessionId, _, _ := session.GetSession()
	if sessionId == "" {
		t.Error("expected session ID")
	}

	if authenticatedUser != user {
		t.Errorf("expected user %+v, got %+v", user, authenticatedUser)
	}
}

func TestAuth_AuthenticateRequest_NoSession(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	// Create request without session cookie
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Authenticate request without session
	_, _, err := auth.AuthenticateRequest(w, r)
	if err == nil {
		t.Error("expected error for request without session")
	}
}

func TestAuth_LogoutRequest(t *testing.T) {
	sessionStore := newMockSessionStore()
	auth, _ := NewAuth(sessionStore)

	user := UserData{Email: "test@example.com", Name: "Test"}
	sessionStore.users[user.Email] = user

	// Login to get session
	w := httptest.NewRecorder()
	createdSession, _ := auth.LoginRequest(w, user)

	// Create request with session cookie
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	// Logout
	w2 := httptest.NewRecorder()
	session, loggedOutUser, err := auth.LogoutRequest(w2, r)
	if err != nil {
		t.Fatalf("failed to logout: %v", err)
	}

	if loggedOutUser != user {
		t.Errorf("expected user %+v, got %+v", user, loggedOutUser)
	}

	// Verify session was deleted from store
	sessionId, _, _ := session.GetSession()
	createdSessionId, _, _ := createdSession.GetSession()

	if sessionId != createdSessionId {
		t.Error("session IDs should match")
	}

	if _, ok := sessionStore.sessions[sessionId]; ok {
		t.Error("session should be deleted after logout")
	}

	// Verify deletion cookie was set
	cookies := w2.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set")
	}

	if cookies[0].MaxAge != -1 {
		t.Error("expected deletion cookie with MaxAge -1")
	}
}

func TestAuth_SessionRefresh(t *testing.T) {
	sessionStore := newMockSessionStore()

	// Create auth with short duration and refresh threshold
	auth, _ := NewAuth(sessionStore,
		WithDuration(time.Hour),
		// Refresh when 50% of duration has passed (default half-time threshold)
	)

	user := UserData{Email: "test@example.com", Name: "Test"}
	sessionStore.users[user.Email] = user
	sessionToken, createdSession, _ := auth.CreateSession(user)

	// Manually set session to be past refresh threshold
	sessionId, _, _ := createdSession.GetSession()
	storedSession := sessionStore.sessions[sessionId]
	oldExpiresAt := time.Now().Add(time.Minute * 20) // Less than half of 1 hour remaining
	storedSession.ExpiresAt = oldExpiresAt
	sessionStore.sessions[sessionId] = storedSession

	// Validate session token - should trigger refresh
	session, _, _, err := auth.ValidateSessionToken(sessionToken)
	if err != nil {
		t.Fatalf("failed to validate session: %v", err)
	}

	_, _, newExpiresAt := session.GetSession()

	// New expiration should be later than old expiration
	if !newExpiresAt.After(oldExpiresAt) {
		t.Errorf("expected session to be refreshed, old=%v, new=%v", oldExpiresAt, newExpiresAt)
	}
}

func TestHalfTimeRefreshThreshold(t *testing.T) {
	duration := time.Hour * 24 // 24 hours

	tests := []struct {
		name          string
		expiresAt     time.Time
		shouldRefresh bool
	}{
		{
			name:          "fresh session - no refresh needed",
			expiresAt:     time.Now().Add(duration - time.Hour), // 23 hours remaining
			shouldRefresh: false,
		},
		{
			name:          "at half-time - refresh needed",
			expiresAt:     time.Now().Add(duration / 2), // Exactly at half
			shouldRefresh: true,
		},
		{
			name:          "past half-time - refresh needed",
			expiresAt:     time.Now().Add(duration/2 - time.Hour), // 11 hours remaining (less than half)
			shouldRefresh: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newExpiresAt, shouldRefresh := halfTimeRefreshThreshold(tt.expiresAt, duration)

			if shouldRefresh != tt.shouldRefresh {
				t.Errorf("expected shouldRefresh=%v, got %v", tt.shouldRefresh, shouldRefresh)
			}

			if shouldRefresh && newExpiresAt.IsZero() {
				t.Error("expected new expiration time when refresh is needed")
			}
		})
	}
}
