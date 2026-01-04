#!/bin/bash

# 报名系统快速启动脚本
# 支持前台和后台启动模式，自动处理端口占用

set -e  # 遇到错误立即退出

# ==================== 颜色定义 ====================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ==================== 配置变量 ====================
BACKEND_PORT=8080
FRONTEND_PORT=8000
BACKEND_DIR="backend"
LOG_FILE="backend.log"
PID_FILE="backend.pid"

# ==================== 日志函数 ====================
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# ==================== 横幅显示 ====================
show_banner() {
    echo ""
    echo "========================================="
    echo "       报名系统 - 快速启动"
    echo "========================================="
    echo ""
}

# ==================== 检查必要的命令 ====================
check_requirements() {
    log_info "检查系统环境..."

    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "未检测到 Go 环境，请先安装 Go 1.21 或更高版本"
        log_info "安装方法: 访问 https://golang.org/dl/"
        exit 1
    fi

    # 检查 Python3（用于前端服务器）
    if ! command -v python3 &> /dev/null; then
        log_warn "未检测到 Python3，前端服务器将无法启动"
    fi

    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "✓ Go 版本: $go_version"
    log_info "✓ 环境检查完成"
    echo ""
}

# ==================== 终止占用端口的进程 ====================
kill_port_process() {
    local port=$1
    local service_name=$2

    log_info "检查端口 $port..."

    # 查找占用端口的进程
    local pid=$(lsof -ti:$port 2>/dev/null || true)

    if [ -n "$pid" ]; then
        log_warn "发现端口 $port 被进程 $pid 占用"

        # 获取进程信息
        local process_info=$(ps -p $pid -o comm= 2>/dev/null || echo "unknown")

        log_info "进程信息: $process_info (PID: $pid)"
        log_info "正在终止进程..."

        # 尝试优雅终止
        kill $pid 2>/dev/null || true

        # 等待进程结束
        local count=0
        while [ $count -lt 5 ]; do
            if ! lsof -ti:$port &> /dev/null; then
                log_info "✓ 进程已优雅终止"
                echo ""
                return 0
            fi
            sleep 1
            count=$((count + 1))
        done

        # 如果优雅终止失败，强制终止
        if lsof -ti:$port &> /dev/null; then
            log_warn "优雅终止失败，强制终止进程..."
            kill -9 $pid 2>/dev/null || true
            sleep 1

            if ! lsof -ti:$port &> /dev/null; then
                log_info "✓ 进程已强制终止"
            else
                log_error "无法终止占用端口的进程"
                exit 1
            fi
        fi
    else
        log_info "✓ 端口 $port 未被占用"
    fi
    echo ""
}

# ==================== 清理旧的 PID 文件和日志 ====================
cleanup_old_files() {
    log_info "清理旧的运行文件..."

    if [ -f "$PID_FILE" ]; then
        local old_pid=$(cat "$PID_FILE")
        if ps -p $old_pid > /dev/null 2>&1; then
            log_warn "发现旧的后端进程 (PID: $old_pid)，正在终止..."
            kill $old_pid 2>/dev/null || true
            sleep 1
        fi
        rm -f "$PID_FILE"
    fi

    # 可选：清理或备份旧日志
    if [ -f "$LOG_FILE" ]; then
        mv "$LOG_FILE" "${LOG_FILE}.old" 2>/dev/null || true
    fi

    log_info "✓ 清理完成"
    echo ""
}

# ==================== 安装依赖 ====================
install_dependencies() {
    log_step "安装 Go 依赖..."

    if [ ! -d "$BACKEND_DIR" ]; then
        log_error "后端目录 '$BACKEND_DIR' 不存在"
        exit 1
    fi

    if [ ! -f "$BACKEND_DIR/go.mod" ]; then
        log_error "未找到 go.mod 文件"
        exit 1
    fi

    # 在当前目录执行依赖安装，不改变工作目录
    (cd "$BACKEND_DIR" && go mod tidy)

    if [ $? -ne 0 ]; then
        log_error "依赖安装失败"
        exit 1
    fi

    log_info "✓ 依赖安装完成"
    echo ""
}

