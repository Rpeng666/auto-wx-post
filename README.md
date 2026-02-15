# 微信公众号自动发布工具 (Go版)

这是一个用Go语言编写的微信公众号文章自动发布工具，支持三种使用方式：

- 📝 **命令行模式**: 批量扫描和发布文章
- 🤖 **MCP 服务器**: 通过 AI 助手（如 Claude Desktop）集成
- 🌐 **HTTP API**: 提供 RESTful API 供外部系统调用

吐槽：Fuck weixin！发布工具这么不好用，就那封闭的公众号生态怎么好得起来？

## ✨ 主要改进

### 架构优化
- **模块化设计**: 采用清晰的包结构，职责分离
- **并发处理**: 使用goroutine并发上传图片，提高效率
- **单例模式**: 微信客户端使用单例，避免重复初始化
- **线程安全**: 缓存管理器使用互斥锁保证并发安全

### 功能增强
- **智能Token管理**: 自动刷新，提前5分钟过期避免边界问题
- **重试机制**: 支持指数退避的自动重试
- **上下文管理**: 支持超时控制和取消操作
- **资源清理**: 自动清理临时文件，防止泄露
- **结构化日志**: 使用slog提供JSON/Text格式日志

### 错误处理
- 完善的错误包装和传播
- 优雅的错误恢复机制
- 详细的日志记录

## 📁 项目结构

```
auto-wx-post/
├── main.go                    # 主程序入口
├── config.yaml                # 配置文件
├── go.mod                     # 依赖管理
├── internal/                  # 内部包
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── wechat/               # 微信客户端
│   │   ├── client.go         # 客户端和Token管理
│   │   └── media.go          # 素材管理
│   ├── cache/                # 缓存管理
│   │   └── manager.go
│   ├── media/                # 媒体管理
│   │   └── manager.go
│   ├── markdown/             # Markdown处理
│   │   ├── parser.go         # 解析器
│   │   └── beautifier.go     # HTML美化
│   ├── publisher/            # 发布器
│   │   └── publisher.go
│   ├── mcp/                  # MCP服务器
│   │   ├── types.go          # 协议类型定义
│   │   ├── server.go         # 服务器实现
│   │   └── handler.go        # stdio处理器
│   ├── api/                  # HTTP API服务器
│   │   └── server.go         # RESTful API实现
│   └── logger/               # 日志
│       └── logger.go
└── assets/                    # CSS模板 (可选)
    ├── para.tmpl
    ├── sub.tmpl
    ├── link.tmpl
    ├── ref_header.tmpl
    ├── ref_link.tmpl
    ├── figure.tmpl
    ├── code.tmpl
    └── header.tmpl
```

## 🚀 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量

```bash
# Windows
set WECHAT_APP_ID=your_app_id
set WECHAT_APP_SECRET=your_app_secret

# Linux/Mac
export WECHAT_APP_ID=your_app_id
export WECHAT_APP_SECRET=your_app_secret
```

### 3. 修改配置文件

编辑 `config.yaml` 文件，根据需要调整配置。

### 4. 运行程序

#### 命令行模式

```bash
# 正常运行（批量发布）
go run main.go

# 使用自定义配置文件
go run main.go -config=custom_config.yaml

# 模拟运行 (不实际发布)
go run main.go -dry-run

# 清空缓存
go run main.go -clear-cache
```

#### MCP 服务器模式（AI 助手集成）

```bash
# 启动 MCP 服务器（用于 Claude Desktop 等）
go run main.go -mcp
```

#### HTTP API 服务器模式（外部调用）

```bash
# 启动 HTTP API（默认端口 8080，无认证）
go run main.go -http

# 指定端口
go run main.go -http -port=3000

# 启用 API 认证
go run main.go -http -api-key=your_secret_key

# 完整示例
go run main.go -http -port=8080 -api-key=my-secret-123
```

### 使用 Makefile（推荐）

```bash
# 构建项目
make build

# 运行项目
make run

# 模拟运行
make run-dry

# 运行 MCP 服务器
make run-mcp

# 运行 HTTP API 服务器
make run-http

# 运行 HTTP API（带认证）
make run-http-auth

# 清空缓存
make clear-cache

# 查看所有命令
make help
```

### 使用 Makefile（推荐）

```bash
# 构建项目
make build

# 运行项目
make run

# 模拟运行
make run-dry

# 运行 MCP 服务器
make run-mcp

# 清空缓存
make clear-cache

# 运行测试
make test

# 代码格式化
make fmt

# 查看所有命令
make help
```

### 5. 编译

```bash
# 编译当前平台
go build -o auto-wx-post.exe

# 交叉编译 Linux
GOOS=linux GOARCH=amd64 go build -o auto-wx-post

# 交叉编译 Mac
GOOS=darwin GOARCH=amd64 go build -o auto-wx-post
```

## ⚙️ 配置说明

### config.yaml 主要配置项

```yaml
wechat:
  app_id: "${WECHAT_APP_ID}"        # 微信公众号AppID
  app_secret: "${WECHAT_APP_SECRET}" # 微信公众号AppSecret

blog:
  source_path: "./blog-source/source/_posts"  # 博客文章目录
  base_url: "https://fuckweixin.com/p/"        # 文章基础URL
  author: "fuckweixin"                            # 默认作者

cache:
  store_file: "cache.json"  # 缓存文件路径

image:
  temp_dir: "./temp"                          # 临时文件目录
  placeholder_service: "https://picsum.photos/seed"
  default_cover_size: "400/600"               # 默认封面尺寸

publish:
  days_before: 7              # 扫描过去7天的文章
  days_after: 2               # 扫描未来2天的文章
  concurrent_uploads: 5       # 并发上传图片数
  max_retries: 3              # 最大重试次数
  timeout: 30                 # 请求超时(秒)

log:
  level: "info"               # debug, info, warn, error
  format: "json"              # json, text
  output: "stdout"            # stdout, file
  file_path: "./logs/app.log" # 日志文件路径
```

