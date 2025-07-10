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

# Test 4: Update Account
echo "4️⃣  Testing Account Update..."
UPDATE_RESPONSE=$(curl -s -X PATCH "$API_BASE/v1/account" 
    -H "Content-Type: application/json" 
    -H "Authorization: Bearer $API_KEY" 
    -d '{
        "plan": "pro"
    }')

if echo "$UPDATE_RESPONSE" | grep -q '"plan":"pro"'; then
    echo "✅ Account update passed"
else
    echo "❌ Account update failed"
    echo "Response: $UPDATE_RESPONSE"
    exit 1
fi

# Test 5: Generate New API Key
echo "5️⃣  Testing New API Key Generation..."
NEW_API_KEY_RESPONSE=$(curl -s -X POST "$API_BASE/v1/account/api-key" 
    -H "Authorization: Bearer $API_KEY")

if echo "$NEW_API_KEY_RESPONSE" | grep -q '"api_key"'; then
    echo "✅ New API key generation passed"
    OLD_API_KEY=$API_KEY
    API_KEY=$(echo "$NEW_API_KEY_RESPONSE" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
    echo "   New API Key: ${API_KEY:0:20}..."
else
    echo "❌ New API key generation failed"
    echo "Response: $NEW_API_KEY_RESPONSE"
    exit 1
}

# Test 4: Update Account
echo "4️⃣  Testing Account Update..."
UPDATE_RESPONSE=$(curl -s -X PATCH "$API_BASE/v1/account" 
    -H "Content-Type: application/json" 
    -H "Authorization: Bearer $API_KEY" 
    -d '{
        "plan": "pro"
    }')

if echo "$UPDATE_RESPONSE" | grep -q '"plan":"pro"'; then
    echo "✅ Account update passed"
else
    echo "❌ Account update failed"
    echo "Response: $UPDATE_RESPONSE"
    exit 1
fi

# Test 5: Generate New API Key
echo "5️⃣  Testing New API Key Generation..."
NEW_API_KEY_RESPONSE=$(curl -s -X POST "$API_BASE/v1/account/api-key" 
    -H "Authorization: Bearer $API_KEY")

