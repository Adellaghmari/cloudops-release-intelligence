package runtime

import (
	"context"
	"net"
	"net/http"
	"time"
)

// ServeHTTP listens until ctx is cancelled (SIGINT/SIGTERM via signal.NotifyContext),
// then shuts the server down. This is the process model used both on a developer
// machine and as PID 1 inside a Linux container.
func ServeHTTP(ctx context.Context, srv *http.Server, shutdownWait time.Duration) error {
	if srv.BaseContext == nil {
		srv.BaseContext = func(_ net.Listener) context.Context { return ctx }
	}
	return serve(ctx, srv, shutdownWait, srv.ListenAndServe)
}

// ServeListener is the same shutdown contract as ServeHTTP, with a caller-owned listener.
// Tests use it so the bound address is known before the first request.
func ServeListener(ctx context.Context, srv *http.Server, ln net.Listener, shutdownWait time.Duration) error {
	if srv.BaseContext == nil {
		srv.BaseContext = func(_ net.Listener) context.Context { return ctx }
	}
	return serve(ctx, srv, shutdownWait, func() error { return srv.Serve(ln) })
}

func serve(ctx context.Context, srv *http.Server, shutdownWait time.Duration, listen func() error) error {
	errCh := make(chan error, 1)
	go func() {
		err := listen()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownWait)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}
