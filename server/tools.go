package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	yrs "github.com/cnosuke/go-yahoo-realtime-search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// SearchArgs - Arguments for the search tool
type SearchArgs struct {
	Query string `json:"query" jsonschema:"Search keyword or phrase"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of tweets to return (0 = all, default 0)"`
}

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *mcp.Server, yrsClient *yrs.Client) error {
	if err := registerSearchTool(mcpServer, yrsClient); err != nil {
		return err
	}

	return nil
}

func registerSearchTool(mcpServer *mcp.Server, yrsClient *yrs.Client) error {
	zap.S().Debugw("registering search tool")

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "search",
		Description: "Search tweets via Yahoo! Japan Realtime Search",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchArgs) (*mcp.CallToolResult, any, error) {
		zap.S().Debugw("executing search",
			"query", input.Query,
			"limit", input.Limit)

		if input.Query == "" {
			return nil, nil, fmt.Errorf("missing or empty query parameter")
		}
		if input.Limit < 0 {
			return nil, nil, fmt.Errorf("limit must be non-negative")
		}

		result, err := yrsClient.SearchWithLimit(ctx, input.Query, input.Limit)
		if err != nil {
			zap.S().Errorw("search failed",
				"query", input.Query,
				"error", err)
			return nil, nil, err
		}

		formatted := formatSearchResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatted}},
		}, nil, nil
	})

	return nil
}

func formatSearchResult(result *yrs.SearchResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Search Results for %q\n\n", result.Query))
	sb.WriteString(fmt.Sprintf("Found %d tweets.\n", len(result.Tweets)))

	for _, tw := range result.Tweets {
		sb.WriteString("\n---\n\n")
		sb.WriteString(fmt.Sprintf("### @%s (%s)\n", tw.ScreenName, tw.AuthorName))
		sb.WriteString(fmt.Sprintf("> %s\n\n", tw.Text))
		sb.WriteString(fmt.Sprintf("- Posted: %s\n", tw.CreatedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- Replies: %d | RT: %d | Likes: %d\n", tw.ReplyCount, tw.RTCount, tw.LikeCount))
		sb.WriteString(fmt.Sprintf("- URL: %s\n", tw.URL))
		if len(tw.Images) > 0 {
			urls := make([]string, len(tw.Images))
			for i, img := range tw.Images {
				urls[i] = img.URL
			}
			sb.WriteString(fmt.Sprintf("- Images: %s\n", strings.Join(urls, ", ")))
		}
	}
	return sb.String()
}
