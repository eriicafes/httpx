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
- Cookie-based token storage (default) or custom storage
- Optimistic validation with cached sessions (optional)

### Basic Usage

```go
import (
    "time"
    "github.com/eriicafes/httpx/session"
)

// Define your session and user types
type Session struct {
    ID         string
    UserID     string
    SecretHash []byte
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
```

### Login

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    // Validate credentials
    user := authenticateUser(r)

    // Create session and set cookie
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
    session, user, err := auth.LogoutRequest(w, r)
    if err != nil {
        http.Error(w, "Logout failed", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Logged out successfully")
}
```

## Cookies

Type-safe cookie operations with signing and encryption support.

**Features:**
- Generic types for cookie data (automatic JSON serialization)
- Signed cookies with HMAC-SHA256 (prevents tampering)
- Encrypted cookies with AES-256-GCM (for sensitive data)
- Key rotation support
- Unsigned cookies (for non-sensitive data)

### Basic Usage

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

```go
secret := session.Secret([]byte("your-secret-key"))

// Create an signed cookie
cookie, err := session.NewCookie[SensitiveData](session.CookieOptions{
    Name:      "sensitive",
    Secret:    secret,
})
```

### Encrypted Cookies

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

One-time messages that auto-delete after reading.

**Features:**
- Automatic deletion after being read once
- Session cookies (expire when browser closes) by default
- All cookie features (signing, encryption, etc.)

```go
// Or create flash cookie directly
flash, err := session.NewFlashCookie[string](session.CookieOptions{
    Name:   "flash_message",
    Secret: session.Secret(key),
})

// Set flash message (on redirect)
flash.Set(w, "Account created successfully!")
http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

// Get and delete flash message (on next page)
message, err := flash.Get(w, r)
// message is now deleted

// Or peek without deleting
message, err := flash.Peek(r)
```

## Advanced Usage

### Token Stores

Token stores handle how session tokens are stored and retrieved from HTTP requests/responses.

The package provides two built-in cookie-based token stores:

#### CookieTokenStore (Default)

Basic unsigned cookie storage that stores only the session ID. Requires database lookup on every request.

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

#### TrustedCookieTokenStore (Optimistic)

Encrypted cookie storage that caches session and user data. Enables optimistic authentication without database lookups until the cache becomes stale.

```go
secret := session.Secret([]byte("your-secret-key"))

// Trust cached data for 15 minutes before database revalidation
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
- On authentication: Data is read directly from the cookie (no database lookup)
- After stale duration: Falls back to database validation and refreshes the cache
- Cookie must be signed or encrypted for security (recommended: AES-256-GCM encryption)

**Benefits:**
- Reduces database load by avoiding session lookups on every request
- Faster authentication (no database roundtrip)
- Automatic cache refresh after stale duration
- Still secure - falls back to database validation periodically

#### Custom TokenStore

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

**Example - Header-based token store:**

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

#### Custom TrustedTokenStore

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

**Example - Using the built-in TrustedCookieTokenStore:**

```go
secret := session.Secret([]byte("your-secret-key"))

// TrustedCookieTokenStore caches session/user data in encrypted cookies
tokenStore, err := session.NewTrustedCookieTokenStore[Session, User](
    secret,
    time.Minute * 15, // Cache data for 15 minutes before database revalidation
    session.WithCookieName("auth_session"),
    session.WithCookieSecure(false) // For development only
)

auth, err := session.NewAuth(store,
    session.WithTokenStore(tokenStore),
)
```

**Example - Custom JWT-based trusted token store:**

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
    w.Header().Del("X-Session-Token")
}

func (s *JWTTokenStore) FromTrustedToken(token string) (Session, User, error) {
    // Parse and validate JWT signature
    claims, err := jwt.Parse(token, s.signingKey)
    if err != nil {
        return Session{}, User{}, err
    }

    // Check if JWT is expired - extract original session token for database validation
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

// Use it
auth, err := session.NewAuth(store,
    session.WithTokenStore(&JWTTokenStore{signingKey: key}),
)
```

**Key points:**
- `FromTrustedToken` receives the token from `GetToken()`
- Return `*StaleSessionError` to trigger database revalidation when cache expires
- `ToTrustedToken` creates the trusted token that goes to `SetToken()`
- The original session token should be embedded for fallback validation

### Custom Session Refresh

By default, sessions are automatically refreshed when 50% of their duration has passed (e.g., a 30-day session refreshes after 15 days).

**Disable automatic refresh:**
```go
// Disable session refresh completely
auth, err := session.NewAuth(store,
    session.WithRefreshThreshold(nil),
)
```

**Custom refresh threshold:**
```go
// Refresh when 75% of duration has passed
customRefresh := func(expiresAt time.Time, duration time.Duration) (time.Time, bool) {
    if time.Until(expiresAt) < duration/4 {
        return time.Now().Add(duration), true
    }
    return time.Time{}, false
}

auth, err := session.NewAuth(store,
    session.WithRefreshThreshold(customRefresh),
)
```

### Manual Session Management

```go
// Create session manually
sessionToken, session, err := auth.CreateSession(user)

// Validate session token manually
session, user, optimistic, err := auth.ValidateSessionToken(sessionToken)

// Invalidate single session
auth.InvalidateSession(sessionId)

// Invalidate all user sessions
auth.InvalidateAllSessions(user)
```

## Security Features

- Session secrets are hashed with SHA-256 before storage
- Cookie signing with HMAC-SHA256 prevents tampering
- Optional AES-256-GCM encryption for sensitive cookie data
- Constant-time secret comparison prevents timing attacks
- Automatic session refresh to extend active sessions
- Support for key rotation
- Secure defaults for auth cookies (HttpOnly, Secure, SameSite)
