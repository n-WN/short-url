# Makefile 使用指南

## 概述
短链接服务提供了丰富的 Makefile 命令来简化开发、测试、部署和监控流程。

## 🚀 快速开始

### 开发环境快速启动
```bash
make dev
```
这个命令会：
- 启动所有 Docker 服务
- 等待服务就绪
- 显示服务状态
- 提供后续操作指引

### 查看所有可用命令
```bash
make help
```

## 📋 命令分类

### 🔧 开发环境
| 命令 | 描述 |
|------|------|
| `make deps` | 安装 Go 依赖 |
| `make build` | 构建应用程序 |
| `make run` | 直接运行应用（非容器） |
| `make migrate` | 运行数据库迁移 |
| `make test` | 运行单元测试 |
| `make clean` | 清理构建文件 |

### 🐳 Docker 操作
| 命令 | 描述 |
|------|------|
| `make docker-build` | 构建 Docker 镜像 |
| `make docker-rebuild` | 强制重新构建镜像（清除缓存） |
| `make docker-up` | 启动所有服务 |
| `make docker-down` | 停止所有服务 |
| `make docker-restart` | 完整重启（停止→重建→启动） |
| `make docker-logs` | 查看容器日志 |

### 💾 数据库操作
| 命令 | 描述 |
|------|------|
| `make db-up` | 仅启动数据库和 Redis |
| `make db-down` | 停止数据库和 Redis |
| `make db-reset` | 完全重置数据库（⚠️ 会删除所有数据） |

### 🔍 监控和调试
| 命令 | 描述 |
|------|------|
| `make status` | 查看服务状态和健康检查 |
| `make stats` | 查看服务统计信息 |
| `make memory-debug` | 查看内存调试信息 |
| `make memory-monitor` | 启动实时内存监控 |
| `make memory-monitor-bg` | 后台启动内存监控 |
| `make memory-analysis` | 内存分析指引 |
| `make logs` | 查看应用日志 |
| `make debug` | 调试模式（查看详细日志） |

### 🧪 测试和验证
| 命令 | 描述 |
|------|------|
| `make api-test` | 基础 API 功能测试 |
| `make functional-test` | 完整功能测试 |
| `make load-test` | 快速负载测试（1分钟） |
| `make benchmark` | 标准压力测试 |
| `make performance-test` | 详细性能分析 |
| `make stress-test` | 运行所有压力测试 |

### 🧹 清理和维护
| 命令 | 描述 |
|------|------|
| `make clean` | 清理构建文件 |
| `make clean-containers` | 清理 Docker 容器和镜像 |
| `make clean-all` | 完全清理所有资源 |
| `make fix-permissions` | 修复脚本执行权限 |

### 📊 报告生成
| 命令 | 描述 |
|------|------|
| `make report` | 生成完整的服务状态报告 |

### ⚙️ 开发工具
| 命令 | 描述 |
|------|------|
| `make fmt` | 格式化 Go 代码 |
| `make lint` | 运行代码检查 |

## 📖 常用场景

### 🎯 日常开发流程
```bash
# 1. 启动开发环境
make dev

# 2. 测试 API
make api-test

# 3. 查看内存使用
make memory-debug

# 4. 运行压力测试
make load-test
```

### 🐛 问题排查
```bash
# 查看服务状态
make status

# 查看详细日志
make debug

# 检查内存使用
make memory-monitor

# 生成诊断报告
make report
```

### 🔄 代码变更后的流程
```bash
# 重新构建和启动
make docker-restart

# 验证功能
make functional-test

# 性能测试
make benchmark
```

### 🧹 环境清理
```bash
# 轻度清理
make clean

# 重置数据库
make db-reset

# 完全清理
make clean-all
```

## 🔧 故障排除

### 权限问题
```bash
make fix-permissions
```

### 服务无法启动
```bash
# 查看详细日志
make debug

# 完全重启
make docker-restart
```

### 数据库问题
```bash
# 重置数据库
make db-reset
```

### 内存问题
```bash
# 实时监控
make memory-monitor

# 详细分析
make memory-debug
```

## 💡 最佳实践

1. **开发前**: 运行 `make dev` 确保环境就绪
2. **代码修改后**: 使用 `make docker-restart` 重新部署
3. **性能测试**: 先运行 `make load-test`，再进行 `make benchmark`
4. **问题排查**: 使用 `make status` 和 `make debug`
5. **定期清理**: 运行 `make clean-containers` 清理 Docker 资源

## 📈 性能监控工作流

1. 启动服务：`make docker-up`
2. 基线测试：`make api-test`
3. 负载测试：`make load-test`
4. 内存监控：`make memory-monitor-bg`
5. 压力测试：`make stress-test`
6. 生成报告：`make report`

这样的工作流能够全面评估服务的性能表现和资源使用情况。 