# Maylng Backend API Design

## Tech Stack Decision: Go + Gin

**Framework**: Go with Gin (faster, more lightweight than Echo for our use case)  
**Database**: PostgreSQL + Redis  
**Email Infrastructure**: Multi-provider approach  

## Email Infrastructure Comparison

### 1. AWS SES (Simple Email Service)

**Pros:**

- Very cost-effective ($0.10 per 1,000 emails)
- High deliverability rates
- Built-in bounce/complaint handling
- Scales automatically
- Good for transactional emails

**Cons:**

- Requires AWS account setup
- Initial sandbox mode restrictions
- Less flexible than custom SMTP
- Requires domain verification

**Best for**: High-volume, cost-sensitive applications

### 2. SendGrid

**Pros:**

- Excellent deliverability reputation
- Rich API and webhooks
- Advanced analytics and reporting
- Easy integration
- Good free tier (100 emails/day)
- Template management

**Cons:**

- More expensive than SES ($14.95/month for 40k emails)
- Can be overkill for simple use cases
- Third-party dependency

**Best for**: Marketing emails, advanced analytics needs

### 3. Postmark

**Pros:**

- Laser-focused on transactional emails
- Excellent deliverability (99%+)
- Fast delivery (usually < 5 seconds)
- Great developer experience
- Detailed delivery tracking

**Cons:**

- More expensive ($1.25 per 1,000 emails)
- Limited marketing email features
- Smaller scale compared to others

**Best for**: Critical transactional emails, developer-focused apps

### 4. Custom SMTP Servers

**Pros:**

- Full control over infrastructure
- No per-email costs (after setup)
- Complete customization
- No third-party dependencies

**Cons:**

- Complex setup and maintenance
- Deliverability challenges
- Need to handle bounces/complaints manually
- Requires email expertise

**Best for**: Large enterprises, specific compliance needs

## Recommended Approach: Multi-Provider Strategy

Start with **SendGrid** for reliability and developer experience, with the architecture designed to easily switch or add providers:

```go
type EmailProvider interface {
    SendEmail(email *Email) (*SendResult, error)
    GetDeliveryStatus(messageID string) (*DeliveryStatus, error)
}

type EmailService struct {
    primary   EmailProvider
    fallback  EmailProvider
}
```

## Database Schema Design

### PostgreSQL Tables

```sql
-- Users/Accounts
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_hash VARCHAR(255) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    email_limit_per_month INTEGER DEFAULT 1000,
    email_address_limit INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Email Addresses
CREATE TABLE email_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('temporary', 'persistent')),
    prefix VARCHAR(100),
    domain VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'expired', 'disabled')),
    expires_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_email_addresses_account_id (account_id),
    INDEX idx_email_addresses_email (email),
    INDEX idx_email_addresses_expires_at (expires_at)
);

-- Sent Emails
CREATE TABLE sent_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    from_email_id UUID NOT NULL REFERENCES email_addresses(id),
    to_recipients JSONB NOT NULL,
    cc_recipients JSONB,
    bcc_recipients JSONB,
    subject VARCHAR(998) NOT NULL,
    text_content TEXT,
    html_content TEXT,
    attachments JSONB,
    headers JSONB,
    thread_id UUID,
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'queued' CHECK (status IN ('queued', 'sent', 'delivered', 'failed', 'scheduled')),
    provider_message_id VARCHAR(255),
    failure_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_sent_emails_account_id (account_id),
    INDEX idx_sent_emails_from_email_id (from_email_id),
    INDEX idx_sent_emails_status (status),
    INDEX idx_sent_emails_scheduled_at (scheduled_at),
    INDEX idx_sent_emails_thread_id (thread_id)
);

-- Email Analytics
CREATE TABLE email_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id UUID NOT NULL REFERENCES sent_emails(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('delivered', 'bounced', 'opened', 'clicked', 'complained', 'unsubscribed')),
    event_data JSONB,
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_email_analytics_email_id (email_id),
    INDEX idx_email_analytics_event_type (event_type),
    INDEX idx_email_analytics_occurred_at (occurred_at)
);

-- Rate Limiting (Redis alternative in PostgreSQL)
CREATE TABLE rate_limits (
    key VARCHAR(255) PRIMARY KEY,
    count INTEGER DEFAULT 1,
    window_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);
```

