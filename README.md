# Maylng Backend API

A powerful and scalable email API service built with Go, Gin, PostgreSQL, and Redis. Maylng provides temporary and persistent email addresses with comprehensive email sending capabilities.

## 🚀 Features

- **Email Address Management**: Create temporary and persistent email addresses
- **Multi-Provider Email Sending**: SendGrid integration with fallback support
- **Scalable Architecture**: Built with Go, PostgreSQL, and Redis
- **Rate Limiting**: Per-account and per-plan rate limiting
- **Scheduled Emails**: Send emails immediately or schedule for later
- **Email Analytics**: Track delivery, opens, clicks, and more
- **RESTful API**: Clean and intuitive REST API design
- **Authentication**: API key-based authentication
- **Background Processing**: Async email processing with worker service

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Redis 6 or higher
- SendGrid API key (for production email sending)

## 🛠️ Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd backend
```

### 2. Install Dependencies

```bash
make deps
```

### 3. Environment Configuration

```bash
cp .env.example .env
# Edit .env with your configuration
```

Required environment variables:

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `SENDGRID_API_KEY`: Your SendGrid API key
- `API_KEY_HASH_SALT`: Random salt for API key hashing

### 4. Database Setup

```bash
# Build the migration tool
make build

# Run migrations
make migrate-up
```

### 5. Run the Application

```bash
# Start the API server
make run

# In another terminal, start the worker
make run-worker
```

The API will be available at `http://localhost:8080`

## 🐳 Docker Development

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## 📚 API Documentation

### Authentication

All API requests (except account creation) require authentication via API key:

```bash
curl -H "Authorization: Bearer your_api_key" http://localhost:8080/v1/account
```

### Core Endpoints

#### Account Management

- `POST /v1/accounts` - Create a new account
- `GET /v1/account` - Get account details

#### Email Addresses

- `POST /v1/email-addresses` - Create email address
- `GET /v1/email-addresses` - List email addresses
- `GET /v1/email-addresses/{id}` - Get email address
- `PATCH /v1/email-addresses/{id}` - Update email address
- `DELETE /v1/email-addresses/{id}` - Delete email address

#### Email Operations

- `POST /v1/emails/send` - Send email
- `GET /v1/emails` - List sent emails
- `GET /v1/emails/{id}` - Get email details
- `GET /v1/emails/{id}/status` - Get email status

### Example Usage

#### 1. Create Account

```bash
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"plan": "starter"}'
```

Response:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "plan": "starter",
  "email_limit_per_month": 5000,
  "email_address_limit": 5,
  "api_key": "maylng_1234567890abcdef...",
  "created_at": "2025-07-04T10:00:00Z"
}
```

#### 2. Create Email Address

```bash
curl -X POST http://localhost:8080/v1/email-addresses \
  -H "Authorization: Bearer your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "temporary",
    "prefix": "test"
  }'
```

#### 3. Send Email

```bash
curl -X POST http://localhost:8080/v1/emails/send \
  -H "Authorization: Bearer your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "from_email_id": "email_address_id",
    "to_recipients": ["recipient@example.com"],
    "subject": "Hello from Maylng!",
    "text_content": "This is a test email.",
    "html_content": "<p>This is a <strong>test</strong> email.</p>"
  }'
```

## 🏗️ Architecture

### Project Structure

```md
backend/
├── cmd/                    # Application entry points
│   ├── api/               # API server
│   ├── worker/            # Background worker
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── api/              # HTTP handlers and routes
│   ├── auth/             # Authentication logic
│   ├── config/           # Configuration management
│   ├── database/         # Database connections
│   ├── email/            # Email service and providers
│   ├── models/           # Data models
│   └── services/         # Business logic
├── migrations/           # Database migrations
└── pkg/                 # Public packages (if any)
```

### Key Components

- **API Server**: Handles HTTP requests and responses
- **Worker Service**: Processes scheduled emails and cleanup tasks
- **Email Service**: Multi-provider email sending with fallback
- **Database Layer**: PostgreSQL for persistence, Redis for caching
- **Authentication**: API key-based auth with hashed storage

## 🔧 Development

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build all binaries
make test              # Run tests
make test-coverage     # Run tests with coverage
make format            # Format Go code
make lint              # Run golangci-lint
make clean             # Clean build artifacts
make setup             # Setup development environment
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/services/...
```

### Database Operations

```bash
# Create new migration
make migrate-create NAME=add_new_feature

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## 📊 Monitoring and Logging

The application provides comprehensive logging and health check endpoints:

- `GET /health` - Basic health check
- `GET /v1/health` - Detailed health check with version info

Logs are structured and include:

- Request/response logging
- Database operation logs
- Email sending logs
- Error tracking

## 🔒 Security

- API keys are hashed using SHA-256 with salt
- Rate limiting per account and plan
- Input validation on all endpoints
- SQL injection protection with parameterized queries
- CORS support for web applications

## 📈 Rate Limits

### Starter Plan

- 250 emails/hour
- 2,500 emails/day
- 5,000 emails/month
- 5 email addresses

### Pro Plan

- 2,500 emails/hour
- 25,000 emails/day
- 50,000 emails/month
- 50 email addresses

### Enterprise Plan

- 10,000 emails/hour
- 100,000 emails/day
- 175,000 emails/month
- 500 email addresses

## 🚀 Production Deployment

### Environment Variables

```bash
# Required
DATABASE_URL=postgres://user:pass@host:5432/dbname
REDIS_URL=redis://host:6379
SENDGRID_API_KEY=your_sendgrid_key
API_KEY_HASH_SALT=random_salt_string

# Optional
GIN_MODE=release
LOG_LEVEL=info
PORT=8080
ENVIRONMENT=production
```

### Docker Deployment

```bash
# Build image
docker build -t maylng/backend .

# Run container
docker run -d \
  --name maylng-api \
  -p 8080:8080 \
  --env-file .env \
  maylng/backend
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support, please create an issue in the GitHub repository or contact the development team.

---

Built with ❤️ using Go, PostgreSQL, and Redis.
