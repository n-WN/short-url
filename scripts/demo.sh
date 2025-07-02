#!/bin/bash

# 短链接服务演示脚本

echo "==================== Go 短链接服务演示 ===================="
echo

# 检查服务是否运行
echo "1. 检查服务健康状态..."
curl -s http://localhost:8080/health | jq '.' || echo "服务未运行，请先启动服务"
echo

# 创建短链接
echo "2. 创建短链接..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/golang/go"}')

echo "响应: $RESPONSE"

# 提取短码
SHORT_CODE=$(echo $RESPONSE | jq -r '.data.short_code')
echo "生成的短码: $SHORT_CODE"
echo

# 获取短链接信息
echo "3. 获取短链接信息..."
curl -s http://localhost:8080/api/v1/info/$SHORT_CODE | jq '.'
echo

# 测试重定向
echo "4. 测试重定向..."
echo "访问短链接: http://localhost:8080/$SHORT_CODE"
REDIRECT_URL=$(curl -s -I http://localhost:8080/$SHORT_CODE | grep -i location | cut -d' ' -f2 | tr -d '\r')
echo "重定向到: $REDIRECT_URL"
echo

# 创建自定义短码
echo "5. 创建自定义短码..."
curl -s -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://golang.org", "custom_code": "golang"}' | jq '.'
echo

# 创建带过期时间的短链接
echo "6. 创建带过期时间的短链接..."
EXPIRE_TIME=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ")
curl -s -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"https://pkg.go.dev\", \"expires_at\": \"$EXPIRE_TIME\"}" | jq '.'
echo

# 获取统计信息
echo "7. 获取统计信息..."
curl -s http://localhost:8080/api/v1/stats | jq '.'
echo

echo "==================== 演示完成 ===================="
echo "更多 API 使用方法请查看 README.md" 