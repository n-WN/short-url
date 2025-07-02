# 部署指南

## 概述

本文档提供短链接服务在不同环境下的部署方案和最佳实践。

## 🐳 Docker 部署（推荐）

### 快速部署

```bash
# 1. 克隆项目
git clone <repository-url>
cd short-url

# 2. 一键启动
make docker-up

# 3. 验证部署
make status
make api-test
```

### 生产环境配置

1. **创建生产配置**
   ```bash
   cp config.env.example config.prod.env
   ```

2. **编辑生产配置**
   ```env
   # 应用配置
   APP_PORT=8080
   BASE_URL=https://your-domain.com
   
   # 数据库配置
   DB_HOST=postgres
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your-secure-password
   DB_NAME=shorturl
   
   # Redis 配置
   REDIS_HOST=redis
   REDIS_PORT=6379
   REDIS_PASSWORD=your-redis-password
   
   # 布隆过滤器配置
   BLOOM_FILTER_CAPACITY=10000000
   BLOOM_FILTER_ERROR_RATE=0.001
   ```

3. **生产环境 docker-compose.yml**
   ```yaml
   version: '3.8'
   
   services:
     app:
       image: short-url:latest
       ports:
         - "8080:8080"
       environment:
         - BASE_URL=https://your-domain.com
       env_file:
         - config.prod.env
       depends_on:
         - postgres
         - redis
       deploy:
         resources:
           limits:
             memory: 512M
             cpus: '1.0'
         restart_policy:
           condition: on-failure
           max_attempts: 3
       healthcheck:
         test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
         interval: 30s
         timeout: 10s
         retries: 3
   
     postgres:
       image: postgres:15-alpine
       environment:
         POSTGRES_DB: shorturl
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: your-secure-password
       volumes:
         - postgres_data:/var/lib/postgresql/data
         - ./sql/migrations:/docker-entrypoint-initdb.d
       deploy:
         resources:
           limits:
             memory: 1G
             cpus: '1.0'
   
     redis:
       image: redis/redis-stack:latest
       command: redis-stack-server --protected-mode yes --requirepass your-redis-password
       volumes:
         - redis_data:/data
       deploy:
         resources:
           limits:
             memory: 512M
             cpus: '0.5'
   
     nginx:
       image: nginx:alpine
       ports:
         - "80:80"
         - "443:443"
       volumes:
         - ./nginx/nginx.conf:/etc/nginx/nginx.conf
         - ./ssl:/etc/nginx/ssl
       depends_on:
         - app
   
   volumes:
     postgres_data:
     redis_data:
   ```

## ☁️ 云平台部署

### AWS 部署

1. **使用 ECS**
   ```bash
   # 构建镜像并推送到 ECR
   aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin <account>.dkr.ecr.us-west-2.amazonaws.com
   
   docker build -t short-url .
   docker tag short-url:latest <account>.dkr.ecr.us-west-2.amazonaws.com/short-url:latest
   docker push <account>.dkr.ecr.us-west-2.amazonaws.com/short-url:latest
   ```

2. **RDS 和 ElastiCache 配置**
   ```env
   # 使用托管服务
   DB_HOST=your-rds-endpoint.amazonaws.com
   REDIS_HOST=your-elasticache-endpoint.amazonaws.com
   ```

### Google Cloud 部署

