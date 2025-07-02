#!/bin/bash

# 短链接服务性能测试和分析脚本
# 使用多种工具进行综合性能测试

BASE_URL="http://localhost:8080"
TEST_DURATION="60s"
WARMUP_DURATION="10s"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 创建结果目录
RESULT_DIR="performance_results_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RESULT_DIR"

# 日志函数
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] $1${NC}" | tee -a "$RESULT_DIR/test.log"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" | tee -a "$RESULT_DIR/test.log"
}

log_info() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] INFO: $1${NC}" | tee -a "$RESULT_DIR/test.log"
}

# 检查依赖工具
check_tools() {
    log "检查测试工具..."
    
    local tools=("wrk" "curl" "jq" "docker")
    local missing_tools=()
    
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "缺少以下工具: ${missing_tools[*]}"
        log_error "请安装缺少的工具后重试"
        exit 1
    fi
    
    log "所有工具检查通过"
}

# 预热系统
warmup_system() {
    log "系统预热中 (${WARMUP_DURATION})..."
    
    # 简单的健康检查预热
    wrk -t4 -c100 -d"$WARMUP_DURATION" \
        --timeout=10s \
        "$BASE_URL/health" > /dev/null 2>&1
    
    log "系统预热完成"
}

# 创建大量测试数据
create_test_data() {
    log "创建测试数据..."
    
    local test_codes_file="$RESULT_DIR/test_codes.txt"
    
    # 创建100个测试短链接
    for i in {1..100}; do
        local response=$(curl -s -X POST "$BASE_URL/api/v1/shorten" \
            -H "Content-Type: application/json" \
            -d "{\"url\": \"https://testsite.com/page$i\"}")
        
        local short_code=$(echo "$response" | jq -r '.data.short_code // empty')
        if [ -n "$short_code" ]; then
            echo "$short_code" >> "$test_codes_file"
        fi
        
        # 避免过快请求
        [ $((i % 10)) -eq 0 ] && sleep 0.1
    done
    
    local code_count=$(wc -l < "$test_codes_file" 2>/dev/null || echo 0)
    log "创建了 $code_count 个测试短链接"
}

# 基准测试 - 健康检查
benchmark_health() {
    log "开始健康检查基准测试..."
    
    wrk -t12 -c400 -d"$TEST_DURATION" \
        --timeout=30s \
        --latency \
        "$BASE_URL/health" > "$RESULT_DIR/health_benchmark.txt" 2>&1
    
    log "健康检查基准测试完成"
}

# 基准测试 - 创建短链接
benchmark_create() {
    log "开始创建短链接基准测试..."
    
    cat > "$RESULT_DIR/create_test.lua" << 'EOF'
math.randomseed(os.time())

request = function()
    local timestamp = os.time()
    local rand = math.random(1000000)
    local url = "https://benchmark.com/test/" .. timestamp .. "/" .. rand
    local body = string.format('{"url": "%s"}', url)
    
    return wrk.format("POST", "/api/v1/shorten", {
        ["Content-Type"] = "application/json"
    }, body)
end

response = function(status, headers, body)
    if status ~= 201 then
        print("Error: HTTP " .. status .. " - " .. body)
    end
end
EOF
    
    wrk -t8 -c200 -d"$TEST_DURATION" \
        --timeout=30s \
        --latency \
        -s "$RESULT_DIR/create_test.lua" \
        "$BASE_URL" > "$RESULT_DIR/create_benchmark.txt" 2>&1
    
    log "创建短链接基准测试完成"
}

# 基准测试 - 重定向
benchmark_redirect() {
    log "开始重定向基准测试..."
    
    local test_codes_file="$RESULT_DIR/test_codes.txt"
    if [ ! -f "$test_codes_file" ]; then
        log_error "测试数据文件不存在，跳过重定向测试"
        return
    fi
    
    # 将测试码复制到 Lua 脚本可访问的位置
    cp "$test_codes_file" /tmp/benchmark_codes.txt
    
    cat > "$RESULT_DIR/redirect_test.lua" << 'EOF'
math.randomseed(os.time())

local codes = {}
local file = io.open("/tmp/benchmark_codes.txt", "r")
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

response = function(status, headers, body)
    if status ~= 302 and status ~= 404 then
        print("Error: HTTP " .. status .. " - " .. body)
    end
end
EOF
    
    wrk -t12 -c500 -d"$TEST_DURATION" \
        --timeout=30s \
        --latency \
        -s "$RESULT_DIR/redirect_test.lua" \
        "$BASE_URL" > "$RESULT_DIR/redirect_benchmark.txt" 2>&1
    
    log "重定向基准测试完成"
}

