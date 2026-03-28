package server

import (
	"testing"
	"time"

	yrs "github.com/cnosuke/go-yahoo-realtime-search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestRegisterTools(t *testing.T) {
	logger := zaptest.NewLogger(t)
	zap.ReplaceGlobals(logger)

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "0.0.1"}, nil)
	yrsClient := yrs.NewClient()

	assert.NotPanics(t, func() {
		RegisterTools(mcpServer, yrsClient)
	})
}

func TestFormatSearchResult_Empty(t *testing.T) {
	result := &yrs.SearchResult{
		Query:  "test",
		Tweets: []yrs.Tweet{},
	}

	formatted := formatSearchResult(result)
	assert.Contains(t, formatted, `## Search Results for "test"`)
	assert.Contains(t, formatted, "Found 0 tweets.")
}

func TestFormatSearchResult_WithTweets(t *testing.T) {
	result := &yrs.SearchResult{
		Query: "golang",
		Tweets: []yrs.Tweet{
			{
				ScreenName: "user1",
				AuthorName: "User One",
				Text:       "Hello from Go!",
				CreatedAt:  time.Date(2026, 3, 27, 10, 30, 0, 0, time.UTC),
				ReplyCount: 5,
				RTCount:    12,
				LikeCount:  42,
				URL:        "https://x.com/user1/status/123",
			},
		},
	}

	formatted := formatSearchResult(result)
	assert.Contains(t, formatted, `## Search Results for "golang"`)
	assert.Contains(t, formatted, "Found 1 tweets.")
	assert.Contains(t, formatted, "### @user1 (User One)")
	assert.Contains(t, formatted, "> Hello from Go!")
	assert.Contains(t, formatted, "2026-03-27T10:30:00Z")
	assert.Contains(t, formatted, "Replies: 5 | RT: 12 | Likes: 42")
	assert.Contains(t, formatted, "https://x.com/user1/status/123")
}

func TestFormatSearchResult_WithImages(t *testing.T) {
	result := &yrs.SearchResult{
		Query: "photo",
		Tweets: []yrs.Tweet{
			{
				ScreenName: "photographer",
				AuthorName: "Photo User",
				Text:       "Nice shot!",
				CreatedAt:  time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC),
				URL:        "https://x.com/photographer/status/456",
				Images: []yrs.Image{
					{URL: "https://pbs.twimg.com/media/img1.jpg"},
					{URL: "https://pbs.twimg.com/media/img2.jpg"},
				},
			},
		},
	}

	formatted := formatSearchResult(result)
	assert.Contains(t, formatted, "- Images: https://pbs.twimg.com/media/img1.jpg, https://pbs.twimg.com/media/img2.jpg")
}

func TestSearchArgs(t *testing.T) {
	args := SearchArgs{Query: "test", Limit: 10}
	assert.Equal(t, "test", args.Query)
	assert.Equal(t, 10, args.Limit)

	argsDefault := SearchArgs{Query: "test"}
	assert.Equal(t, 0, argsDefault.Limit)
}

func TestBuildQuery_KeywordOnly(t *testing.T) {
	q := buildQuery(SearchArgs{Query: "golang"})
	built, err := q.Build()
	require.NoError(t, err)
	assert.Equal(t, "golang", built)
}

func TestBuildQuery_WithAllFilters(t *testing.T) {
	q := buildQuery(SearchArgs{
		Query:    "golang",
		Not:      []string{"tutorial"},
		Or:       []string{"rust", "python"},
		FromUser: "cnosuke",
		ToUser:   "someone",
		Hashtags: []string{"go"},
		URL:      "github.com",
	})
	built, err := q.Build()
	require.NoError(t, err)
	assert.Contains(t, built, "golang")
	assert.Contains(t, built, "-tutorial")
	assert.Contains(t, built, "(rust python)")
	assert.Contains(t, built, "ID:cnosuke")
	assert.Contains(t, built, "@someone")
	assert.Contains(t, built, "#go")
	assert.Contains(t, built, "URL:github.com")
}

func TestBuildQuery_FilterOnly(t *testing.T) {
	q := buildQuery(SearchArgs{FromUser: "cnosuke"})
	built, err := q.Build()
	require.NoError(t, err)
	assert.Equal(t, "ID:cnosuke", built)
}

func TestBuildQuery_EmptyQuery(t *testing.T) {
	q := buildQuery(SearchArgs{})
	_, err := q.Build()
	assert.Error(t, err)
}
