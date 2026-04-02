package httpx

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ServerConfig configures the HTTP server and graceful shutdown.
type ServerConfig struct {
	// Server is the HTTP server configuration.
	// If nil, a default http.Server will be used.
	Server *http.Server

	// ShutdownTimeout is the maximum duration to wait for graceful shutdown.
	// Default: 30 seconds.
	ShutdownTimeout time.Duration
}

// ListenAndServe starts an HTTP server with graceful shutdown.
// It listens on the TCP network address addr and calls Serve with handler to handle requests.
//
// If handler is nil, [http.DefaultServeMux] is used.
// If config is nil, default configuration is used (default server, 30s shutdown timeout).
//
// The server will shutdown gracefully when it receives SIGINT or SIGTERM signals.
// During shutdown, the server will stop accepting new connections and wait for
// existing connections to finish, up to config.ShutdownTimeout.
//
// ListenAndServe returns when the server has shut down completely.
// If the server fails to start, it returns the error immediately.
// If shutdown completes successfully, it returns nil.
// If shutdown exceeds the timeout, it returns context.DeadlineExceeded.
func ListenAndServe(addr string, handler http.Handler, config *ServerConfig) error {
	// Use default config if nil
	if config == nil {
		config = &ServerConfig{}
	}

	// Use default server if nil
	server := config.Server
	if server == nil {
		server = &http.Server{}
	}
	server.Addr = addr
	server.Handler = handler

	// Use default shutdown timeout if zero
	shutdownTimeout := config.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second
	}

	// Channel to listen for errors from the server
	serverErr := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErr:
		// Server failed to start or stopped unexpectedly
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-shutdown:
		// Graceful shutdown initiated
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			server.Close()
			return err
		}

		// Wait for server to finish shutting down or timeout
		select {
		case err := <-serverErr:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
