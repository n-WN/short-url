#!/bin/bash

# 快速负载测试脚本
# 适用于开发阶段的快速压测

BASE_URL="http://localhost:8080"

echo "🚀 快速负载测试 - 短链接服务"
echo "目标服务: $BASE_URL"
echo

# 检查服务是否运行
if ! curl -f -s "$BASE_URL/health" > /dev/null; then
    echo "❌ 服务未响应，请先启动服务"
    echo "启动命令: make docker-up"
    exit 1
fi

echo "✅ 服务运行正常"
echo

# 1. 快速创建测试数据
echo "📝 创建测试数据..."
for i in {1..5}; do
    RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/shorten" \
        -H "Content-Type: application/json" \
        -d "{\"url\": \"https://quicktest.com/page$i\"}")
    SHORT_CODE=$(echo "$RESPONSE" | jq -r '.data.short_code // empty')
    if [ -n "$SHORT_CODE" ]; then
        echo "$SHORT_CODE" >> /tmp/quick_test_codes.txt
    fi
done

echo "✅ 测试数据创建完成"
echo

# 2. 健康检查快速压测 (10秒)
echo "🏥 健康检查接口压测 (10秒, 100并发)..."
wrk -t4 -c100 -d10s --timeout=10s "$BASE_URL/health" | grep -E "(Requests/sec|Latency)"
echo

# 3. 创建短链接压测 (15秒)
echo "➕ 创建短链接接口压测 (15秒, 50并发)..."
cat > /tmp/quick_create.lua << 'EOF'
math.randomseed(os.time())
request = function()
    local url = "https://loadtest.com/page/" .. math.random(10000)
    local body = string.format('{"url": "%s"}', url)
    return wrk.format("POST", "/api/v1/shorten", {
        ["Content-Type"] = "application/json"
    }, body)
end
EOF

wrk -t4 -c50 -d15s --timeout=15s -s /tmp/quick_create.lua "$BASE_URL" | grep -E "(Requests/sec|Latency)"
echo

# 4. 重定向压测 (10秒)
echo "🔄 重定向接口压测 (10秒, 200并发)..."
if [ -f /tmp/quick_test_codes.txt ]; then
    cat > /tmp/quick_redirect.lua << 'EOF'
math.randomseed(os.time())
local codes = {}
local file = io.open("/tmp/quick_test_codes.txt", "r")
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

    wrk -t4 -c200 -d10s --timeout=10s -s /tmp/quick_redirect.lua "$BASE_URL" | grep -E "(Requests/sec|Latency)"
else
    echo "⚠️  跳过重定向测试 (无测试数据)"
fi
echo

# 5. 信息查询压测 (10秒)
echo "📊 信息查询接口压测 (10秒, 100并发)..."
if [ -f /tmp/quick_test_codes.txt ]; then
    cat > /tmp/quick_info.lua << 'EOF'
math.randomseed(os.time())
local codes = {}
local file = io.open("/tmp/quick_test_codes.txt", "r")
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

    wrk -t4 -c100 -d10s --timeout=10s -s /tmp/quick_info.lua "$BASE_URL" | grep -E "(Requests/sec|Latency)"
else
    echo "⚠️  跳过信息查询测试 (无测试数据)"
fi
echo

# 6. 显示资源使用情况
echo "💻 当前资源使用情况:"
echo "=== Docker 容器状态 ==="
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" 2>/dev/null || echo "Docker 统计信息获取失败"
echo

echo "=== Redis 内存使用 ==="
docker exec shorturl_redis redis-cli info memory 2>/dev/null | grep -E "(used_memory_human|used_memory_peak_human)" || echo "Redis 信息获取失败"
echo

echo "=== 数据库统计 ==="
curl -s "$BASE_URL/api/v1/stats" | jq . 2>/dev/null || echo "统计信息获取失败"
echo

# 清理测试文件
rm -f /tmp/quick_test_codes.txt /tmp/quick_create.lua /tmp/quick_redirect.lua /tmp/quick_info.lua

echo "🎉 快速负载测试完成！"
echo
echo "💡 提示："
echo "  - 如需更详细的测试，请运行: ./scripts/performance_test.sh"
echo "  - 如需长时间压测，请运行: ./scripts/benchmark.sh" 