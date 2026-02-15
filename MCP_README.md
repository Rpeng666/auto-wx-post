# MCP Server for WeChat Official Account Automation

This project implements an MCP (Model Context Protocol) server that allows AI assistants like Claude to interact with WeChat Official Account automation tools.

## Quick Start

### 1. Build the application

```bash
make build
```

### 2. Configure your WeChat credentials

Edit `config.yaml` or set environment variables:

```bash
export WECHAT_APP_ID="your_app_id"
export WECHAT_APP_SECRET="your_app_secret"
```

### 3. Start the MCP server

```bash
./auto-wx-post -mcp
```

Or use the Makefile:

```bash
make run-mcp
```

## Configuration for Claude Desktop

Add this to your Claude Desktop configuration file:

**macOS/Linux:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "auto-wx-post": {
      "command": "/path/to/auto-wx-post",
      "args": ["-mcp"],
      "env": {
        "WECHAT_APP_ID": "your_app_id_here",
        "WECHAT_APP_SECRET": "your_app_secret_here"
      }
    }
  }
}
```

## Available MCP Tools

### 1. list_articles

列出指定日期范围内的 Markdown 文章。

**Parameters:**
- `start_date` (optional): 开始日期 (YYYY-MM-DD)
- `end_date` (optional): 结束日期 (YYYY-MM-DD)
- `show_published` (optional): 是否显示已发布的文章 (默认: false)

**Example:**
```
List all unpublished articles from 2024-01-01 to 2024-12-31
```

### 2. parse_article

解析指定的 Markdown 文章，返回元数据和内容预览。

**Parameters:**
- `file_path` (required): Markdown 文件的完整路径

**Example:**
```
Parse the article at /path/to/article.md
```

### 3. upload_image

上传图片到微信公众号。

**Parameters:**
- `image_path` (required): 本地路径或远程 URL

**Example:**
```
Upload image from /path/to/image.jpg
```

### 4. publish_article

发布文章到微信公众号草稿箱。

**Parameters:**
- `file_path` (required): Markdown 文件路径
- `force` (optional): 强制发布，即使已经发布过 (默认: false)

**Example:**
```
Publish the article at /path/to/article.md
```

### 5. get_cache_status

获取缓存状态，查看已发布的文章数量。

**Parameters:** None

**Example:**
```
Show cache status
```

### 6. clear_cache

清空缓存。警告：这将清除所有已发布文章的记录。

**Parameters:** None

**Example:**
```
Clear the cache
```

## Usage Examples with Claude

Once configured, you can ask Claude to:

1. **List articles:**
   ```
   List all unpublished articles from last month
   ```

2. **Parse and review:**
   ```
   Parse the article at blog-source/source/_posts/my-article.md and show me the metadata
   ```

3. **Upload images:**
   ```
   Upload all images from my-images/ directory to WeChat
   ```

4. **Publish:**
   ```
   Publish the article blog-source/source/_posts/new-post.md to WeChat
   ```

5. **Check status:**
   ```
   Show me the cache status and list all published articles
   ```

## Troubleshooting

### MCP server not appearing in Claude

1. Check that the path to `auto-wx-post` is correct and absolute
2. Verify your WeChat credentials are set
3. Restart Claude Desktop after updating the configuration
4. Check Claude Desktop logs for errors

### Permission denied errors

Make sure the executable has proper permissions:

```bash
chmod +x auto-wx-post
```

### Connection issues

Test the MCP server manually:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./auto-wx-post -mcp
```

## Architecture

The MCP server exposes the following functionality:

```
┌─────────────────┐
│   AI Assistant  │ (Claude, etc.)
│   (MCP Client)  │
└────────┬────────┘
         │ JSON-RPC over stdio
         │
┌────────▼────────┐
│   MCP Handler   │ (internal/mcp/handler.go)
└────────┬────────┘
         │
┌────────▼────────┐
│   MCP Server    │ (internal/mcp/server.go)
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┐
    │         │          │          │
┌───▼───┐ ┌──▼───┐ ┌────▼────┐ ┌──▼──────┐
│Parser │ │Media │ │Publisher│ │Cache    │
└───────┘ └──────┘ └─────────┘ └─────────┘
```

## Protocol Details

This implementation follows the MCP specification:
- Protocol Version: 2024-11-05
- Transport: stdio (standard input/output)
- Format: JSON-RPC 2.0

## License

Same as the main project.
