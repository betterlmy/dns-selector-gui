# DNS Selector GUI - 跨平台构建
# 支持: Windows / macOS / Linux × amd64 / arm64

APP_NAME := dns-selector-gui
VERSION  := 0.1.0
BUILD    := build/bin
GOCACHE ?= $(CURDIR)/.cache/go-build
HOST_OS := $(shell uname -s)
HOST_ARCH := $(shell uname -m)
GO_BUILD_TAGS := desktop,wv2runtime.download,production
DARWIN_ARM64_LDFLAGS := -w -s -extldflags '-framework UniformTypeIdentifiers'
DARWIN_AMD64_LDFLAGS := -w -s -extldflags '-arch x86_64 -framework UniformTypeIdentifiers'

export GOCACHE

ifeq ($(OS),Windows_NT)
  NPM := npm.cmd
else
  NPM := npm
endif

# --- 通用构建函数 ---
# $(1)=platform $(2)=output filename
define wails_build
	@echo ">>> 构建 $(2)..."
	wails build -clean -platform $(1) -o $(2)
	@echo ">>> 完成: $(BUILD)/$(2)"
endef

# 刷新前端产物和 wailsjs 绑定
define prepare_frontend_build
	@mkdir -p frontend/dist
	@touch frontend/dist/.gitkeep
	wails generate module
	@cd frontend && $(NPM) run build
endef

# $(1)=output filename $(2)=display name $(3)=extra environment $(4)=ldflags
define go_build_mac_binary
	@echo ">>> 构建 $(2)..."
	$(3) go build -buildvcs=false -tags "$(GO_BUILD_TAGS)" -ldflags "$(4)" -o "$(BUILD)/$(1)" .
	@echo ">>> 完成: $(BUILD)/$(1)"
endef

# $(1)=binary filename
define package_mac_app
	@rm -rf "$(BUILD)/DNS Selector.app"
	@mkdir -p "$(BUILD)/DNS Selector.app/Contents/MacOS"
	@mkdir -p "$(BUILD)/DNS Selector.app/Contents/Resources"
	@cp "$(BUILD)/$(1)" "$(BUILD)/DNS Selector.app/Contents/MacOS/dns-selector-gui"
	@cp assets/icons.icns "$(BUILD)/DNS Selector.app/Contents/Resources/iconfile.icns"
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<plist version="1.0"><dict>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<key>CFBundleExecutable</key><string>dns-selector-gui</string>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<key>CFBundleIdentifier</key><string>com.betterlmy.dns-selector-gui</string>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<key>CFBundleName</key><string>DNS Selector</string>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<key>CFBundleIconFile</key><string>iconfile</string>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<key>CFBundlePackageType</key><string>APPL</string>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<key>CFBundleVersion</key><string>$(VERSION)</string>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '<key>CFBundleShortVersionString</key><string>$(VERSION)</string>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@echo '</dict></plist>' >> "$(BUILD)/DNS Selector.app/Contents/Info.plist"
	@xattr -cr "$(BUILD)/DNS Selector.app"
	@codesign --force --deep --sign - "$(BUILD)/DNS Selector.app"
	@echo ">>> 完成: $(BUILD)/DNS Selector.app"
endef

# $(1)=arch
define make_dmg
	@echo ">>> 生成 DMG ($(1))..."
	create-dmg \
		--volname "DNS Selector GUI" \
		--window-pos 200 120 --window-size 600 400 \
		--icon-size 100 --app-drop-link 425 178 \
		"$(BUILD)/$(APP_NAME)-$(VERSION)-darwin-$(1).dmg" \
		"$(BUILD)/DNS Selector.app"
	@echo ">>> 完成: $(BUILD)/$(APP_NAME)-$(VERSION)-darwin-$(1).dmg"
endef

.PHONY: all deps frontend build dev test test-unit test-integration clean help \
        build-all \
        build-windows build-windows-amd64 build-windows-arm64 \
        build-mac build-mac-amd64 build-mac-arm64 build-mac-universal \
        build-linux build-linux-amd64 build-linux-arm64 \
        dmg-amd64 dmg-arm64

# --- 基础目标 ---

all: build

deps:
	@cd frontend && $(NPM) install
	@go mod download
	@mkdir -p "$(GOCACHE)"
	@mkdir -p build/bin build/darwin build/windows
	@cp assets/appicon.png build/appicon.png
	@cp assets/icons.icns build/darwin/icons.icns
	@cp assets/icon.ico build/windows/icon.ico

frontend: deps
	$(call prepare_frontend_build)

build: deps
ifeq ($(HOST_OS),Darwin)
	$(call prepare_frontend_build)
ifeq ($(HOST_ARCH),arm64)
	$(call go_build_mac_binary,$(APP_NAME),$(APP_NAME),GOOS=darwin GOARCH=arm64 CGO_ENABLED=1,$(DARWIN_ARM64_LDFLAGS))
else
	$(call go_build_mac_binary,$(APP_NAME),$(APP_NAME),GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CGO_CFLAGS='-arch x86_64' CGO_LDFLAGS='-arch x86_64 -framework UniformTypeIdentifiers',$(DARWIN_AMD64_LDFLAGS))
endif
else
	wails build -clean
endif

dev:
	@mkdir -p build/bin
	@xattr -cr build/bin 2>/dev/null || true
	@xattr -w com.apple.fileprovider.ignore#P 1 build/bin 2>/dev/null || true
	wails dev

test:
	go test ./backend/... -short -v -count=1

test-unit:
	go test ./backend/... -v -count=1

test-integration:
	go test ./backend/... -v -count=1 -run 'TestIntegration'

