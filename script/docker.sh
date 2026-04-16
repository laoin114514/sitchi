#!/bin/bash

# Sitchi Docker 管理脚本
# 从 configs/.env 读取环境配置

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_DIR/configs/.env"

# 当前环境配置（从.env文件读取）
CURRENT_ENV=""
CURRENT_PORT=""
PROJECT_NAME=""
MODE=""

# 从.env文件加载配置
load_env_config() {
    if [ -f "$ENV_FILE" ]; then
        # 读取MODE
        CURRENT_ENV=$(grep -E '^MODE=' "$ENV_FILE" | cut -d'=' -f2 | tr -d ' \r\n')
        # 读取PORT
        CURRENT_PORT=$(grep -E '^PORT=' "$ENV_FILE" | cut -d'=' -f2 | tr -d ' \r\n')
    fi

    # 设置默认值
    CURRENT_ENV="${CURRENT_ENV:-dev}"
    CURRENT_PORT="${CURRENT_PORT:-5050}"

    # 根据MODE设置项目名
    case "$CURRENT_ENV" in
        dev)
            PROJECT_NAME="sitchi-dev"
            MODE="dev"
            ;;
        prod)
            PROJECT_NAME="sitchi-prod"
            MODE="prod"
            ;;
        *)
            echo -e "${YELLOW}警告: .env文件中未知的MODE '$CURRENT_ENV'，将使用自定义模式${NC}"
            PROJECT_NAME="sitchi-$CURRENT_ENV"
            MODE="$CURRENT_ENV"
            ;;
    esac
}

# 清屏并显示标题
show_header() {
    clear
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}           ${GREEN}Sitchi Docker 管理脚本${NC}                       ${CYAN}║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# 显示当前环境信息
show_env_info() {
    load_env_config
    echo -e "${BLUE}当前环境配置 (从 $ENV_FILE 读取):${NC}"
    echo -e "  环境(MODE): ${GREEN}$CURRENT_ENV${NC}"
    echo -e "  项目名:     ${GREEN}$PROJECT_NAME${NC}"
    echo -e "  端口(PORT): ${GREEN}$CURRENT_PORT${NC}"
    echo -e "  配置文件:   ${GREEN}configs/configs.$CURRENT_ENV.yml${NC}"
    echo ""
}

# 显示主菜单
show_menu() {
    show_header
    show_env_info
    echo -e "${YELLOW}请选择操作:${NC}"
    echo ""
    echo -e "  ${CYAN}1)${NC} build       - 构建镜像"
    echo -e "  ${CYAN}2)${NC} start       - 启动容器"
    echo -e "  ${CYAN}3)${NC} stop        - 停止容器"
    echo -e "  ${CYAN}4)${NC} restart     - 重启容器"
    echo -e "  ${CYAN}5)${NC} status      - 查看容器状态"
    echo -e "  ${CYAN}6)${NC} logs        - 查看容器日志"
    echo -e "  ${CYAN}7)${NC} clean       - 停止并删除容器、镜像和数据卷"
    echo -e "  ${CYAN}8)${NC} shell       - 进入容器内部"
    echo ""
    echo -e "  ${CYAN}9)${NC} 刷新配置    - 重新读取.env文件"
    echo -e "  ${CYAN}0)${NC} 退出"
    echo ""
}

# 启动容器
cmd_start() {
    load_env_config
    echo -e "${BLUE}正在启动 Sitchi $CURRENT_ENV 环境...${NC}"
    echo -e "  项目名: $PROJECT_NAME"
    echo -e "  端口: $CURRENT_PORT"
    echo -e "  模式: $CURRENT_ENV"
    echo ""

    cd "$PROJECT_DIR"

    if MODE=$CURRENT_ENV PORT=$CURRENT_PORT docker-compose --project-name "$PROJECT_NAME" up -d; then
        echo ""
        echo -e "${GREEN}✓ $CURRENT_ENV 环境启动成功！${NC}"
        echo -e "  访问地址: http://localhost:$CURRENT_PORT"
    else
        echo ""
        echo -e "${RED}✗ $CURRENT_ENV 环境启动失败${NC}"
    fi
    echo ""
    read -p "按回车键继续..."
}

