# AGENTS.md - Coding Agent Guidelines

This document provides essential information for AI coding agents working in this repository.

## Project Overview

**Language:** Go 1.24.0  
**Type:** CLI application for WeChat Official Account automation  
**Architecture:** Internal packages with clear separation of concerns

## Build & Development Commands

### Building
```bash
make build              # Build for current platform
make build-release      # Build optimized release version
make build-all          # Cross-compile for all platforms (Linux, Windows, macOS)
make run                # Run the application
make run-dry            # Run in dry-run mode (no publishing)
make run-mcp            # Run as MCP (Model Context Protocol) server
```

### Testing
```bash
make test               # Run all tests: go test -v ./...
make test-coverage      # Generate HTML coverage report

# Run a single test
go test -v ./internal/config -run TestFunctionName
go test -v ./internal/wechat -run TestTokenRefresh
```

### Code Quality
```bash
make fmt                # Format code: go fmt ./...
make lint               # Run golangci-lint (must be installed)
go vet ./...            # Run Go vet for suspicious constructs
```

### Dependency Management
```bash
make deps               # Download and tidy dependencies
go mod tidy             # Clean up go.mod and go.sum
go mod vendor           # Vendor dependencies (if needed)
```

### Cleanup
```bash
make clean              # Remove build artifacts, cache, temp files
make clear-cache        # Clear application cache only
make clean-all          # Full cleanup including Go build cache
```

## Code Style Guidelines

### File Organization
- Use `internal/` for packages that should not be imported by external projects
- One package per directory with clear single responsibility
- Package names: lowercase, single word (e.g., `cache`, `wechat`, `publisher`)
- Test files: `*_test.go` in the same directory as the code

### Naming Conventions
- **Types:** PascalCase (e.g., `Manager`, `Client`, `Article`)
- **Exported functions:** PascalCase (e.g., `NewManager`, `UploadImage`)
- **Unexported functions:** camelCase (e.g., `refreshToken`, `validateConfig`)
- **Constants:** PascalCase (e.g., `MediaTypeImage`, `DefaultTimeout`)
- **Variables:** camelCase for locals, PascalCase for package-level exports
- **Interfaces:** Noun or adjective form (e.g., `Publisher`, `Cacheable`)

### Import Organization
Group imports in standard Go order:
```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "time"

    // 2. External dependencies
    "github.com/PuerkitoBio/goquery"
    "gopkg.in/yaml.v3"

    // 3. Internal packages
    "github.com/yourusername/auto-wx-post/internal/config"
    "github.com/yourusername/auto-wx-post/internal/logger"
)
```

### Types and Structs
```go
// Document types with clear descriptions
type Client struct {
    appID       string          // Unexported fields: camelCase
    appSecret   string
    token       string
    tokenExpiry time.Time
    mu          sync.RWMutex    // Protect concurrent access
    once        sync.Once       // For singleton initialization
}

// Config structs with yaml tags
type Config struct {
    AppID     string `yaml:"app_id"`
    AppSecret string `yaml:"app_secret"`
}
```

### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to upload image: %w", err)
}

// Log non-fatal errors as warnings
if err := cleanup(); err != nil {
    slog.Warn("cleanup failed", "error", err)
}

// Don't panic in library code - return errors
// Panic only in main.go for unrecoverable initialization failures
```

### Context Usage
```go
// Always accept context.Context as first parameter
func (c *Client) UploadImage(ctx context.Context, imagePath string) (string, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }
    
    // Pass context to downstream calls
    resp, err := c.makeRequest(ctx, url, data)
}
```

### Logging
Use structured logging with `log/slog`:
```go
slog.Info("uploading image", "path", imagePath, "size", fileSize)
slog.Warn("retry attempt", "attempt", i, "error", err)
slog.Error("upload failed", "error", err, "mediaID", mediaID)
slog.Debug("cache hit", "key", cacheKey)
```

### Concurrency Patterns
```go
// Use sync.RWMutex for shared state
c.mu.RLock()
token := c.token
c.mu.RUnlock()

