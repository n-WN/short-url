#!/bin/bash

# 短链接服务压力测试脚本
# 使用 wrk 进行 HTTP 负载测试

BASE_URL="http://localhost:8080"
DURATION="30s"
THREADS=12
CONNECTIONS=400

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}=== 检查测试依赖 ===${NC}"
    
    if ! command -v wrk &> /dev/null; then
        echo -e "${RED}错误: wrk 未安装${NC}"
        echo "请安装 wrk: brew install wrk (macOS) 或 apt-get install wrk (Ubuntu)"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}错误: jq 未安装${NC}"
        echo "请安装 jq: brew install jq (macOS) 或 apt-get install jq (Ubuntu)"
        exit 1
    fi
    
    echo -e "${GREEN}依赖检查通过${NC}"
}

# 检查服务状态
check_service() {
    echo -e "${BLUE}=== 检查服务状态 ===${NC}"
    
    if ! curl -f -s "$BASE_URL/health" > /dev/null; then
        echo -e "${RED}错误: 服务未响应，请确保服务正在运行${NC}"
        echo "启动服务: make docker-up"
        exit 1
    fi
    
    echo -e "${GREEN}服务运行正常${NC}"
}

# 创建测试数据
setup_test_data() {
    echo -e "${BLUE}=== 创建测试数据 ===${NC}"
    
    # 创建一些短链接用于重定向测试
    echo "创建测试短链接..."
    
    for i in {1..10}; do
        curl -s -X POST "$BASE_URL/api/v1/shorten" \
            -H "Content-Type: application/json" \
            -d "{\"url\": \"https://example.com/page$i\"}" | jq -r '.data.short_code' >> /tmp/test_codes.txt
    done
    
    echo -e "${GREEN}测试数据创建完成${NC}"
}

# 健康检查压测
test_health_check() {
    echo -e "${BLUE}=== 健康检查接口压测 ===${NC}"
    echo "测试参数: ${THREADS}线程, ${CONNECTIONS}连接, 持续${DURATION}"
    echo
    
    wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} \
        --timeout=30s \
        "$BASE_URL/health"
    
    echo
}

# 创建短链接压测
test_create_shortlink() {
    echo -e "${BLUE}=== 创建短链接接口压测 ===${NC}"
    echo "测试参数: ${THREADS}线程, ${CONNECTIONS}连接, 持续${DURATION}"
    echo
    
    # 创建 Lua 脚本用于随机URL
    cat > /tmp/create_test.lua << 'EOF'
math.randomseed(os.time())

request = function()
    local url = "https://example.com/test/" .. math.random(1000000)
    local body = string.format('{"url": "%s"}', url)
    
    return wrk.format("POST", "/api/v1/shorten", {
        ["Content-Type"] = "application/json"
    }, body)
end
EOF
    
    wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} \
        --timeout=30s \
        -s /tmp/create_test.lua \
        "$BASE_URL"
    
    echo
}

