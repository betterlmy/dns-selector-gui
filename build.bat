@echo off
echo === DNS Selector GUI 构建脚本 ===

:: 设置 Node.js 路径
set PATH=C:\Program Files\nodejs;%PATH%

:: 1. 安装前端依赖
echo [1/3] 安装前端依赖...
cd frontend
npm.cmd install
if %errorlevel% neq 0 (echo 前端依赖安装失败 & pause & exit /b 1)

:: 2. 构建前端
echo [2/3] 构建前端...
npm.cmd run build
if %errorlevel% neq 0 (echo 前端构建失败 & pause & exit /b 1)
cd ..

:: 3. 编译 Go + 打包 exe（跳过前端步骤，已手动完成）
echo [3/3] 编译应用...
wails build -skipbindings -s
if %errorlevel% neq 0 (echo 应用编译失败 & pause & exit /b 1)

echo.
echo 构建完成！输出: build\bin\dns-selector-gui.exe
pause