clean:
	rm -rf $(BUILD) frontend/dist

# --- Windows ---

build-windows: build-windows-amd64

build-windows-amd64: deps
	$(call wails_build,windows/amd64,$(APP_NAME)-$(VERSION)-windows-amd64.exe)

build-windows-arm64: deps
	$(call wails_build,windows/arm64,$(APP_NAME)-$(VERSION)-windows-arm64.exe)

# --- macOS ---

build-mac: build-mac-universal

build-mac-universal: deps
	$(call prepare_frontend_build)
	$(call go_build_mac_binary,$(APP_NAME)-$(VERSION)-darwin-arm64,$(APP_NAME)-$(VERSION)-darwin-arm64,GOOS=darwin GOARCH=arm64 CGO_ENABLED=1,$(DARWIN_ARM64_LDFLAGS))
	$(call go_build_mac_binary,$(APP_NAME)-$(VERSION)-darwin-amd64,$(APP_NAME)-$(VERSION)-darwin-amd64,GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CGO_CFLAGS='-arch x86_64' CGO_LDFLAGS='-arch x86_64 -framework UniformTypeIdentifiers',$(DARWIN_AMD64_LDFLAGS))
	lipo -create -output "$(BUILD)/$(APP_NAME)-$(VERSION)-darwin-universal" "$(BUILD)/$(APP_NAME)-$(VERSION)-darwin-arm64" "$(BUILD)/$(APP_NAME)-$(VERSION)-darwin-amd64"
	$(call package_mac_app,$(APP_NAME)-$(VERSION)-darwin-universal)

build-mac-amd64: deps
	$(call prepare_frontend_build)
	$(call go_build_mac_binary,$(APP_NAME)-$(VERSION)-darwin-amd64,$(APP_NAME)-$(VERSION)-darwin-amd64,GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CGO_CFLAGS='-arch x86_64' CGO_LDFLAGS='-arch x86_64 -framework UniformTypeIdentifiers',$(DARWIN_AMD64_LDFLAGS))
	$(call package_mac_app,$(APP_NAME)-$(VERSION)-darwin-amd64)

build-mac-arm64: deps
	$(call prepare_frontend_build)
	$(call go_build_mac_binary,$(APP_NAME)-$(VERSION)-darwin-arm64,$(APP_NAME)-$(VERSION)-darwin-arm64,GOOS=darwin GOARCH=arm64 CGO_ENABLED=1,$(DARWIN_ARM64_LDFLAGS))
	$(call package_mac_app,$(APP_NAME)-$(VERSION)-darwin-arm64)

dmg-amd64: build-mac-amd64
	$(call make_dmg,amd64)

dmg-arm64: build-mac-arm64
	$(call make_dmg,arm64)

# --- Linux ---

build-linux: build-linux-amd64

build-linux-amd64: deps
ifeq ($(HOST_OS),Darwin)
	@echo "Linux cross-build is not supported by Wails on macOS hosts. Run this target on a Linux runner."
	@exit 1
else
	$(call wails_build,linux/amd64,$(APP_NAME)-$(VERSION)-linux-amd64)
endif

build-linux-arm64: deps
ifeq ($(HOST_OS),Darwin)
	@echo "Linux cross-build is not supported by Wails on macOS hosts. Run this target on a Linux runner."
	@exit 1
else
	$(call wails_build,linux/arm64,$(APP_NAME)-$(VERSION)-linux-arm64)
endif

# --- 全平台 ---

ifeq ($(HOST_OS),Darwin)
build-all: build-windows-amd64 build-windows-arm64 \
           build-mac-amd64 build-mac-arm64
	@echo ">>> 当前 macOS 主机支持的 4 个目标构建完成"
else
build-all: build-windows-amd64 build-windows-arm64 \
           build-mac-amd64 build-mac-arm64 \
           build-linux-amd64 build-linux-arm64
	@echo ">>> 全部 6 个目标构建完成"
endif

# --- 帮助 ---

help:
	@echo ""
	@echo "  基础命令:"
	@echo "    make                     构建当前平台"
	@echo "    make dev                 开发模式（热重载）"
	@echo "    make deps                安装所有依赖"
	@echo "    make frontend            构建前端（生成 frontend/dist）"
	@echo "    make test                运行稳定后端测试（short）"
	@echo "    make test-unit           运行后端完整测试集"
	@echo "    make test-integration    仅运行集成测试"
	@echo "    make clean               清理构建产物"
	@echo ""
	@echo "  全平台构建:"
ifeq ($(HOST_OS),Darwin)
	@echo "    make build-all           构建当前 macOS 主机支持的 4 个目标"
else
	@echo "    make build-all           构建全部 6 个目标"
endif
	@echo ""
	@echo "  Windows:"
	@echo "    make build-windows       Windows amd64（默认）"
	@echo "    make build-windows-amd64"
	@echo "    make build-windows-arm64"
	@echo ""
	@echo "  macOS:"
	@echo "    make build-mac           macOS universal（默认）"
	@echo "    make build-mac-amd64     macOS Intel"
	@echo "    make build-mac-arm64     macOS Apple Silicon"
	@echo "    make dmg-amd64           DMG 安装包（Intel）"
	@echo "    make dmg-arm64           DMG 安装包（Apple Silicon）"
	@echo ""
	@echo "  Linux:"
	@echo "    make build-linux         Linux amd64（默认）"
	@echo "    make build-linux-amd64"
	@echo "    make build-linux-arm64"
ifeq ($(HOST_OS),Darwin)
	@echo "    注: Wails 不支持在 macOS 主机上交叉编译 Linux 目标"
endif
	@echo ""
