# Maylng API Documentation

A powerful and scalable email API service for creating temporary and persistent email addresses with comprehensive email sending capabilities.

## 🚀 Overview

The Maylng API allows you to:

- Create and manage temporary/persistent email addresses
- Send emails with full HTML/text support
- Track email delivery status
- Manage account settings and rate limits
- Schedule emails for future delivery

**Base URL**: `https://api.mayl.ng:8080`  
**Current Version**: `v1`  
**Protocol**: HTTPS  
**Format**: JSON  

## 🔐 Authentication

All API requests (except account creation) require authentication using an API key in the Authorization header:

```bash
Authorization: Bearer maylng_your_api_key_here
```

### Getting an API Key

Create an account to receive your API key:

```bash
curl -X POST https://api.mayl.ng/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"plan": "free"}'
```

## 📊 Rate Limits

Rate limits vary by plan:

| Plan | Emails/Hour | Emails/Day | Emails/Month | Email Addresses | API Calls/Hour |
|------|-------------|------------|--------------|-----------------|----------------|
| **Free** | 100 | 1,000 | 1,000 | 5 | 1,000 |
| **Pro** | 1,000 | 10,000 | 50,000 | 50 | 10,000 |
| **Enterprise** | 10,000 | 100,000 | 1,000,000 | 500 | 100,000 |

Rate limit information is included in response headers:

- `X-RateLimit-Limit`: Request limit per time window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the current window resets

## 📚 API Endpoints

### Health Check

#### Check API Health

```http
GET /health
```

**Response:**

```json
{
  "status": "ok",
  "service": "maylng-api",
  "version": "1.0.0"
}
```

---

### Account Management

#### Create Account

```http
POST /v1/accounts
```

**Request Body:**

```json
{
  "plan": "free"  // Optional: "free", "pro", "enterprise"
}
```

**Response:**

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "plan": "free",
  "email_limit_per_month": 1000,
  "email_address_limit": 5,
  "api_key": "maylng_1234567890abcdef...",
  "created_at": "2025-07-06T10:00:00Z",
  "updated_at": "2025-07-06T10:00:00Z"
}
```

#### Get Account Details

```http
GET /v1/account
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:**

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "plan": "free",
  "email_limit_per_month": 1000,
  "email_address_limit": 5,
  "created_at": "2025-07-06T10:00:00Z",
  "updated_at": "2025-07-06T10:00:00Z"
}
```

#### Update Account

```http
PATCH /v1/account
```

**Headers:**

```http
Authorization: Bearer your_api_key
Content-Type: application/json
```

**Request Body:**

```json
{
  "plan": "pro"  // Optional: "free", "pro", "enterprise"
}
```

**Response:**

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "plan": "pro",
  "email_limit_per_month": 50000,
  "email_address_limit": 50,
  "created_at": "2025-07-06T10:00:00Z",
  "updated_at": "2025-07-06T10:00:00Z"
}
```

#### Delete Account

```http
DELETE /v1/account
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:** `204 No Content`

#### Generate New API Key

```http
POST /v1/account/api-key
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:**

```json
{
  "api_key": "maylng_new_api_key_here"
}
```

---

### Email Address Management

#### Create Email Address

```http
POST /v1/email-addresses
```

**Headers:**

```http
Authorization: Bearer your_api_key
Content-Type: application/json
```

**Request Body:**

```json
{
  "type": "temporary",           // Required: "temporary" or "persistent"
  "prefix": "my-custom-prefix",  // Optional: custom prefix (auto-generated if omitted)
  "domain": "mayl.ng",          // Optional: domain (uses default if omitted)
  "expires_at": "2025-07-07T10:00:00Z",  // Optional: custom expiration (temporary only)
  "metadata": {                  // Optional: custom metadata
    "purpose": "newsletter",
    "category": "marketing"
  }
}
```

**Response:**

```json
{
  "id": "456e7890-e89b-12d3-a456-426614174111",
  "email": "my-custom-prefix@mayl.ng",
  "type": "temporary",
  "prefix": "my-custom-prefix",
  "domain": "mayl.ng",
  "status": "active",
  "expires_at": "2025-07-07T10:00:00Z",
  "metadata": {
    "purpose": "newsletter",
    "category": "marketing"
  },
  "created_at": "2025-07-06T10:00:00Z",
  "updated_at": "2025-07-06T10:00:00Z"
}
```

#### List Email Addresses

```http
GET /v1/email-addresses
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:**

```json
{
  "email_addresses": [
    {
      "id": "456e7890-e89b-12d3-a456-426614174111",
      "email": "my-custom-prefix@mayl.ng",
      "type": "temporary",
      "prefix": "my-custom-prefix",
      "domain": "mayl.ng",
      "status": "active",
      "expires_at": "2025-07-07T10:00:00Z",
      "metadata": {},
      "created_at": "2025-07-06T10:00:00Z",
      "updated_at": "2025-07-06T10:00:00Z"
    }
  ]
}
```

#### Get Email Address

```http
GET /v1/email-addresses/{id}
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:**

