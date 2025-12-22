package session

import (
	"net/http"
)

// FlashCookie provides flash message functionality using cookies.
// Flash cookies are deleted immediately after being read once.
// By default, they are session cookies that expire when the browser closes.
//
// Example usage:
//
//	flash, _ := NewFlashCookie[string](FlashCookieOptions{
//	    Name: "flash_message",
//	    Secret: Secret(key),
//	})
//
//	// Set a flash message
//	flash.Set(w, "Account created successfully!")
//
//	// Read and delete in one operation
//	message, _ := flash.Get(w, r)
//
//	// Peek without deleting
//	message, _ := flash.Peek(r)
type FlashCookie[T any] interface {
	// Peek reads the flash message without deleting it.
	// Returns an error if the cookie is not found, malformed, or has an invalid signature.
	Peek(r *http.Request) (T, error)

	// Get reads the flash message and immediately deletes it from the response.
	// Returns an error if the cookie is not found, malformed, or has an invalid signature.
	Get(w http.ResponseWriter, r *http.Request) (T, error)

	// Set writes a flash message to the response cookies.
	// Optional cookie modifier functions can be passed to override defaults.
	Set(w http.ResponseWriter, data T, options ...func(*http.Cookie)) error
}

// NewFlashCookie creates a new flash cookie.
// Flash cookies are deleted immediately after being read once.
//
// The Duration field in CookieOptions is ignored for flash cookies.
// By default, flash cookies are session cookies (expire when browser closes).
func NewFlashCookie[T any](options CookieOptions) (FlashCookie[T], error) {
	options.Duration = 0

	cookie, err := NewCookie[T](options)
	if err != nil {
		return nil, err
	}
	return &flash[T]{cookie}, nil
}

type flash[T any] struct{ cookie Cookie[T] }

// Peek reads the flash message without deleting it.
func (f *flash[T]) Peek(r *http.Request) (T, error) {
	return f.cookie.Get(r)
}

// Get reads the flash message and immediately deletes it.
func (f *flash[T]) Get(w http.ResponseWriter, r *http.Request) (T, error) {
	data, err := f.cookie.Get(r)
	if err != nil {
		return data, err
	}
	// Delete the cookie after reading
	f.cookie.Delete(w)
	return data, nil
}

// Set writes a flash message to cookies.
// The cookie is created as a session cookie (expires when browser closes) by default.
func (f *flash[T]) Set(w http.ResponseWriter, data T, options ...func(*http.Cookie)) error {
	return f.cookie.Set(w, data, options...)
}
