# 🚀 Maylng Backend Deployment Guide

## 🛡️ Security Status: FIXED

✅ **All vulnerabilities addressed!** Your Docker images now use:

- Distroless base images (no shell/package manager)
- Non-root user execution
- Minimal attack surface
- Security build flags

## 📦 Quick Start

### 1. Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit with your values
SENDGRID_API_KEY=your_sendgrid_key
DATABASE_URL=postgres://user:pass@localhost:5432/maylng
REDIS_URL=redis://localhost:6379
```

### 2. Database Setup

```bash
# Run migrations
go run cmd/migrate/main.go

# Or with Docker
docker run --rm -e DATABASE_URL="$DATABASE_URL" maylng/backend:secure migrate
```

### 3. Secure Docker Deployment

#### Production (Recommended)

```bash
# Build secure image
docker build -t maylng/backend:secure .

# Run API server
docker run -d \
  --name maylng-api \
  -p 8080:8080 \
  -e DATABASE_URL="$DATABASE_URL" \
  -e REDIS_URL="$REDIS_URL" \
  -e SENDGRID_API_KEY="$SENDGRID_API_KEY" \
  --health-cmd="wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=5s \
  --health-retries=3 \
  maylng/backend:secure

# Run background worker
docker run -d \
  --name maylng-worker \
  -e DATABASE_URL="$DATABASE_URL" \
  -e REDIS_URL="$REDIS_URL" \
  -e SENDGRID_API_KEY="$SENDGRID_API_KEY" \
  maylng/backend:secure worker
```

#### Alternative (Alpine)

```bash
# Build Alpine version
docker build -f Dockerfile.alpine -t maylng/backend:alpine .

# Use same run commands as above, but with :alpine tag
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
