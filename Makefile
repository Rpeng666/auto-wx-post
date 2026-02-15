# Makefile for auto-wx-post

.PHONY: all build run clean test deps help

# 变量定义
BINARY_NAME=auto-wx-post
MAIN_FILE=main.go
BUILD_DIR=./build

# 默认目标
all: deps build

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod download
	go mod tidy

# 构建项目
build:
	@echo "构建项目..."
	go build -o $(BINARY_NAME) $(MAIN_FILE)

# 构建到指定目录
build-release:
	@echo "构建发布版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# 交叉编译
build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)

build-windows:
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_FILE)

build-mac:
	@echo "构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-mac $(MAIN_FILE)

build-all: build-linux build-windows build-mac
	@echo "所有平台构建完成"

# 运行项目
run:
	@echo "运行项目..."
	go run $(MAIN_FILE)

# 运行（模拟模式）
run-dry:
	@echo "运行项目（模拟模式）..."
	go run $(MAIN_FILE) -dry-run

# 运行 MCP 服务器
run-mcp:
	@echo "运行 MCP 服务器..."
	go run $(MAIN_FILE) -mcp

# 运行 HTTP API 服务器
run-http:
	@echo "运行 HTTP API 服务器..."
	go run $(MAIN_FILE) -http -port=8080

# 运行 HTTP API 服务器（带认证）
run-http-auth:
	@echo "运行 HTTP API 服务器（带认证）..."
	go run $(MAIN_FILE) -http -port=8080 -api-key=dev-secret-key

# 清空缓存
clear-cache:
	@echo "清空缓存..."
	go run $(MAIN_FILE) -clear-cache

# 测试
test:
	@echo "运行测试..."
	go test -v ./...

# 测试覆盖率
test-coverage:
	@echo "生成测试覆盖率..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码格式化
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "运行代码检查..."
	golangci-lint run

# 清理
clean:
	@echo "清理构建文件..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -rf temp/
	rm -f cache.json
	rm -f result.html origi.html

# 完全清理（包括依赖）
clean-all: clean
	@echo "清理所有文件..."
	go clean -cache -modcache -testcache

# 安装到系统
install:
	@echo "安装到系统..."
	go install

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  make deps           - 安装依赖"
	@echo "  make build          - 构建项目"
	@echo "  make build-release  - 构建优化版本"
	@echo "  make build-all      - 构建所有平台版本"
	@echo "  make run            - 运行项目"
	@echo "  make run-dry        - 模拟运行"
	@echo "  make run-mcp        - 运行 MCP 服务器"
	@echo "  make clear-cache    - 清空缓存"
	@echo "  make test           - 运行测试"
	@echo "  make test-coverage  - 生成测试覆盖率"
	@echo "  make fmt            - 格式化代码"
	@echo "  make lint           - 代码检查"
	@echo "  make clean          - 清理构建文件"
	@echo "  make clean-all      - 清理所有文件"
	@echo "  make install        - 安装到系统"
	@echo "  make help           - 显示帮助信息"
