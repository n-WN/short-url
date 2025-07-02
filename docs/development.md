# 开发指南

## 概述

本文档为短链接服务的开发者提供详细的开发环境搭建、代码贡献和项目维护指南。

## 🏗️ 项目结构

```
short-url/
├── cmd/                    # 应用入口点
│   ├── migrate/           # 数据库迁移工具
│   └── server/            # HTTP 服务器主程序
├── internal/              # 内部业务逻辑包
│   ├── cache/             # Redis 缓存和布隆过滤器
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接和操作
│   ├── handler/           # HTTP 请求处理器
│   ├── models/            # 数据模型定义
│   ├── service/           # 业务逻辑服务层
│   └── utils/             # 通用工具函数
├── pkg/                   # 可被外部引用的公共包
│   ├── logger/            # 结构化日志包
│   └── validator/         # 数据验证包
├── scripts/               # 开发和部署脚本
├── sql/migrations/        # 数据库迁移文件
├── docs/                  # 项目文档
└── configs/               # 配置文件
```

## 🛠️ 开发环境搭建

### 环境要求

- **Go**: 1.24+ 
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **Git**: 2.30+

### 本地开发环境

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd short-url
   ```

2. **启动依赖服务**
   ```bash
   # 启动 PostgreSQL 和 Redis
   make db-up
   ```

3. **配置环境变量**
   ```bash
   # 复制配置模板
   cp config.env.example config.env
   
   # 编辑配置（可选，默认配置适用于本地开发）
   vim config.env
   ```

4. **安装依赖和运行迁移**
   ```bash
   # 安装 Go 依赖
   make deps
   
   # 运行数据库迁移
   make migrate
   ```

5. **启动应用**
   ```bash
   # 开发模式启动
   make run
   
   # 或使用一键开发环境
   make dev
   ```

### Docker 开发环境

```bash
# 一键启动完整开发环境
make docker-up

# 查看服务状态
make status

# 测试 API
make api-test
```

## 🔧 开发工具和命令

### 代码质量
```bash
# 格式化代码
make fmt

# 代码检查（需要安装 golangci-lint）
make lint

# 运行测试
make test
```

### 开发调试
```bash
# 查看应用日志
make logs

# 查看详细调试信息
make debug

# 内存使用监控
make memory-monitor
```

### 数据库操作
```bash
# 连接到数据库
docker-compose exec postgres psql -U postgres -d shorturl

# 重置数据库（谨慎使用）
make db-reset
```

## 🏛️ 架构设计

### 分层架构

1. **Handler 层** (`internal/handler/`)
   - HTTP 请求处理
   - 路由定义
   - 请求验证和响应

2. **Service 层** (`internal/service/`)
   - 业务逻辑实现
   - 数据处理和转换
   - 缓存策略

3. **Repository 层** (`internal/service/repository.go`)
   - 数据访问抽象
   - 数据库操作封装

4. **Model 层** (`internal/models/`)
   - 数据结构定义
   - 数据验证规则

### 关键组件

#### 布隆过滤器 (`internal/cache/bloomfilter.go`)
```go
type BloomFilter struct {
    client   *redis.Client
    capacity int64
    errorRate float64
}

// 用于快速检测短码重复
func (bf *BloomFilter) Exists(key string) (bool, error)
func (bf *BloomFilter) Add(key string) error
```

#### 缓存层 (`internal/cache/redis.go`)
```go
type Cache struct {
    client *redis.Client
    ttl    time.Duration
}

// 热点数据缓存
func (c *Cache) Get(key string) (string, error)
func (c *Cache) Set(key, value string) error
```

#### 短码编码 (`internal/utils/encoder.go`)
```go
// Base62 编码生成短码
func GenerateShortCode(length int) string
func EncodeBase62(num int64) string
```

## 📝 编码规范

### Go 代码风格

1. **命名规范**
   - 包名：小写，简洁，无下划线
   - 函数名：驼峰命名，导出函数首字母大写
   - 变量名：驼峰命名，避免缩写

2. **错误处理**
   ```go
   // 推荐：明确的错误处理
   if err != nil {
       logger.Error("operation failed", zap.Error(err))
       return nil, fmt.Errorf("failed to process: %w", err)
   }
   ```

3. **日志记录**
   ```go
   // 使用结构化日志
   logger.Info("short link created",
       zap.String("code", shortCode),
       zap.String("url", originalURL),
       zap.Duration("duration", duration))
   ```

### API 设计规范

1. **RESTful 接口**
   - 使用标准 HTTP 方法
   - 资源导向的 URL 设计
   - 一致的错误响应格式

2. **响应格式**
   ```go
   type Response struct {
       Data    interface{} `json:"data,omitempty"`
       Message string      `json:"message,omitempty"`
       Error   string      `json:"error,omitempty"`
       Code    int         `json:"code,omitempty"`
   }
   ```

## 🧪 测试策略

### 单元测试
```bash
# 运行所有单元测试
go test ./...

# 运行特定包的测试
go test ./internal/service/

# 带覆盖率的测试
go test -cover ./...
```

### 集成测试
```bash
# API 功能测试
make functional-test

# 完整的端到端测试
./scripts/test_api.sh
```

### 性能测试
```bash
# 快速负载测试
make load-test

# 标准压力测试
make benchmark

# 完整性能测试套件
make stress-test
```

## 🚀 部署和发布

### 本地构建
```bash
# 构建二进制文件
make build

# 构建 Docker 镜像
make docker-build
```

### CI/CD 流程

1. **代码检查**
   - 代码格式化检查
   - 静态代码分析
   - 单元测试覆盖率

2. **集成测试**
   - API 功能测试
   - 数据库集成测试
   - Redis 缓存测试

3. **性能验证**
   - 基准性能测试
   - 负载测试验证

4. **部署**
   - Docker 镜像构建
   - 容器化部署

## 🤝 贡献指南

### 提交代码

1. **Fork 项目**
   ```bash
   git clone https://github.com/your-username/short-url.git
   cd short-url
   ```

2. **创建功能分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **开发和测试**
   ```bash
   # 开发功能
   # ...
   
   # 运行测试
   make test
   make api-test
   ```

4. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   git push origin feature/your-feature-name
   ```

5. **创建 Pull Request**
   - 描述清楚功能变更
   - 确保测试通过
   - 添加必要的文档

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

类型：
- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更新
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建配置等

### 代码审查

所有代码变更需要经过代码审查：

1. **自检清单**
   - [ ] 代码格式化 (`make fmt`)
   - [ ] 通过所有测试 (`make test`)
   - [ ] 添加必要的测试用例
   - [ ] 更新相关文档
   - [ ] 性能影响评估

2. **审查重点**
   - 代码质量和可维护性
   - 安全性考虑
   - 性能影响
   - API 设计一致性

## 🔍 调试和故障排除

### 常见问题

1. **依赖问题**
   ```bash
   # 清理和重新下载依赖
   go clean -modcache
   go mod download
   ```

2. **数据库连接问题**
   ```bash
   # 检查数据库状态
   make status
   
   # 重置数据库
   make db-reset
   ```

3. **Redis 连接问题**
   ```bash
   # 检查 Redis 模块
   docker-compose exec redis redis-cli MODULE LIST
   ```

### 开发环境重置
```bash
# 完全清理和重新开始
make clean-all
make dev
```

## 📚 学习资源

- [Go 官方文档](https://golang.org/doc/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [Redis 文档](https://redis.io/documentation)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)
- [Docker 文档](https://docs.docker.com/)

欢迎加入我们的开发社区，一起构建更好的短链接服务！ 