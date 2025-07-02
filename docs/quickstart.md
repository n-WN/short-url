# 快速启动指南

这是一个基于 Go、布隆过滤器、Redis 和 PostgreSQL 构建的高性能短链接服务的快速启动指南。

## 🚀 一键启动

### 使用 Docker Compose（推荐）

```bash
# 1. 克隆项目并进入目录
cd short-url

# 2. 启动所有服务（PostgreSQL、Redis、应用）
make docker-up

# 3. 等待服务启动完成（约30秒）
docker-compose logs -f app

# 4. 运行数据库迁移
docker-compose exec app /app/migrate

# 5. 测试API
make api-test
```

### 本地开发

```bash
# 1. 启动数据库和Redis
make db-up

# 2. 配置环境变量
cp config.env.example config.env
# 编辑 config.env 设置数据库连接信息

# 3. 安装依赖
make deps

# 4. 运行数据库迁移
make migrate

# 5. 启动应用
make run
```

## 📝 API 使用示例

### 创建短链接

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com"}'
```

响应：
```json
{
  "data": {
    "short_url": "http://localhost:8080/abc123",
    "short_code": "abc123",
    "original_url": "https://www.example.com",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### 访问短链接

```bash
curl -I http://localhost:8080/abc123
```

### 获取短链接信息

```bash
curl http://localhost:8080/api/v1/info/abc123
```

### 创建自定义短码

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://golang.org", "custom_code": "golang"}'
```

### 创建带过期时间的短链接

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://pkg.go.dev", "expires_at": "2024-12-31T23:59:59Z"}'
```

### 获取统计信息

```bash
curl http://localhost:8080/api/v1/stats
```

## 🎯 演示脚本

运行完整的功能演示：

```bash
./scripts/demo.sh
```

## 🛠️ 管理命令

```bash
# 查看帮助
make help

# 构建应用
make build

# 运行测试
make test

# 查看日志
make logs

# 清理过期链接
curl -X POST http://localhost:8080/api/v1/admin/clean

# 停止服务
make docker-down
```

## 🔧 配置说明

主要配置项（在 `config.env` 中）：

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=shorturl

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379

# 应用配置
APP_PORT=8080
BASE_URL=http://localhost:8080

# 布隆过滤器配置
BLOOM_FILTER_CAPACITY=1000000
BLOOM_FILTER_ERROR_RATE=0.001
```

## 📊 性能特性

- **布隆过滤器**: 99.9% 的重复检查在内存中完成
- **Redis 缓存**: 热点数据毫秒级响应
- **并发处理**: 支持高并发读写
- **数据库连接池**: 优化数据库访问

## 🐛 故障排除

### 常见问题

1. **数据库连接失败**
   ```bash
   # 检查数据库是否启动
   docker-compose ps postgres
   
   # 查看数据库日志
   docker-compose logs postgres
   ```

2. **Redis 连接失败**
   ```bash
   # 检查 Redis 是否支持 RedisBloom
   docker-compose exec redis redis-cli MODULE LIST
   ```

3. **布隆过滤器初始化失败**
   ```bash
   # 检查 Redis 版本和模块
   docker-compose exec redis redis-cli INFO modules
   ```

### 日志查看

```bash
# 查看应用日志
docker-compose logs -f app

# 查看所有服务日志
docker-compose logs -f
```

## 🔗 相关链接

- [完整文档](README.md)
- [API 文档](README.md#api-文档)
- [架构设计](README.md#架构设计)

---

如果遇到问题，请查看完整的 [README.md](README.md) 或提交 Issue。 