# 重定向压测
test_redirect() {
    echo -e "${BLUE}=== 短链接重定向压测 ===${NC}"
    echo "测试参数: ${THREADS}线程, ${CONNECTIONS}连接, 持续${DURATION}"
    echo
    
    if [ ! -f /tmp/test_codes.txt ]; then
        echo -e "${YELLOW}警告: 未找到测试短码，先创建测试数据${NC}"
        setup_test_data
    fi
    
    # 创建 Lua 脚本用于随机选择短码
    cat > /tmp/redirect_test.lua << 'EOF'
math.randomseed(os.time())

-- 读取短码列表
local codes = {}
local file = io.open("/tmp/test_codes.txt", "r")
if file then
    for line in file:lines() do
        if line and line ~= "" then
            table.insert(codes, line)
        end
    end
    file:close()
end

request = function()
    if #codes > 0 then
        local code = codes[math.random(#codes)]
        return wrk.format("GET", "/" .. code)
    else
        return wrk.format("GET", "/notfound")
    end
end
EOF
    
    wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} \
        --timeout=30s \
        -s /tmp/redirect_test.lua \
        "$BASE_URL"
    
    echo
}

# 信息查询压测
test_info_query() {
    echo -e "${BLUE}=== 短链接信息查询压测 ===${NC}"
    echo "测试参数: ${THREADS}线程, ${CONNECTIONS}连接, 持续${DURATION}"
    echo
    
    if [ ! -f /tmp/test_codes.txt ]; then
        echo -e "${YELLOW}警告: 未找到测试短码，先创建测试数据${NC}"
        setup_test_data
    fi
    
    # 创建 Lua 脚本
    cat > /tmp/info_test.lua << 'EOF'
math.randomseed(os.time())

local codes = {}
local file = io.open("/tmp/test_codes.txt", "r")
if file then
    for line in file:lines() do
        if line and line ~= "" then
            table.insert(codes, line)
        end
    end
    file:close()
end

request = function()
    if #codes > 0 then
        local code = codes[math.random(#codes)]
        return wrk.format("GET", "/api/v1/info/" .. code)
    else
        return wrk.format("GET", "/api/v1/info/notfound")
    end
end
EOF
    
    wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} \
        --timeout=30s \
        -s /tmp/info_test.lua \
        "$BASE_URL"
    
    echo
}

# 统计信息压测
test_stats() {
    echo -e "${BLUE}=== 统计信息接口压测 ===${NC}"
    echo "测试参数: ${THREADS}线程, ${CONNECTIONS}连接, 持续${DURATION}"
    echo
    
    wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} \
        --timeout=30s \
        "$BASE_URL/api/v1/stats"
    
    echo
}

# 混合负载测试
test_mixed_load() {
    echo -e "${BLUE}=== 混合负载测试 ===${NC}"
    echo "测试参数: ${THREADS}线程, ${CONNECTIONS}连接, 持续${DURATION}"
    echo "模拟真实使用场景: 20%创建, 70%重定向, 10%查询"
    echo
    
    cat > /tmp/mixed_test.lua << 'EOF'
math.randomseed(os.time())

local codes = {}
local file = io.open("/tmp/test_codes.txt", "r")
if file then
    for line in file:lines() do
        if line and line ~= "" then
            table.insert(codes, line)
        end
    end
    file:close()
end

request = function()
    local rand = math.random(100)
    
    if rand <= 20 then
        -- 20% 创建短链接
        local url = "https://example.com/mixed/" .. math.random(1000000)
        local body = string.format('{"url": "%s"}', url)
        return wrk.format("POST", "/api/v1/shorten", {
            ["Content-Type"] = "application/json"
        }, body)
    elseif rand <= 90 then
        -- 70% 重定向
        if #codes > 0 then
            local code = codes[math.random(#codes)]
            return wrk.format("GET", "/" .. code)
        else
            return wrk.format("GET", "/notfound")
        end
    else
        -- 10% 信息查询
        if #codes > 0 then
            local code = codes[math.random(#codes)]
            return wrk.format("GET", "/api/v1/info/" .. code)
        else
            return wrk.format("GET", "/api/v1/info/notfound")
        end
    end
end
EOF
    
    wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} \
        --timeout=30s \
        -s /tmp/mixed_test.lua \
        "$BASE_URL"
    
    echo
}

# 获取系统指标
get_system_metrics() {
    echo -e "${BLUE}=== 系统资源使用情况 ===${NC}"
    
    echo "=== Docker 容器状态 ==="
    docker-compose ps
    echo
    
    echo "=== 内存使用情况 ==="
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
    echo
    
    echo "=== Redis 信息 ==="
    docker exec shorturl_redis redis-cli info memory | grep -E "(used_memory_human|used_memory_peak_human)"
    echo
    
    echo "=== PostgreSQL 连接数 ==="
    docker exec shorturl_postgres psql -U postgres -d shorturl -c "SELECT count(*) as active_connections FROM pg_stat_activity WHERE state = 'active';"
    echo
}

# 清理测试文件
cleanup() {
    echo -e "${BLUE}=== 清理测试文件 ===${NC}"
    rm -f /tmp/test_codes.txt
    rm -f /tmp/create_test.lua
    rm -f /tmp/redirect_test.lua
    rm -f /tmp/info_test.lua
    rm -f /tmp/mixed_test.lua
    echo -e "${GREEN}清理完成${NC}"
}

# 主函数
main() {
    echo -e "${GREEN}=== 短链接服务压力测试 ===${NC}"
    echo "测试目标: $BASE_URL"
    echo "测试配置: ${THREADS}线程, ${CONNECTIONS}并发连接, 持续${DURATION}"
    echo
    
    # 检查依赖和服务
    check_dependencies
    check_service
    
    # 设置测试数据
    setup_test_data
    
    # 执行各项测试
    test_health_check
    test_create_shortlink
    test_redirect
    test_info_query
    test_stats
    test_mixed_load
    
    # 获取系统指标
    get_system_metrics
    
    # 清理
    cleanup
    
    echo -e "${GREEN}=== 压力测试完成 ===${NC}"
    echo "测试报告已生成，请查看上述输出结果"
}

# 处理命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--duration)
            DURATION="$2"
            shift 2
            ;;
        -t|--threads)
            THREADS="$2"
            shift 2
            ;;
        -c|--connections)
            CONNECTIONS="$2"
            shift 2
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -h|--help)
            echo "使用方法: $0 [选项]"
            echo "选项:"
            echo "  -d, --duration      测试持续时间 (默认: 30s)"
            echo "  -t, --threads       线程数 (默认: 12)"
            echo "  -c, --connections   并发连接数 (默认: 400)"
            echo "  -u, --url          服务URL (默认: http://localhost:8080)"
            echo "  -h, --help         显示帮助信息"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            exit 1
            ;;
    esac
done

# 运行主函数
main 