// Goroutine pools with semaphores
sem := make(chan struct{}, maxConcurrency)
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        sem <- struct{}{}        // Acquire
        defer func() { <-sem }() // Release
        processItem(ctx, item)
    }(item)
}
wg.Wait()
```

### Resource Management
```go
// Always use defer for cleanup
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()

// Explicit cleanup functions when defer isn't enough
func (m *Manager) Close() error {
    m.cleanup()
    return m.saveCache()
}
```

### Testing Best Practices
```go
// Table-driven tests
func TestParseMarkdown(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Article
        wantErr bool
    }{
        {name: "valid markdown", input: "---\ntitle: Test\n---\n# Hello", want: &Article{Title: "Test"}, wantErr: false},
        {name: "no frontmatter", input: "# Hello", want: nil, wantErr: true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseMarkdown(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseMarkdown() error = %v, wantErr %v", err, tt.wantErr)
            }
            // Add more assertions
        })
    }
}

// Use testdata/ for fixtures
// Mock external dependencies (e.g., WeChat API)
```

## Configuration

- **Main config:** `config.yaml` with environment variable expansion `${VAR_NAME}`
- **Environment variables:** Use `.env` file (see `.env.example`)
- **Sensitive data:** Never commit secrets - use environment variables
- **Cache:** `cache.json` (gitignored) for article MD5 tracking

## Project Structure

```
internal/
├── config/         # Configuration management
├── wechat/         # WeChat API client (token, media upload)
├── cache/          # Thread-safe cache manager
├── media/          # Media file handling
├── markdown/       # MD parser with YAML front matter
├── publisher/      # Article publishing orchestration
├── mcp/            # MCP (Model Context Protocol) server
└── logger/         # Structured logging setup
```

## Common Pitfalls

1. **Don't use `panic` in library code** - return errors instead
2. **Always check context cancellation** in long-running operations
3. **Protect shared state with mutexes** - this codebase uses concurrent operations
4. **Close all resources** - use defer for files, HTTP responses, etc.
5. **Test token expiry logic** - WeChat tokens expire after 2 hours
6. **Handle UTF-8 BOM** - markdown parser strips BOM bytes
7. **Clean up temp files** - images are downloaded to `temp/` directory

## Git Workflow

- **Branch:** Create feature branches from `main`
- **Commits:** Write clear, descriptive commit messages
- **Testing:** Run `make test` before committing (add tests if missing)
- **Formatting:** Run `make fmt` before committing
- **CI/CD:** GitHub Actions builds releases on `v*` tags

## External Dependencies

- `github.com/PuerkitoBio/goquery` - HTML/XML parsing
- `github.com/gomarkdown/markdown` - Markdown to HTML conversion
- `gopkg.in/yaml.v3` - YAML parsing for config and front matter

## Notes for Agents

- **Language:** Code is in English, docs/comments mix Chinese and English
- **No tests currently exist** - write tests when modifying code
- **Token management is critical** - auto-refreshes 5 minutes before expiry
- **Concurrent uploads** - respects rate limits with semaphore pattern
- **File encoding:** Always UTF-8, CRLF normalized to LF

## MCP Server (Model Context Protocol)

This application can run as an MCP server, allowing AI assistants to interact with WeChat automation tools.

### Starting MCP Server

```bash
./auto-wx-post -mcp        # Start MCP server
make run-mcp               # Or use Makefile
```

### Available MCP Tools

1. **list_articles** - List Markdown articles in date range
2. **parse_article** - Parse article metadata and content
3. **upload_image** - Upload single image to WeChat
4. **publish_article** - Publish article to WeChat draft box
5. **get_cache_status** - View cache status
6. **clear_cache** - Clear all cache

### Claude Desktop Configuration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "auto-wx-post": {
      "command": "/path/to/auto-wx-post",
      "args": ["-mcp"],
      "env": {
        "WECHAT_APP_ID": "your_app_id",
        "WECHAT_APP_SECRET": "your_app_secret"
      }
    }
  }
}
```

See `MCP_README.md` for detailed documentation.

