# MCP 服务器使用指南（中文）

## 📖 什么是 MCP？

MCP (Model Context Protocol，模型上下文协议) 是 Anthropic 推出的开放标准协议，用于连接 AI 助手（如 Claude）与外部工具和数据源。

通过 MCP，你可以：
- 让 AI 助手直接操作你的微信公众号
- 用自然语言管理文章发布
- 自动化重复性工作

## 🚀 快速开始

### 第一步：构建项目

```bash
make build
# 或者
go build -o auto-wx-post.exe main.go
```

### 第二步：配置环境变量

确保设置了微信公众号的凭证：

```bash
# Windows
set WECHAT_APP_ID=你的AppID
set WECHAT_APP_SECRET=你的AppSecret

# Linux/Mac
export WECHAT_APP_ID=你的AppID
export WECHAT_APP_SECRET=你的AppSecret
```

或者在 `config.yaml` 中配置。

### 第三步：启动 MCP 服务器

```bash
# 使用 Makefile（推荐）
make run-mcp

# 或直接运行
./auto-wx-post -mcp
```

### 第四步：配置 Claude Desktop

找到 Claude Desktop 的配置文件：

**Windows:**  
`%APPDATA%\Claude\claude_desktop_config.json`

**macOS:**  
`~/Library/Application Support/Claude/claude_desktop_config.json`

**Linux:**  
`~/.config/Claude/claude_desktop_config.json`

编辑配置文件，添加以下内容（**注意修改路径**）：

```json
{
  "mcpServers": {
    "auto-wx-post": {
      "command": "C:\\path\\to\\auto-wx-post.exe",
      "args": ["-mcp"],
      "env": {
        "WECHAT_APP_ID": "你的AppID",
        "WECHAT_APP_SECRET": "你的AppSecret"
      }
    }
  }
}
```

**重要提示：**
- Windows 路径要使用双反斜杠 `\\` 或单斜杠 `/`
- 路径必须是**绝对路径**，不能使用相对路径
- 保存后需要**重启 Claude Desktop**

### 第五步：开始使用

重启 Claude Desktop 后，你就可以直接和 Claude 对话来管理公众号了！

## 💬 使用示例

### 1. 列出待发布的文章

```
请帮我列出最近 7 天内所有未发布的文章
```

或者

```
List all unpublished articles from 2024-01-01 to 2024-12-31
```

### 2. 解析文章内容

```
帮我看看 blog-source/source/_posts/我的文章.md 这篇文章的信息
```

Claude 会返回：
- 文章标题
- 作者
- 发布日期
- 副标题
- 图片数量
- 内容预览

### 3. 上传图片

```
把 D:\images\封面.jpg 上传到微信公众号
```

Claude 会返回：
- Media ID（素材 ID）
- 图片 URL

### 4. 发布文章

```
发布文章 blog-source/source/_posts/新文章.md 到微信公众号草稿箱
```

Claude 会自动：
1. 解析 Markdown 文件
2. 上传文章中的所有图片
3. 转换 Markdown 为微信支持的 HTML
4. 应用 CSS 样式美化
5. 添加到草稿箱

### 5. 查看缓存状态

```
查看缓存状态
```

或者

```
Show cache status
```

### 6. 清空缓存

```
清空缓存
```

**注意：** 清空缓存后，已发布的文章记录会被删除，可能导致重复发布。

## 🛠️ 可用工具

| 工具名称 | 功能说明 | 必需参数 | 可选参数 |
|---------|---------|---------|---------|
| **list_articles** | 列出文章 | - | `start_date`, `end_date`, `show_published` |
| **parse_article** | 解析文章 | `file_path` | - |
| **upload_image** | 上传图片 | `image_path` | - |
| **publish_article** | 发布文章 | `file_path` | `force` |
| **get_cache_status** | 查看缓存 | - | - |
| **clear_cache** | 清空缓存 | - | - |

### 工具详细说明

#### list_articles - 列出文章
列出指定日期范围内的 Markdown 文章。

**参数：**
- `start_date` (可选): 开始日期，格式 `YYYY-MM-DD`，留空使用配置中的 `days_before`
- `end_date` (可选): 结束日期，格式 `YYYY-MM-DD`，留空使用配置中的 `days_after`
- `show_published` (可选): 是否显示已发布的文章，默认 `false`

**示例：**
```
列出 2024-01-01 到 2024-12-31 之间所有未发布的文章
```

#### parse_article - 解析文章
解析 Markdown 文件，返回文章的元数据和内容预览。

**参数：**
- `file_path` (必需): Markdown 文件的完整路径

**示例：**
```
解析文章 blog-source/source/_posts/测试文章.md
```

#### upload_image - 上传图片
上传单张图片到微信公众号素材库。

**参数：**
- `image_path` (必需): 图片的本地路径或远程 URL

**示例：**
```
上传图片 C:\Users\myname\Pictures\cover.jpg
```

#### publish_article - 发布文章
发布文章到微信公众号草稿箱。

**参数：**
- `file_path` (必需): Markdown 文件路径
- `force` (可选): 强制发布，即使已发布过，默认 `false`

