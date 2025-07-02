# 性能优化指南

## 概述

本文档介绍短链接服务的性能特性、测试方法和优化策略。

## 🚀 性能表现

### 基准性能指标
- **重定向响应**: < 5ms (缓存命中)
- **创建短链**: < 50ms 平均响应时间  
- **并发支持**: 10K+ RPS (标准配置)
- **内存效率**: 100万短码仅需 ~1.2MB 布隆过滤器

### 实际测试结果
基于标准硬件配置的测试结果：

| 操作类型 | RPS | 平均延迟 | 99% 延迟 |
|----------|-----|----------|----------|
| 健康检查 | 31,159 | 1ms | 3ms |
| 短链重定向 | 23,888 | 2ms | 8ms |
| 创建短链 | 3,199 | 15ms | 45ms |
| 信息查询 | 15,000+ | 3ms | 12ms |

## 🧪 性能测试

### 快速负载测试
```bash
# 1分钟快速验证
make load-test
```

特点：
- 快速验证基本性能
- 实时结果显示
- 资源使用监控

### 标准压力测试
```bash
# 全面性能测试
make benchmark
```

测试场景：
- 健康检查高并发
- 创建短链性能
- 重定向服务压测
- 信息查询测试
- 混合负载模拟

### 详细性能分析
```bash
# 60秒深度分析
make performance-test
```

功能：
- 详细性能指标
- 自动报告生成
- 系统资源监控

### 完整压力测试套件
```bash
# 运行所有测试
make stress-test
```

## 🔍 内存监控

### 实时内存监控
```bash
# 启动实时监控
make memory-monitor

# 或后台监控
make memory-monitor-bg
```

### 内存调试信息
```bash
# 查看内存状态
make memory-debug
```

返回信息：
```json
{
  "data": {
    "after_gc": {
      "alloc_mb": 2,
      "heap_alloc_mb": 2,
      "heap_idle_mb": 3,
      "heap_inuse_mb": 3,
      "heap_sys_mb": 7,
      "num_gc": 10,
      "num_goroutine": 7,
      "stack_inuse_mb": 0,
      "stack_sys_mb": 0,
      "sys_mb": 12,
      "total_alloc_mb": 4
    },
    "gc_triggered": true
  }
}
```

### 内存使用分析

#### 正常状态
- **启动状态**: 15-30 MiB
- **轻量负载**: 13-20 MiB  
- **稳定运行**: 20-50 MiB

#### 高负载状态
- **高强度压测**: 可能达到 1GB+
- **原因**: 大量临时对象 + GC延迟
- **解决**: 压测后内存会自动回收

#### 内存优化建议
- 生产环境设置内存限制（推荐 512MB）
- 定期监控内存使用趋势
- 高并发场景可调整GC参数

## ⚡ 性能优化策略

### 1. 布隆过滤器优化
```go
// 配置参数
BLOOM_FILTER_CAPACITY=1000000    // 容量
BLOOM_FILTER_ERROR_RATE=0.001    // 错误率
```

优势：
- 99.9% 重复检查在内存中完成
- O(k) 常数时间复杂度
- 仅 1.2MB 内存占用

### 2. Redis 缓存策略
```go
// 缓存配置
REDIS_TTL=3600              // 1小时缓存
REDIS_MAX_CONNECTIONS=100   // 连接池
```

优化：
- 热点数据缓存
- 异步访问计数
- 连接池复用

### 3. 数据库优化
```sql
-- 关键索引
CREATE INDEX idx_short_links_code ON short_links(short_code);
CREATE INDEX idx_short_links_url ON short_links(original_url);
CREATE INDEX idx_short_links_created ON short_links(created_at);
```

特性：
- 多重索引优化
- 连接池管理
- 异步写入操作

## 📊 性能监控

### 系统指标监控
```bash
# 生成完整报告
make report
```

包含内容：
- 服务状态检查
- 内存使用分析
- 统计信息汇总

### 关键性能指标（KPI）

#### 响应时间
- 重定向: < 10ms (99%)
- 创建: < 100ms (99%)
- 查询: < 50ms (99%)

#### 吞吐量
- 重定向: > 20K RPS
- 创建: > 3K RPS
- 查询: > 10K RPS

#### 可用性
- 服务可用性: > 99.9%
- 缓存命中率: > 80%
- 错误率: < 0.1%

## 🔧 调优建议

### 硬件配置
```yaml
# 推荐最低配置
CPU: 2 cores
RAM: 4GB
Disk: SSD 20GB

# 生产环境配置  
CPU: 4+ cores
RAM: 8GB+
Disk: SSD 100GB+
```

### 容器配置
```yaml
# docker-compose.yml
services:
  app:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.5'
```

### 应用优化
```bash
# Go 运行时调优
GOGC=100                    # GC 目标百分比
GOMAXPROCS=4               # 最大 CPU 核数
```

## 🚨 性能问题排查

### 常见性能问题

1. **内存泄漏**
   ```bash
   # 监控内存增长
   make memory-monitor
   ```

2. **缓存未命中**
   ```bash
   # 检查Redis连接
   docker-compose logs redis
   ```

3. **数据库慢查询**
   ```bash
   # 检查数据库日志
   docker-compose logs postgres
   ```

### 故障排除流程

1. **确认问题**：`make status`
2. **查看日志**：`make debug`  
3. **内存检查**：`make memory-debug`
4. **性能测试**：`make load-test`
5. **生成报告**：`make report`

## 📈 性能测试最佳实践

### 测试前准备
1. 确保服务稳定运行
2. 清理历史数据影响
3. 预热缓存和连接池

### 测试流程
1. **基线测试**：`make api-test`
2. **负载测试**：`make load-test`  
3. **压力测试**：`make benchmark`
4. **性能分析**：`make performance-test`

### 结果分析
- 关注平均响应时间和99%分位
- 监控系统资源使用情况
- 记录错误率和成功率
- 分析瓶颈和优化方向

这样的性能优化策略能够确保短链接服务在高并发场景下保持稳定和高效的运行。 