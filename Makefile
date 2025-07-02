.PHONY: help build run test clean migrate docker-up docker-down deps

# 默认目标
help: ## 显示帮助信息
	@echo "🔧 短链接服务 - 可用命令:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发环境
deps: ## 安装依赖
	go mod download
	go mod tidy

build: ## 构建应用
	go build -o bin/server cmd/server/main.go
	go build -o bin/migrate cmd/migrate/main.go

run: ## 运行应用
	go run cmd/server/main.go

migrate: ## 运行数据库迁移
	go run cmd/migrate/main.go

test: ## 运行测试
	go test -v ./...

clean: ## 清理构建文件
	rm -rf bin/

# Docker 相关
docker-build: ## 构建 Docker 镜像
	docker build -t short-url:latest .

docker-rebuild: ## 强制重新构建 Docker 镜像
	docker-compose build --no-cache

docker-up: ## 启动所有服务
	docker-compose up -d

docker-down: ## 停止所有服务
	docker-compose down

docker-logs: ## 查看日志
	docker-compose logs -f

docker-restart: ## 重启所有服务
	$(MAKE) docker-down
	$(MAKE) docker-rebuild
	$(MAKE) docker-up

# 开发工具
fmt: ## 格式化代码
	go fmt ./...

lint: ## 代码检查
	golangci-lint run

# 数据库操作
db-up: ## 启动数据库和 Redis
	docker-compose up -d postgres redis

db-down: ## 停止数据库和 Redis
	docker-compose stop postgres redis

db-reset: ## 重置数据库
	docker-compose down postgres
	docker volume rm short-url_postgres_data || true
	docker-compose up -d postgres
	sleep 5
	$(MAKE) migrate

# API 测试
api-test: ## 测试 API 接口
	@echo "🧪 测试 API 接口..."
	@echo "创建短链接:"
	curl -X POST http://localhost:8080/api/v1/shorten \
		-H "Content-Type: application/json" \
		-d '{"url": "https://www.example.com"}'
	@echo "\n\n获取统计信息:"
	curl http://localhost:8080/api/v1/stats
	@echo "\n\n健康检查:"
	curl http://localhost:8080/health

# 内存监控和调试
memory-debug: ## 查看内存调试信息
	@echo "🔍 内存调试信息:"
	curl -s http://localhost:8080/debug/memory | jq .

memory-monitor: ## 启动内存监控
	@echo "🔍 启动内存监控 (Ctrl+C 停止):"
	@chmod +x scripts/memory_monitor.sh
	@./scripts/memory_monitor.sh

memory-monitor-bg: ## 后台启动内存监控
	@echo "🔍 后台启动内存监控..."
	@chmod +x scripts/memory_monitor.sh
	@./scripts/memory_monitor.sh &

memory-analysis: ## 生成内存分析报告
	@echo "📊 生成内存分析报告..."
	@echo "查看 MEMORY_ANALYSIS.md 了解详细信息"

# 生产部署
deploy: ## 部署到生产环境
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 监控
logs: ## 查看应用日志
	docker-compose logs -f app

status: ## 查看服务状态
	@echo "📊 服务状态:"
	docker-compose ps
	@echo ""
	@echo "🌐 健康检查:"
	curl -s http://localhost:8080/health | jq . || echo "服务未运行"

stats: ## 查看服务统计
	@echo "📈 服务统计:"
	curl -s http://localhost:8080/api/v1/stats | jq . || echo "服务未运行"

# 压力测试
benchmark: ## 运行标准压力测试
	@echo "🚀 运行标准压力测试..."
	@chmod +x scripts/benchmark.sh
	@./scripts/benchmark.sh

load-test: ## 运行快速负载测试
	@echo "⚡ 运行快速负载测试..."
	@chmod +x scripts/quick_load_test.sh
	@./scripts/quick_load_test.sh

performance-test: ## 运行详细性能测试
	@echo "📊 运行详细性能测试..."
	@chmod +x scripts/performance_test.sh
	@./scripts/performance_test.sh

functional-test: ## 运行功能测试
	@echo "✅ 运行功能测试..."
	@chmod +x scripts/test_api.sh
	@./scripts/test_api.sh

stress-test: ## 运行所有压力测试
	@echo "💪 运行完整压力测试套件..."
	$(MAKE) load-test
	$(MAKE) benchmark
	$(MAKE) performance-test

# 清理和维护
clean-containers: ## 清理所有容器和镜像
	docker-compose down -v
	docker system prune -f

clean-all: ## 完全清理
	$(MAKE) clean
	$(MAKE) clean-containers
	docker volume prune -f

# 快速启动
dev: ## 快速开发环境启动
	$(MAKE) docker-up
	@echo "⏳ 等待服务启动..."
	sleep 10
	$(MAKE) status
	@echo ""
	@echo "🎉 开发环境已就绪!"
	@echo "📖 API 文档: http://localhost:8080/api/v1/"
	@echo "🔍 内存监控: make memory-monitor"
	@echo "🧪 API 测试: make api-test"

# 故障排除
debug: ## 调试模式启动
	@echo "🐛 调试模式..."
	$(MAKE) docker-logs

fix-permissions: ## 修复脚本权限
	@echo "🔧 修复脚本权限..."
	chmod +x scripts/*.sh

# 报告生成
report: ## 生成完整报告
	@echo "📋 生成服务报告..."
	@echo "=== 服务状态 ===" > report.txt
	$(MAKE) status >> report.txt 2>&1
	@echo "" >> report.txt
	@echo "=== 内存信息 ===" >> report.txt
	$(MAKE) memory-debug >> report.txt 2>&1
	@echo "" >> report.txt
	@echo "=== 统计信息 ===" >> report.txt
	$(MAKE) stats >> report.txt 2>&1
	@echo "�� 报告已生成: report.txt" 