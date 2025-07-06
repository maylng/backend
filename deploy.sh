#!/bin/bash

# Maylng Backend Production Deployment Script
# This script sets up and deploys the Maylng email API in production

set -e

echo "🚀 Starting Maylng Backend Production Deployment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo "❌ Docker Compose is not installed. Please install it and try again."
    exit 1
fi

# Check if the secure image exists
if ! docker images | grep -q "maylng/backend.*secure"; then
    echo "🔨 Building secure Docker image..."
    docker build -t maylng/backend:secure .
    echo "✅ Secure image built successfully"
else
    echo "✅ Secure Docker image already exists"
fi

# Create production environment file if it doesn't exist
if [ ! -f .env.production ]; then
    echo "📝 Creating production environment file..."
    cp .env.production .env
    echo "⚠️  Please edit .env file with your production values before continuing!"
    echo "   Required: SENDGRID_API_KEY"
    echo "   Recommended: Change default passwords"
    read -p "Press Enter when you've configured .env file..."
fi

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Validate required environment variables
if [ -z "$SENDGRID_API_KEY" ] || [ "$SENDGRID_API_KEY" = "your_sendgrid_api_key_here" ]; then
    echo "❌ SENDGRID_API_KEY is not configured. Please set it in .env file."
    exit 1
fi

echo "🏗️  Starting production services..."

# Stop any existing services
docker-compose -f docker-compose.prod.yml down --remove-orphans

# Start production services
docker-compose -f docker-compose.prod.yml up -d

echo "⏳ Waiting for services to be healthy..."

# Wait for services to be ready
sleep 10

# Check service health
echo "🔍 Checking service health..."

# Check PostgreSQL
if docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -U maylng > /dev/null 2>&1; then
    echo "✅ PostgreSQL is healthy"
else
    echo "❌ PostgreSQL is not healthy"
fi

# Check Redis
if docker-compose -f docker-compose.prod.yml exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is healthy"
else
    echo "❌ Redis is not healthy"
fi

# Check API health endpoint
sleep 5
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ API is healthy and responding"
else
    echo "❌ API health check failed"
fi

echo ""
echo "🎉 Maylng Backend Production Deployment Complete!"
echo ""
echo "📊 Service Status:"
docker-compose -f docker-compose.prod.yml ps
echo ""
echo "🔗 API Endpoints:"
echo "   Health Check: http://localhost:8080/health"
echo "   API Base URL: http://localhost:8080/api/v1"
echo ""
echo "📋 Next Steps:"
echo "   1. Test API endpoints"
echo "   2. Configure your SDK to use http://localhost:8080"
echo "   3. Set up SSL certificate for HTTPS"
echo "   4. Configure domain name"
echo ""
echo "🔧 Management Commands:"
echo "   View logs: docker-compose -f docker-compose.prod.yml logs -f"
echo "   Stop services: docker-compose -f docker-compose.prod.yml down"
echo "   Restart: docker-compose -f docker-compose.prod.yml restart"
echo ""