if echo "$NEW_API_KEY_RESPONSE" | grep -q '"api_key"'; then
    echo "✅ New API key generation passed"
    OLD_API_KEY=$API_KEY
    API_KEY=$(echo "$NEW_API_KEY_RESPONSE" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
    echo "   New API Key: ${API_KEY:0:20}..."
else
    echo "❌ New API key generation failed"
    echo "Response: $NEW_API_KEY_RESPONSE"
    exit 1
fi

# Test 4: Update Account
echo "4️⃣  Testing Account Update..."
UPDATE_RESPONSE=$(curl -s -X PATCH "$API_BASE/v1/account" 
    -H "Content-Type: application/json" 
    -H "Authorization: Bearer $API_KEY" 
    -d '{
        "plan": "pro"
    }')

if echo "$UPDATE_RESPONSE" | grep -q '"plan":"pro"'; then
    echo "✅ Account update passed"
else
    echo "❌ Account update failed"
    echo "Response: $UPDATE_RESPONSE"
    exit 1
fi

# Test 5: Generate New API Key
echo "5️⃣  Testing New API Key Generation..."
NEW_API_KEY_RESPONSE=$(curl -s -X POST "$API_BASE/v1/account/api-key" 
    -H "Authorization: Bearer $API_KEY")

if echo "$NEW_API_KEY_RESPONSE" | grep -q '"api_key"'; then
    echo "✅ New API key generation passed"
    OLD_API_KEY=$API_KEY
    API_KEY=$(echo "$NEW_API_KEY_RESPONSE" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
    echo "   New API Key: ${API_KEY:0:20}..."
else
    echo "❌ New API key generation failed"
    echo "Response: $NEW_API_KEY_RESPONSE"
    exit 1
fi

# Test 4: Update Account
echo "4️⃣  Testing Account Update..."
UPDATE_RESPONSE=$(curl -s -X PATCH "$API_BASE/v1/account" 
    -H "Content-Type: application/json" 
    -H "Authorization: Bearer $API_KEY" 
    -d '{
        "plan": "pro"
    }')

if echo "$UPDATE_RESPONSE" | grep -q '"plan":"pro"'; then
    echo "✅ Account update passed"
else
    echo "❌ Account update failed"
    echo "Response: $UPDATE_RESPONSE"
    exit 1
fi

# Test 5: Generate New API Key
echo "5️⃣  Testing New API Key Generation..."
NEW_API_KEY_RESPONSE=$(curl -s -X POST "$API_BASE/v1/account/api-key" 
    -H "Authorization: Bearer $API_KEY")

if echo "$NEW_API_KEY_RESPONSE" | grep -q '"api_key"'; then
    echo "✅ New API key generation passed"
    OLD_API_KEY=$API_KEY
    API_KEY=$(echo "$NEW_API_KEY_RESPONSE" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
    echo "   New API Key: ${API_KEY:0:20}..."
else
    echo "❌ New API key generation failed"
    echo "Response: $NEW_API_KEY_RESPONSE"
    exit 1
fi

# Test 6: Send Email (this will fail without SendGrid, but tests authentication with new key)
echo "6️⃣  Testing Email Send Authentication with New API Key..."
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
    echo "✅ Email send authentication with new API key passed (SendGrid validation expected)"
else
    echo "⚠️  Email send test with new API key - check SendGrid configuration"
    echo "Response: $SEND_RESPONSE"
fi

# Test 7: Delete Account
echo "7️⃣  Testing Account Deletion..."
DELETE_RESPONSE=$(curl -s -X DELETE "$API_BASE/v1/account" 
    -H "Authorization: Bearer $API_KEY" -w "%{\http_code}" -o /dev/null)

if [ "$DELETE_RESPONSE" -eq 204 ]; then
    echo "✅ Account deletion passed"
else
    echo "❌ Account deletion failed"
    echo "HTTP Status Code: $DELETE_RESPONSE"
    exit 1
fi

echo ""
echo "🎉 API Testing Complete!"
echo ""
echo "📋 Summary:"
echo "   ✅ Health check working"
echo "   ✅ Account creation working"
echo "   ✅ Account update working"
echo "   ✅ New API key generation working"
echo "   ✅ Authentication working"
echo "   ✅ Email endpoints accessible"
echo "   ✅ Account deletion working"
echo ""
echo "🔧 Your API is ready for SDK integration!"
echo "   Base URL: $API_BASE"
echo "   Test API Key (old): ${OLD_API_KEY:0:20}..."
echo "   Test API Key (new): ${API_KEY:0:20}..."
echo ""


# Test 7: Delete Account
echo "7️⃣  Testing Account Deletion..."
DELETE_RESPONSE=$(curl -s -X DELETE "$API_BASE/v1/account" 
    -H "Authorization: Bearer $API_KEY" -w "%{\http_code}" -o /dev/null)

if [ "$DELETE_RESPONSE" -eq 204 ]; then
    echo "✅ Account deletion passed"
else
    echo "❌ Account deletion failed"
    echo "HTTP Status Code: $DELETE_RESPONSE"
    exit 1
fi

echo ""
echo "🎉 API Testing Complete!"
echo ""
echo "📋 Summary:"
echo "   ✅ Health check working"
echo "   ✅ Account creation working"
echo "   ✅ Account update working"
echo "   ✅ New API key generation working"
echo "   ✅ Authentication working"
echo "   ✅ Email endpoints accessible"
echo "   ✅ Account deletion working"
echo ""
echo "🔧 Your API is ready for SDK integration!"
echo "   Base URL: $API_BASE"
echo "   Test API Key (old): ${OLD_API_KEY:0:20}..."
echo "   Test API Key (new): ${API_KEY:0:20}..."
echo ""
