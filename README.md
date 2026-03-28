# mcp-yahoo-realtime-search

An MCP (Model Context Protocol) server that exposes Yahoo! Japan Realtime Search (tweet search) via the [go-yahoo-realtime-search](https://github.com/cnosuke/go-yahoo-realtime-search) library.

No API key required - uses web scraping under the hood.

## Installation

```bash
go install github.com/cnosuke/mcp-yahoo-realtime-search@latest
```

## Usage

### STDIO mode (for Claude Desktop, Claude Code, etc.)

```bash
mcp-yahoo-realtime-search stdioserver -c config.yml
```

### HTTP mode (Streamable HTTP transport)

```bash
mcp-yahoo-realtime-search httpserver -c config.yml
```

## Claude Desktop Configuration

```json
{
  "mcpServers": {
    "yahoo-realtime-search": {
      "command": "/path/to/mcp-yahoo-realtime-search",
      "args": ["stdioserver", "-c", "/path/to/config.yml"]
    }
  }
}
```

## Tools

### `search`

Search tweets via Yahoo! Japan Realtime Search.

| Parameter | Type   | Required | Description                                       |
|-----------|--------|----------|---------------------------------------------------|
| `query`   | string | Yes      | Search keyword or phrase                           |
| `limit`   | number | No       | Maximum number of tweets to return (0 = all)       |

## Configuration

See `config.yml` for all available options. Environment variables can override config file values:

| Env Var              | Config Key             |
|----------------------|------------------------|
| `LOG_PATH`           | `log`                  |
| `LOG_LEVEL`          | `log_level`            |
| `HTTP_BINDING`       | `http.binding`         |
| `HTTP_ENDPOINT_PATH` | `http.endpoint_path`   |
| `HTTP_AUTH_TOKEN`    | `http.auth_token`      |
| `HTTP_ALLOWED_ORIGINS` | `http.allowed_origins` |
| `YRS_USER_AGENT`     | `search.user_agent`    |
| `YRS_REQUEST_TIMEOUT` | `search.request_timeout` |

## License

MIT
