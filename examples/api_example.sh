#!/bin/bash

# Maylng API Example Usage Script
# This script demonstrates how to use the Maylng API

API_BASE="http://localhost:8080/v1"
API_KEY=""

echo "=== Maylng API Example Usage ==="
echo ""

# 1. Create an account
echo "1. Creating account..."
ACCOUNT_RESPONSE=$(curl -s -X POST "$API_BASE/accounts" \
  -H "Content-Type: application/json" \
  -d '{"plan": "starter"}') -d '{"plan": "starter"})aylng API Example Usage Script
# This script demonstrates how to use the Maylng API

API_BASE="http://localhost:8080/v1"
API_KEY=""

echo "=== Maylng API Example Usage ==="
echo ""

# 1. Create an account
echo "1. Creating account..."
ACCOUNT_RESPONSE=$(curl -s -X POST "$API_BASE/accounts" \
  -H "Content-Type: application/json" \
  -d '{"plan": "free"}')

echo "Account created: $ACCOUNT_RESPONSE"
API_KEY=$(echo $ACCOUNT_RESPONSE | jq -r '.api_key')
echo "API Key: $API_KEY"
echo ""

if [ "$API_KEY" = "null" ] || [ -z "$API_KEY" ]; then
  echo "Failed to create account or extract API key"
  exit 1
fi

# 2. Get account details
echo "2. Getting account details..."
curl -s -X GET "$API_BASE/account" \
  -H "Authorization: Bearer $API_KEY" | jq '.'
echo ""

# 3. Create email address
echo "3. Creating email address..."
EMAIL_ADDRESS_RESPONSE=$(curl -s -X POST "$API_BASE/email-addresses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "temporary",
    "prefix": "demo"
  }')

echo "Email address created: $EMAIL_ADDRESS_RESPONSE"
EMAIL_ADDRESS_ID=$(echo $EMAIL_ADDRESS_RESPONSE | jq -r '.id')
echo "Email Address ID: $EMAIL_ADDRESS_ID"
echo ""

# 4. List email addresses
echo "4. Listing email addresses..."
curl -s -X GET "$API_BASE/email-addresses" \
  -H "Authorization: Bearer $API_KEY" | jq '.'
echo ""

# 5. Send email
echo "5. Sending email..."
if [ "$EMAIL_ADDRESS_ID" != "null" ] && [ -n "$EMAIL_ADDRESS_ID" ]; then
  SEND_EMAIL_RESPONSE=$(curl -s -X POST "$API_BASE/emails/send" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
      \"from_email_id\": \"$EMAIL_ADDRESS_ID\",
      \"to_recipients\": [\"test@example.com\"],
      \"subject\": \"Hello from Maylng!\",
      \"text_content\": \"This is a test email from Maylng API.\",
      \"html_content\": \"<p>This is a <strong>test email</strong> from Maylng API.</p>\"
    }")
  
  echo "Email sent: $SEND_EMAIL_RESPONSE"
  EMAIL_ID=$(echo $SEND_EMAIL_RESPONSE | jq -r '.id')
  echo "Email ID: $EMAIL_ID"
  echo ""

  # 6. Check email status
  if [ "$EMAIL_ID" != "null" ] && [ -n "$EMAIL_ID" ]; then
    echo "6. Checking email status..."
    curl -s -X GET "$API_BASE/emails/$EMAIL_ID/status" \
      -H "Authorization: Bearer $API_KEY" | jq '.'
    echo ""
  fi
fi

# 7. List sent emails
echo "7. Listing sent emails..."
curl -s -X GET "$API_BASE/emails?limit=10" \
  -H "Authorization: Bearer $API_KEY" | jq '.'
echo ""

echo "=== Example completed ==="
