package server

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"errors"

	"github.com/cnosuke/mcp-yahoo-realtime-search/config"
	ierrors "github.com/cnosuke/mcp-yahoo-realtime-search/internal/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// RunHTTP - Execute the MCP server with Streamable HTTP transport
func RunHTTP(cfg *config.Config, name string, version string, revision string) error {
	zap.S().Infow("starting MCP Yahoo Realtime Search Server with Streamable HTTP transport")

	mcpSrv := createMCPServer(cfg, name, version, revision)

	httpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return mcpSrv },
		nil,
	)

	var handler http.Handler = httpHandler
	handler = withAuthMiddleware(handler, cfg.HTTP.AuthToken)
	handler = withOriginValidation(handler, cfg.HTTP.AllowedOrigins)
	handler = withRequestLogging(handler)

	mux := http.NewServeMux()
	mux.Handle(cfg.HTTP.EndpointPath, handler)
	mux.HandleFunc("/health", handleHealth)

	srv := &http.Server{
		Addr:              cfg.HTTP.Binding,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		zap.S().Infow("starting Streamable HTTP server",
			"binding", cfg.HTTP.Binding,
			"endpoint", cfg.HTTP.EndpointPath,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- ierrors.Wrap(err, "HTTP server error")
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		zap.S().Infow("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			zap.S().Warnw("graceful shutdown timed out, forcing close", "error", err)
			srv.Close()
		}
	}

	return nil
}
