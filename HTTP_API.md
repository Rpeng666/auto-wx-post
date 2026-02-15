# HTTP API 文档

## 概览

auto-wx-post 提供 HTTP REST API，允许外部系统通过网络调用微信公众号自动发布功能。

## 启动 HTTP API 服务器

### 基本启动

```bash
# 默认端口 8080，无认证
./auto-wx-post -http

# 指定端口
./auto-wx-post -http -port=3000

# 启用 API 认证
./auto-wx-post -http -api-key=your_secret_key

# 完整示例
./auto-wx-post -http -port=8080 -api-key=my-secret-key-123
```

### 使用 Makefile

在 `Makefile` 中添加：

```makefile
run-http:
	@echo "运行 HTTP API 服务器..."
	go run $(MAIN_FILE) -http -port=8080
```

然后运行：

```bash
make run-http
```

## 认证

如果启动时指定了 `-api-key`，所有 API 请求（除 `/health`）都需要在 HTTP 头中包含认证信息：

```http
Authorization: Bearer your_secret_key
```

或简写为：

```http
Authorization: your_secret_key
```

## API 端点

### 1. 健康检查

**端点：** `GET /health`  
**认证：** 不需要  
**描述：** 检查服务器状态

**请求示例：**

```bash
curl http://localhost:8080/health
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "version": "1.0.0",
    "time": "2024-02-15T12:00:00Z"
  }
}
```

---

### 2. 列出文章

**端点：** `POST /api/articles/list`  
**认证：** 需要（如果启用）  
**描述：** 列出指定日期范围内的文章

**请求体：**

```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "show_published": false
}
```

**参数说明：**

| 参数 | 类型 | 必需 | 说明 |
|-----|------|------|------|
| `start_date` | string | 否 | 开始日期 (YYYY-MM-DD) |
| `end_date` | string | 否 | 结束日期 (YYYY-MM-DD) |
| `show_published` | boolean | 否 | 是否显示已发布文章，默认 false |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/articles/list \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "show_published": false
  }'
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "count": 5,
    "articles": [
      {
        "path": "blog-source/source/_posts/article1.md",
        "title": "我的第一篇文章",
        "author": "张三",
        "date": "2024-01-15",
        "subtitle": "这是副标题",
        "published": false
      },
      {
        "path": "blog-source/source/_posts/article2.md",
        "title": "第二篇文章",
        "author": "李四",
        "date": "2024-02-01",
        "subtitle": "",
        "published": false
      }
    ]
  }
}
```

---

### 3. 解析文章

**端点：** `POST /api/articles/parse`  
**认证：** 需要（如果启用）  
**描述：** 解析 Markdown 文章，返回元数据和内容预览

**请求体：**

```json
{
  "file_path": "blog-source/source/_posts/my-article.md"
}
```

**参数说明：**

| 参数 | 类型 | 必需 | 说明 |
|-----|------|------|------|
| `file_path` | string | 是 | Markdown 文件的完整路径 |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/articles/parse \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "blog-source/source/_posts/my-article.md"
  }'
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "title": "我的文章标题",
    "author": "作者名",
    "date": "2024-01-15",
    "subtitle": "副标题",
    "gen_cover": "false",
    "image_count": 3,
    "content_size": 1500,
    "content": "文章内容预览（前 500 字符）..."
  }
}
```

---

### 4. 上传图片

**端点：** `POST /api/images/upload`  
**认证：** 需要（如果启用）  
**描述：** 上传图片到微信公众号素材库

**请求体：**

```json
{
  "image_path": "C:\\images\\cover.jpg"
}
```

**参数说明：**

| 参数 | 类型 | 必需 | 说明 |
|-----|------|------|------|
| `image_path` | string | 是 | 图片的本地路径或远程 URL |

**请求示例：**

```bash
# 上传本地图片
curl -X POST http://localhost:8080/api/images/upload \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "image_path": "/path/to/image.jpg"
  }'

# 上传远程图片
curl -X POST http://localhost:8080/api/images/upload \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "image_path": "https://example.com/image.jpg"
  }'
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "media_id": "abc123xyz789",
    "url": "http://mmbiz.qpic.cn/..."
  }
}
```

---

### 5. 发布文章

**端点：** `POST /api/articles/publish`  
**认证：** 需要（如果启用）  
**描述：** 发布文章到微信公众号草稿箱

**请求体：**

```json
{
  "file_path": "blog-source/source/_posts/new-article.md",
  "force": false
}
```

**参数说明：**

| 参数 | 类型 | 必需 | 说明 |
|-----|------|------|------|
| `file_path` | string | 是 | Markdown 文件路径 |
| `force` | boolean | 否 | 是否强制发布（即使已发布过），默认 false |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/articles/publish \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "blog-source/source/_posts/new-article.md",
    "force": false
  }'
