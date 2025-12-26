# session

Session-based authentication, type-safe cookies, and flash messages.

## Installation

```bash
go get github.com/eriicafes/httpx
```

## Session-based Authentication

Session-based authentication following the [Lucia Auth guide](https://lucia-auth.com/sessions/cookies/) for secure session management.

**Features:**
- Secure session token generation and validation
- Generic types for session and user data
- Automatic session refresh
- Token store (default cookie store or custom)
- Optimistic validation with cached sessions (optional)

### Basic Usage

```go
import (
    "net/http"
    "time"
    "github.com/eriicafes/httpx"
    "github.com/eriicafes/httpx/session"
)

// Define your session and user types
type Session struct {
    ID         string
    UserID     string
    SecretHash []byte `json:"-"` // Never send secret hash to client
    ExpiresAt  time.Time
}

func (s Session) GetSession() (string, []byte, time.Time) {
    return s.ID, s.SecretHash, s.ExpiresAt
}

type User struct {
    ID   string
    Name string
}

// Implement SessionStore
type MySessionStore struct {
    // your database connection
}

func (s *MySessionStore) GetSessionAndUser(sessionId string) (Session, User, error) {
    // retrieve session and user from database
}

func (s *MySessionStore) CreateSession(sessionData session.Session, user User) (Session, error) {
    // create session in database
}

func (s *MySessionStore) UpdateSession(sessionId string, expiresAt time.Time) (Session, error) {
    // update session expiration
}

func (s *MySessionStore) DeleteSession(sessionId string) error {
    // delete session from database
}

func (s *MySessionStore) DeleteUserSessions(user User) error {
    // delete all sessions for user
}

// Create auth instance
store := &MySessionStore{}
auth, err := session.NewAuth(store)

// CSRF protection
csrf := http.NewCrossOriginProtection()

mux := httpx.Use(http.NewServeMux(), csrf.Handler)

mux.Route("POST /login", loginHandler)
mux.Route("POST /logout", logoutHandler)
```

### Login

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    // Validate credentials
    user := authenticateUser(r)

    // LoginRequest creates a new session and sets the session token in the response
    session, err := auth.LoginRequest(w, user)
    if err != nil {
        http.Error(w, "Login failed", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Logged in successfully")
}
```

### Protected Routes

```go
func protectedHandler(w http.ResponseWriter, r *http.Request) {
    // AuthenticateRequest gets the session token from request, validates it, and returns session and user
    // Automatically refreshes the session and updates the token in the response
    session, user, err := auth.AuthenticateRequest(w, r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Use session and user data
    fmt.Fprintf(w, "Welcome, %s!", user.Name)
}
```

### Logout

```go
func logoutHandler(w http.ResponseWriter, r *http.Request) {
    // LogoutRequest gets the session token from request, validates it, deletes the session, and removes the token
    session, user, err := auth.LogoutRequest(w, r)
    if err != nil {
        http.Error(w, "Logout failed", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Logged out successfully")
}
```

### Session Management

For custom flows, use the manual session methods below. The `LoginRequest`, `AuthenticateRequest`, and `LogoutRequest` methods shown above are convenience helpers that combine these operations.

```go
// Login flow
sessionToken, session, err := auth.CreateSession(user)
err = auth.SetSessionToken(w, sessionToken, session, user)

// Authentication flow
sessionToken, err := auth.GetSessionToken(r)
session, user, optimistic, err := auth.ValidateSessionToken(sessionToken)
// Update token store if not optimistic
if !optimistic {
    err = auth.SetSessionToken(w, sessionToken, session, user)
}

// Logout flow - single session
sessionToken, err := auth.GetSessionToken(r)
sessionId, _ := session.ParseSessionToken(sessionToken)
auth.InvalidateSession(sessionId)
auth.DeleteSessionToken(w)

// Logout flow - all user sessions (e.g., "logout from all devices")
sessionToken, err := auth.GetSessionToken(r)
session, user, _, err := auth.ValidateSessionToken(sessionToken)
auth.InvalidateAllSessions(user)
auth.DeleteSessionToken(w)
```

### CSRF Protection

For cookie-based authentication, `SameSite=Lax` limits when session cookies are sent, while Go's origin-based CSRF protection blocks non-safe cross-origin browser requests (including those from other subdomains). Together, they provide protection against CSRF, provided that safe HTTP methods are never used to perform state-changing actions.

Use Go 1.25's `http.NewCrossOriginProtection()` to block cross-origin requests:

```go
csrf := http.NewCrossOriginProtection()

// Trust additional origins
csrf.AddTrustedOrigin("https://app.example.com")

// Use a custom error response
csrf.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    httpx.SendStatus(w, http.StatusForbidden, httpx.JSON{
        "error": "invalid_origin",
    })
}))

mux := httpx.Use(http.NewServeMux(), csrf.Handler)
```

Cross-origin requests are detected with the `Sec-Fetch-Site` header or by comparing the hostname of the `Origin` header with `Host` header. GET requests are always allowed.

For more details, see the [http.CrossOriginProtection](https://pkg.go.dev/net/http#CrossOriginProtection) documentation.

### Advanced Usage

#### Token Stores

Token stores handle how session tokens are stored and retrieved from HTTP requests/responses.

The package provides two built-in cookie-based token stores:

##### CookieTokenStore (Default)

Basic unsigned cookie storage that stores only the session ID. Requires lookup on every request.

```go
tokenStore, err := session.NewCookieTokenStore(
    session.WithCookieName("auth_session"),
    session.WithCookieSecure(false) // For development only
)

auth, err := session.NewAuth(store,
    session.WithTokenStore(tokenStore),
)
```

**Note:** If no token store is specified, `NewAuth` uses `CookieTokenStore` by default.

##### TrustedCookieTokenStore (Optimistic)

Encrypted cookie storage that caches session and user data. Enables optimistic authentication without lookups until the cache becomes stale.

```go
secret := session.Secret([]byte("your-secret-key"))

// Trust cached data for 15 minutes before session revalidation
tokenStore, err := session.NewTrustedCookieTokenStore[Session, User](
    secret,
    time.Minute * 15, // Stale duration
    session.WithCookieName("auth_session"),
    session.WithCookieSecure(false) // For development only
)

auth, err := session.NewAuth(store,
    session.WithTokenStore(tokenStore),
)
```

**How it works:**
- On login: Session and user data are stored in a signed or encrypted cookie
- On authentication: Data is read directly from the cookie
- After stale duration: Falls back to session validation and refreshes the cache
- Cookie must be signed or encrypted for security

**Security note:** When using trusted token stores, ensure the session's `SecretHash` field is never serialized to the client. Use the `json:"-"` tag to omit it from JSON encoding. The secret hash must remain in the database only

##### Custom TokenStore

Implement this interface for custom token storage (cookies, headers, etc.):

```go
type TokenStore interface {
    // GetToken retrieves the session token from the request
    GetToken(r *http.Request) (sessionToken string, err error)

    // SetToken stores the session token in the response
    SetToken(w http.ResponseWriter, sessionToken string, expiresAt time.Time) error

    // DeleteToken removes the session token from the response
    DeleteToken(w http.ResponseWriter)
}
```

Example - Header-based token store

```go
type HeaderTokenStore struct{}

func (s *HeaderTokenStore) GetToken(r *http.Request) (string, error) {
    token := r.Header.Get("Authorization")
    if token == "" {
        return "", http.ErrNoCookie // Use standard error
    }
    return strings.TrimPrefix(token, "Bearer "), nil
}

func (s *HeaderTokenStore) SetToken(w http.ResponseWriter, token string, expiresAt time.Time) error {
    w.Header().Set("X-Session-Token", token)
    return nil
}

func (s *HeaderTokenStore) DeleteToken(w http.ResponseWriter) {
    w.Header().Del("X-Session-Token")
}
```

##### Custom TrustedTokenStore

Implement this interface to add optimistic validation with session/user caching (like JWT, signed tokens, etc.):

```go
type TrustedTokenStore[S Session, U any] interface {
    TokenStore // Embed TokenStore methods

    // FromTrustedToken validates and extracts session/user data from a trusted token
    // Return *StaleSessionError to fallback to session store validation
    FromTrustedToken(trustedToken string) (session S, user U, err error)

    // ToTrustedToken creates a trusted token containing session and user data
    // The returned token will be passed to TokenStore.SetToken()
    ToTrustedToken(sessionToken string, session S, user U) (trustedToken string, err error)
}
```

Example - JWT-based trusted token store:

```go
type JWTTokenStore struct {
    signingKey []byte
}

func (s *JWTTokenStore) GetToken(r *http.Request) (string, error) {
    // Get JWT from Authorization header
    auth := r.Header.Get("Authorization")
    return strings.TrimPrefix(auth, "Bearer "), nil
}

func (s *JWTTokenStore) SetToken(w http.ResponseWriter, token string, expiresAt time.Time) error {
    // Return JWT in response header for client to store
    w.Header().Set("X-Session-Token", token)
    return nil
}

func (s *JWTTokenStore) DeleteToken(w http.ResponseWriter) {
    // Signal client to clear the token
    w.Header().Set("X-Session-Token", "")
}

func (s *JWTTokenStore) FromTrustedToken(token string) (Session, User, error) {
    // Parse and validate JWT signature
    claims, err := jwt.Parse(token, s.signingKey)
    if err != nil {
        return Session{}, User{}, err
    }

    // Check if JWT is expired - extract original session token for fallback session validation
    if time.Now().After(claims.ExpiresAt) {
        return Session{}, User{}, &session.StaleSessionError{
            SessionToken: claims.SessionToken,
        }
    }

    // Return cached session and user data
    return claims.Session, claims.User, nil
}

func (s *JWTTokenStore) ToTrustedToken(sessionToken string, sess Session, user User) (string, error) {
    // Create JWT with session and user data
    claims := JWTClaims{
        Session:                sess,
        User:                   user,
        OriginalSessionToken:   sessionToken,
        StaleAt:                time.Now().Add(15 * time.Minute),
    }
    return jwt.Sign(claims, s.signingKey)
}
```

**Key points:**
- `FromTrustedToken` receives the token from `GetToken()`
- Return `*StaleSessionError` to trigger session revalidation when cache expires
- `ToTrustedToken` creates the trusted token that goes to `SetToken()`
- The original session token should be embedded for fallback validation

#### Key Rotation

Support for rotating encryption/signing keys without invalidating existing sessions. Pass multiple secrets when creating token stores - the first key is used for new operations, older keys are tried for validation.

```go
oldSecret := []byte("old-secret-key")
newSecret := []byte("new-secret-key")

// Use multiple secrets for key rotation
tokenStore, err := session.NewTrustedCookieTokenStore[Session, User](
    session.Secret(newSecret, oldSecret), // New key first, old keys follow
    time.Minute * 15,
    session.WithCookieName("auth_session"),
)

auth, err := session.NewAuth(store,
    session.WithTokenStore(tokenStore),
)
```

This allows seamless key rotation: new sessions use the new key, while existing sessions with the old key remain valid.

## Cookies

Type-safe cookie operations with signing and encryption support. Uses generic types for automatic JSON serialization of your data structures.

### Unsigned Cookies

Suitable for non-sensitive data like user preferences or UI settings. The data is stored as plain JSON and is readable by the client.

```go
// Create cookie (unsigned by default)
cookie, err := session.NewCookie[UserData](session.CookieOptions{
    Name:   "user_prefs",
})

// Set cookie
userData := UserData{Theme: "dark", Lang: "en"}
cookie.Set(w, userData)

// Get cookie
data, err := cookie.Get(r)

// Delete cookie
cookie.Delete(w)
```

### Signed Cookies

Prevents tampering using HMAC-SHA256 signatures. The data is still readable by the client, but any modifications will invalidate the signature. Use this for data that needs integrity protection but doesn't need to be hidden.

Supports key rotation by passing multiple secrets (the first key is used for signing, older keys are tried for verification).

```go
secret := session.Secret([]byte("your-secret-key"))

// Create an signed cookie
cookie, err := session.NewCookie[SensitiveData](session.CookieOptions{
    Name:      "sensitive",
    Secret:    secret,
})
```

### Encrypted Cookies

Encrypts data using AES-256-GCM, which both hides the content and prevents tampering. The data is completely hidden from the client. Use this for sensitive information that must remain confidential (e.g., tokens, personal data).

Supports key rotation by passing multiple secrets (the first key is used for encryption, older keys are tried for decryption).

```go
secret := session.Secret([]byte("your-secret-key"))

// Create an encrypted cookie
cookie, err := session.NewCookie[SensitiveData](session.CookieOptions{
    Name:      "sensitive",
    Secret:    secret,
    Encrypted: true,
})
```

### Cookie Options

```go
// Using DefaultCookieOptions with custom options
opts := session.DefaultCookieOptions("my_cookie",
    session.WithCookieSecret(session.Secret(key1, key2)),
    session.WithCookieDuration(time.Hour * 24),
    session.WithCookieSecure(true),
    session.WithCookieHttpOnly(true),
    session.WithCookieSameSite(http.SameSiteStrictMode),
    session.WithCookiePath("/"),
    session.WithCookieDomain("example.com"),
)
cookie, err := session.NewCookie[T](opts)

// Or using CookieOptions directly
cookie, err := session.NewCookie[T](session.CookieOptions{
    Name:     "my_cookie",
    Secret:   session.Secret(key1, key2),
    Duration: time.Hour * 24,
    Secure:   true,
    HttpOnly: true,
    SameSite: http.SameSiteStrictMode,
    Path:     "/",
    Domain:   "example.com",
})
```

## Flash Messages

Temporary messages that are automatically deleted after being read once. Flash messages are stored as session cookies (expire when browser closes) by default and support all cookie features like signing and encryption.

```go
// Create flash cookie
flash, err := session.NewFlashCookie[string](session.CookieOptions{
    Name:   "flash_message",
    Secret: session.Secret(key),
})

// Set flash message (on redirect)
flash.Set(w, "Account created successfully!")
http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

// Get and delete flash message (on next request)
message, err := flash.Get(w, r)

// Or peek without deleting
message, err := flash.Peek(r)
```
