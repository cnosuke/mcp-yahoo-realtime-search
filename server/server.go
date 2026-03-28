package server

import (
	"context"
	"log/slog"
	"time"

	yrs "github.com/cnosuke/go-yahoo-realtime-search"
	"github.com/cnosuke/mcp-yahoo-realtime-search/config"
	ierrors "github.com/cnosuke/mcp-yahoo-realtime-search/internal/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// RunStdio executes the MCP server with STDIO transport
func RunStdio(cfg *config.Config, name string, version string, revision string) error {
	mcpServer := createMCPServer(cfg, name, version, revision)

	zap.S().Infow("starting MCP server with STDIO")
	if err := mcpServer.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		return ierrors.Wrap(err, "failed to start STDIO server")
	}

	zap.S().Infow("STDIO server shutting down")
	return nil
}

func createMCPServer(cfg *config.Config, name string, version string, revision string) *mcp.Server {
	versionString := version
	if revision != "" && revision != "xxx" {
		versionString = versionString + " (" + revision + ")"
	}

	var opts []yrs.ClientOption
	if cfg.Search.UserAgent != "" {
		opts = append(opts, yrs.WithUserAgent(cfg.Search.UserAgent))
	}
	if cfg.Search.RequestTimeout > 0 {
		opts = append(opts, yrs.WithRequestTimeout(time.Duration(cfg.Search.RequestTimeout)*time.Second))
	}
	yrsClient := yrs.NewClient(opts...)

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: versionString,
	}, &mcp.ServerOptions{
		Logger: slog.Default(),
	})

	RegisterTools(mcpServer, yrsClient)

	return mcpServer
}