1. **使用 Cloud Run**
   ```bash
   # 部署到 Cloud Run
   gcloud run deploy short-url \
     --image gcr.io/your-project/short-url:latest \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

2. **使用 Cloud SQL 和 Memorystore**
   ```env
   DB_HOST=/cloudsql/your-project:us-central1:your-instance
   REDIS_HOST=your-memorystore-ip
   ```

### Kubernetes 部署

1. **创建 Kubernetes 清单**
   ```yaml
   # deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: short-url
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: short-url
     template:
       metadata:
         labels:
           app: short-url
       spec:
         containers:
         - name: short-url
           image: short-url:latest
           ports:
           - containerPort: 8080
           env:
           - name: DB_HOST
             value: "postgres"
           - name: REDIS_HOST
             value: "redis"
           resources:
             requests:
               memory: "256Mi"
               cpu: "250m"
             limits:
               memory: "512Mi"
               cpu: "500m"
           livenessProbe:
             httpGet:
               path: /health
               port: 8080
             initialDelaySeconds: 30
             periodSeconds: 10
           readinessProbe:
             httpGet:
               path: /health
               port: 8080
             initialDelaySeconds: 5
             periodSeconds: 5
   
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: short-url-service
   spec:
     selector:
       app: short-url
     ports:
     - protocol: TCP
       port: 80
       targetPort: 8080
     type: LoadBalancer
   ```

2. **部署到集群**
   ```bash
   kubectl apply -f deployment.yaml
   kubectl apply -f service.yaml
   ```

## 🔧 环境配置

### 环境变量

| 变量名 | 描述 | 默认值 | 必需 |
|--------|------|--------|------|
| `APP_PORT` | 应用端口 | 8080 | 否 |
| `BASE_URL` | 基础URL | http://localhost:8080 | 是 |
| `DB_HOST` | 数据库主机 | localhost | 是 |
| `DB_PORT` | 数据库端口 | 5432 | 否 |
| `DB_USER` | 数据库用户 | postgres | 是 |
| `DB_PASSWORD` | 数据库密码 | - | 是 |
| `DB_NAME` | 数据库名 | shorturl | 是 |
| `REDIS_HOST` | Redis主机 | localhost | 是 |
| `REDIS_PORT` | Redis端口 | 6379 | 否 |
| `REDIS_PASSWORD` | Redis密码 | - | 否 |
| `BLOOM_FILTER_CAPACITY` | 布隆过滤器容量 | 1000000 | 否 |
| `BLOOM_FILTER_ERROR_RATE` | 错误率 | 0.001 | 否 |

### SSL/TLS 配置

1. **Nginx 配置示例**
   ```nginx
   server {
       listen 80;
       server_name your-domain.com;
       return 301 https://$server_name$request_uri;
   }
   
   server {
       listen 443 ssl;
       server_name your-domain.com;
   
       ssl_certificate /etc/nginx/ssl/cert.pem;
       ssl_certificate_key /etc/nginx/ssl/key.pem;
   
       location / {
           proxy_pass http://app:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   ```

2. **Let's Encrypt 自动证书**
   ```bash
   # 使用 Certbot
   certbot --nginx -d your-domain.com
   ```

## 📊 监控和健康检查

### 健康检查端点

- **基础健康检查**: `GET /health`
- **内存状态**: `GET /debug/memory`
- **统计信息**: `GET /api/v1/stats`

### 监控指标

1. **应用指标**
   ```bash
   # 服务状态
   curl http://localhost:8080/health
   
   # 系统统计
   curl http://localhost:8080/api/v1/stats
   
   # 内存使用
   curl http://localhost:8080/debug/memory
   ```

2. **系统指标**
   ```bash
   # Docker 资源使用
   docker stats
   
   # 系统资源
   top, htop, iostat
   ```

### 日志配置

1. **应用日志**
   ```bash
   # 查看应用日志
   docker-compose logs -f app
   
   # 或使用 make 命令
   make logs
   ```

2. **日志聚合**
   ```yaml
   # 添加到 docker-compose.yml
   logging:
     driver: "json-file"
     options:
       max-size: "200m"
       max-file: "10"
   ```

## 🛡️ 安全配置

### 网络安全

1. **防火墙配置**
   ```bash
   # 只开放必要端口
   ufw allow 22      # SSH
   ufw allow 80      # HTTP
   ufw allow 443     # HTTPS
   ufw enable
   ```

2. **Docker 安全**
   ```yaml
   # 非 root 用户运行
   services:
     app:
       user: "1000:1000"
       security_opt:
         - no-new-privileges:true
   ```

### 数据安全

1. **数据库安全**
   ```sql
   -- 创建专用用户
   CREATE USER shorturl_user WITH PASSWORD 'secure_password';
   GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO shorturl_user;
   ```

2. **Redis 安全**
   ```bash
   # 设置密码
   requirepass your-redis-password
   
   # 禁用危险命令
   rename-command FLUSHDB ""
   rename-command FLUSHALL ""
   ```

## 📈 性能优化

### 资源配置

1. **容器资源限制**
   ```yaml
   deploy:
     resources:
       limits:
         memory: 512M
         cpus: '1.0'
       reservations:
         memory: 256M
         cpus: '0.5'
   ```

2. **数据库优化**
   ```sql
   -- 调整 PostgreSQL 配置
   shared_buffers = 256MB
   effective_cache_size = 1GB
   work_mem = 4MB
   ```

### 缓存优化

1. **Redis 配置**
   ```bash
   # redis.conf
   maxmemory 512mb
   maxmemory-policy allkeys-lru
   ```

2. **CDN 配置**
   ```nginx
   # 静态资源缓存
   location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg)$ {
       expires 1y;
       add_header Cache-Control "public, immutable";
   }
   ```

## 🔄 备份和恢复

### 数据库备份

1. **自动备份脚本**
   ```bash
   #!/bin/bash
   # backup.sh
   DATE=$(date +%Y%m%d_%H%M%S)
   docker-compose exec postgres pg_dump -U postgres shorturl > backup_${DATE}.sql
   ```

2. **恢复数据**
   ```bash
   docker-compose exec -T postgres psql -U postgres shorturl < backup_file.sql
   ```

### Redis 备份

```bash
# 创建 RDB 快照
docker-compose exec redis redis-cli BGSAVE

# 复制备份文件
docker cp container_name:/data/dump.rdb ./redis_backup.rdb
```

## 🚀 CI/CD 集成

### GitHub Actions 示例

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Build and Deploy
      run: |
        docker build -t short-url .
        docker tag short-url your-registry/short-url:latest
        docker push your-registry/short-url:latest
        
    - name: Deploy to Production
      run: |
        ssh user@server 'docker pull your-registry/short-url:latest'
        ssh user@server 'docker-compose up -d'
```

### 滚动更新

```bash
# 零停机更新
docker-compose pull
docker-compose up -d --no-deps app
```

这个部署指南提供了从开发到生产的完整部署流程，确保服务的稳定性和安全性。 