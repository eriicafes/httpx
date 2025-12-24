package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Secret creates a slice of keys for signing and encrypting cookies.
// The first key is used for signing/encrypting new cookies.
// All keys are tried when verifying/decrypting existing cookies (for key rotation support).
//
// Example usage:
//
//	Secret(key1)                    // Single key
//	Secret(key1, key2, key3)        // Key rotation (key1 is primary)
func Secret(b ...[]byte) [][]byte {
	return b
}

// Cookie provides high-level cookie operations for HTTP request/response handling.
// This is the primary interface for typical cookie usage - reading from requests
// and writing to responses directly.
type Cookie[T any] interface {
	// Get reads and parses a cookie value from an HTTP request.
	// Returns an error if the cookie is not found, malformed, or has an invalid signature.
	Get(r *http.Request) (T, error)

	// Set creates a cookie with the given data and writes it to the HTTP response.
	// Optional cookie modifier functions can be passed to override defaults.
	Set(w http.ResponseWriter, data T, options ...func(*http.Cookie)) error

	// Delete removes the cookie from the HTTP response by setting MaxAge to -1.
	Delete(w http.ResponseWriter)
}

// RawCookie provides low-level cookie operations for working with cookie values.
// Use this interface when you need manual control over cookie creation and parsing.
//
// Example usage:
//
//	cookie, _ := NewRawCookie[UserSession](CookieOptions{...})
//
//	// Get cookie name
//	name := cookie.Name()
//
//	// Encode data to cookie value
//	sessionData := SessionData{...}
//	value, _ := cookie.Data(sessionData)
//
//	// Parse raw cookie value
//	parsed, _ := cookie.Parse(value)
//
//	// Create cookie object with value
//	httpCookie := cookie.SetCookie(value)
//
//	// Create removal cookie
//	removeCookie := cookie.RemoveCookie()
type RawCookie[T any] interface {
	// Name returns the configured cookie name.
	Name() string

	// Data encodes cookie data into a raw cookie value string.
	Data(data T) (string, error)

	// Parse decodes and validates a raw cookie value string.
	Parse(value string) (T, error)

	// SetCookie returns an http.Cookie with the given value.
	SetCookie(value string, options ...func(*http.Cookie)) *http.Cookie

	// RemoveCookie creates an http.Cookie configured to delete itself by setting MaxAge to -1.
	RemoveCookie() *http.Cookie
}

// NewCookie creates a new cookie instance that provides high-level cookie operations
// for HTTP request/response handling.
//
// The generic type T is JSON-marshaled and optionally signed (HMAC-SHA256) and/or encrypted (AES-256-GCM).
//
// Options:
//   - Name: Cookie name (required)
//   - Duration: If > 0, persistent cookie. Default: 0 (session cookie, expires when browser closes).
//   - Secure: Only send over HTTPS. Default: false
//   - HttpOnly: Block JavaScript access. Default: false
//   - SameSite: Lax/Strict/None mode. Default: 0 (browser default, typically Lax)
//
// Cryptographic security:
//   - Signing: Prevents tampering. Default: enabled. Disable with Unsigned=true.
//   - Encryption: Hides data from clients. Default: disabled. Enable with Encrypted=true.
//   - Secret: Required if signed or encrypted. Use Secret(key) for a single key or Secret(key1, key2, ...) for key rotation.
//
// Example usage:
//
//	// Using DefaultCookieOptions for secure defaults
//	opts := DefaultCookieOptions("user_session",
//	    WithCookieSecret(Secret(key)),
//	    WithCookieSecure(false), // For development
//	)
//	cookie, _ := NewCookie[UserData](opts)
//
//	// Or using CookieOptions directly
//	cookie, _ := NewCookie[UserData](CookieOptions{
//	    Name:   "user_session",
//	    Secret: Secret(key),
//	})
func NewCookie[T any](options CookieOptions) (Cookie[T], error) {
	raw, err := NewRawCookie[T](options)
	if err != nil {
		return nil, err
	}
	return &cookie[T]{raw}, nil
}

func FromRawCookie[T any](raw RawCookie[T]) Cookie[T] {
	return &cookie[T]{raw}
}

type cookie[T any] struct{ raw RawCookie[T] }

// Get reads a cookie value.
// Returns an error if the cookie is not found, malformed, or has an invalid signature.
func (c *cookie[T]) Get(r *http.Request) (T, error) {
	var data T
	cookieData, err := r.Cookie(c.raw.Name())
	if err != nil {
		return data, err
	}
	return c.raw.Parse(cookieData.Value)
}

// Set writes a cookie value.
// Options can be used to override cookie defaults except from name and value.
func (c *cookie[T]) Set(w http.ResponseWriter, data T, options ...func(*http.Cookie)) error {
	value, err := c.raw.Data(data)
	if err != nil {
		return err
	}
	cookie := c.raw.SetCookie(value, options...)
	http.SetCookie(w, cookie)
	return nil
}

// Delete deletes the cookie.
func (c *cookie[T]) Delete(w http.ResponseWriter) {
	http.SetCookie(w, c.raw.RemoveCookie())
}