# 基准测试 - 信息查询
benchmark_info() {
    log "开始信息查询基准测试..."
    
    local test_codes_file="$RESULT_DIR/test_codes.txt"
    if [ ! -f "$test_codes_file" ]; then
        log_error "测试数据文件不存在，跳过信息查询测试"
        return
    fi
    
    cat > "$RESULT_DIR/info_test.lua" << 'EOF'
math.randomseed(os.time())

local codes = {}
local file = io.open("/tmp/benchmark_codes.txt", "r")
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
    
    wrk -t8 -c300 -d"$TEST_DURATION" \
        --timeout=30s \
        --latency \
        -s "$RESULT_DIR/info_test.lua" \
        "$BASE_URL" > "$RESULT_DIR/info_benchmark.txt" 2>&1
    
    log "信息查询基准测试完成"
}

# 监控系统资源
monitor_resources() {
    log "开始监控系统资源..."
    
    # 创建资源监控脚本
    cat > "$RESULT_DIR/monitor.sh" << 'EOF'
#!/bin/bash
while true; do
    echo "=== $(date) ===" 
    
    # Docker 容器状态
    docker stats --no-stream --format "{{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}"
    
    # Redis 内存使用
    echo "Redis Memory:"
    docker exec shorturl_redis redis-cli info memory | grep -E "(used_memory_human|used_memory_peak_human|used_memory_rss_human)"
    
    # PostgreSQL 连接和性能
    echo "PostgreSQL:"
    docker exec shorturl_postgres psql -U postgres -d shorturl -c "
    SELECT 
        count(*) as connections,
        count(*) FILTER (WHERE state = 'active') as active_connections,
        count(*) FILTER (WHERE state = 'idle') as idle_connections
    FROM pg_stat_activity;
    "
    
    echo "---"
    sleep 5
done
EOF
    
    chmod +x "$RESULT_DIR/monitor.sh"
    
    # 在后台运行监控
    "$RESULT_DIR/monitor.sh" > "$RESULT_DIR/resource_monitor.log" 2>&1 &
    local monitor_pid=$!
    echo $monitor_pid > "$RESULT_DIR/monitor.pid"
    
    log "资源监控已启动 (PID: $monitor_pid)"
}

# 停止资源监控
stop_monitoring() {
    if [ -f "$RESULT_DIR/monitor.pid" ]; then
        local monitor_pid=$(cat "$RESULT_DIR/monitor.pid")
        if kill -0 "$monitor_pid" 2>/dev/null; then
            kill "$monitor_pid"
            log "资源监控已停止"
        fi
        rm -f "$RESULT_DIR/monitor.pid"
    fi
}

