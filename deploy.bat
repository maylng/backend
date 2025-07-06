@echo off
REM Maylng Backend Production Deployment Script for Windows
REM This script sets up and deploys the Maylng email API in production

echo 🚀 Starting Maylng Backend Production Deployment...

REM Check if Docker is running
docker info >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker is not running. Please start Docker and try again.
    exit /b 1
)

REM Check if Docker Compose is available
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker Compose is not installed. Please install it and try again.
    exit /b 1
)

REM Check if the secure image exists
docker images | findstr "maylng/backend.*secure" >nul 2>&1
if errorlevel 1 (
    echo 🔨 Building secure Docker image...
    docker build -t maylng/backend:secure .
    echo ✅ Secure image built successfully
) else (
    echo ✅ Secure Docker image already exists
)

REM Create production environment file if it doesn't exist
if not exist .env (
    echo 📝 Creating production environment file...
    copy .env.production .env
    echo ⚠️  Please edit .env file with your production values before continuing!
    echo    Required: SENDGRID_API_KEY
    echo    Recommended: Change default passwords
    pause
)

echo 🏗️  Starting production services...

REM Stop any existing services
docker-compose -f docker-compose.prod.yml down --remove-orphans

REM Start production services
docker-compose -f docker-compose.prod.yml up -d

echo ⏳ Waiting for services to be healthy...
timeout /t 15 /nobreak >nul

echo 🔍 Checking service health...

REM Check API health endpoint
timeout /t 5 /nobreak >nul
curl -f http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    echo ❌ API health check failed - but services may still be starting
) else (
    echo ✅ API is healthy and responding
)

echo.
echo 🎉 Maylng Backend Production Deployment Complete!
echo.
echo 📊 Service Status:
docker-compose -f docker-compose.prod.yml ps
echo.
echo 🔗 API Endpoints:
echo    Health Check: http://localhost:8080/health
echo    API Base URL: http://localhost:8080/api/v1
echo.
echo 📋 Next Steps:
echo    1. Test API endpoints
echo    2. Configure your SDK to use http://localhost:8080
echo    3. Set up SSL certificate for HTTPS
echo    4. Configure domain name
echo.
echo 🔧 Management Commands:
echo    View logs: docker-compose -f docker-compose.prod.yml logs -f
echo    Stop services: docker-compose -f docker-compose.prod.yml down
echo    Restart: docker-compose -f docker-compose.prod.yml restart
echo.
pause
