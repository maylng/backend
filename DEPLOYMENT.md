# 🚀 Maylng Backend Deployment Guide

## 🛡️ Security Status: FIXED

✅ **All vulnerabilities addressed!** Your Docker images now use:

- Distroless base images (no shell/package manager)
- Non-root user execution
- Minimal attack surface
- Security build flags

## 📦 Quick Start

### 🚀 One-Click Production Deployment

```bash
# Windows
./deploy.bat

# Linux/Mac
chmod +x deploy.sh
./deploy.sh
```

This will automatically:

- ✅ Build the secure Docker image
- ✅ Set up PostgreSQL + Redis
- ✅ Run database migrations  
- ✅ Start API server + background worker
- ✅ Configure Nginx reverse proxy
- ✅ Set up health checks

### 🧪 Test Your Deployment

```bash
# Run API tests
chmod +x test-api.sh
./test-api.sh
```

### 📝 Manual Setup (Alternative)

#### 1. Environment Setup

```bash
# Copy environment template
cp .env.production .env

# Edit with your values (EMAIL PROVIDER REQUIRED)
# For SendGrid (legacy):
SENDGRID_API_KEY=your_sendgrid_key
EMAIL_PROVIDER=sendgrid

# For AWS SES (recommended):
EMAIL_PROVIDER=ses
AWS_REGION=us-east-1
# AWS credentials can be provided via:
# - AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables
# - IAM roles (if running on EC2)
# - AWS credential files

POSTGRES_PASSWORD=your_secure_password
REDIS_PASSWORD=your_redis_password
```

#### 2. Start Services

```bash
# Start all services
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps
```

### 3. Secure Docker Deployment

Your production environment is now running at:

- **API**: <http://localhost:8080>
- **Health Check**: <http://localhost:8080/health>  
- **Database**: PostgreSQL on port 5432
- **Cache**: Redis on port 6379
- **Proxy**: Nginx on port 80

#### 🔧 Management Commands

```bash
# View all service logs
docker-compose -f docker-compose.prod.yml logs -f

# View specific service logs
docker-compose -f docker-compose.prod.yml logs -f api
docker-compose -f docker-compose.prod.yml logs -f worker

# Restart services
docker-compose -f docker-compose.prod.yml restart

# Stop all services
docker-compose -f docker-compose.prod.yml down

# Update and redeploy
docker build -t maylng/backend:secure .
docker-compose -f docker-compose.prod.yml up -d
```

## 🔍 Security Verification

```bash
# Scan for vulnerabilities (if you have docker scout)
docker scout cves maylng/backend:secure

# Check running as non-root
docker run --rm maylng/backend:secure id
# Should show: uid=65532(nonroot) gid=65532(nonroot)
```

## 📊 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/accounts` | POST | Create account |
| `/api/v1/emails` | POST | Send email |
| `/api/v1/emails/addresses` | GET/POST | Manage email addresses |

## 🎯 Production Checklist

- ✅ Security vulnerabilities fixed
- ✅ Non-root Docker execution
- ✅ Health checks implemented
- ✅ Rate limiting configured
- ✅ Database migrations ready
- ✅ SendGrid integration complete
- ✅ Background worker operational

## 🔧 Monitoring

```bash
# Check container health
docker ps --filter "name=maylng"

# View logs
docker logs maylng-api
docker logs maylng-worker

# Monitor health endpoint
curl http://localhost:8080/health
```

Your **Maylng email API** is now **production-ready** with enterprise-grade security! 🛡️

## 📞 Need Help?

Check `SECURITY.md` for detailed security documentation and best practices.