# 生成性能报告
generate_report() {
    log "生成性能测试报告..."
    
    local report_file="$RESULT_DIR/performance_report.md"
    
    cat > "$report_file" << EOF
# 短链接服务性能测试报告

**测试时间**: $(date)
**测试持续时间**: $TEST_DURATION
**测试目标**: $BASE_URL

## 测试环境

### 系统信息
- **操作系统**: $(uname -s) $(uname -r)
- **CPU**: $(sysctl -n machdep.cpu.brand_string 2>/dev/null || grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
- **内存**: $(free -h 2>/dev/null | grep Mem | awk '{print $2}' || sysctl -n hw.memsize | awk '{print $1/1024/1024/1024 "GB"}')

### Docker 容器状态
\`\`\`
$(docker-compose ps)
\`\`\`

## 测试结果

### 1. 健康检查接口
EOF
    
    if [ -f "$RESULT_DIR/health_benchmark.txt" ]; then
        echo -e "\n\`\`\`" >> "$report_file"
        cat "$RESULT_DIR/health_benchmark.txt" >> "$report_file"
        echo -e "\`\`\`\n" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

### 2. 创建短链接接口
EOF
    
    if [ -f "$RESULT_DIR/create_benchmark.txt" ]; then
        echo -e "\n\`\`\`" >> "$report_file"
        cat "$RESULT_DIR/create_benchmark.txt" >> "$report_file"
        echo -e "\`\`\`\n" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

### 3. 重定向接口
EOF
    
    if [ -f "$RESULT_DIR/redirect_benchmark.txt" ]; then
        echo -e "\n\`\`\`" >> "$report_file"
        cat "$RESULT_DIR/redirect_benchmark.txt" >> "$report_file"
        echo -e "\`\`\`\n" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

### 4. 信息查询接口
EOF
    
    if [ -f "$RESULT_DIR/info_benchmark.txt" ]; then
        echo -e "\n\`\`\`" >> "$report_file"
        cat "$RESULT_DIR/info_benchmark.txt" >> "$report_file"
        echo -e "\`\`\`\n" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## 性能分析总结

### 关键指标
EOF
    
    # 提取关键性能指标
    local health_rps=$(grep "Requests/sec:" "$RESULT_DIR/health_benchmark.txt" 2>/dev/null | awk '{print $2}')
    local create_rps=$(grep "Requests/sec:" "$RESULT_DIR/create_benchmark.txt" 2>/dev/null | awk '{print $2}')
    local redirect_rps=$(grep "Requests/sec:" "$RESULT_DIR/redirect_benchmark.txt" 2>/dev/null | awk '{print $2}')
    local info_rps=$(grep "Requests/sec:" "$RESULT_DIR/info_benchmark.txt" 2>/dev/null | awk '{print $2}')
    
    cat >> "$report_file" << EOF

| 接口 | 每秒请求数 (RPS) | 说明 |
|------|------------------|------|
| 健康检查 | $health_rps | 最高性能，纯内存操作 |
| 创建短链接 | $create_rps | 涉及数据库写入和缓存 |
| 重定向 | $redirect_rps | 主要业务逻辑，缓存优化 |
| 信息查询 | $info_rps | 数据库查询操作 |

### 建议

1. **缓存优化**: 重定向接口性能表现良好，说明 Redis 缓存有效
2. **数据库优化**: 如果创建短链接性能较低，考虑异步写入优化
3. **连接池**: 监控数据库连接池使用情况
4. **资源扩展**: 根据实际负载调整容器资源配置

### 资源使用情况

详细的资源监控数据请查看 \`resource_monitor.log\` 文件。

EOF
    
    log "性能测试报告已生成: $report_file"
}

# 清理测试文件
cleanup() {
    log "清理测试文件..."
    rm -f /tmp/benchmark_codes.txt
    stop_monitoring
    log "清理完成"
}

# 主函数
main() {
    log "开始短链接服务性能测试"
    log "测试目标: $BASE_URL"
    log "测试持续时间: $TEST_DURATION"
    log "结果目录: $RESULT_DIR"
    
    # 设置清理陷阱
    trap cleanup EXIT
    
    # 检查工具和服务
    check_tools
    
    # 检查服务状态
    if ! curl -f -s "$BASE_URL/health" > /dev/null; then
        log_error "服务未响应，请确保服务正在运行"
        exit 1
    fi
    
    # 开始测试流程
    warmup_system
    create_test_data
    monitor_resources
    
    # 执行基准测试
    benchmark_health
    sleep 5  # 间隔时间避免系统过载
    
    benchmark_create
    sleep 5
    
    benchmark_redirect
    sleep 5
    
    benchmark_info
    
    # 停止监控并生成报告
    stop_monitoring
    generate_report
    
    log "性能测试完成！结果保存在: $RESULT_DIR/"
    echo
    echo "主要文件："
    echo "  - performance_report.md  : 性能测试报告"
    echo "  - resource_monitor.log   : 资源使用监控"
    echo "  - *_benchmark.txt        : 详细测试结果"
    echo
}

# 处理命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--duration)
            TEST_DURATION="$2"
            shift 2
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -h|--help)
            echo "使用方法: $0 [选项]"
            echo "选项:"
            echo "  -d, --duration  测试持续时间 (默认: 60s)"
            echo "  -u, --url       服务URL (默认: http://localhost:8080)"
            echo "  -h, --help      显示帮助信息"
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