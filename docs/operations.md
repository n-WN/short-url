# 运维监控指南

## 概述

本文档提供短链接服务的运维监控、故障排除和维护指南。

## 📊 监控体系

### 系统监控指标

#### 应用层指标
- **响应时间**: 平均响应时间和99%分位数
- **吞吐量**: 每秒请求数 (RPS)
- **错误率**: HTTP 4xx/5xx 错误百分比
- **可用性**: 服务正常运行时间百分比

#### 基础设施指标
- **CPU 使用率**: 容器和主机 CPU 使用情况
- **内存使用**: 应用内存、堆内存、系统内存
- **磁盘 I/O**: 读写速度和 IOPS
- **网络**: 带宽使用和连接数

#### 业务指标
- **短链创建数**: 每分钟/小时创建的短链数量
- **访问次数**: 短链点击次数
- **缓存命中率**: Redis 缓存效率
- **数据库连接数**: PostgreSQL 活跃连接

### 监控命令

```bash
# 快速状态检查
make status

# 详细内存监控
make memory-monitor

# 生成完整报告
make report

# 查看实时日志
make debug
```

## 🔍 健康检查

### 内置健康检查端点

1. **基础健康检查**
   ```bash
   curl http://localhost:8080/health
   ```
   
   响应示例：
   ```json
   {
     "data": {
       "status": "ok",
       "timestamp": "2025-07-02T20:13:30Z"
     }
   }
   ```

2. **内存状态检查**
   ```bash
   curl http://localhost:8080/debug/memory
   ```

3. **服务统计信息**
   ```bash
   curl http://localhost:8080/api/v1/stats
   ```

### 自动化健康检查

```bash
#!/bin/bash
# health_check.sh

ENDPOINT="http://localhost:8080/health"
TIMEOUT=10

response=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT $ENDPOINT)

if [ $response -eq 200 ]; then
    echo "✅ 服务运行正常"
    exit 0
else
    echo "❌ 服务异常，HTTP状态码: $response"
    exit 1
fi
```

## 📈 性能监控

### 关键性能指标 (KPI)

| 指标 | 目标值 | 警告阈值 | 严重阈值 |
|------|--------|----------|----------|
| 重定向响应时间 | < 10ms | > 50ms | > 100ms |
| 创建响应时间 | < 100ms | > 500ms | > 1000ms |
| 错误率 | < 0.1% | > 1% | > 5% |
| 可用性 | > 99.9% | < 99.5% | < 99% |
| 内存使用 | < 256MB | > 400MB | > 500MB |
| CPU 使用 | < 50% | > 80% | > 90% |

### 实时性能监控

```bash
# 启动性能监控脚本
#!/bin/bash
# performance_monitor.sh

while true; do
    echo "=== $(date) ==="
    
    # 健康检查
    curl -s http://localhost:8080/health | jq .
    
    # 内存使用
    echo "内存使用:"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"
    
    # 服务统计
    echo "服务统计:"
    curl -s http://localhost:8080/api/v1/stats | jq .
    
    echo "---"
    sleep 30
done
```

## 🚨 告警和通知

### 告警规则

1. **服务不可用**
   - 条件: 健康检查失败
   - 处理: 立即重启服务

2. **响应时间过长**
   - 条件: 99%分位响应时间 > 1s
   - 处理: 检查资源使用和数据库性能

3. **内存使用过高**
   - 条件: 内存使用 > 80%
   - 处理: 检查内存泄漏，考虑重启

4. **错误率过高**
   - 条件: 错误率 > 5%
   - 处理: 检查日志，分析错误原因

### 告警脚本示例

```bash
#!/bin/bash
# alert.sh

# 检查内存使用
memory_usage=$(docker stats --no-stream --format "{{.MemPerc}}" shorturl_app | sed 's/%//')
if (( $(echo "$memory_usage > 80" | bc -l) )); then
    echo "🚨 内存使用过高: ${memory_usage}%"
    # 发送告警通知
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"短链接服务内存使用过高: '${memory_usage}'%"}' \
        $SLACK_WEBHOOK_URL
fi

# 检查服务可用性
if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "🚨 服务不可用"
    # 尝试重启服务
    make docker-restart
    
    # 发送告警
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"短链接服务不可用，已尝试重启"}' \
        $SLACK_WEBHOOK_URL
fi
```

## 📋 日志管理

### 日志类型

1. **应用日志**
   - 位置: 容器内或宿主机挂载目录
   - 格式: 结构化 JSON 日志
   - 级别: ERROR, WARN, INFO, DEBUG

2. **访问日志**
   - 内容: HTTP 请求记录
   - 包含: IP、方法、路径、状态码、响应时间

3. **系统日志**
   - Docker 容器日志
   - 系统级错误和警告

### 日志查看和分析

```bash
# 查看应用日志
docker-compose logs -f app

# 查看最近的错误日志
docker-compose logs app | grep "level\":\"error"

# 查看特定时间段的日志
docker-compose logs --since "2025-01-01T10:00:00" --until "2025-01-01T11:00:00" app

# 统计错误日志
docker-compose logs app | grep "error" | wc -l
```

### 日志轮转配置

```yaml
# docker-compose.yml
services:
  app:
    logging:
      driver: "json-file"
      options:
        max-size: "200m"
        max-file: "10"
        labels: "service=short-url"
```