**示例：**
```
发布文章 blog-source/source/_posts/新文章.md
```

强制重新发布：
```
强制发布文章 blog-source/source/_posts/已发布的文章.md
```

#### get_cache_status - 查看缓存
获取缓存状态，包括已发布的文章数量。

**示例：**
```
显示缓存状态
```

#### clear_cache - 清空缓存
清空所有缓存。**警告：** 这会清除已发布文章的记录！

**示例：**
```
清空所有缓存
```

## 🔧 故障排除

### Claude 中看不到 MCP 工具

**解决方法：**
1. 确认配置文件路径正确
2. 确认 `auto-wx-post` 的路径是**绝对路径**
3. 确认配置文件 JSON 格式正确（可以用 JSON 验证器检查）
4. 重启 Claude Desktop
5. 查看 Claude Desktop 的日志文件

**Windows 日志位置:**  
`%APPDATA%\Claude\logs\mcp*.log`

**macOS 日志位置:**  
`~/Library/Logs/Claude/mcp*.log`

### 权限错误

**Windows:**
确保 `auto-wx-post.exe` 可以执行

**Linux/Mac:**
```bash
chmod +x auto-wx-post
```

### Token 错误

确认环境变量设置正确：
```bash
# Windows
echo %WECHAT_APP_ID%
echo %WECHAT_APP_SECRET%

# Linux/Mac
echo $WECHAT_APP_ID
echo $WECHAT_APP_SECRET
```

### 测试 MCP 服务器

手动测试 MCP 服务器是否正常：

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./auto-wx-post -mcp
```

应该返回类似：
```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05",...}}
```

## 📊 工作流程

```
┌─────────────────────────────────────────────┐
│          用户（通过 Claude 对话）              │
└───────────────┬─────────────────────────────┘
                │
                │ 自然语言请求
                ▼
┌─────────────────────────────────────────────┐
│           Claude Desktop (MCP 客户端)         │
│  - 理解用户意图                                │
│  - 选择合适的工具                              │
│  - 构造工具参数                                │
└───────────────┬─────────────────────────────┘
                │
                │ JSON-RPC over stdio
                ▼
┌─────────────────────────────────────────────┐
│      auto-wx-post MCP 服务器 (本项目)         │
│  - 接收工具调用请求                            │
│  - 执行相应操作                                │
│  - 返回结果                                    │
└───────────────┬─────────────────────────────┘
                │
        ┌───────┴────────┐
        ▼                ▼
┌──────────────┐  ┌──────────────┐
│  本地文件系统  │  │  微信公众号 API │
│  - 读取文章    │  │  - 上传素材     │
│  - 解析 MD     │  │  - 添加草稿     │
└──────────────┘  └──────────────┘
```

## 🎯 使用场景

### 场景 1：批量发布文章

```
我需要发布最近 3 天内所有未发布的文章到草稿箱
```

Claude 会：
1. 列出最近 3 天的未发布文章
2. 逐个解析和发布每篇文章
3. 报告成功和失败的文章

### 场景 2：检查待发布内容

```
帮我看看有哪些文章还没发布，以及它们的标题和日期
```

Claude 会列出所有未发布文章的详细信息。

### 场景 3：重新发布文章

```
我修改了文章 blog-source/source/_posts/旧文章.md，请强制重新发布
```

Claude 会使用 `force=true` 参数重新发布。

### 场景 4：上传多张图片

```
帮我把 D:\images\ 目录下的所有 jpg 图片上传到微信
```

Claude 会遍历目录并逐个上传图片。

## 🌟 高级技巧

### 1. 批量操作

你可以要求 Claude 执行批量操作：
```
列出所有未发布的文章，然后发布前 5 篇
```

### 2. 条件发布

```
列出所有未发布的文章，只发布标题包含"教程"的文章
```

### 3. 定期检查

```
每天早上 9 点检查是否有新文章需要发布
```
（需要配合系统定时任务）

### 4. 智能分析

```
分析最近发布的 10 篇文章，统计使用的图片数量和平均字数
```

## 📚 相关资源

- [MCP 官方文档](https://modelcontextprotocol.io/)
- [Claude Desktop 下载](https://claude.ai/download)
- [微信公众平台开发文档](https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html)

## 📝 注意事项

1. **数据安全**: MCP 服务器会访问你的微信公众号，请保管好 AppID 和 AppSecret
2. **API 限制**: 微信公众号 API 有调用频率限制，避免短时间内大量请求
3. **缓存机制**: 清空缓存后可能导致文章重复发布，请谨慎操作
4. **文件路径**: 在 Windows 上使用路径时注意转义字符
5. **网络连接**: MCP 服务器需要能访问微信 API，确保网络畅通

## 🆘 获取帮助

如果遇到问题：

1. 查看日志文件 `logs/app.log`
2. 检查 Claude Desktop 的 MCP 日志
3. 提交 Issue 到 GitHub 仓库
4. 查阅 `MCP_README.md` 英文文档

---

**祝你使用愉快！如果觉得有用，请给项目点个 Star ⭐**
