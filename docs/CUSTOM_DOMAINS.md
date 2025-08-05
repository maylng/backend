# Custom Domain API Documentation

## Overview

The Custom Domain API allows users to add, verify, and manage their own domains for sending emails. This enables users to send emails from their own domain (e.g., `user@example.com`) instead of the default `mayl.ng` domain.

## Authentication

All endpoints require API key authentication via the `Authorization: Bearer <API_KEY>` header.

## Endpoints

### 1. Add Custom Domain

**POST** `/v1/custom-domains`

Add a new custom domain to your account and initiate SES verification.

#### Request Body

```json
{
  "domain": "example.com"
}
```

#### Response (201 Created)

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "domain": "example.com",
  "status": "pending",
  "dns_records": [
    {
      "type": "CNAME",
      "name": "abc123._domainkey.example.com",
      "value": "abc123.dkim.amazonses.com",
      "ttl": 1800
    },
    {
      "type": "CNAME", 
      "name": "def456._domainkey.example.com",
      "value": "def456.dkim.amazonses.com",
      "ttl": 1800
    }
  ],
  "ses_verification_status": "Pending",
  "ses_dkim_verification_status": "Pending",
  "created_at": "2025-08-05T20:00:00Z",
  "updated_at": "2025-08-05T20:00:00Z"
}
```

#### Error Responses

- `400 Bad Request` - Invalid domain format
- `409 Conflict` - Domain already exists

### 2. List Custom Domains

**GET** `/v1/custom-domains`

List all custom domains for your account.

#### Response (200 OK)

```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "domain": "example.com",
    "status": "verified",
    "verified_at": "2025-08-05T21:00:00Z",
    "created_at": "2025-08-05T20:00:00Z",
    "updated_at": "2025-08-05T21:00:00Z"
  }
]
```

### 3. Get Custom Domain Details

**GET** `/v1/custom-domains/{id}`

Get detailed information about a specific custom domain.

#### Response (200 OK)'

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "domain": "example.com",
  "status": "verified",
  "dns_records": [
    {
      "type": "CNAME",
      "name": "abc123._domainkey.example.com",
      "value": "abc123.dkim.amazonses.com",
      "ttl": 1800
    }
  ],
  "ses_verification_status": "Success",
  "ses_dkim_verification_status": "Success",
  "verified_at": "2025-08-05T21:00:00Z",
  "created_at": "2025-08-05T20:00:00Z",
  "updated_at": "2025-08-05T21:00:00Z"
}
```

#### Error Responses-

- `404 Not Found` - Domain not found
- `403 Forbidden` - Access denied

### 4. Delete Custom Domain

**DELETE** `/v1/custom-domains/{id}`

Delete a custom domain from your account and remove it from SES.

#### Response (204 No Content)

#### Error Responses'

- `404 Not Found` - Domain not found
- `403 Forbidden` - Access denied

### 5. Trigger/Retry Verification

**POST** `/v1/custom-domains/{id}/verify`

Trigger or retry the verification process for a custom domain.

#### Response (200 OK)-

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "domain": "example.com",
  "status": "pending",
  "dns_records": [
    {
      "type": "CNAME",
      "name": "abc123._domainkey.example.com",
      "value": "abc123.dkim.amazonses.com",
      "ttl": 1800
    }
  ],
  "verification_attempted_at": "2025-08-05T22:00:00Z"
}
```

### 6. Check Verification Status

**GET** `/v1/custom-domains/{id}/status`

Check the current verification status of a custom domain.

#### Response (200 OK)=

```json
{
  "status": "verified",
  "ses_verification_status": "Success",
  "ses_dkim_verification_status": "Success",
  "verified_at": "2025-08-05T21:00:00Z"
}
```

## Domain Status Values

- `pending` - Domain verification is in progress
- `verified` - Domain is verified and can be used for sending emails
- `failed` - Domain verification failed
- `disabled` - Domain has been disabled

## Setup Process

1. **Add Domain**: Call `POST /v1/custom-domains` with your domain
2. **Get DNS Records**: Use the returned `dns_records` to configure your DNS
3. **Add DNS Records**: Add the CNAME records to your domain's DNS settings
4. **Wait for Verification**: Verification can take up to 72 hours
5. **Check Status**: Use `GET /v1/custom-domains/{id}/status` to check progress
6. **Start Sending**: Once verified, create email addresses using the custom domain

## DNS Configuration

After adding a domain, you'll receive DKIM CNAME records that must be added to your DNS:

```txt
Type: CNAME
Name: abc123._domainkey.example.com
Value: abc123.dkim.amazonses.com
TTL: 1800 (30 minutes)
```

Add all provided CNAME records to your DNS provider. Verification typically completes within a few hours but can take up to 72 hours.

## Integration with Email Addresses

Once a domain is verified, you can create email addresses that use the custom domain by setting the `custom_domain_id` field when creating email addresses via the `/v1/email-addresses` endpoint.

## Error Handling

- Always check the `failure_reason` field if a domain status is `failed`
- Retry verification using the `/verify` endpoint if initial verification fails
- Ensure DNS records are correctly configured before retrying