## 🔧 故障排除

### 常见问题和解决方案

#### 1. 服务无法启动

**症状**: 容器启动失败或立即退出

**排查步骤**:
```bash
# 查看容器状态
docker-compose ps

# 查看启动日志
docker-compose logs app

# 检查配置文件
cat config.env

# 检查端口占用
netstat -tlnp | grep 8080
```

**常见原因**:
- 端口冲突
- 配置文件错误
- 依赖服务未启动
- 权限问题

**解决方案**:
```bash
# 重置环境
make clean-all
make dev

# 或手动修复
make fix-permissions
make db-reset
```

#### 2. 数据库连接失败

**症状**: 应用启动时报数据库连接错误

**排查步骤**:
```bash
# 检查数据库状态
docker-compose ps postgres

# 检查数据库日志
docker-compose logs postgres

# 测试连接
docker-compose exec postgres psql -U postgres -d shorturl -c "SELECT 1;"
```

**解决方案**:
```bash
# 重启数据库
docker-compose restart postgres

# 重置数据库
make db-reset
```

#### 3. Redis 连接问题

**症状**: 缓存功能失效，布隆过滤器错误

**排查步骤**:
```bash
# 检查 Redis 状态
docker-compose ps redis

# 检查 Redis 模块
docker-compose exec redis redis-cli MODULE LIST

# 测试连接
docker-compose exec redis redis-cli ping
```

**解决方案**:
```bash
# 重启 Redis
docker-compose restart redis

# 检查镜像版本
docker-compose pull redis
```

#### 4. 性能问题

**症状**: 响应时间长，吞吐量低

**排查步骤**:
```bash
# 检查资源使用
docker stats

# 运行性能测试
make load-test

# 检查内存使用
make memory-debug

# 查看数据库慢查询
docker-compose exec postgres psql -U postgres -d shorturl -c "SELECT query, calls, total_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
```

### 故障恢复流程

1. **问题识别**
   ```bash
   make status
   make memory-debug
   ```

2. **收集信息**
   ```bash
   make report
   docker-compose logs --tail=100 app
   ```

3. **尝试修复**
   ```bash
   # 轻度重启
   docker-compose restart app
   
   # 或完全重启
   make docker-restart
   ```

4. **验证修复**
   ```bash
   make api-test
   make load-test
   ```

## 🔄 定期维护

### 日常维护任务

#### 每日检查
```bash
#!/bin/bash
# daily_check.sh

echo "📊 每日健康检查 - $(date)"

# 服务状态
make status

# 资源使用
docker stats --no-stream

# 错误日志统计
error_count=$(docker-compose logs --since "24h" app | grep -c "error")
echo "过去24小时错误数: $error_count"

# 磁盘使用
df -h

echo "✅ 每日检查完成"
```

#### 每周维护
```bash
#!/bin/bash
# weekly_maintenance.sh

echo "🔧 每周维护 - $(date)"

# 清理过期链接
curl -X POST http://localhost:8080/api/v1/admin/clean

# 清理 Docker 资源
docker system prune -f

# 备份数据库
make backup

# 性能测试
make benchmark

echo "✅ 每周维护完成"
```

#### 每月任务
- 更新依赖和镜像
- 审查监控指标和趋势
- 优化配置参数
- 容量规划评估

### 数据备份

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
echo "备份数据库..."
docker-compose exec -T postgres pg_dump -U postgres shorturl > $BACKUP_DIR/db_backup_$DATE.sql

# 备份 Redis
echo "备份 Redis..."
docker-compose exec redis redis-cli BGSAVE
docker cp $(docker-compose ps -q redis):/data/dump.rdb $BACKUP_DIR/redis_backup_$DATE.rdb

# 备份配置文件
echo "备份配置..."
cp config.env $BACKUP_DIR/config_backup_$DATE.env

# 清理旧备份（保留30天）
find $BACKUP_DIR -name "*backup*" -mtime +30 -delete

echo "✅ 备份完成: $BACKUP_DIR"
```

## 📱 监控工具集成

### Prometheus 监控

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'short-url'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

### Grafana 仪表板

关键图表：
- 响应时间趋势
- QPS (每秒查询数)
- 错误率变化
- 内存使用情况
- 数据库连接池状态

### 日志聚合 (ELK Stack)

```yaml
# logstash.conf
input {
  docker {
    path => "/var/lib/docker/containers/*/*.log"
  }
}

filter {
  if [docker][name] == "shorturl_app" {
    json {
      source => "message"
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "short-url-logs-%{+YYYY.MM.dd}"
  }
}
```

## 🎯 最佳实践

### 监控最佳实践

1. **建立基线**: 记录正常运行时的指标基线
2. **分层监控**: 应用层、基础设施层、业务层全面监控
3. **主动告警**: 预防性告警而非被动响应
4. **自动化**: 尽可能自动化监控和修复流程
5. **文档化**: 记录所有故障和解决方案

### 运维最佳实践

1. **定期备份**: 自动化数据备份和恢复测试
2. **容量规划**: 基于增长趋势进行容量规划
3. **安全更新**: 定期更新依赖和修复安全漏洞
4. **性能优化**: 持续监控和优化性能瓶颈
5. **灾难恢复**: 制定和测试灾难恢复计划

这个运维指南确保了短链接服务的稳定运行和高可用性。 