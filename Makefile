# ==== 基本配置 ====
APP_NAME_MAC = wxt_mac
APP_NAME_LNX = wxt_lnx
FRONTEND_DIR = frontend
SERVER_DIR   = server
RELEASE_DIR  = release
STATIC_DIR   = $(RELEASE_DIR)/static

# 自动检测系统（保留你已有的，用于单平台 build）
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	BIN_NAME := $(APP_NAME_MAC)
endif
ifeq ($(UNAME_S),Linux)
	BIN_NAME := $(APP_NAME_LNX)
endif

# ---- 新增：默认架构，可通过命令行覆盖 ----
UNAME_M := $(shell uname -m)
# Mac 默认根据本机架构决定（arm64/amd64）
ifeq ($(UNAME_M),arm64)
	DEFAULT_MAC_ARCH = arm64
else
	DEFAULT_MAC_ARCH = amd64
endif

MAC_ARCH ?= $(DEFAULT_MAC_ARCH)  # 覆盖: make build-mac MAC_ARCH=amd64
LNX_ARCH ?= amd64                # 覆盖: make build-linux LNX_ARCH=arm64

# ==== 构建前端 ====
build-frontend:
	cd frontend && npm ci --legacy-peer-deps && npm run build

# ==== 同步静态文件 ====
sync-dist:build-frontend
	mkdir -p $(STATIC_DIR)
	rm -rf $(STATIC_DIR)/*
	cp -r $(FRONTEND_DIR)/dist/* $(STATIC_DIR)/

# ==== 构建后端（当前系统）====
build-server:
	mkdir -p $(RELEASE_DIR)
	cd $(SERVER_DIR) && GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -ldflags "-s -w" -o ../$(RELEASE_DIR)/$(BIN_NAME)

# ---- 新增：分别为 mac / linux 交叉编译 ----
build-mac:
	mkdir -p $(RELEASE_DIR)
	cd $(SERVER_DIR) && GOOS=darwin GOARCH=$(MAC_ARCH) CGO_ENABLED=0 go build -ldflags "-s -w" -o ../$(RELEASE_DIR)/$(APP_NAME_MAC)

build-linux:
	mkdir -p $(RELEASE_DIR)
	cd $(SERVER_DIR) && GOOS=linux GOARCH=$(LNX_ARCH) CGO_ENABLED=0 go build -ldflags "-s -w" -o ../$(RELEASE_DIR)/$(APP_NAME_LNX)

# ---- 新增：一次性同时编译 mac + linux ----
build-all: build-mac build-linux
	@echo "✅ 同时生成: $(RELEASE_DIR)/$(APP_NAME_MAC) 和 $(RELEASE_DIR)/$(APP_NAME_LNX)"

# ==== 一键打包（单平台）====
deploy: build-frontend sync-dist build-server
	@echo "✅ 完成构建: $(RELEASE_DIR)/$(BIN_NAME)"
	@echo "静态资源: $(STATIC_DIR)"
	@echo "可直接上传整个 $(RELEASE_DIR) 目录到服务器运行。"

# ---- 新增：一键打包 + 同时产出两套二进制 ----
deploy-all: build-frontend sync-dist build-all
	@echo "✅ 已生成 Mac 与 Linux 两套可执行文件到 $(RELEASE_DIR)/"
	@echo "静态资源: $(STATIC_DIR)"

# ==== 本地启动 ====
run: build-server
	cd $(RELEASE_DIR) && ./$(BIN_NAME)

clean:
	rm -rf $(RELEASE_DIR)
