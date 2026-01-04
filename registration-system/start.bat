@echo off
chcp 65001 >nul
cls

echo =========================================
echo        报名系统 - 快速启动
echo =========================================
echo.

REM 检查 Go 是否安装
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误：未检测到 Go 环境，请先安装 Go 1.21 或更高版本
    pause
    exit /b 1
)

echo √ Go 环境检测通过
echo.

REM 进入后端目录
cd backend

REM 安装依赖
echo 正在安装 Go 依赖...
call go mod tidy

if %errorlevel% neq 0 (
    echo 错误：依赖安装失败
    pause
    exit /b 1
)

echo √ 依赖安装完成
echo.

REM 启动后端服务
echo 正在启动后端服务...
echo 后端将运行在 http://localhost:8080
echo.
echo =========================================
echo 后端启动中...
echo =========================================
echo.
echo 请在前端浏览器中打开：
echo   管理员界面: frontend\admin.html
echo   用户界面:   frontend\user.html
echo.
echo 或使用本地服务器：
echo   cd frontend
echo   python -m http.server 8000
echo.
echo 按 Ctrl+C 停止服务器
echo.

go run main.go

pause