### Redis Schema

```md
# Rate limiting
rate_limit:api:{account_id}:{window} -> count
rate_limit:email:{account_id}:{window} -> count

# Caching
account:{api_key_hash} -> account_data (TTL: 1 hour)
email_address:{id} -> email_address_data (TTL: 30 minutes)

# Email queue
email_queue:immediate -> [email_ids]
email_queue:scheduled -> sorted_set(email_id, scheduled_timestamp)

# Temporary email cleanup
temp_email_cleanup -> sorted_set(email_id, expires_at)
```

## API Endpoints Design

### Authentication

```md
Header: Authorization: Bearer {api_key}
```

### Core Endpoints

```http
# Health Check
GET /health
GET /v1/health

# Account Management
GET /v1/account

# Email Address Management
POST /v1/email-addresses
GET /v1/email-addresses
GET /v1/email-addresses/{id}
PATCH /v1/email-addresses/{id}
DELETE /v1/email-addresses/{id}
POST /v1/email-addresses/{id}/extend

# Email Operations
POST /v1/emails/send
GET /v1/emails
GET /v1/emails/{id}
POST /v1/emails/{id}/cancel
POST /v1/emails/{id}/resend
GET /v1/emails/{id}/status

# Webhooks (for email providers)
POST /webhooks/sendgrid
POST /webhooks/ses
POST /webhooks/postmark
```

## Go Project Structure

```md
maylng-api/
├── cmd/
│   ├── api/
│   │   └── main.go
│   ├── worker/
│   │   └── main.go
│   └── migrate/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── routes/
│   │   └── server.go
│   ├── auth/
│   │   └── auth.go
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── migrations/
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── email/
│   │   ├── providers/
│   │   │   ├── sendgrid.go
│   │   │   ├── ses.go
│   │   │   └── postmark.go
│   │   ├── queue.go
│   │   └── service.go
│   ├── models/
│   │   ├── account.go
│   │   ├── email_address.go
│   │   └── sent_email.go
│   ├── services/
│   │   ├── account.go
│   │   ├── email_address.go
│   │   └── email.go
│   └── utils/
│       ├── validator.go
│       └── logger.go
├── pkg/
│   ├── errors/
│   └── types/
├── migrations/
├── docker/
├── deployments/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Rate Limiting Strategy

```go
// Rate limits per plan
var RateLimits = map[string]RateLimit{
    "free": {
        EmailsPerHour:     100,
        EmailsPerDay:     1000,
        EmailsPerMonth:   1000,
        APICallsPerHour: 1000,
        EmailAddresses:    5,
    },
    "pro": {
        EmailsPerHour:     1000,
        EmailsPerDay:     10000,
        EmailsPerMonth:   50000,
        APICallsPerHour: 10000,
        EmailAddresses:    50,
    },
    "enterprise": {
        EmailsPerHour:     10000,
        EmailsPerDay:     100000,
        EmailsPerMonth:   1000000,
        APICallsPerHour: 100000,
        EmailAddresses:    500,
    },
}
```

## Deployment Architecture

```yaml
# Docker Compose for development
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/maylng
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  worker:
    build: .
    command: ./worker
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: maylng
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

## Security Considerations

1. **API Key Security**
   - Hash API keys in database
   - Use secure random generation
   - Implement key rotation

2. **Rate Limiting**
   - Per-account limits
   - IP-based limits for abuse prevention
   - Sliding window counters

3. **Email Security**
   - SPF/DKIM/DMARC validation
   - Content filtering
   - Attachment scanning

4. **Infrastructure**
   - TLS everywhere
   - Database encryption at rest
   - Secure webhook validation

## Next Steps

1. **MVP Features** (Week 1)
   - Basic API structure
   - Email address CRUD
   - Simple email sending with SendGrid
   - PostgreSQL setup

2. **Core Features** (Week 2)
   - Rate limiting
   - Authentication system
   - Email scheduling
   - Basic analytics

3. **Advanced Features** (Week 3)
   - Multi-provider email
   - Webhook handling
   - Advanced rate limiting
   - Monitoring/logging

Would you like me to start implementing any specific part of this backend architecture?
