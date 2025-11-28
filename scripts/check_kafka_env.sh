#!/bin/bash

echo "=================================================="
echo "Kafka Environment Variables Checker"
echo "=================================================="
echo ""

# Cargar el archivo .env si existe
if [ -f .env ]; then
    echo "✅ Found .env file"
    echo ""
else
    echo "❌ No .env file found in current directory"
    exit 1
fi

# Función para mostrar variable con formato
check_var() {
    var_name=$1
    var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        echo "❌ $var_name: NOT SET"
    else
        # Mostrar primeros y últimos caracteres para passwords
        if [[ $var_name == *"PASSWORD"* ]] || [[ $var_name == *"SECRET"* ]]; then
            length=${#var_value}
            if [ $length -gt 50 ]; then
                echo "✅ $var_name: ${var_value:0:30}...${var_value: -20} (length: $length)"
            else
                echo "✅ $var_name: ${var_value:0:20}... (length: $length)"
            fi
        else
            echo "✅ $var_name: $var_value"
        fi
    fi
}

echo "📋 Loading variables from .env..."
export $(grep -v '^#' .env | grep -v '^$' | xargs)
echo ""

echo "=================================================="
echo "Critical Kafka Variables for Azure Event Hub"
echo "=================================================="
echo ""

check_var "KAFKA_BOOTSTRAP_SERVERS"
check_var "KAFKA_SECURITY_PROTOCOL"
check_var "KAFKA_SASL_MECHANISM"
check_var "KAFKA_SASL_USERNAME"
check_var "KAFKA_SASL_PASSWORD"

echo ""
echo "=================================================="
echo "Topic Configuration"
echo "=================================================="
echo ""

check_var "KAFKA_TOPIC"
check_var "KAFKA_GROUP_ID"
check_var "KAFKA_USER_REGISTRATION_TOPIC"
check_var "KAFKA_USER_REGISTRATION_GROUP_ID"

echo ""
echo "=================================================="
echo "Validation Checks"
echo "=================================================="
echo ""

# Validar KAFKA_SASL_USERNAME
if [ "$KAFKA_SASL_USERNAME" = "\$ConnectionString" ]; then
    echo "✅ KAFKA_SASL_USERNAME is correct: \$ConnectionString"
elif [ "$KAFKA_SASL_USERNAME" = "ConnectionString" ]; then
    echo "❌ KAFKA_SASL_USERNAME is missing the '\$' symbol!"
    echo "   Current: $KAFKA_SASL_USERNAME"
    echo "   Should be: \$ConnectionString"
elif [ "${KAFKA_SASL_USERNAME:0:1}" != "\$" ]; then
    echo "❌ KAFKA_SASL_USERNAME does not start with '\$'"
    echo "   Current: '$KAFKA_SASL_USERNAME'"
    echo "   First char: '${KAFKA_SASL_USERNAME:0:1}'"
    echo "   Length: ${#KAFKA_SASL_USERNAME}"
    echo "   Should be: \$ConnectionString"
else
    echo "⚠️  KAFKA_SASL_USERNAME format: $KAFKA_SASL_USERNAME"
fi

echo ""

# Validar KAFKA_SASL_PASSWORD (connection string)
if [[ $KAFKA_SASL_PASSWORD == Endpoint=sb://* ]]; then
    if [[ $KAFKA_SASL_PASSWORD == *SharedAccessKeyName=* ]]; then
        if [[ $KAFKA_SASL_PASSWORD == *SharedAccessKey=* ]]; then
            echo "✅ KAFKA_SASL_PASSWORD format looks correct (Azure Connection String)"
        else
            echo "⚠️  KAFKA_SASL_PASSWORD missing SharedAccessKey"
        fi
    else
        echo "⚠️  KAFKA_SASL_PASSWORD missing SharedAccessKeyName"
    fi
else
    echo "❌ KAFKA_SASL_PASSWORD does not look like an Azure connection string"
    echo "   Should start with: Endpoint=sb://"
fi

echo ""

# Validar KAFKA_SECURITY_PROTOCOL
if [ "$KAFKA_SECURITY_PROTOCOL" = "SASL_SSL" ]; then
    echo "✅ KAFKA_SECURITY_PROTOCOL is correct: SASL_SSL"
else
    echo "❌ KAFKA_SECURITY_PROTOCOL should be SASL_SSL for Azure Event Hub"
    echo "   Current: $KAFKA_SECURITY_PROTOCOL"
fi

echo ""

# Validar KAFKA_SASL_MECHANISM
if [ "$KAFKA_SASL_MECHANISM" = "PLAIN" ]; then
    echo "✅ KAFKA_SASL_MECHANISM is correct: PLAIN"
else
    echo "❌ KAFKA_SASL_MECHANISM should be PLAIN for Azure Event Hub"
    echo "   Current: $KAFKA_SASL_MECHANISM"
fi

echo ""

# Validar Bootstrap Servers
if [[ $KAFKA_BOOTSTRAP_SERVERS == *.servicebus.windows.net:9093 ]]; then
    echo "✅ KAFKA_BOOTSTRAP_SERVERS format looks correct"
else
    echo "⚠️  KAFKA_BOOTSTRAP_SERVERS should end with .servicebus.windows.net:9093"
    echo "   Current: $KAFKA_BOOTSTRAP_SERVERS"
fi

echo ""
echo "=================================================="
echo "Common Issues and Solutions"
echo "=================================================="
echo ""

if [ "$KAFKA_SASL_USERNAME" != "\$ConnectionString" ]; then
    echo "🔧 Issue: Username is not '\$ConnectionString'"
    echo ""
    echo "   Fix in .env (remove quotes):"
    echo "   KAFKA_SASL_USERNAME=\$ConnectionString"
    echo ""
fi

if [[ $KAFKA_SASL_PASSWORD != Endpoint=sb://* ]]; then
    echo "🔧 Issue: Connection string format incorrect"
    echo ""
    echo "   Fix in .env (remove quotes):"
    echo "   KAFKA_SASL_PASSWORD=Endpoint=sb://your-namespace.servicebus.windows.net/;SharedAccessKeyName=...;SharedAccessKey=..."
    echo ""
fi

echo "=================================================="
echo "⚠️  IMPORTANT: Do NOT use quotes in .env file!"
echo "=================================================="
echo ""
echo "❌ WRONG:"
echo "   KAFKA_SASL_USERNAME='\$ConnectionString'"
echo "   KAFKA_SASL_PASSWORD='Endpoint=sb://...'"
echo ""
echo "✅ CORRECT:"
echo "   KAFKA_SASL_USERNAME=\$ConnectionString"
echo "   KAFKA_SASL_PASSWORD=Endpoint=sb://..."
echo ""
