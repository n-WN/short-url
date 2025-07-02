#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== 短链接服务 API 测试 ==="
echo

# 1. 健康检查
echo "1. 健康检查..."
curl -s "$BASE_URL/health" | jq .
echo

# 2. 创建短链接
echo "2. 创建短链接 (随机短码)..."
RESPONSE1=$(curl -s -X POST "$BASE_URL/api/v1/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com"}')
echo $RESPONSE1 | jq .
SHORT_CODE1=$(echo $RESPONSE1 | jq -r '.data.short_code')
echo

# 3. 创建自定义短链接
echo "3. 创建短链接 (自定义短码)..."
RESPONSE2=$(curl -s -X POST "$BASE_URL/api/v1/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com", "custom_code": "google"}')
echo $RESPONSE2 | jq .
echo

# 4. 测试重定向
echo "4. 测试重定向..."
echo "重定向到 $SHORT_CODE1:"
curl -s -w "%{http_code} %{redirect_url}\n" -o /dev/null "$BASE_URL/$SHORT_CODE1"
echo "重定向到 google:"
curl -s -w "%{http_code} %{redirect_url}\n" -o /dev/null "$BASE_URL/google"
echo

# 5. 获取短链接信息
echo "5. 获取短链接信息..."
curl -s "$BASE_URL/api/v1/info/$SHORT_CODE1" | jq .
echo

# 6. 获取统计信息
echo "6. 获取统计信息..."
curl -s "$BASE_URL/api/v1/stats" | jq .
echo

echo "=== 测试完成 ===" 