#!/bin/bash

# Simple Banking API - Quick Integration Test
API_URL="http://localhost:8080"

echo "🗑️  Cleaning database..."
rm -rf ../data/ 2>/dev/null
echo ""

echo "🧪 Testing Banking API..."
echo ""

# Test 1: Health Check
echo "1️⃣ Health Check"
curl -s $API_URL/health | jq
echo ""

# Test 2: Create Account
echo "2️⃣ Create Account"
ACCOUNT=$(curl -s -X POST $API_URL/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number":"12345678900"}')
echo $ACCOUNT | jq
ACCOUNT_ID=$(echo $ACCOUNT | jq -r '.account_id')
echo "✅ Account ID: $ACCOUNT_ID"
echo ""

# Test 3: Get Account
echo "3️⃣ Get Account"
curl -s $API_URL/v1/accounts/$ACCOUNT_ID | jq
echo ""

# Test 4: Create Purchase Transaction (should be negative)
echo "4️⃣ Create Purchase Transaction"
curl -s -X POST $API_URL/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: purchase-test-1" \
  -d "{\"account_id\":$ACCOUNT_ID,\"operation_type_id\":1,\"amount\":50.0}" | jq
echo ""

# Test 5: Create Credit Voucher (should be positive)
echo "5️⃣ Create Credit Voucher"
curl -s -X POST $API_URL/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: credit-test-1" \
  -d "{\"account_id\":$ACCOUNT_ID,\"operation_type_id\":4,\"amount\":100.0}" | jq
echo ""

# Test 6: Get Account Transactions
echo "6️⃣ Get Account Transactions"
curl -s "$API_URL/v1/accounts/$ACCOUNT_ID/transactions" | jq
echo ""

# Test 7: Validation Tests
echo "7️⃣ Validation Tests"
echo "   - Missing Idempotency-Key (should be 400):"
curl -s -w "\nHTTP Status: %{http_code}\n" -X POST $API_URL/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"account_id\":$ACCOUNT_ID,\"operation_type_id\":1,\"amount\":50.0}" | jq
echo ""

echo "   - Duplicate account (should be 409):"
curl -s -w "\nHTTP Status: %{http_code}\n" -X POST $API_URL/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number":"12345678900"}' | jq
echo ""

echo "   - Invalid document (should be 400):"
curl -s -w "\nHTTP Status: %{http_code}\n" -X POST $API_URL/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number":"123"}' | jq
echo ""

echo "   - Account not found (should be 404):"
curl -s -w "\nHTTP Status: %{http_code}\n" $API_URL/v1/accounts/9999 | jq
echo ""

echo "✅ All tests completed!"