```json
{
  "id": "456e7890-e89b-12d3-a456-426614174111",
  "email": "my-custom-prefix@mayl.ng",
  "type": "temporary",
  "prefix": "my-custom-prefix",
  "domain": "mayl.ng",
  "status": "active",
  "expires_at": "2025-07-07T10:00:00Z",
  "metadata": {},
  "created_at": "2025-07-06T10:00:00Z",
  "updated_at": "2025-07-06T10:00:00Z"
}
```

#### Update Email Address

```http
PATCH /v1/email-addresses/{id}
```

**Headers:**

```http
Authorization: Bearer your_api_key
Content-Type: application/json
```

**Request Body:**

```json
{
  "status": "disabled",          // Optional: "active", "expired", "disabled"
  "expires_at": "2025-08-07T10:00:00Z",  // Optional: update expiration
  "metadata": {                  // Optional: update metadata
    "updated_purpose": "testing"
  }
}
```

#### Delete Email Address

```http
DELETE /v1/email-addresses/{id}
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:** `204 No Content`

---

### Email Operations

#### Send Email

```http
POST /v1/emails/send
```

**Headers:**

```http
Authorization: Bearer your_api_key
Content-Type: application/json
```

**Request Body:**

```json
{
  "from_email_id": "456e7890-e89b-12d3-a456-426614174111",  // Required: ID of your email address
  "to_recipients": [                    // Required: recipient email addresses
    "recipient1@example.com",
    "recipient2@example.com"
  ],
  "cc_recipients": [                    // Optional: CC recipients
    "cc@example.com"
  ],
  "bcc_recipients": [                   // Optional: BCC recipients
    "bcc@example.com"
  ],
  "subject": "Hello from Maylng!",      // Required: email subject
  "text_content": "This is the plain text version of the email.",  // Optional: plain text content
  "html_content": "<h1>Hello!</h1><p>This is the <strong>HTML</strong> version.</p>",  // Optional: HTML content
  "attachments": {                      // Optional: file attachments (metadata)
    "file1.pdf": {
      "size": 1024,
      "type": "application/pdf"
    }
  },
  "headers": {                          // Optional: custom email headers
    "X-Priority": "1",
    "X-Custom-Header": "custom-value"
  },
  "thread_id": "789e0123-e89b-12d3-a456-426614174222",  // Optional: thread ID for grouping
  "scheduled_at": "2025-07-07T15:00:00Z",  // Optional: schedule for future delivery
  "metadata": {                         // Optional: custom metadata for tracking
    "campaign_id": "newsletter_2025_07",
    "source": "api"
  }
}
```

**Response:**

```json
{
  "id": "789e0123-e89b-12d3-a456-426614174333",
  "from_email_id": "456e7890-e89b-12d3-a456-426614174111",
  "to_recipients": ["recipient1@example.com", "recipient2@example.com"],
  "cc_recipients": ["cc@example.com"],
  "bcc_recipients": ["bcc@example.com"],
  "subject": "Hello from Maylng!",
  "text_content": "This is the plain text version of the email.",
  "html_content": "<h1>Hello!</h1><p>This is the <strong>HTML</strong> version.</p>",
  "thread_id": "789e0123-e89b-12d3-a456-426614174222",
  "scheduled_at": "2025-07-07T15:00:00Z",
  "sent_at": null,
  "status": "scheduled",
  "provider_message_id": null,
  "failure_reason": null,
  "created_at": "2025-07-06T10:00:00Z",
  "updated_at": "2025-07-06T10:00:00Z"
}
```

#### List Sent Emails

```http
GET /v1/emails?limit=50&offset=0
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Query Parameters:**

- `limit` (optional): Number of emails to return (default: 50, max: 100)
- `offset` (optional): Number of emails to skip (default: 0)

**Response:**

```json
{
  "emails": [
    {
      "id": "789e0123-e89b-12d3-a456-426614174333",
      "from_email_id": "456e7890-e89b-12d3-a456-426614174111",
      "to_recipients": ["recipient@example.com"],
      "cc_recipients": [],
      "bcc_recipients": [],
      "subject": "Hello from Maylng!",
      "text_content": "This is a test email.",
      "html_content": "<p>This is a <strong>test</strong> email.</p>",
      "thread_id": null,
      "scheduled_at": null,
      "sent_at": "2025-07-06T10:05:00Z",
      "status": "sent",
      "provider_message_id": "sg.abc123",
      "failure_reason": null,
      "created_at": "2025-07-06T10:00:00Z",
      "updated_at": "2025-07-06T10:05:00Z"
    }
  ],
  "pagination": {
    "limit": 50,
    "offset": 0
  }
}
```

