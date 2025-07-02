# 短链接服务 API 文档

## 概述

这是一个高性能的短链接服务，支持创建、管理和访问短链接。服务使用 Go 语言开发，集成了 Redis 缓存、布隆过滤器和 PostgreSQL 数据库。

## 基础信息

- **基础URL**: `http://localhost:8080`
- **API版本**: v1
- **内容类型**: `application/json`

## API 端点

### 1. 健康检查

**端点**: `GET /health`

**描述**: 检查服务运行状态

**响应示例**:
```json
{
  "data": {
    "status": "ok",
    "timestamp": "2025-07-02T20:13:30Z"
  }
}
```

### 2. 创建短链接

**端点**: `POST /api/v1/shorten`

**请求体**:
```json
{
  "url": "https://www.example.com",           // 必需：原始URL
  "custom_code": "mycustom",                  // 可选：自定义短码
  "expires_at": "2025-12-31T23:59:59Z"       // 可选：过期时间
}
```

**成功响应 (201)**:
```json
{
  "data": {
    "short_url": "http://localhost:8080/abc123",
    "short_code": "abc123",
    "original_url": "https://www.example.com",
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2025-07-02T20:13:30.775473Z"
  },
  "message": "short link created successfully"
}
```

**错误响应**:
- `400 Bad Request`: 无效的URL格式
- `409 Conflict`: 自定义短码已存在

### 3. 短链接重定向

**端点**: `GET /{short_code}`

**描述**: 重定向到原始URL

**响应**: 
- `302 Found`: 成功重定向
- `404 Not Found`: 短码不存在
- `410 Gone`: 短链接已过期

### 4. 获取短链接信息

**端点**: `GET /api/v1/info/{short_code}`

**响应示例**:
```json
{
  "data": {
    "short_code": "abc123",
    "original_url": "https://www.example.com",
    "access_count": 42,
    "created_at": "2025-07-02T20:13:30.775473Z",
    "expires_at": "2025-12-31T23:59:59Z"
  }
}
```

### 5. 获取统计信息

**端点**: `GET /api/v1/stats`

**响应示例**:
```json
{
  "data": {
    "total_links": 1000,
    "total_accesses": 50000,
    "active_links": 950,
    "expired_links": 50,
    "permanent_links": 900
  }
}
```

### 6. 清理过期链接 (管理员)

**端点**: `POST /api/v1/admin/clean`

**描述**: 删除所有过期的短链接

**响应示例**:
```json
{
  "data": {
    "deleted_count": 25,
    "timestamp": "2025-07-02T20:13:30Z"
  },
  "message": "expired links cleaned successfully"
}
```

## 错误响应格式

所有错误响应遵循统一格式：

```json
{
  "error": "Bad Request",
  "message": "具体错误描述",
  "code": 400
}
```

## 状态码说明

- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `302 Found`: 重定向
- `400 Bad Request`: 请求格式错误
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突
- `410 Gone`: 资源已过期
- `500 Internal Server Error`: 服务器内部错误

## 使用示例

### 创建短链接
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com"}'
```

### 访问短链接
```bash
curl -L http://localhost:8080/abc123
```

### 获取链接信息
```bash
curl http://localhost:8080/api/v1/info/abc123
```

## 特性

- 🚀 **高性能**: 使用 Redis 缓存和布隆过滤器优化
- 🔒 **防重复**: 布隆过滤器快速检测重复短码
- ⏰ **过期控制**: 支持设置链接过期时间
- 📊 **访问统计**: 记录每个链接的访问次数
- 🛡️ **错误处理**: 完善的错误处理和响应
- 🔍 **URL验证**: 严格的URL格式验证 