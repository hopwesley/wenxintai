# 目标二进制名字
BINARY_NAME=wxt_bin

# 输出目录
DIST_DIR=dist

# Go 源文件入口
MAIN=main.go

# 默认目标（编译并压缩当前系统）
build:
	@mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY_NAME) $(MAIN)

# 为 Linux 编译并压缩
build-linux:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)_linux $(MAIN)
	@upx --best --lzma $(DIST_DIR)/$(BINARY_NAME)_linux || true

# 为 macOS (Intel) 编译并压缩
build-mac:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)_mac $(MAIN)

# 为 macOS (Apple Silicon) 编译并压缩
build-mac-arm:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)_mac_arm $(MAIN)

# 一次性编译所有平台
build-all: build-linux build-mac build-mac-arm

# 清理编译产物
clean:
	rm -rf $(DIST_DIR)
