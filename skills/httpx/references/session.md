# Session Reference

Use `github.com/eriicafes/httpx/session` for session-based authentication, cookie token storage, and flash messages.

## Auth Flow

Create an auth instance with `session.NewAuth(store)`, where `store` implements the package's session store interface.

High-level request helpers:

- `auth.LoginRequest(w, user)` creates a session and writes the token
- `auth.AuthenticateRequest(w, r)` loads and validates the current session, optionally refreshing it
- `auth.LogoutRequest(w, r)` invalidates the current session and removes the token

These helpers are the default choice for handler code.

## Token Storage

If you do not pass a token store, `NewAuth` uses the default cookie token store.

Use:

- `session.NewCookieTokenStore(...)` for simple cookie-backed session tokens
- `session.NewTrustedCookieTokenStore[Session, User](secret, staleDuration, ...)` when you want optimistic authentication with signed or encrypted cached session data

When using trusted cookies, ensure sensitive fields like session secret hashes are not serialized to clients.

## CSRF

For cookie auth, use Go's `http.NewCrossOriginProtection()` for unsafe cross-origin requests and keep state-changing actions off safe HTTP methods.

```go
csrf := http.NewCrossOriginProtection()
csrf.AddTrustedOrigin("https://app.example.com")

mux := httpx.Use(http.NewServeMux(), csrf.Handler)
```

## Flash Messages

Use the flash helpers in this package when you need short-lived messages persisted across redirects or the next request.
