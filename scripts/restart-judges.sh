#!/bin/bash

# 判题机重启脚本
# 作者: 自动生成
# 用途: 一次性重启所有判题机

set -e  # 遇到错误立即退出

# 判题机列表
JUDGES=(
    "ubuntu@120.53.239.108"
    "ubuntu@82.156.29.203"
)

# 登录凭据
PASSWORD="n208966737."

# 工作目录（两个判题机相同）
WORKSPACE="/workspace/hoj-deploy/distributed/judgeserver"

# 日志函数
log_info() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] INFO: $1"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >&2
}

log_success() {
    echo "[$(date '+%H:%M:%S')] SUCCESS: $1"
}

# 远程执行命令的函数
# 参数：主机名、命令
remote_exec() {
    local host=$1
    local command=$2
    log_info "正在执行 [$host]: $command"
    
    # 使用 sshpass 远程执行命令
    sshpass -p "$PASSWORD" ssh "$host" "$command"
}

# 重启单个判题机
restart_judge() {
    local judge_ip=$1
    local judge_name=$2
    
    log_info "========================================"
    log_info "开始重启判题机: $judge_name ($judge_ip)"
    log_info "========================================"
    
    # 1. 停止容器
    log_info "[$judge_name] 正在停止容器..."
    remote_exec "$judge_host" "cd $WORKSPACE && sudo docker compose down"

    # 2. 启动容器
    log_info "[$judge_name] 正在启动容器..."
    remote_exec "$judge_host" "cd $WORKSPACE && sudo docker compose up -d"
    
    log_success "[$judge_name] 重启完成!"
    echo ""
}

# 主函数
main() {
    log_info "========================================"
    log_info "开始批量重启判题机"
    log_info "判题机数量: ${#JUDGES[@]}"
    log_info "========================================"
    echo ""
    
    # 遍历所有判题机并重启
    for i in "${!JUDGES[@]}"; do
        judge_host="${JUDGES[$i]}"
        judge_name="判题机$((i+1))"
        restart_judge "$judge_host" "$judge_name"
    done
    
    log_info "========================================"
    log_success "所有判题机重启完成！"
    log_info "========================================"
    
    # 等待5秒让容器完全启动
    log_info "等待5秒让容器完全启动..."
    sleep 5
    
    log_info "判题机重启流程执行完毕！"
}

# 执行主函数
main "$@"
