#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="http://localhost:8080"
LOG_FILE="test-results.log"

# Clear previous log
> $LOG_FILE

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║       Simple Banking API - Local Run & Test           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to print section header
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Function to make API call and check response
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -e "${YELLOW}▶ Testing: ${description}${NC}"
    echo "  ${method} ${endpoint}"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X ${method} "${API_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "${data}")
    else
        response=$(curl -s -w "\n%{http_code}" -X ${method} "${API_URL}${endpoint}")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo "$description" >> $LOG_FILE
    echo "Request: $method $endpoint" >> $LOG_FILE
    echo "Data: $data" >> $LOG_FILE
    echo "Response Code: $http_code" >> $LOG_FILE
    echo "Response Body: $body" >> $LOG_FILE
    echo "---" >> $LOG_FILE
    
    if [ "$http_code" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        echo "  Response: $body"
    else
        echo -e "${RED}✗ FAILED${NC} (Expected HTTP $expected_status, got $http_code)"
        echo "  Response: $body"
    fi
    echo ""
    
    # Return the response body for further use
    echo "$body"
}

# Check if server is running
print_header "1. Starting Application"
echo -e "${YELLOW}Building application...${NC}"
make build

echo -e "\n${YELLOW}Starting server...${NC}"
./bin/banking-api > server.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server to start
echo -e "${YELLOW}Waiting for server to start...${NC}"
for i in {1..30}; do
    if curl -s "${API_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Server is running!${NC}\n"
        break
    fi
    sleep 1
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗ Server failed to start within 30 seconds${NC}"
        exit 1
    fi
done

# Health Check
print_header "2. Health Check"
test_endpoint "GET" "/health" "" "200" "Health check endpoint"

# Test 1: Create Account (Success)
print_header "3. Account Endpoints - Success Scenarios"
account1=$(test_endpoint "POST" "/api/v1/accounts" \
    '{"document_number":"12345678900"}' \
    "201" \
    "Create account with valid document number")

account1_id=$(echo $account1 | grep -o '"account_id":[0-9]*' | cut -d':' -f2)
echo "Created Account ID: $account1_id"

# Test 2: Get Account by ID (Success)
test_endpoint "GET" "/api/v1/accounts/${account1_id}" "" "200" \
    "Retrieve account by ID"

# Test 3: Create Another Account
account2=$(test_endpoint "POST" "/api/v1/accounts" \
    '{"document_number":"98765432100"}' \
    "201" \
    "Create second account")

account2_id=$(echo $account2 | grep -o '"account_id":[0-9]*' | cut -d':' -f2)

# Test 4: Get All Accounts
test_endpoint "GET" "/api/v1/accounts" "" "200" \
    "List all accounts"

# Validation Tests for Accounts
print_header "4. Account Endpoints - Validation Scenarios"

# Test 5: Create Account with Duplicate Document Number (Conflict)
test_endpoint "POST" "/api/v1/accounts" \
    '{"document_number":"12345678900"}' \
    "409" \
    "Attempt to create account with duplicate document number (should fail)"

# Test 6: Create Account with Short Document Number (Bad Request)
test_endpoint "POST" "/api/v1/accounts" \
    '{"document_number":"123"}' \
    "400" \
    "Attempt to create account with invalid document number (too short)"

# Test 7: Create Account with Empty Document Number (Bad Request)
test_endpoint "POST" "/api/v1/accounts" \
    '{"document_number":""}' \
    "400" \
    "Attempt to create account with empty document number"

# Test 8: Create Account with Invalid JSON (Bad Request)
test_endpoint "POST" "/api/v1/accounts" \
    '{invalid json}' \
    "400" \
    "Attempt to create account with invalid JSON"

# Test 9: Get Non-existent Account (Not Found)
test_endpoint "GET" "/api/v1/accounts/9999" "" "404" \
    "Attempt to get non-existent account"

# Test 10: Get Account with Invalid ID (Bad Request)
test_endpoint "GET" "/api/v1/accounts/invalid" "" "400" \
    "Attempt to get account with invalid ID format"

