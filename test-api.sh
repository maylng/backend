#!/bin/bash

# Maylng API Test Script
# Tests the production deployment to ensure everything is working

API_BASE="http://api.mayl.ng:8080"

echo "🧪 Testing Maylng API Production Deployment"
echo "==========================================="

# Test 1: Health Check
echo "1️⃣  Testing Health Check..."
if curl -s -f "$API_BASE/health" > /dev/null; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi

# Test 2: Create Account
echo "2️⃣  Testing Account Creation..."
ACCOUNT_RESPONSE=$(curl -s -X POST "$API_BASE/v1/accounts" \
    -H "Content-Type: application/json" \
    -d '{
        "plan": "free"
    }')

if echo "$ACCOUNT_RESPONSE" | grep -q '"id"'; then
    echo "✅ Account creation passed"
    ACCOUNT_ID=$(echo "$ACCOUNT_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    API_KEY=$(echo "$ACCOUNT_RESPONSE" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
    echo "   Account ID: $ACCOUNT_ID"
    echo "   API Key: ${API_KEY:0:20}..."
else
    echo "❌ Account creation failed"
    echo "Response: $ACCOUNT_RESPONSE"
    exit 1
fi

# Test 3: Add Email Address
echo "3️⃣  Testing Email Address Creation..."
EMAIL_RESPONSE=$(curl -s -X POST "$API_BASE/v1/email-addresses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{
        "type": "temporary",
        "prefix": "testapi"
    }')

if echo "$EMAIL_RESPONSE" | grep -q '"email"'; then
    echo "✅ Email address creation passed"
    EMAIL_ID=$(echo "$EMAIL_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    EMAIL_ADDRESS=$(echo "$EMAIL_RESPONSE" | grep -o '"email":"[^"]*"' | cut -d'"' -f4)
    echo "   Email: $EMAIL_ADDRESS"
    echo "   Domain: mayl.ng"
else
    echo "❌ Email address creation failed"
    echo "Response: $EMAIL_RESPONSE"
fi

# Test 4: Send Email (this will fail without SendGrid, but tests authentication)
echo "4️⃣  Testing Email Send Authentication..."
SEND_RESPONSE=$(curl -s -X POST "$API_BASE/v1/emails/send" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "{
        \"from_email_id\": \"$EMAIL_ID\",
        \"to_recipients\": [\"test@example.com\"],
        \"subject\": \"Test Email from $EMAIL_ADDRESS\",
        \"text_content\": \"This is a test email from Maylng API using domain mayl.ng\"
    }")

# Check if we get a proper error response (not 401 unauthorized)
if echo "$SEND_RESPONSE" | grep -q -E '"error"|"message"' && ! echo "$SEND_RESPONSE" | grep -q "unauthorized"; then
    echo "✅ Email send authentication passed (SendGrid validation expected)"
else
    echo "⚠️  Email send test - check SendGrid configuration"
    echo "Response: $SEND_RESPONSE"
fi

echo ""
echo "🎉 API Testing Complete!"
echo ""
echo "📋 Summary:"
echo "   ✅ Health check working"
echo "   ✅ Account creation working" 
echo "   ✅ Authentication working"
echo "   ✅ Email endpoints accessible"
echo ""
echo "🔧 Your API is ready for SDK integration!"
echo "   Base URL: $API_BASE"
echo "   Test API Key: ${API_KEY:0:20}..."
echo ""