#### Get Email Details

```http
GET /v1/emails/{id}
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:**

```json
{
  "id": "789e0123-e89b-12d3-a456-426614174333",
  "from_email_id": "456e7890-e89b-12d3-a456-426614174111",
  "to_recipients": ["recipient@example.com"],
  "cc_recipients": [],
  "bcc_recipients": [],
  "subject": "Hello from Maylng!",
  "text_content": "This is a test email.",
  "html_content": "<p>This is a <strong>test</strong> email.</p>",
  "thread_id": null,
  "scheduled_at": null,
  "sent_at": "2025-07-06T10:05:00Z",
  "status": "sent",
  "provider_message_id": "sg.abc123",
  "failure_reason": null,
  "created_at": "2025-07-06T10:00:00Z",
  "updated_at": "2025-07-06T10:05:00Z"
}
```

#### Get Email Status

```http
GET /v1/emails/{id}/status
```

**Headers:**

```http
Authorization: Bearer your_api_key
```

**Response:**

```json
{
  "id": "789e0123-e89b-12d3-a456-426614174333",
  "status": "sent",
  "sent_at": "2025-07-06T10:05:00Z",
  "provider_message_id": "sg.abc123",
  "failure_reason": null
}
```

---

## 📋 Email Status Values

| Status | Description |
|--------|-------------|
| `queued` | Email is queued for sending |
| `scheduled` | Email is scheduled for future delivery |
| `sent` | Email has been sent to the email provider |
| `delivered` | Email has been delivered to the recipient |
| `failed` | Email sending failed (check `failure_reason`) |

---

## 🚨 Error Handling

### Error Response Format

```json
{
  "error": "Detailed error message"
}
```

### Common HTTP Status Codes

| Code | Status | Description |
|------|--------|-------------|
| `200` | OK | Request succeeded |
| `201` | Created | Resource created successfully |
| `204` | No Content | Request succeeded, no response body |
| `400` | Bad Request | Invalid request parameters |
| `401` | Unauthorized | Invalid or missing API key |
| `403` | Forbidden | Request forbidden (rate limit exceeded) |
| `404` | Not Found | Resource not found |
| `422` | Unprocessable Entity | Invalid request data |
| `429` | Too Many Requests | Rate limit exceeded |
| `500` | Internal Server Error | Server error |

### Common Error Examples

#### Invalid API Key

```json
{
  "error": "Invalid API key"
}
```

#### Rate Limit Exceeded

```json
{
  "error": "Rate limit exceeded"
}
```

#### Email Address Not Found

```json
{
  "error": "from email address not found or not active"
}
```

#### Validation Error

```json
{
  "error": "At least one of text_content or html_content must be provided"
}
```

---

## 💡 Quick Start Examples

### 1. Complete Workflow Example

```bash
#!/bin/bash

