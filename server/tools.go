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

type SearchArgs struct {
	Query    string   `json:"query,omitempty" jsonschema:"Search keyword or phrase"`
	Limit    int      `json:"limit,omitempty" jsonschema:"Maximum number of tweets to return (0 = all, default 0)"`
	Not      []string `json:"not,omitempty" jsonschema:"Exclude tweets containing these keywords"`
	Or       []string `json:"or,omitempty" jsonschema:"Match any of these keywords (OR group)"`
	FromUser string   `json:"from_user,omitempty" jsonschema:"Filter by posting account (screen name without @)"`
	ToUser   string   `json:"to_user,omitempty" jsonschema:"Filter by mention target (screen name without @)"`
	Hashtags []string `json:"hashtags,omitempty" jsonschema:"Filter by hashtags (without #)"`
	URL      string   `json:"url,omitempty" jsonschema:"Filter by URL or domain"`
}

// RegisterTools registers all tools with the server
func RegisterTools(mcpServer *mcp.Server, yrsClient *yrs.Client) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "search",
		Description: "Search tweets on X (formerly Twitter), primarily Japanese-language content. Supports filtering by user, hashtag, URL, and keyword exclusion. Powered by Yahoo! Japan Realtime Search.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchArgs) (*mcp.CallToolResult, any, error) {
		zap.S().Debugw("executing search",
			"query", input.Query,
			"limit", input.Limit,
			"not", input.Not,
			"or", input.Or,
			"from_user", input.FromUser,
			"to_user", input.ToUser,
			"hashtags", input.Hashtags,
			"url", input.URL)

		if input.Limit < 0 {
			return nil, nil, fmt.Errorf("limit must be non-negative")
		}
		if input.Query == "" && len(input.Not) == 0 && len(input.Or) == 0 &&
			input.FromUser == "" && input.ToUser == "" && len(input.Hashtags) == 0 && input.URL == "" {
			return nil, nil, fmt.Errorf("at least one search parameter (query or filter) is required")
		}

		q := buildQuery(input)

		result, err := yrsClient.SearchWithQueryAndLimit(ctx, q, input.Limit)
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
}

func buildQuery(input SearchArgs) *yrs.Query {
	var q *yrs.Query
	if input.Query != "" {
		q = yrs.NewQuery(input.Query)
	} else {
		q = yrs.NewQuery()
	}

	if len(input.Not) > 0 {
		q.Not(input.Not...)
	}
	if len(input.Or) > 0 {
		q.Or(input.Or...)
	}
	if input.FromUser != "" {
		q.FromUser(input.FromUser)
	}
	if input.ToUser != "" {
		q.ToUser(input.ToUser)
	}
	if len(input.Hashtags) > 0 {
		q.Hashtag(input.Hashtags...)
	}
	if input.URL != "" {
		q.URL(input.URL)
	}

	return q
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