# ==================== 启动后端服务（前台） ====================
start_backend_foreground() {
    log_step "启动后端服务 (前台模式)..."

    cd "$BACKEND_DIR"

    log_info "后端服务将运行在: ${GREEN}http://localhost:${BACKEND_PORT}${NC}"
    echo ""
    echo "========================================="
    echo "后端服务启动中..."
    echo "按 ${RED}Ctrl+C${NC} 停止服务器"
    echo "========================================="
    echo ""

    go run main.go
}

# ==================== 启动后端服务（后台） ====================
start_backend_background() {
    log_step "启动后端服务 (后台模式)..."

    # 启动后端并记录日志
    (cd "$BACKEND_DIR" && nohup go run main.go > "../$LOG_FILE" 2>&1 &)
    local pid=$!

    # 保存 PID
    echo $pid > "$PID_FILE"

    # 等待启动
    sleep 2

    # 验证进程是否运行
    if ps -p $pid > /dev/null 2>&1; then
        log_info "✓ 后端服务已在后台启动"
        log_info "  - PID: $pid"
        log_info "  - 地址: http://localhost:${BACKEND_PORT}"
        log_info "  - 日志: $LOG_FILE"
        echo ""

        # 健康检查
        log_info "正在进行健康检查..."
        sleep 2

        if curl -s http://localhost:${BACKEND_PORT}/health > /dev/null 2>&1 || \
           curl -s http://localhost:${BACKEND_PORT}/ > /dev/null 2>&1; then
            log_info "✓ 后端服务健康检查通过"
        else
            log_warn "健康检查失败，但服务可能正在启动中..."
            log_info "请查看日志: tail -f $LOG_FILE"
        fi
    else
        log_error "后端服务启动失败"
        log_info "请查看日志: cat $LOG_FILE"
        exit 1
    fi
    echo ""
}

# ==================== 显示访问信息 ====================
show_access_info() {
    echo ""
    echo "========================================="
    echo "  服务访问地址"
    echo "========================================="
    echo ""
    echo "后端 API:"
    echo "  - 地址: ${GREEN}http://localhost:${BACKEND_PORT}${NC}"
    echo ""
    echo "前端界面:"
    echo "  1. 直接打开 HTML 文件:"
    echo "     ${BLUE}frontend/admin.html${NC}  (管理员界面)"
    echo "     ${BLUE}frontend/user.html${NC}    (用户界面)"
    echo ""
    echo "  2. 使用 HTTP 服务器:"
    echo "     ${YELLOW}cd frontend && python3 -m http.server ${FRONTEND_PORT}${NC}"
    echo ""
    echo "常用命令:"
    if [ "$1" = "background" ]; then
        echo "  - 查看日志: ${YELLOW}tail -f $LOG_FILE${NC}"
        echo "  - 停止服务: ${YELLOW}kill \$(cat $PID_FILE)${NC}"
        echo "  - 查看进程: ${YELLOW}ps \$(cat $PID_FILE)${NC}"
    fi
    echo "========================================="
    echo ""
}

# ==================== 主函数 ====================
main() {
    show_banner

    # 解析参数
    MODE="foreground"
    SKIP_PORT_CHECK=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--daemon)
                MODE="background"
                shift
                ;;
            -s|--skip-port-check)
                SKIP_PORT_CHECK=true
                shift
                ;;
            -h|--help)
                echo "用法: $0 [选项]"
                echo ""
                echo "选项:"
                echo "  -d, --daemon          后台模式运行"
                echo "  -s, --skip-port-check 跳过端口检查"
                echo "  -h, --help            显示帮助信息"
                echo ""
                echo "示例:"
                echo "  $0                # 前台模式运行（默认）"
                echo "  $0 -d            # 后台模式运行"
                echo "  $0 -d -s         # 后台模式，跳过端口检查"
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                log_info "使用 -h 查看帮助信息"
                exit 1
                ;;
        esac
    done

    # 执行启动流程
    check_requirements

    if [ "$SKIP_PORT_CHECK" = false ]; then
        kill_port_process $BACKEND_PORT "后端服务"
    fi

    cleanup_old_files
    install_dependencies

    # 根据模式启动服务
    if [ "$MODE" = "background" ]; then
        start_backend_background
        show_access_info "background"
    else
        show_access_info "foreground"
        start_backend_foreground
    fi
}

# ==================== 执行主函数 ====================
main "$@"
