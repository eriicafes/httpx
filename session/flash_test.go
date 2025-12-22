package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewFlashCookie(t *testing.T) {
	flash, err := NewFlashCookie[string](CookieOptions{
		Name:     "flash_message",
	})
	if err != nil {
		t.Fatalf("failed to create flash cookie: %v", err)
	}

	w := httptest.NewRecorder()
	testMessage := "Account created successfully!"

	// Set flash message
	err = flash.Set(w, testMessage)
	if err != nil {
		t.Fatalf("failed to set flash message: %v", err)
	}

	// Read flash message (should delete it)
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	// Create new recorder to capture deletion
	w2 := httptest.NewRecorder()
	retrievedMessage, err := flash.Get(w2, r)
	if err != nil {
		t.Fatalf("failed to get flash message: %v", err)
	}

	if retrievedMessage != testMessage {
		t.Errorf("expected %q, got %q", testMessage, retrievedMessage)
	}

	// Check that deletion cookie was set
	cookies := w2.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected deletion cookie to be set")
	}
	if cookies[0].MaxAge != -1 {
		t.Errorf("expected MaxAge -1 for deletion, got %d", cookies[0].MaxAge)
	}
}

func TestFlashCookie_Peek(t *testing.T) {
	flash, err := NewFlashCookie[string](CookieOptions{
		Name:     "flash_message",
	})
	if err != nil {
		t.Fatalf("failed to create flash cookie: %v", err)
	}

	w := httptest.NewRecorder()
	testMessage := "Peeked message"

	// Set flash message
	flash.Set(w, testMessage)

	// Peek at flash message (should NOT delete it)
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	peekedMessage, err := flash.Peek(r)
	if err != nil {
		t.Fatalf("failed to peek flash message: %v", err)
	}

	if peekedMessage != testMessage {
		t.Errorf("expected %q, got %q", testMessage, peekedMessage)
	}

	// Peek again - should still work
	peeked2, err := flash.Peek(r)
	if err != nil {
		t.Fatalf("failed to peek flash message second time: %v", err)
	}

	if peeked2 != testMessage {
		t.Errorf("expected %q, got %q", testMessage, peeked2)
	}
}

func TestFlashCookie_GetAfterPeek(t *testing.T) {
	flash, err := NewFlashCookie[string](CookieOptions{
		Name:     "flash_message",
	})
	if err != nil {
		t.Fatalf("failed to create flash cookie: %v", err)
	}

	w := httptest.NewRecorder()
	testMessage := "Message to peek then get"

	// Set flash message
	flash.Set(w, testMessage)

	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	// Peek first
	peekedMessage, _ := flash.Peek(r)
	if peekedMessage != testMessage {
		t.Errorf("peek: expected %q, got %q", testMessage, peekedMessage)
	}

	// Then Get (should delete)
	w2 := httptest.NewRecorder()
	gotMessage, err := flash.Get(w2, r)
	if err != nil {
		t.Fatalf("failed to get flash message: %v", err)
	}

	if gotMessage != testMessage {
		t.Errorf("get: expected %q, got %q", testMessage, gotMessage)
	}

	// Check deletion cookie was set
	cookies := w2.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected deletion cookie")
	}
	if cookies[0].MaxAge != -1 {
		t.Error("expected deletion cookie with MaxAge -1")
	}
}

func TestFlashCookie_SessionCookie(t *testing.T) {
	// Flash cookies should always be session cookies (Duration=0)
	flash, err := NewFlashCookie[string](CookieOptions{
		Name:     "flash_message",
		Duration: 3600, // This should be ignored
	})
	if err != nil {
		t.Fatalf("failed to create flash cookie: %v", err)
	}

	w := httptest.NewRecorder()
	flash.Set(w, "test")

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set")
	}

	// Session cookies have zero Expires time
	if !cookies[0].Expires.IsZero() {
		t.Error("expected session cookie (zero Expires), got persistent cookie")
	}
}

func TestFlashCookie_MissingMessage(t *testing.T) {
	flash, err := NewFlashCookie[string](CookieOptions{
		Name:     "flash_message",
	})
	if err != nil {
		t.Fatalf("failed to create flash cookie: %v", err)
	}

	// Try to get flash message when none exists
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	_, err = flash.Get(w, r)
	if err == nil {
		t.Error("expected error when flash message not found")
	}
	if err != http.ErrNoCookie {
		t.Errorf("expected http.ErrNoCookie, got %v", err)
	}
}

func TestFlashCookie_ComplexType(t *testing.T) {
	type Notification struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	flash, err := NewFlashCookie[Notification](CookieOptions{
		Name:     "flash_notification",
	})
	if err != nil {
		t.Fatalf("failed to create flash cookie: %v", err)
	}

	w := httptest.NewRecorder()
	notification := Notification{
		Type:    "success",
		Message: "Operation completed",
	}

	// Set flash notification
	flash.Set(w, notification)

	// Get flash notification
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}

	w2 := httptest.NewRecorder()
	retrieved, err := flash.Get(w2, r)
	if err != nil {
		t.Fatalf("failed to get flash notification: %v", err)
	}

	if retrieved != notification {
		t.Errorf("expected %+v, got %+v", notification, retrieved)
	}
}
