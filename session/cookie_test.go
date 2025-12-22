package session

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func generateTestSecret() []byte {
	secret := make([]byte, 32)
	rand.Read(secret)
	return secret
}

func TestSecret(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")
	key3 := []byte("key3")

	// Single key
	secrets := Secret(key1)
	if len(secrets) != 1 {
		t.Errorf("expected 1 key, got %d", len(secrets))
	}

	// Multiple keys for rotation
	secrets = Secret(key1, key2, key3)
	if len(secrets) != 3 {
		t.Errorf("expected 3 keys, got %d", len(secrets))
	}
	if string(secrets[0]) != "key1" || string(secrets[1]) != "key2" || string(secrets[2]) != "key3" {
		t.Error("keys not in expected order")
	}
}

func TestDefaultCookieOptions(t *testing.T) {
	// Test with just name (unsigned by default - no secret)
	opts := DefaultCookieOptions("session_id")

	if opts.Name != "session_id" {
		t.Errorf("expected name 'session_id', got %q", opts.Name)
	}
	if len(opts.Secret) != 0 {
		t.Error("expected Secret to be empty (unsigned)")
	}
	if !opts.Secure {
		t.Error("expected Secure to be true")
	}
	if !opts.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if opts.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %v", opts.SameSite)
	}
	if opts.Path != "/" {
		t.Errorf("expected Path '/', got %q", opts.Path)
	}

	// Test with custom options
	opts = DefaultCookieOptions("custom_session",
		WithCookieSecure(false),
		WithCookiePath("/api"),
		WithCookieDomain("example.com"),
		WithCookieSameSite(http.SameSiteStrictMode),
	)

	if opts.Name != "custom_session" {
		t.Errorf("expected name 'custom_session', got %q", opts.Name)
	}
	if opts.Secure {
		t.Error("expected Secure to be false")
	}
	if opts.Path != "/api" {
		t.Errorf("expected Path '/api', got %q", opts.Path)
	}
	if opts.Domain != "example.com" {
		t.Errorf("expected Domain 'example.com', got %q", opts.Domain)
	}
	if opts.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite Strict, got %v", opts.SameSite)
	}
}

