# DNS Selector GUI - 跨平台构建
# 支持: Windows / macOS / Linux × amd64 / arm64

APP_NAME := dns-selector-gui
VERSION  := 0.1.0
BUILD    := build/bin

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

# macOS 构建：用 -nopackage 跳过 wails 自签名，手动创建 .app bundle 并签名
# $(1)=platform $(2)=output filename
define wails_build_mac
	@echo ">>> 构建 $(2)..."
	wails build -clean -platform $(1) -o $(2) -nopackage
	@APP="$(BUILD)/DNS Selector.app"; \
	 rm -rf "$$APP"; \
	 mkdir -p "$$APP/Contents/MacOS" "$$APP/Contents/Resources"; \
	 cp "$(BUILD)/$(2)" "$$APP/Contents/MacOS/dns-selector-gui"; \
	 echo '<?xml version="1.0" encoding="UTF-8"?>' > "$$APP/Contents/Info.plist"; \
	 echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> "$$APP/Contents/Info.plist"; \
	 echo '<plist version="1.0"><dict>' >> "$$APP/Contents/Info.plist"; \
	 echo '<key>CFBundleExecutable</key><string>dns-selector-gui</string>' >> "$$APP/Contents/Info.plist"; \
	 echo '<key>CFBundleIdentifier</key><string>com.betterlmy.dns-selector-gui</string>' >> "$$APP/Contents/Info.plist"; \
	 echo '<key>CFBundleName</key><string>DNS Selector</string>' >> "$$APP/Contents/Info.plist"; \
	 echo '<key>CFBundlePackageType</key><string>APPL</string>' >> "$$APP/Contents/Info.plist"; \
	 echo '<key>CFBundleVersion</key><string>$(VERSION)</string>' >> "$$APP/Contents/Info.plist"; \
	 echo '<key>CFBundleShortVersionString</key><string>$(VERSION)</string>' >> "$$APP/Contents/Info.plist"; \
	 echo '</dict></plist>' >> "$$APP/Contents/Info.plist"; \
	 xattr -cr "$$APP"; \
	 codesign --force --deep --sign - "$$APP"
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

.PHONY: all deps frontend build dev test clean help \
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

frontend: deps
	@mkdir -p frontend/dist
	@touch frontend/dist/.gitkeep
	wails generate module
	@cd frontend && $(NPM) run build

build: deps
	wails build -clean

dev:
	@mkdir -p build/bin
	@xattr -cr build/bin 2>/dev/null || true
	@xattr -w com.apple.fileprovider.ignore#P 1 build/bin 2>/dev/null || true
	wails dev

test:
	go test ./backend/... -v -count=1

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
	$(call wails_build_mac,darwin/universal,$(APP_NAME)-$(VERSION)-darwin-universal)

build-mac-amd64: deps
	$(call wails_build_mac,darwin/amd64,$(APP_NAME)-$(VERSION)-darwin-amd64)

build-mac-arm64: deps
	$(call wails_build_mac,darwin/arm64,$(APP_NAME)-$(VERSION)-darwin-arm64)

dmg-amd64: build-mac-amd64
	$(call make_dmg,amd64)

dmg-arm64: build-mac-arm64
	$(call make_dmg,arm64)

# --- Linux ---

build-linux: build-linux-amd64

build-linux-amd64: deps
	$(call wails_build,linux/amd64,$(APP_NAME)-$(VERSION)-linux-amd64)

build-linux-arm64: deps
	$(call wails_build,linux/arm64,$(APP_NAME)-$(VERSION)-linux-arm64)

# --- 全平台 ---

build-all: build-windows-amd64 build-windows-arm64 \
           build-mac-amd64 build-mac-arm64 \
           build-linux-amd64 build-linux-arm64
	@echo ">>> 全部 6 个目标构建完成"

# --- 帮助 ---

help:
	@echo ""
	@echo "  基础命令:"
	@echo "    make                     构建当前平台"
	@echo "    make dev                 开发模式（热重载）"
	@echo "    make deps                安装所有依赖"
	@echo "    make frontend            构建前端（生成 frontend/dist）"
	@echo "    make test                运行后端测试"
	@echo "    make clean               清理构建产物"
	@echo ""
	@echo "  全平台构建:"
	@echo "    make build-all           构建全部 6 个目标"
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
	@echo ""
