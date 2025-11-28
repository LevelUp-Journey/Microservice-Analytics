package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("KAFKA Environment Variables Debugger")
	fmt.Println("========================================")
	fmt.Println()

	// Cargar archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	} else {
		fmt.Println("✅ Loaded .env file")
	}
	fmt.Println()

	// Obtener variables
	username := os.Getenv("KAFKA_SASL_USERNAME")
	password := os.Getenv("KAFKA_SASL_PASSWORD")
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	securityProtocol := os.Getenv("KAFKA_SECURITY_PROTOCOL")
	saslMechanism := os.Getenv("KAFKA_SASL_MECHANISM")

	fmt.Println("📊 Raw Environment Variables:")
	fmt.Println("-------------------------------------------")
	fmt.Printf("KAFKA_SASL_USERNAME: '%s'\n", username)
	fmt.Printf("Length: %d characters\n", len(username))
	if len(username) > 0 {
		fmt.Printf("First character: '%c' (ASCII: %d)\n", username[0], username[0])
		fmt.Printf("Last character: '%c' (ASCII: %d)\n", username[len(username)-1], username[len(username)-1])
	}
	fmt.Println()

	// Analizar caracteres
	fmt.Println("🔍 Character Analysis (KAFKA_SASL_USERNAME):")
	fmt.Println("-------------------------------------------")
	for i, char := range username {
		fmt.Printf("  [%d] = '%c' (ASCII: %d, Unicode: U+%04X)\n", i, char, char, char)
	}
	fmt.Println()

	// Verificar si tiene el formato correcto
	fmt.Println("✅ Validation:")
	fmt.Println("-------------------------------------------")

	expectedUsername := "$ConnectionString"
	if username == expectedUsername {
		fmt.Println("✅ KAFKA_SASL_USERNAME is CORRECT: $ConnectionString")
	} else {
		fmt.Println("❌ KAFKA_SASL_USERNAME is INCORRECT!")
		fmt.Printf("   Expected: '%s'\n", expectedUsername)
		fmt.Printf("   Got:      '%s'\n", username)
		fmt.Println()
		fmt.Println("   Comparison:")
		for i := 0; i < len(expectedUsername) && i < len(username); i++ {
			if expectedUsername[i] == username[i] {
				fmt.Printf("     [%d] ✅ '%c' == '%c'\n", i, expectedUsername[i], username[i])
			} else {
				fmt.Printf("     [%d] ❌ '%c' != '%c'\n", i, expectedUsername[i], username[i])
			}
		}
		if len(username) < len(expectedUsername) {
			fmt.Printf("     ⚠️  Missing %d characters\n", len(expectedUsername)-len(username))
		} else if len(username) > len(expectedUsername) {
			fmt.Printf("     ⚠️  %d extra characters\n", len(username)-len(expectedUsername))
		}
	}
	fmt.Println()

	// Password validation
	fmt.Println("📝 Password (Connection String):")
	fmt.Println("-------------------------------------------")
	if len(password) == 0 {
		fmt.Println("❌ KAFKA_SASL_PASSWORD is empty!")
	} else {
		fmt.Printf("Length: %d characters\n", len(password))
		fmt.Printf("Starts with: %s...\n", password[:min(40, len(password))])
		fmt.Printf("Ends with:   ...%s\n", password[max(0, len(password)-30):])

		if password[:10] == "Endpoint=s" {
			fmt.Println("✅ Connection string format looks correct")
		} else {
			fmt.Println("⚠️  Connection string doesn't start with 'Endpoint=sb://'")
		}

		if len(password) > 0 && (password[0] == '\'' || password[0] == '"') {
			fmt.Printf("❌ Password starts with quote character: '%c'\n", password[0])
		}
		if len(password) > 0 && (password[len(password)-1] == '\'' || password[len(password)-1] == '"') {
			fmt.Printf("❌ Password ends with quote character: '%c'\n", password[len(password)-1])
		}
	}
	fmt.Println()

	// Other variables
	fmt.Println("🔐 Other Kafka Configuration:")
	fmt.Println("-------------------------------------------")
	fmt.Printf("KAFKA_BOOTSTRAP_SERVERS: %s\n", bootstrapServers)
	fmt.Printf("KAFKA_SECURITY_PROTOCOL: %s\n", securityProtocol)
	fmt.Printf("KAFKA_SASL_MECHANISM: %s\n", saslMechanism)
	fmt.Println()

	// Final recommendations
	fmt.Println("========================================")
	fmt.Println("💡 Recommendations:")
	fmt.Println("========================================")

	if username != expectedUsername {
		fmt.Println()
		fmt.Println("🔧 Fix your .env file:")
		fmt.Println()
		fmt.Println("   Remove ALL quotes around KAFKA_SASL_USERNAME and KAFKA_SASL_PASSWORD")
		fmt.Println()
		fmt.Println("   ❌ WRONG:")
		fmt.Println("      KAFKA_SASL_USERNAME='$ConnectionString'")
		fmt.Println("      KAFKA_SASL_PASSWORD='Endpoint=sb://...'")
		fmt.Println()
		fmt.Println("   ✅ CORRECT:")
		fmt.Println("      KAFKA_SASL_USERNAME=$ConnectionString")
		fmt.Println("      KAFKA_SASL_PASSWORD=Endpoint=sb://...")
		fmt.Println()
		fmt.Println("   Then restart: go run main.go")
		fmt.Println()
	} else {
		fmt.Println()
		fmt.Println("✅ Configuration looks good!")
		fmt.Println("   If you still get authentication errors, check:")
		fmt.Println("   1. Connection string is valid in Azure Portal")
		fmt.Println("   2. Event Hubs (topics) exist in Azure")
		fmt.Println("   3. Network connectivity to Azure")
		fmt.Println()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
