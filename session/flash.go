package session

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// Flash sets and gets flash messages with session cookies (cookies that are deleted once the browser session ends).
type Flash[T any] struct {
	options FlashOptions
}

type FlashOptions struct {
	// Cookie configures the cookie name.
	Cookie string

	// Other cookie options

	SameSite http.SameSite // default is http.SameSiteLaxMode
	Secure   bool          // default is false
	Path     string        // optional
	Domain   string        // optional
}

func NewFlash[T any](options FlashOptions) *Flash[T] {
	if options.Cookie == "" {
		options.Cookie = "auth_session"
	}
	if options.SameSite == 0 {
		options.SameSite = http.SameSiteLaxMode
	}
	return &Flash[T]{
		options: options,
	}
}

// Set writes a flash message to cookies.
func (f *Flash[T]) Set(w http.ResponseWriter, data T) {
	// marshal json and encode with base64
	jsonData, _ := json.Marshal(data)
	cookieData := base64.URLEncoding.EncodeToString(jsonData)
	// set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     f.options.Cookie,
		Value:    cookieData,
		HttpOnly: true,
		SameSite: f.options.SameSite,
		Secure:   f.options.Secure,
		Path:     f.options.Path,
		Domain:   f.options.Domain,
	})
}

// Has returns a boolean indicating if a flash message exists in cookies.
func (f *Flash[T]) Has(r *http.Request) bool {
	var data T
	cookieData, err := r.Cookie(f.options.Cookie)
	if err != nil {
		return false
	}
	// decode base64 and unmarshal json
	jsonString, err := base64.URLEncoding.DecodeString(cookieData.Value)
	if err != nil {
		// malformed cookie
		return false
	}
	if err := json.Unmarshal(jsonString, &data); err != nil {
		// malformed cookie
		return false
	}
	return true
}

// Get reads a flash message from cookies and deletes it immediately.
// Get returns the zero value if it does not exist.
func (f *Flash[T]) Get(w http.ResponseWriter, r *http.Request) T {
	var data T
	cookieData, err := r.Cookie(f.options.Cookie)
	if err != nil {
		return data
	}
	// clear cookie
	defer func() {
		http.SetCookie(w, &http.Cookie{
			Name:     f.options.Cookie,
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: f.options.SameSite,
			Secure:   f.options.Secure,
			Path:     f.options.Path,
			Domain:   f.options.Domain,
		})
	}()
	// decode base64 and unmarshal json
	jsonString, err := base64.URLEncoding.DecodeString(cookieData.Value)
	if err != nil {
		// malformed cookie
		return data
	}
	if err := json.Unmarshal(jsonString, &data); err != nil {
		// malformed cookie
		return data
	}
	return data
}