# 1. Create account
ACCOUNT_RESPONSE=$(curl -s -X POST https://api.mayl.ng:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"plan": "free"}')

API_KEY=$(echo $ACCOUNT_RESPONSE | jq -r '.api_key')
echo "API Key: $API_KEY"

# 2. Create email address
EMAIL_RESPONSE=$(curl -s -X POST https://api.mayl.ng:8080/v1/email-addresses \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"type": "temporary", "prefix": "demo"}')

EMAIL_ID=$(echo $EMAIL_RESPONSE | jq -r '.id')
EMAIL_ADDRESS=$(echo $EMAIL_RESPONSE | jq -r '.email')
echo "Email Address: $EMAIL_ADDRESS"

# 3. Send email
SEND_RESPONSE=$(curl -s -X POST https://api.mayl.ng:8080/v1/emails/send \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"from_email_id\": \"$EMAIL_ID\",
    \"to_recipients\": [\"test@example.com\"],
    \"subject\": \"Hello from Maylng!\",
    \"html_content\": \"<h1>Welcome!</h1><p>This is a test email from the Maylng API.</p>\"
  }")

EMAIL_SEND_ID=$(echo $SEND_RESPONSE | jq -r '.id')
echo "Email sent with ID: $EMAIL_SEND_ID"

# 4. Check email status
curl -s -X GET "https://api.mayl.ng:8080/v1/emails/$EMAIL_SEND_ID/status" \
  -H "Authorization: Bearer $API_KEY" | jq '.'
```

### 2. JavaScript/Node.js Example

```javascript
const API_BASE = 'https://api.mayl.ng:8080/v1';

class MaylngAPI {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }

  async request(method, endpoint, data = null) {
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.apiKey}`
      }
    };

    if (data) {
      options.body = JSON.stringify(data);
    }

    const response = await fetch(`${API_BASE}${endpoint}`, options);
    return response.json();
  }

  // Create email address
  async createEmailAddress(type = 'temporary', prefix = null) {
    return this.request('POST', '/email-addresses', {
      type,
      prefix
    });
  }

  // Send email
  async sendEmail(fromEmailId, toRecipients, subject, htmlContent, textContent = null) {
    return this.request('POST', '/emails/send', {
      from_email_id: fromEmailId,
      to_recipients: toRecipients,
      subject,
      html_content: htmlContent,
      text_content: textContent
    });
  }

  // Get email status
  async getEmailStatus(emailId) {
    return this.request('GET', `/emails/${emailId}/status`);
  }
}

// Usage
const api = new MaylngAPI('your_api_key_here');

async function example() {
  // Create email address
  const emailAddress = await api.createEmailAddress('temporary', 'my-prefix');
  console.log('Created email:', emailAddress.email);

  // Send email
  const email = await api.sendEmail(
    emailAddress.id,
    ['recipient@example.com'],
    'Hello from JavaScript!',
    '<h1>Hello!</h1><p>This email was sent via the Maylng API.</p>'
  );
  console.log('Email sent:', email.id);

  // Check status
  const status = await api.getEmailStatus(email.id);
  console.log('Email status:', status.status);
}

example().catch(console.error);
```

### 3. Python Example

```python
import requests
import json

class MaylngAPI:
    def __init__(self, api_key):
        self.api_key = api_key
        self.base_url = 'https://api.mayl.ng:8080/v1'
        self.headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {api_key}'
        }

    def create_email_address(self, address_type='temporary', prefix=None):
        data = {'type': address_type}
        if prefix:
            data['prefix'] = prefix
        
        response = requests.post(
            f'{self.base_url}/email-addresses',
            headers=self.headers,
            json=data
        )
        return response.json()

    def send_email(self, from_email_id, to_recipients, subject, html_content, text_content=None):
        data = {
            'from_email_id': from_email_id,
            'to_recipients': to_recipients,
            'subject': subject,
            'html_content': html_content
        }
        if text_content:
            data['text_content'] = text_content
        
        response = requests.post(
            f'{self.base_url}/emails/send',
            headers=self.headers,
            json=data
        )
        return response.json()

    def get_email_status(self, email_id):
        response = requests.get(
            f'{self.base_url}/emails/{email_id}/status',
            headers=self.headers
        )
        return response.json()

# Usage
api = MaylngAPI('your_api_key_here')

# Create email address
email_address = api.create_email_address('temporary', 'python-demo')
print(f"Created email: {email_address['email']}")

# Send email
email = api.send_email(
    email_address['id'],
    ['recipient@example.com'],
    'Hello from Python!',
    '<h1>Hello!</h1><p>This email was sent via the Maylng API from Python.</p>'
)
print(f"Email sent: {email['id']}")

# Check status
status = api.get_email_status(email['id'])
print(f"Email status: {status['status']}")
```

---

## 🔧 Advanced Features

### Scheduled Emails

Send emails at a future time by including the `scheduled_at` field:

```json
{
  "from_email_id": "your-email-id",
  "to_recipients": ["recipient@example.com"],
  "subject": "Scheduled Email",
  "html_content": "<p>This email was scheduled!</p>",
  "scheduled_at": "2025-07-07T15:00:00Z"
}
```

### Email Threading

Group related emails using the `thread_id` field:

```json
{
  "from_email_id": "your-email-id",
  "to_recipients": ["recipient@example.com"],
  "subject": "Re: Previous conversation",
  "html_content": "<p>This is a follow-up email.</p>",
  "thread_id": "original-email-id"
}
```

### Custom Headers

Add custom email headers for tracking or client-specific requirements:

```json
{
  "from_email_id": "your-email-id",
  "to_recipients": ["recipient@example.com"],
  "subject": "Email with custom headers",
  "html_content": "<p>This email has custom headers.</p>",
  "headers": {
    "X-Priority": "1",
    "X-Campaign-ID": "newsletter-2025-07",
    "List-Unsubscribe": "<mailto:unsubscribe@yoursite.com>"
  }
}
```

---

## 📞 Support

- **Documentation**: [https://docs.mayl.ng](https://docs.mayl.ng)
- **Status Page**: [https://status.mayl.ng](https://status.mayl.ng)
- **Support Email**: <support@mayl.ng>
- **GitHub**: [https://github.com/maylng/backend](https://github.com/maylng/backend)

---

## 🔄 Changelog

### v1.0.0 (2025-07-06)

- Initial API release
- Account management
- Email address creation and management
- Email sending with HTML/text support
- Email status tracking
- Rate limiting implementation
- Multi-provider email delivery (SendGrid)

---

## *Last updated: July 6, 2025*