func TestNewCookie_UnsignedCookie(t *testing.T) {
	cookie, err := NewCookie[string](CookieOptions{
		Name:     "test_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	w := httptest.NewRecorder()
	testData := "hello world"

	// Set cookie
	err = cookie.Set(w, testData)
	if err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}

	// Read cookie back
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	retrievedData, err := cookie.Get(r)
	if err != nil {
		t.Fatalf("failed to get cookie: %v", err)
	}

	if retrievedData != testData {
		t.Errorf("expected %q, got %q", testData, retrievedData)
	}
}

func TestNewCookie_SignedCookie(t *testing.T) {
	secret := generateTestSecret()
	cookie, err := NewCookie[map[string]string](CookieOptions{
		Name:   "test_cookie",
		Secret: Secret(secret),
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	w := httptest.NewRecorder()
	testData := map[string]string{"message": "signed data"}

	// Set cookie
	err = cookie.Set(w, testData)
	if err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}

	// Read cookie back
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	retrievedData, err := cookie.Get(r)
	if err != nil {
		t.Fatalf("failed to get cookie: %v", err)
	}

	if retrievedData["message"] != testData["message"] {
		t.Errorf("expected %v, got %v", testData, retrievedData)
	}
}

func TestNewCookie_EncryptedCookie(t *testing.T) {
	secret := generateTestSecret()
	cookie, err := NewCookie[map[string]string](CookieOptions{
		Name:      "test_cookie",
		Secret:    Secret(secret),
		Encrypted: true,
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	w := httptest.NewRecorder()
	testData := map[string]string{"user": "john", "role": "admin"}

	// Set cookie
	err = cookie.Set(w, testData)
	if err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}

	// Read cookie back
	r := &http.Request{Header: http.Header{"Cookie": w.Header()["Set-Cookie"]}}
	retrievedData, err := cookie.Get(r)
	if err != nil {
		t.Fatalf("failed to get cookie: %v", err)
	}

	if retrievedData["user"] != "john" || retrievedData["role"] != "admin" {
		t.Errorf("expected %v, got %v", testData, retrievedData)
	}
}

func TestNewCookie_KeyRotation(t *testing.T) {
	oldSecret := generateTestSecret()
	newSecret := generateTestSecret()

	// Create cookie with old secret
	oldCookie, err := NewCookie[map[string]string](CookieOptions{
		Name:   "test_cookie",
		Secret: Secret(oldSecret),
	})
	if err != nil {
		t.Fatalf("failed to create old cookie: %v", err)
	}

	// Set cookie with old secret
	w := httptest.NewRecorder()
	err = oldCookie.Set(w, map[string]string{"data": "test"})
	if err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}

	// Create new cookie with key rotation (new key first, old key for verification)
	newCookie, err := NewCookie[map[string]string](CookieOptions{
		Name:   "test_cookie",
		Secret: Secret(newSecret, oldSecret),
	})
	if err != nil {
		t.Fatalf("failed to create new cookie: %v", err)
	}

	// Should be able to read cookie signed with old key
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	retrievedData, err := newCookie.Get(r)
	if err != nil {
		t.Fatalf("failed to get cookie with key rotation: %v", err)
	}

	if retrievedData["data"] != "test" {
		t.Errorf("expected 'test', got %q", retrievedData["data"])
	}
}

func TestNewCookie_InvalidSignature(t *testing.T) {
	secret1 := generateTestSecret()
	secret2 := generateTestSecret()

	// Create cookie with secret1
	cookie1, _ := NewCookie[map[string]string](CookieOptions{
		Name:   "test_cookie",
		Secret: Secret(secret1),
	})

	w := httptest.NewRecorder()
	cookie1.Set(w, map[string]string{"data": "test"})

	// Try to read with different secret
	cookie2, _ := NewCookie[map[string]string](CookieOptions{
		Name:   "test_cookie",
		Secret: Secret(secret2),
	})

	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	_, err := cookie2.Get(r)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestNewCookie_Delete(t *testing.T) {
	cookie, err := NewCookie[string](CookieOptions{
		Name:     "test_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	w := httptest.NewRecorder()

	// Set cookie
	cookie.Set(w, "test data")

	// Delete cookie
	cookie.Delete(w)

	// Check that deletion cookie was set
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies found")
	}

	deleteCookie := cookies[len(cookies)-1]
	if deleteCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1 for deletion, got %d", deleteCookie.MaxAge)
	}
}

func TestNewCookie_Duration(t *testing.T) {
	cookie, err := NewCookie[string](CookieOptions{
		Name:     "test_cookie",
		Duration: time.Hour * 24,
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	w := httptest.NewRecorder()
	cookie.Set(w, "test data")

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies found")
	}

	// Check that Expires is set
	if cookies[0].Expires.IsZero() {
		t.Error("expected Expires to be set for persistent cookie")
	}

	// Check that Expires is approximately 24 hours from now
	expectedExpiry := time.Now().Add(time.Hour * 24)
	if cookies[0].Expires.Before(expectedExpiry.Add(-time.Second)) ||
		cookies[0].Expires.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("expected Expires around %v, got %v", expectedExpiry, cookies[0].Expires)
	}
}

func TestNewCookie_CustomOptions(t *testing.T) {
	cookie, err := NewCookie[string](CookieOptions{
		Name:     "test_cookie",
		Path:     "/api",
		Domain:   "example.com",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	w := httptest.NewRecorder()
	cookie.Set(w, "test data")

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies found")
	}

	c := cookies[0]
	if c.Path != "/api" {
		t.Errorf("expected Path /api, got %s", c.Path)
	}
	if c.Domain != "example.com" {
		t.Errorf("expected Domain example.com, got %s", c.Domain)
	}
	if !c.Secure {
		t.Error("expected Secure to be true")
	}
	if !c.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if c.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite Strict, got %v", c.SameSite)
	}
}

func TestNewCookie_ComplexType(t *testing.T) {
	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	cookie, err := NewCookie[User](CookieOptions{
		Name:     "user_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	w := httptest.NewRecorder()
	testUser := User{ID: 123, Name: "John", Email: "john@example.com"}

	// Set cookie
	err = cookie.Set(w, testUser)
	if err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}

	// Read cookie back
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	retrievedUser, err := cookie.Get(r)
	if err != nil {
		t.Fatalf("failed to get cookie: %v", err)
	}

	if retrievedUser != testUser {
		t.Errorf("expected %+v, got %+v", testUser, retrievedUser)
	}
}

func TestNewCookie_MissingCookie(t *testing.T) {
	cookie, err := NewCookie[string](CookieOptions{
		Name:     "test_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	_, err = cookie.Get(r)
	if err == nil {
		t.Error("expected error when cookie not found")
	}
	if err != http.ErrNoCookie {
		t.Errorf("expected http.ErrNoCookie, got %v", err)
	}
}

func TestValidateSecret(t *testing.T) {
	tests := []struct {
		name      string
		secret    [][]byte
		encrypted bool
		wantErr   bool
	}{
		{
			name:      "empty secret",
			secret:    [][]byte{},
			encrypted: false,
			wantErr:   true,
		},
		{
			name:      "secret too short",
			secret:    Secret([]byte("short")),
			encrypted: false,
			wantErr:   true,
		},
		{
			name:      "valid secret for signing",
			secret:    Secret(generateTestSecret()),
			encrypted: false,
			wantErr:   false,
		},
		{
			name:      "valid secret for encryption",
			secret:    Secret(generateTestSecret()),
			encrypted: true,
			wantErr:   false,
		},
		{
			name:      "secret wrong size for encryption",
			secret:    Secret(make([]byte, 64)),
			encrypted: true,
			wantErr:   true,
		},
		{
			name:      "secret valid size for signing",
			secret:    Secret(make([]byte, 64)),
			encrypted: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecret(tt.secret, tt.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCookie_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		options CookieOptions
		wantErr bool
	}{
		{
			name:    "missing name",
			options: CookieOptions{},
			wantErr: true,
		},
		{
			name:    "unsigned",
			options: CookieOptions{Name: "test"},
			wantErr: false,
		},
		{
			name:    "encrypted without secret",
			options: CookieOptions{Name: "test", Encrypted: true},
			wantErr: true,
		},
		{
			name: "valid unsigned",
			options: CookieOptions{
				Name:     "test",
			},
			wantErr: false,
		},
		{
			name: "valid signed",
			options: CookieOptions{
				Name:   "test",
				Secret: Secret(generateTestSecret()),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCookie[string](tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCookie() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRawCookie_Name(t *testing.T) {
	cookie, err := NewRawCookie[string](CookieOptions{
		Name:     "test_raw_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	if cookie.Name() != "test_raw_cookie" {
		t.Errorf("expected name 'test_raw_cookie', got %q", cookie.Name())
	}
}

func TestNewRawCookie_DataAndParse_Unsigned(t *testing.T) {
	type Data struct {
		Key string `json:"key"`
		Foo string `json:"foo"`
	}

	cookie, err := NewRawCookie[Data](CookieOptions{
		Name:     "test_raw_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	testData := Data{Key: "value", Foo: "bar"}

	// Encode data
	value, err := cookie.Data(testData)
	if err != nil {
		t.Fatalf("failed to encode data: %v", err)
	}

	// Parse value
	parsed, err := cookie.Parse(value)
	if err != nil {
		t.Fatalf("failed to parse value: %v", err)
	}

	if parsed != testData {
		t.Errorf("expected %v, got %v", testData, parsed)
	}
}

func TestNewRawCookie_DataAndParse_Signed(t *testing.T) {
	secret := generateTestSecret()
	cookie, err := NewRawCookie[string](CookieOptions{
		Name:   "test_raw_cookie",
		Secret: Secret(secret),
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	testData := "signed data"

	// Encode data
	value, err := cookie.Data(testData)
	if err != nil {
		t.Fatalf("failed to encode data: %v", err)
	}

	// Parse value
	parsed, err := cookie.Parse(value)
	if err != nil {
		t.Fatalf("failed to parse value: %v", err)
	}

	if parsed != testData {
		t.Errorf("expected %q, got %q", testData, parsed)
	}
}

func TestNewRawCookie_DataAndParse_Encrypted(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Admin bool   `json:"admin"`
	}

	secret := generateTestSecret()
	cookie, err := NewRawCookie[User](CookieOptions{
		Name:      "test_raw_cookie",
		Secret:    Secret(secret),
		Encrypted: true,
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	testData := User{Name: "alice", Admin: true}

	// Encode data
	value, err := cookie.Data(testData)
	if err != nil {
		t.Fatalf("failed to encode data: %v", err)
	}

	// Parse value
	parsed, err := cookie.Parse(value)
	if err != nil {
		t.Fatalf("failed to parse value: %v", err)
	}

	if parsed != testData {
		t.Errorf("expected %v, got %v", testData, parsed)
	}
}

func TestNewRawCookie_Parse_InvalidBase64(t *testing.T) {
	cookie, err := NewRawCookie[string](CookieOptions{
		Name:     "test_raw_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	_, err = cookie.Parse("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestNewRawCookie_Parse_InvalidSignature(t *testing.T) {
	secret1 := generateTestSecret()
	secret2 := generateTestSecret()

	// Create cookie with secret1
	cookie1, _ := NewRawCookie[string](CookieOptions{
		Name:   "test_raw_cookie",
		Secret: Secret(secret1),
	})

	value, _ := cookie1.Data("test data")

	// Try to parse with different secret
	cookie2, _ := NewRawCookie[string](CookieOptions{
		Name:   "test_raw_cookie",
		Secret: Secret(secret2),
	})

	_, err := cookie2.Parse(value)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestNewRawCookie_SetCookie(t *testing.T) {
	cookie, err := NewRawCookie[string](CookieOptions{
		Name:     "test_raw_cookie",
		Path:     "/api",
		Domain:   "example.com",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Duration: time.Hour * 2,
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	value, _ := cookie.Data("test value")
	httpCookie := cookie.SetCookie(value)

	if httpCookie.Name != "test_raw_cookie" {
		t.Errorf("expected name 'test_raw_cookie', got %q", httpCookie.Name)
	}
	if httpCookie.Value != value {
		t.Errorf("expected value %q, got %q", value, httpCookie.Value)
	}
	if httpCookie.Path != "/api" {
		t.Errorf("expected path '/api', got %q", httpCookie.Path)
	}
	if httpCookie.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", httpCookie.Domain)
	}
	if !httpCookie.Secure {
		t.Error("expected Secure to be true")
	}
	if !httpCookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if httpCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %v", httpCookie.SameSite)
	}
	if httpCookie.Expires.IsZero() {
		t.Error("expected Expires to be set")
	}

	// Check that Expires is approximately 2 hours from now
	expectedExpiry := time.Now().Add(time.Hour * 2)
	if httpCookie.Expires.Before(expectedExpiry.Add(-time.Second)) ||
		httpCookie.Expires.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("expected Expires around %v, got %v", expectedExpiry, httpCookie.Expires)
	}
}

func TestNewRawCookie_SetCookie_WithOptions(t *testing.T) {
	cookie, err := NewRawCookie[string](CookieOptions{
		Name:     "test_raw_cookie",
		Path:     "/",
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	value, _ := cookie.Data("test value")

	// Override path with option
	httpCookie := cookie.SetCookie(value, func(c *http.Cookie) {
		c.Path = "/custom"
		c.Secure = true
	})

	if httpCookie.Path != "/custom" {
		t.Errorf("expected path '/custom', got %q", httpCookie.Path)
	}
	if !httpCookie.Secure {
		t.Error("expected Secure to be true from option")
	}
}

func TestNewRawCookie_RemoveCookie(t *testing.T) {
	cookie, err := NewRawCookie[string](CookieOptions{
		Name:     "test_raw_cookie",
		Path:     "/api",
		Domain:   "example.com",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	removeCookie := cookie.RemoveCookie()

	if removeCookie.Name != "test_raw_cookie" {
		t.Errorf("expected name 'test_raw_cookie', got %q", removeCookie.Name)
	}
	if removeCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", removeCookie.MaxAge)
	}
	if removeCookie.Path != "/api" {
		t.Errorf("expected path '/api', got %q", removeCookie.Path)
	}
	if removeCookie.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", removeCookie.Domain)
	}
	if !removeCookie.Secure {
		t.Error("expected Secure to be true")
	}
	if !removeCookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if removeCookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite Strict, got %v", removeCookie.SameSite)
	}
}

func TestNewRawCookie_KeyRotation(t *testing.T) {
	oldSecret := generateTestSecret()
	newSecret := generateTestSecret()

	// Create cookie with old secret
	oldCookie, err := NewRawCookie[string](CookieOptions{
		Name:   "test_raw_cookie",
		Secret: Secret(oldSecret),
	})
	if err != nil {
		t.Fatalf("failed to create old raw cookie: %v", err)
	}

	// Encode with old secret
	value, err := oldCookie.Data("rotated data")
	if err != nil {
		t.Fatalf("failed to encode data: %v", err)
	}

	// Create new cookie with key rotation
	newCookie, err := NewRawCookie[string](CookieOptions{
		Name:   "test_raw_cookie",
		Secret: Secret(newSecret, oldSecret),
	})
	if err != nil {
		t.Fatalf("failed to create new raw cookie: %v", err)
	}

	// Should be able to parse value signed with old key
	parsed, err := newCookie.Parse(value)
	if err != nil {
		t.Fatalf("failed to parse with key rotation: %v", err)
	}

	if parsed != "rotated data" {
		t.Errorf("expected 'rotated data', got %q", parsed)
	}
}

func TestFromRawCookie(t *testing.T) {
	rawCookie, err := NewRawCookie[string](CookieOptions{
		Name:     "test_cookie",
	})
	if err != nil {
		t.Fatalf("failed to create raw cookie: %v", err)
	}

	// Convert to regular cookie
	cookie := FromRawCookie(rawCookie)

	w := httptest.NewRecorder()
	testData := "from raw cookie"

	// Test Set
	err = cookie.Set(w, testData)
	if err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}

	// Test Get
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	retrievedData, err := cookie.Get(r)
	if err != nil {
		t.Fatalf("failed to get cookie: %v", err)
	}

	if retrievedData != testData {
		t.Errorf("expected %q, got %q", testData, retrievedData)
	}

	// Test Delete
	w2 := httptest.NewRecorder()
	cookie.Delete(w2)
	cookies := w2.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies found after delete")
	}
	if cookies[0].MaxAge != -1 {
		t.Errorf("expected MaxAge -1 for deletion, got %d", cookies[0].MaxAge)
	}
}
