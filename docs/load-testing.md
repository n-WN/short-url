# 短链接服务压力测试指南

本文档描述了如何对短链接服务进行压力测试和性能分析。

## 📋 测试工具概览

我们提供了三种不同级别的压力测试工具：

| 工具 | 用途 | 持续时间 | 适用场景 |
|------|------|----------|----------|
| `quick_load_test.sh` | 快速负载测试 | ~1分钟 | 开发阶段快速验证 |
| `benchmark.sh` | 标准压力测试 | 30秒/场景 | 功能验证和基准测试 |
| `performance_test.sh` | 详细性能分析 | 60秒/场景 | 生产前性能评估 |

## 🚀 快速开始

### 1. 环境准备

确保已安装必要的工具：

```bash
# macOS
brew install wrk jq

# Ubuntu/Debian
sudo apt-get install wrk jq

# 检查 Docker 是否运行
docker --version
docker-compose --version
```

### 2. 启动服务

```bash
# 启动所有服务
make docker-up

# 等待服务就绪（约30秒）
# 检查服务状态
curl http://localhost:8080/health
```

### 3. 运行快速测试

```bash
# 运行快速负载测试（推荐首次使用）
make load-test

# 或直接执行脚本
./scripts/quick_load_test.sh
```

## 📊 详细测试方案

### 标准压力测试

使用 `wrk` 进行多场景压力测试：

```bash
# 运行标准压力测试
make benchmark

# 自定义参数
./scripts/benchmark.sh -d 60s -t 8 -c 300
```

**测试场景包括：**
- 健康检查接口 (高并发)
- 创建短链接 (写入密集)
- 重定向服务 (读取密集)
- 信息查询 (数据库查询)
- 统计接口 (聚合查询)
- 混合负载 (真实使用场景模拟)

### 详细性能分析

进行全面的性能测试和资源监控：

```bash
# 运行详细性能测试
make performance-test

# 自定义测试时长
./scripts/performance_test.sh -d 120s
```

**包含功能：**
- 系统预热
- 多维度性能测试
- 实时资源监控
- 详细性能报告生成
- 性能指标分析

### 功能验证测试

确保所有 API 功能正常：

```bash
# 运行功能测试
make functional-test
```

## 🔧 测试配置

### 自定义测试参数

编辑 `configs/load_test_configs.yaml` 调整测试配置：

```yaml
scenarios:
  custom:
    duration: "90s"
    threads: 10
    connections: 400
    description: "自定义测试场景"
```

### 命令行参数

所有脚本都支持命令行参数：

```bash
# benchmark.sh 参数
./scripts/benchmark.sh \
  -d 60s \          # 测试持续时间
  -t 12 \           # 线程数
  -c 500 \          # 并发连接数
  -u http://localhost:8080  # 服务URL

# performance_test.sh 参数
./scripts/performance_test.sh \
  -d 120s \         # 测试持续时间
  -u http://localhost:8080  # 服务URL
```

## 📈 性能指标解读

### 关键性能指标 (KPI)

| 指标 | 说明 | 良好范围 |
|------|------|----------|
| **RPS** (Requests/sec) | 每秒处理请求数 | 视接口而定 |
| **延迟** (Latency) | 请求响应时间 | <100ms (P99) |
| **错误率** (Error Rate) | 请求失败比例 | <1% |
| **CPU 使用率** | 处理器使用率 | <80% |
| **内存使用率** | 内存消耗 | <85% |

### 各接口性能预期

| 接口 | 预期 RPS | P99 延迟 | 说明 |
|------|----------|----------|------|
| 健康检查 | >5000 | <10ms | 纯内存操作 |
| 创建短链接 | >100 | <500ms | 数据库写入 |
| 重定向 | >1000 | <100ms | 缓存优化 |
| 信息查询 | >500 | <200ms | 数据库查询 |
| 统计接口 | >50 | <1000ms | 聚合计算 |

## 🚨 性能问题排查

### 常见性能瓶颈

1. **数据库连接池不足**
   ```bash
   # 检查 PostgreSQL 连接数
   docker exec shorturl_postgres psql -U postgres -d shorturl \
     -c "SELECT count(*) FROM pg_stat_activity;"
   ```

2. **Redis 内存不足**
   ```bash
   # 检查 Redis 内存使用
   docker exec shorturl_redis redis-cli info memory
   ```

3. **应用程序内存泄漏**
   ```bash
   # 监控容器内存使用
   docker stats --no-stream
   ```

### 优化建议

1. **缓存优化**
   - 调整 Redis 缓存 TTL
   - 优化缓存命中率
   - 使用布隆过滤器减少无效查询

2. **数据库优化**
   - 检查索引使用情况
   - 优化 SQL 查询
   - 调整连接池大小

3. **应用优化**
   - 异步处理非关键操作
   - 减少内存分配
   - 优化 HTTP 处理逻辑

## 📋 测试检查清单

### 测试前准备
- [ ] 确认服务运行正常
- [ ] 检查依赖工具安装
- [ ] 清理历史测试数据
- [ ] 确认系统资源充足

### 测试执行
- [ ] 运行快速负载测试
- [ ] 执行标准压力测试
- [ ] 进行详细性能分析
- [ ] 监控系统资源使用

### 测试后分析
- [ ] 查看性能测试报告
- [ ] 分析资源使用情况
- [ ] 识别性能瓶颈
- [ ] 制定优化方案

## 🔍 测试结果分析

### 测试报告位置

性能测试会在以下位置生成报告：

```
performance_results_YYYYMMDD_HHMMSS/
├── performance_report.md      # 主报告
├── resource_monitor.log       # 资源监控日志
├── health_benchmark.txt       # 健康检查测试结果
├── create_benchmark.txt       # 创建接口测试结果
├── redirect_benchmark.txt     # 重定向接口测试结果
├── info_benchmark.txt         # 信息查询测试结果
└── test.log                   # 测试执行日志
```

### 报告解读

1. **吞吐量分析**：查看各接口的 RPS 是否达到预期
2. **延迟分析**：关注 P99 延迟是否在可接受范围
3. **错误分析**：检查是否有请求失败或超时
4. **资源分析**：评估 CPU、内存使用是否合理

## 🛠️ 故障排除

### 常见问题

1. **wrk 命令找不到**
   ```bash
   # macOS
   brew install wrk
   
   # Ubuntu
   sudo apt-get install wrk
   ```

2. **服务连接失败**
   ```bash
   # 检查服务状态
   docker-compose ps
   curl http://localhost:8080/health
   ```

3. **测试脚本权限错误**
   ```bash
   # 添加执行权限
   chmod +x scripts/*.sh
   ```

## 📚 参考资料

- [wrk 官方文档](https://github.com/wg/wrk)
- [性能测试最佳实践](https://github.com/wg/wrk/wiki)
- [Docker 性能监控](https://docs.docker.com/config/containers/resource_constraints/)

---

如有问题，请查看日志文件或联系开发团队。 