# 停止容器
cmd_stop() {
    load_env_config
    echo -e "${YELLOW}正在停止 Sitchi $MODE 环境...${NC}"
    echo ""

    cd "$PROJECT_DIR"
    if docker-compose --project-name "$PROJECT_NAME" down; then
        echo ""
        echo -e "${GREEN}✓ $MODE 环境已停止${NC}"
    else
        echo ""
        echo -e "${RED}✗ 停止失败${NC}"
    fi
    echo ""
    read -p "按回车键继续..."
}

# 重启容器
cmd_restart() {
    load_env_config
    echo -e "${YELLOW}正在重启 Sitchi $MODE 环境...${NC}"
    echo ""
    cmd_stop_internal
    cmd_start_internal
    echo ""
    read -p "按回车键继续..."
}

# 内部停止（不等待回车）
cmd_stop_internal() {
    cd "$PROJECT_DIR"
    docker-compose --project-name "$PROJECT_NAME" down >/dev/null 2>&1 || true
}

# 内部启动（不等待回车）
cmd_start_internal() {
    cd "$PROJECT_DIR"
    if MODE=$MODE PORT=$CURRENT_PORT docker-compose --project-name "$PROJECT_NAME" up -d; then
        echo -e "${GREEN}✓ $MODE 环境已重启成功！${NC}"
        echo -e "  访问地址: http://localhost:$CURRENT_PORT"
    else
        echo -e "${RED}✗ 重启失败${NC}"
    fi
}

# 查看状态
cmd_status() {
    load_env_config
    echo -e "${BLUE}Sitchi $MODE 环境状态:${NC}"
    echo ""
    docker-compose --project-name "$PROJECT_NAME" ps
    echo ""
    read -p "按回车键继续..."
}

# 查看日志
cmd_logs() {
    load_env_config
    echo -e "${BLUE}查看 Sitchi $MODE 环境日志...${NC}"
    echo -e "${YELLOW}(按 Ctrl+C 退出日志查看)${NC}"
    echo ""
    # 捕获SIGINT信号(Ctrl+C)，防止退出脚本
    trap '' SIGINT
    docker-compose --project-name "$PROJECT_NAME" logs -f || true
    # 恢复默认信号处理
    trap - SIGINT
    echo ""
    read -p "按回车键继续..."
}

# 重新构建
cmd_build() {
    load_env_config
    echo -e "${YELLOW}正在重新构建 Sitchi $MODE 环境镜像...${NC}"
    echo ""

    cd "$PROJECT_DIR"
    echo -e "${BLUE}正在构建镜像...${NC}"
    if docker-compose --project-name "$PROJECT_NAME" build; then
        echo ""
        echo -e "${GREEN}✓ $MODE 环境镜像构建成功！${NC}"
    else
        echo ""
        echo -e "${RED}✗ $MODE 环境镜像构建失败${NC}"
    fi
    echo ""
    read -p "按回车键继续..."
}

# 清理环境
cmd_clean() {
    load_env_config
    echo -e "${RED}警告: 这将删除 $MODE 环境的所有容器、镜像和数据卷！${NC}"
    echo ""
    read -p "确定要继续吗? (y/N): " confirm

    if [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]]; then
        cd "$PROJECT_DIR"
        docker-compose --project-name "$PROJECT_NAME" down --rmi all --volumes
        echo ""
        echo -e "${GREEN}✓ $MODE 环境已清理${NC}"
    else
        echo ""
        echo -e "${YELLOW}已取消清理操作${NC}"
    fi
    echo ""
    read -p "按回车键继续..."
}