```

**响应示例（成功）：**

```json
{
  "success": true,
  "data": {
    "file_path": "blog-source/source/_posts/new-article.md",
    "message": "Article published successfully"
  }
}
```

**响应示例（已发布）：**

```json
{
  "success": false,
  "error": "Article already published. Use force=true to republish."
}
```

---

### 6. 获取缓存状态

**端点：** `GET /api/cache/status`  
**认证：** 需要（如果启用）  
**描述：** 获取缓存状态信息

**请求示例：**

```bash
curl http://localhost:8080/api/cache/status \
  -H "Authorization: Bearer your_secret_key"
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "size": 10,
    "count": 10
  }
}
```

---

### 7. 清空缓存

**端点：** `POST /api/cache/clear`  
**认证：** 需要（如果启用）  
**描述：** 清空所有缓存

**警告：** 这会清除所有已发布文章的记录，可能导致重复发布！

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/cache/clear \
  -H "Authorization: Bearer your_secret_key"
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "message": "Cache cleared successfully"
  }
}
```

---

## 错误响应

所有错误响应都遵循以下格式：

```json
{
  "success": false,
  "error": "错误信息描述"
}
```

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（API key 无效或缺失） |
| 405 | 请求方法不允许 |
| 409 | 冲突（如文章已发布） |
| 500 | 服务器内部错误 |

### 错误示例

```json
{
  "success": false,
  "error": "Invalid API key"
}
```

```json
{
  "success": false,
  "error": "file_path is required"
}
```

---

## 使用示例

### Python

```python
import requests

API_BASE = "http://localhost:8080"
API_KEY = "your_secret_key"

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# 列出文章
response = requests.post(
    f"{API_BASE}/api/articles/list",
    headers=headers,
    json={
        "start_date": "2024-01-01",
        "show_published": False
    }
)
print(response.json())

# 发布文章
response = requests.post(
    f"{API_BASE}/api/articles/publish",
    headers=headers,
    json={
        "file_path": "blog-source/source/_posts/article.md"
    }
)
print(response.json())
```

### JavaScript (Node.js)

```javascript
const axios = require('axios');

const API_BASE = 'http://localhost:8080';
const API_KEY = 'your_secret_key';

const headers = {
  'Authorization': `Bearer ${API_KEY}`,
  'Content-Type': 'application/json'
};

// 列出文章
axios.post(`${API_BASE}/api/articles/list`, {
  start_date: '2024-01-01',
  show_published: false
}, { headers })
  .then(res => console.log(res.data))
  .catch(err => console.error(err));

// 发布文章
axios.post(`${API_BASE}/api/articles/publish`, {
  file_path: 'blog-source/source/_posts/article.md'
}, { headers })
  .then(res => console.log(res.data))
  .catch(err => console.error(err));
```

### cURL 批量操作

```bash
#!/bin/bash

API_BASE="http://localhost:8080"
API_KEY="your_secret_key"

# 1. 列出所有未发布文章
articles=$(curl -s -X POST "$API_BASE/api/articles/list" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"show_published": false}')

echo "未发布文章: $articles"

# 2. 发布第一篇文章
curl -X POST "$API_BASE/api/articles/publish" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "blog-source/source/_posts/article1.md"
  }'
```

---

## 部署建议

### 1. 使用反向代理

生产环境建议使用 Nginx 作为反向代理：

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 2. 使用 HTTPS

```nginx
server {
    listen 443 ssl;
    server_name api.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 3. 使用 systemd 服务

创建 `/etc/systemd/system/auto-wx-post-api.service`:

```ini
[Unit]
Description=Auto WeChat Post HTTP API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/auto-wx-post
Environment="WECHAT_APP_ID=your_app_id"
Environment="WECHAT_APP_SECRET=your_app_secret"
ExecStart=/opt/auto-wx-post/auto-wx-post -http -port=8080 -api-key=your_secret_key
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl start auto-wx-post-api
sudo systemctl enable auto-wx-post-api
```

### 4. 使用 Docker

创建 `Dockerfile`:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o auto-wx-post main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/auto-wx-post .
COPY config.yaml .
EXPOSE 8080
CMD ["./auto-wx-post", "-http", "-port=8080"]
```

运行：

```bash
docker build -t auto-wx-post-api .
docker run -d -p 8080:8080 \
  -e WECHAT_APP_ID=your_app_id \
  -e WECHAT_APP_SECRET=your_app_secret \
  auto-wx-post-api
```

---

## 安全建议

1. **始终使用 API Key**: 生产环境必须启用认证
2. **使用 HTTPS**: 避免 API Key 在网络中明文传输
3. **限制访问**: 使用防火墙限制访问 IP
4. **定期更换密钥**: 定期更换 API Key
5. **日志审计**: 定期检查访问日志
6. **速率限制**: 考虑添加速率限制防止滥用

---

## 常见问题

### Q: 如何更改 API 端口？

使用 `-port` 参数：

```bash
./auto-wx-post -http -port=3000
```

### Q: 可以同时运行多个实例吗？

可以，但需要注意：
- 使用不同端口
- 共享缓存文件可能导致冲突
- 微信 API 有速率限制

### Q: 支持 WebSocket 吗？

当前版本不支持，如有需求可以提 Issue。

### Q: 如何监控服务状态？

使用 `/health` 端点进行健康检查，可以配合监控工具（如 Prometheus）使用。

---

## 更新日志

### v1.0.0 (2024-02-15)

- ✅ 初始版本
- ✅ 实现 7 个 REST API 端点
- ✅ 支持 API Key 认证
- ✅ 支持 CORS
- ✅ 结构化日志

---

## 支持

如有问题或建议，请提交 Issue 到 GitHub 仓库。