# Transaction Tests - Success Scenarios
print_header "5. Transaction Endpoints - Success Scenarios"

# Test 11: Create Purchase Transaction (Amount should be negative)
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":1,\"amount\":50.0}" \
    "201" \
    "Create purchase transaction (amount auto-converted to negative)"

# Test 12: Create Withdrawal Transaction (Amount should be negative)
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":3,\"amount\":23.5}" \
    "201" \
    "Create withdrawal transaction (amount auto-converted to negative)"

# Test 13: Create Credit Voucher Transaction (Amount should be positive)
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":4,\"amount\":60.0}" \
    "201" \
    "Create credit voucher transaction (amount stays positive)"

# Test 14: Create Purchase with Installments
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":2,\"amount\":100.0}" \
    "201" \
    "Create purchase with installments transaction"

# Test 15: Get All Transactions
test_endpoint "GET" "/api/v1/transactions" "" "200" \
    "List all transactions"

# Test 16: Get Account Transactions
test_endpoint "GET" "/api/v1/accounts/${account1_id}/transactions" "" "200" \
    "Get all transactions for account ${account1_id}"

# Transaction Validation Tests
print_header "6. Transaction Endpoints - Validation Scenarios"

# Test 17: Create Transaction for Non-existent Account (Not Found)
test_endpoint "POST" "/api/v1/transactions" \
    '{"account_id":9999,"operation_type_id":1,"amount":50.0}' \
    "404" \
    "Attempt to create transaction for non-existent account"

# Test 18: Create Transaction with Invalid Operation Type (Bad Request)
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":99,\"amount\":50.0}" \
    "400" \
    "Attempt to create transaction with invalid operation type"

# Test 19: Create Transaction with Zero Amount (Bad Request)
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":1,\"amount\":0}" \
    "400" \
    "Attempt to create transaction with zero amount"

# Test 20: Create Transaction with Invalid Account ID (Bad Request)
test_endpoint "POST" "/api/v1/transactions" \
    '{"account_id":0,"operation_type_id":1,"amount":50.0}' \
    "400" \
    "Attempt to create transaction with invalid account ID (0)"

# Test 21: Create Transaction with Invalid JSON (Bad Request)
test_endpoint "POST" "/api/v1/transactions" \
    '{invalid json}' \
    "400" \
    "Attempt to create transaction with invalid JSON"

# Test 22: Get Non-existent Transaction (Not Found)
test_endpoint "GET" "/api/v1/transactions/9999" "" "404" \
    "Attempt to get non-existent transaction"

# Test 23: Get Transactions for Non-existent Account (Not Found)
test_endpoint "GET" "/api/v1/accounts/9999/transactions" "" "404" \
    "Attempt to get transactions for non-existent account"

# Amount Normalization Tests
print_header "7. Amount Normalization Tests"

# Test 24: Purchase with already negative amount (should stay negative)
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":1,\"amount\":-50.0}" \
    "201" \
    "Create purchase with negative amount (should normalize to negative)"

# Test 25: Credit voucher with negative amount (should convert to positive)
test_endpoint "POST" "/api/v1/transactions" \
    "{\"account_id\":${account1_id},\"operation_type_id\":4,\"amount\":-100.0}" \
    "201" \
    "Create credit voucher with negative amount (should normalize to positive)"

# Summary
print_header "8. Test Summary"
echo -e "Test results have been logged to: ${YELLOW}${LOG_FILE}${NC}"
echo -e "Server logs available at: ${YELLOW}server.log${NC}\n"

# Count results
total_tests=$(grep -c "Testing:" <<< "$(cat $LOG_FILE)")
echo -e "${BLUE}Total tests executed: ${total_tests}${NC}\n"

# Stop server
print_header "9. Cleanup"
echo -e "${YELLOW}Stopping server (PID: $SERVER_PID)...${NC}"
kill $SERVER_PID
sleep 2

echo -e "${GREEN}✓ All tests completed!${NC}\n"
echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║            Testing Complete - Check Logs              ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
