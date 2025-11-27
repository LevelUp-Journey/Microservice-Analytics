#!/bin/bash

# ===================================================
# Script de Verificación de Configuración
# Azure Event Hub - Analytics Microservice
# ===================================================

echo "=========================================="
echo "Azure Event Hub Configuration Checker"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load .env file if exists
if [ -f .env ]; then
    echo -e "${GREEN}✓${NC} Found .env file"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}✗${NC} .env file not found"
    echo -e "${YELLOW}→${NC} Create one from .env.example or .env.azure"
    exit 1
fi

echo ""
echo "Checking required environment variables..."
echo ""

# Function to check environment variable
check_var() {
    local var_name=$1
    local var_value=${!var_name}
    local is_required=${2:-true}

    if [ -z "$var_value" ]; then
        if [ "$is_required" = true ]; then
            echo -e "${RED}✗${NC} $var_name: NOT SET (required)"
            return 1
        else
            echo -e "${YELLOW}⚠${NC} $var_name: NOT SET (optional)"
            return 0
        fi
    else
        # Mask sensitive data
        if [[ $var_name == *"PASSWORD"* ]] || [[ $var_name == *"SECRET"* ]] || [[ $var_name == *"CONNECTION_STRING"* ]]; then
            local masked_value="${var_value:0:10}...${var_value: -10}"
            echo -e "${GREEN}✓${NC} $var_name: $masked_value"
        else
            echo -e "${GREEN}✓${NC} $var_name: $var_value"
        fi
        return 0
    fi
}

# Track errors
errors=0

echo "=== Server Configuration ==="
check_var "SERVER_PORT" || ((errors++))
check_var "SERVER_IP" || ((errors++))

echo ""
echo "=== Database Configuration ==="
check_var "DB_HOST" || ((errors++))
check_var "DB_PORT" || ((errors++))
check_var "DB_USER" || ((errors++))
check_var "DB_PASSWORD" || ((errors++))
check_var "DB_NAME" || ((errors++))
check_var "DB_SSLMODE" || ((errors++))

echo ""
echo "=== Kafka/Azure Event Hub Configuration ==="
check_var "KAFKA_BOOTSTRAP_SERVERS" || ((errors++))
check_var "KAFKA_SECURITY_PROTOCOL" || ((errors++))

# Check if Azure Event Hub configuration is present
if [ "$KAFKA_SECURITY_PROTOCOL" = "SASL_SSL" ]; then
    echo ""
    echo -e "${GREEN}→ Azure Event Hub configuration detected${NC}"
    echo ""
    check_var "KAFKA_SASL_MECHANISM" || ((errors++))
    check_var "KAFKA_SASL_USERNAME" || ((errors++))
    check_var "AZURE_EVENTHUB_CONNECTION_STRING" || ((errors++))

    # Validate SASL username
    if [ "$KAFKA_SASL_USERNAME" != "\$ConnectionString" ]; then
        echo -e "${RED}✗${NC} KAFKA_SASL_USERNAME should be '\$ConnectionString' for Azure Event Hub"
        ((errors++))
    fi

    # Validate connection string format
    if [[ ! "$AZURE_EVENTHUB_CONNECTION_STRING" =~ ^Endpoint=sb:// ]]; then
        echo -e "${RED}✗${NC} AZURE_EVENTHUB_CONNECTION_STRING format appears invalid"
        echo -e "${YELLOW}→${NC} Expected format: Endpoint=sb://..."
        ((errors++))
    fi

    # Check bootstrap server format
    if [[ ! "$KAFKA_BOOTSTRAP_SERVERS" =~ \.servicebus\.windows\.net:9093 ]]; then
        echo -e "${YELLOW}⚠${NC} KAFKA_BOOTSTRAP_SERVERS may not be correct for Azure Event Hub"
        echo -e "${YELLOW}→${NC} Expected format: <namespace>.servicebus.windows.net:9093"
    fi
else
    echo ""
    echo -e "${YELLOW}→ Local Kafka configuration detected${NC}"
    echo ""
fi

echo ""
echo "=== Topics Configuration ==="
check_var "KAFKA_TOPIC" || ((errors++))
check_var "KAFKA_USER_REGISTRATION_TOPIC" || ((errors++))

echo ""
echo "=== Consumer Groups ==="
check_var "KAFKA_GROUP_ID" || ((errors++))
check_var "KAFKA_USER_REGISTRATION_GROUP_ID" || ((errors++))

echo ""
echo "=== Timeout Configuration ==="
check_var "KAFKA_REQUEST_TIMEOUT_MS" false
check_var "KAFKA_SESSION_TIMEOUT_MS" false
check_var "KAFKA_ENABLE_AUTO_COMMIT" false

echo ""
echo "=== Service Discovery ==="
check_var "SERVICE_DISCOVERY_URL" false
check_var "SERVICE_NAME" false
check_var "SERVICE_DISCOVERY_ENABLED" false

echo ""
echo "=========================================="
echo "Verification Summary"
echo "=========================================="

if [ $errors -eq 0 ]; then
    echo -e "${GREEN}✓ All required variables are set correctly!${NC}"
    echo ""
    echo "You can now start the service with:"
    echo "  go run main.go"
    echo ""
    exit 0
else
    echo -e "${RED}✗ Found $errors error(s) in configuration${NC}"
    echo ""
    echo "Please fix the errors above and try again."
    echo ""
    echo "For Azure Event Hub configuration, refer to:"
    echo "  docs/AZURE_EVENT_HUB_CONFIG.md"
    echo ""
    exit 1
fi
