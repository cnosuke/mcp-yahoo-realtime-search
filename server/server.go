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

// RunStdio - Execute the MCP server with STDIO transport
func RunStdio(cfg *config.Config, name string, version string, revision string) error {
	zap.S().Infow("starting MCP Yahoo Realtime Search Server with STDIO transport")

	mcpServer, err := createMCPServer(cfg, name, version, revision)
	if err != nil {
		return err
	}

	zap.S().Infow("starting MCP server with STDIO")
	err = mcpServer.Run(context.Background(), &mcp.StdioTransport{})
	if err != nil {
		zap.S().Errorw("failed to start STDIO server", "error", err)
		return ierrors.Wrap(err, "failed to start STDIO server")
	}

	zap.S().Infow("STDIO server shutting down")
	return nil
}

// createMCPServer - Create MCP server instance with common configuration
func createMCPServer(cfg *config.Config, name string, version string, revision string) (*mcp.Server, error) {
	versionString := version
	if revision != "" && revision != "xxx" {
		versionString = versionString + " (" + revision + ")"
	}

	zap.S().Debugw("creating yrs.Client")
	var opts []yrs.ClientOption
	if cfg.Search.UserAgent != "" {
		opts = append(opts, yrs.WithUserAgent(cfg.Search.UserAgent))
	}
	if cfg.Search.RequestTimeout > 0 {
		opts = append(opts, yrs.WithRequestTimeout(time.Duration(cfg.Search.RequestTimeout)*time.Second))
	}
	yrsClient := yrs.NewClient(opts...)

	zap.S().Debugw("creating MCP server",
		"name", name,
		"version", versionString,
	)
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: versionString,
	}, &mcp.ServerOptions{
		Logger: slog.Default(),
	})

	zap.S().Debugw("registering tools")
	if err := RegisterAllTools(mcpServer, yrsClient); err != nil {
		zap.S().Errorw("failed to register tools", "error", err)
		return nil, err
	}

	return mcpServer, nil
}