## 🎯 主要特性

### 1. Token自动管理
- 自动获取和刷新access_token
- 提前5分钟刷新避免过期
- 线程安全的token缓存

### 2. 并发图片上传
- 使用goroutine池并发上传
- 可配置并发数量
- 自动错误收集和处理

### 3. 智能缓存
- 基于文件MD5的缓存机制
- 避免重复上传已处理的文章
- 图片URL缓存减少API调用

### 4. 重试机制
- HTTP请求自动重试
- 指数退避策略
- 可配置重试次数

### 5. 资源管理
- 自动清理临时文件
- 优雅的资源释放
- 防止资源泄露

### 6. 日志系统
- 结构化日志输出
- 支持JSON/Text格式
- 可配置日志级别
- 支持文件和控制台输出

### 7. 🆕 MCP 服务器 (AI 助手集成)
- 实现 Model Context Protocol 规范
- 支持 Claude Desktop 等 AI 助手调用
- 提供 6 个实用工具：
  - **list_articles** - 列出待发布文章
  - **parse_article** - 解析文章元数据
  - **upload_image** - 上传图片到微信
  - **publish_article** - 发布文章到草稿箱
  - **get_cache_status** - 查看缓存状态
  - **clear_cache** - 清空缓存

### 8. 🆕 HTTP API (外部系统集成)
- RESTful API 接口
- 支持 API Key 认证
- CORS 跨域支持
- 提供 7 个端点：
  - `GET /health` - 健康检查
  - `POST /api/articles/list` - 列出文章
  - `POST /api/articles/parse` - 解析文章
  - `POST /api/articles/publish` - 发布文章
  - `POST /api/images/upload` - 上传图片
  - `GET /api/cache/status` - 缓存状态
  - `POST /api/cache/clear` - 清空缓存

## 🤖 MCP 服务器使用指南

### 什么是 MCP？

MCP (Model Context Protocol) 是 Anthropic 推出的开放协议，允许 AI 助手（如 Claude）连接到外部工具和数据源。通过 MCP，你可以让 AI 助手帮你管理微信公众号文章。

### 快速开始

#### 1. 启动 MCP 服务器

```bash
# 方式 1: 使用 Makefile
make run-mcp

# 方式 2: 直接运行
./auto-wx-post -mcp

# 方式 3: 使用 go run
go run main.go -mcp
```

#### 2. 配置 Claude Desktop

编辑 Claude Desktop 配置文件：

**macOS/Linux:**  
`~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows:**  
`%APPDATA%\Claude\claude_desktop_config.json`

添加以下配置：

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

#### 3. 使用 AI 助手管理文章

配置完成后，重启 Claude Desktop，然后你就可以：

**列出文章：**
```
列出从 2024-01-01 到现在所有未发布的文章
```

**解析文章：**
```
帮我解析 blog-source/source/_posts/my-article.md 这篇文章
```

**上传图片：**
```
把 /path/to/image.jpg 上传到微信公众号
```

**发布文章：**
```
发布文章 blog-source/source/_posts/new-post.md 到微信公众号
```

**查看状态：**
```
显示缓存状态
```

### MCP 工具详情

| 工具名称 | 描述 | 参数 |
|---------|------|------|
| `list_articles` | 列出指定日期范围的文章 | `start_date`, `end_date`, `show_published` |
| `parse_article` | 解析 Markdown 文章 | `file_path` (必需) |
| `upload_image` | 上传图片到微信 | `image_path` (必需) |
| `publish_article` | 发布文章到草稿箱 | `file_path` (必需), `force` |
| `get_cache_status` | 查看缓存状态 | 无 |
| `clear_cache` | 清空缓存 | 无 |

详细文档请查看：
- [MCP_README.md](MCP_README.md) - 英文文档
- [MCP_使用指南.md](MCP_使用指南.md) - 中文详细指南




## 🔧 开发指南

### 添加新的CSS模板

在 `assets/` 目录下创建 `.tmpl` 文件，使用Go的格式化字符串语法：

```html
<!-- para.tmpl -->
<p style="margin: 10px 0; line-height: 1.75em; color: #333;">

<!-- sub.tmpl -->
<h%s style="font-size: %dpx; font-weight: bold; margin: 20px 0;">%s</h%s>
```

### 扩展功能

1. **添加新的素材类型**: 在 `wechat/media.go` 中扩展
2. **自定义渲染器**: 在 `markdown/beautifier.go` 中添加
3. **新的缓存策略**: 修改 `cache/manager.go`

## 🐛 故障排除

### 问题：Token获取失败
- 检查环境变量是否正确设置
- 验证AppID和AppSecret的有效性
- 检查网络连接

### 问题：图片上传失败
- 检查图片URL是否可访问
- 验证图片格式和大小限制
- 查看日志中的详细错误信息

### 问题：文章未找到
- 检查 `blog.source_path` 配置
- 确认文章的date字段格式正确
- 检查文件权限

## 📄 License

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request！
