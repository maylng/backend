#!/bin/bash

# Maylng API Test Script
API_BASE="https://api.mayl.ng"

echo "🚀 Testing Maylng API at $API_BASE"
echo ""

# 1. Health Check
echo "1. Health Check..."
curl -s "$API_BASE/health" | jq '.'
echo ""

# 2. Create Account
echo "2. Creating account..."
ACCOUNT_RESPONSE=$(curl -s -X POST "$API_BASE/v1/accounts" \
  -H "Content-Type: application/json" \
  -d '{"plan": "free"}')

echo $ACCOUNT_RESPONSE | jq '.'
API_KEY=$(echo $ACCOUNT_RESPONSE | jq -r '.api_key')
echo "API Key: $API_KEY"
echo ""

# 3. Create Email Address
echo "3. Creating email address..."
EMAIL_RESPONSE=$(curl -s -X POST "$API_BASE/v1/email-addresses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "temporary",
    "prefix": "test"
  }')

echo $EMAIL_RESPONSE | jq '.'
EMAIL_ID=$(echo $EMAIL_RESPONSE | jq -r '.id')
EMAIL_ADDRESS=$(echo $EMAIL_RESPONSE | jq -r '.email')
echo "Email ID: $EMAIL_ID"
echo "Email Address: $EMAIL_ADDRESS"
echo ""

# 4. List Email Addresses
echo "4. Listing email addresses..."
curl -s -X GET "$API_BASE/v1/email-addresses" \
  -H "Authorization: Bearer $API_KEY" | jq '.'
echo ""

# 5. Send Email (optional - requires valid recipient)
if [ ! -z "$EMAIL_ID" ] && [ "$EMAIL_ID" != "null" ]; then
  echo "5. Sending test email..."
  SEND_RESPONSE=$(curl -s -X POST "$API_BASE/v1/emails/send" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
      \"from_email_id\": \"$EMAIL_ID\",
      \"to_recipients\": [\"test@example.com\"],
      \"subject\": \"Hello from Maylng!\",
      \"text_content\": \"This is a test email from $EMAIL_ADDRESS\",
      \"html_content\": \"<p>This is a test email from <strong>$EMAIL_ADDRESS</strong></p>\"
    }")
  
  echo $SEND_RESPONSE | jq '.'
  SENT_EMAIL_ID=$(echo $SEND_RESPONSE | jq -r '.id')
  echo "Sent Email ID: $SENT_EMAIL_ID"
  echo ""
fi

echo "✅ API test complete!"
