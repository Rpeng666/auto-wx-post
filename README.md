# 微信公众号自动发布工具 (Go版)

这是一个用Go语言写的微信公众号文章自动发布工具，

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

```bash
# 正常运行
go run main.go

# 使用自定义配置文件
go run main.go -config=custom_config.yaml

# 模拟运行 (不实际发布)
go run main.go -dry-run

# 清空缓存
go run main.go -clear-cache
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