# 进入容器shell
cmd_shell() {
    load_env_config
    echo -e "${BLUE}进入 Sitchi $MODE 环境容器...${NC}"
    echo -e "${YELLOW}(输入 'exit' 退出容器)${NC}"
    echo ""
    docker-compose --project-name "$PROJECT_NAME" exec sitchi /bin/sh || true
    echo ""
    read -p "按回车键继续..."
}

# 刷新配置
refresh_config() {
    echo -e "${BLUE}正在刷新配置...${NC}"
    echo ""

    if [ -f "$ENV_FILE" ]; then
        echo -e "${GREEN}✓ 已重新读取 $ENV_FILE${NC}"
        load_env_config
        echo ""
        echo -e "${BLUE}当前配置:${NC}"
        echo -e "  MODE: $CURRENT_ENV"
        echo -e "  PORT: $CURRENT_PORT"
        echo -e "  PROJECT_NAME: $PROJECT_NAME"
    else
        echo -e "${RED}✗ 配置文件不存在: $ENV_FILE${NC}"
        echo -e "${YELLOW}使用默认配置: MODE=dev, PORT=5050${NC}"
    fi
    echo ""
    read -p "按回车键继续..."
}

# 命令行模式 - 显示帮助信息
show_help() {
    echo -e "${BLUE}Sitchi Docker 管理脚本${NC}"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "说明:"
    echo "  脚本从 configs/.env 读取 MODE 和 PORT 配置"
    echo ""
    echo "命令:"
    echo "  start       启动容器"
    echo "  stop        停止容器"
    echo "  restart     重启容器"
    echo "  status      查看容器状态"
    echo "  logs        查看容器日志"
    echo "  build       重新构建镜像并启动"
    echo "  clean       停止并删除容器、镜像和数据卷"
    echo "  shell       进入容器内部"
    echo ""
    echo "示例:"
    echo "  $0                          # 启动交互式菜单"
    echo "  $0 start                    # 根据.env配置启动"
    echo "  $0 logs                     # 查看日志"
    echo "  $0 clean                    # 清理环境"
    echo ""
    echo "配置文件示例 (configs/.env):"
    echo "  MODE=dev"
    echo "  PORT=5050"
}

# 命令行模式 - 解析参数
parse_args() {
    COMMAND=""
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                COMMAND="$1"
                shift
                ;;
        esac
    done
}

# 命令行模式 - 执行命令
run_command() {
    load_env_config
    case "$COMMAND" in
        start)
            cmd_start
            ;;
        stop)
            cmd_stop
            ;;
        restart)
            cmd_restart
            ;;
        status)
            cmd_status
            ;;
        logs)
            cmd_logs
            ;;
        build)
            cmd_build
            ;;
        clean)
            cmd_clean
            ;;
        shell)
            cmd_shell
            ;;
        *)
            echo -e "${RED}错误: 未知命令 '$COMMAND'${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 交互式菜单模式
interactive_mode() {
    load_env_config
    while true; do
        show_menu
        read -p "请输入选项 (0-9): " choice

        # ✅ 修复完成：选项与功能一一对应
        case "$choice" in
            1) cmd_build ;;
            2) cmd_start ;;
            3) cmd_stop ;;
            4) cmd_restart ;;
            5) cmd_status ;;
            6) cmd_logs ;;
            7) cmd_clean ;;
            8) cmd_shell ;;
            9) refresh_config ;;
            0)
                echo -e "${GREEN}再见!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}无效选项，请重新选择${NC}"
                sleep 1
                ;;
        esac
    done
}

# 主函数
main() {
    # 如果没有参数，进入交互式菜单模式
    if [ $# -eq 0 ]; then
        interactive_mode
    else
        # 有参数，进入命令行模式
        parse_args "$@"
        if [ -z "$COMMAND" ]; then
            show_help
            exit 1
        fi
        run_command
    fi
}

main "$@"