// NewCooNewRawCookiekie creates a new raw cookie instance that provides low-level cookie operations
// for working with cookie values.
//
// The generic type T is JSON-marshaled and optionally signed (HMAC-SHA256) and/or encrypted (AES-256-GCM).
//
// Options:
//   - Name: Cookie name (required)
//   - Duration: If > 0, persistent cookie. Default: 0 (session cookie, expires when browser closes).
//   - Secure: Only send over HTTPS. Default: false
//   - HttpOnly: Block JavaScript access. Default: false
//   - SameSite: Lax/Strict/None mode. Default: 0 (browser default, typically Lax)
//
// Cryptographic security:
//   - Signing: Prevents tampering. Default: enabled. Disable with Unsigned=true.
//   - Encryption: Hides data from clients. Default: disabled. Enable with Encrypted=true.
//   - Secret: Required if signed or encrypted. Use Secret(key) for a single key or Secret(key1, key2, ...) for key rotation.
func NewRawCookie[T any](options CookieOptions) (RawCookie[T], error) {
	// Validate secret if signing or encryption is requested
	if len(options.Secret) > 0 || options.Encrypted {
		if err := validateSecret(options.Secret, options.Encrypted); err != nil {
			return nil, err
		}
	}
	if options.Name == "" {
		return nil, errors.New("cookie name is required")
	}
	return &rawCookie[T]{options}, nil
}

type rawCookie[T any] struct{ options CookieOptions }

func (c *rawCookie[T]) Name() string {
	return c.options.Name
}

func (c *rawCookie[T]) Data(data T) (string, error) {
	// marshal json
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	var cookieData []byte
	if c.options.Encrypted {
		// encrypted cookie - encrypt with primary key
		// Format: [12 bytes nonce][remaining bytes encrypted data]
		ciphertext, nonce, err := encrypt(c.options.Secret[0], jsonData)
		if err != nil {
			return "", err
		}
		cookieData = append(nonce, ciphertext...)
	} else if len(c.options.Secret) > 0 {
		// signed cookie - sign with primary key
		// Format: [32 bytes HMAC signature][remaining bytes JSON data]
		signature := sign(c.options.Secret[0], jsonData)
		cookieData = append(signature, jsonData...)
	} else {
		// unsigned cookie - data is just the JSON
		cookieData = jsonData
	}

	return base64.URLEncoding.EncodeToString(cookieData), nil
}

func (c *rawCookie[T]) Parse(value string) (T, error) {
	var data T

	// decode base64
	decoded, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return data, errors.New("malformed cookie: invalid base64")
	}

	var jsonData []byte
	if c.options.Encrypted {
		// encrypted cookie - decrypt with key rotation support
		// Format: [12 bytes nonce][remaining bytes encrypted data]
		if len(decoded) < 12 {
			return data, errors.New("malformed cookie: data too short")
		}
		nonce, ciphertext := decoded[:12], decoded[12:]

		// Try all keys
		for _, key := range c.options.Secret {
			if jsonData, err = decrypt(key, nonce, ciphertext); err == nil {
				break
			}
		}
		if err != nil {
			return data, errors.New("malformed cookie: decryption failed")
		}
	} else if len(c.options.Secret) > 0 {
		// signed cookie - verify signature with key rotation support
		// Format: [32 bytes HMAC signature][remaining bytes JSON data]
		if len(decoded) < sha256.Size {
			return data, errors.New("malformed cookie: data too short")
		}
		var signature []byte
		signature, jsonData = decoded[:sha256.Size], decoded[sha256.Size:]

		// Try all keys
		verified := false
		for _, key := range c.options.Secret {
			if verify(key, jsonData, signature) {
				verified = true
				break
			}
		}
		if !verified {
			return data, errors.New("malformed cookie: invalid signature")
		}
	} else {
		// unsigned cookie - data is just the JSON
		jsonData = decoded
	}

	// unmarshal json
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return data, errors.New("malformed cookie: invalid json")
	}
	return data, nil
}

func (c *rawCookie[T]) SetCookie(value string, options ...func(*http.Cookie)) *http.Cookie {
	// build cookie with defaults
	cookie := &http.Cookie{
		Path:     c.options.Path,
		Domain:   c.options.Domain,
		Secure:   c.options.Secure,
		HttpOnly: c.options.HttpOnly,
		SameSite: c.options.SameSite,
	}
	// Only set Expires if Duration is not zero (session cookie if zero)
	if c.options.Duration > 0 {
		cookie.Expires = time.Now().Add(c.options.Duration).UTC()
	}
	// override cookie defaults
	for _, opt := range options {
		opt(cookie)
	}
	cookie.Name = c.options.Name
	cookie.Value = value
	return cookie
}

func (c *rawCookie[T]) RemoveCookie() *http.Cookie {
	return &http.Cookie{
		Name:     c.options.Name,
		Path:     c.options.Path,
		Domain:   c.options.Domain,
		MaxAge:   -1,
		Secure:   c.options.Secure,
		HttpOnly: c.options.HttpOnly,
		SameSite: c.options.SameSite,
	}
}

func validateSecret(secret [][]byte, encrypted bool) error {
	if len(secret) == 0 {
		return errors.New("secret must have at least one key")
	}
	for _, key := range secret {
		// For encryption, ensure keys are exactly 32 bytes (AES-256)
		if encrypted {
			if len(key) == 32 {
				return nil
			}
			return fmt.Errorf("secret key for encrypted cookie must be 32 bytes, got %d bytes", len(key))
		}
		if len(key) < 32 {
			return fmt.Errorf("secret key for signed cookie must be at least 32 bytes, got %d bytes", len(key))
		}
	}
	return nil
}

// sign creates an HMAC signature for data.
func sign(secret []byte, data []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	return h.Sum(nil)
}

// verify checks if the HMAC signature is valid for data.
func verify(secret, data, signature []byte) bool {
	expected := sign(secret, data)
	return hmac.Equal(signature, expected)
}

// encrypt encrypts data using AES-256-GCM with the provided key and nonce.
func encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// decrypt decrypts data using AES-256-GCM with the provided key and nonce.
func decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
