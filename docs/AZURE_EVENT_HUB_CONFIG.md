# Configuración de Azure Event Hub para el Microservicio de Analytics

Este documento explica cómo configurar el microservicio de Analytics (Go) para conectarse a Azure Event Hub usando el protocolo Kafka.

## Tabla de Contenidos

1. [Introducción](#introducción)
2. [Requisitos Previos](#requisitos-previos)
3. [Configuración de Variables de Entorno](#configuración-de-variables-de-entorno)
4. [Diferencias entre Kafka y Azure Event Hub](#diferencias-entre-kafka-y-azure-event-hub)
5. [Configuración de Seguridad](#configuración-de-seguridad)
6. [Consumer Groups](#consumer-groups)
7. [Topics (Event Hubs)](#topics-event-hubs)
8. [Timeouts y Performance](#timeouts-y-performance)
9. [Troubleshooting](#troubleshooting)
10. [Ejemplo de Configuración Completa](#ejemplo-de-configuración-completa)

---

## Introducción

Azure Event Hub es compatible con el protocolo Apache Kafka, lo que permite usar clientes Kafka existentes para conectarse a Event Hub. Este microservicio utiliza la librería **IBM Sarama** para Go, configurada para trabajar con Azure Event Hub.

### Características Principales

- ✅ Protocolo Kafka 1.0+ compatible
- ✅ Autenticación SASL/PLAIN con TLS
- ✅ Consumer Groups nativos
- ✅ Soporte para múltiples topics (Event Hubs)
- ✅ Auto-commit de offsets
- ✅ Timeouts configurables para alta disponibilidad

---

## Requisitos Previos

### En Azure Portal

1. **Event Hub Namespace** creado
   - Ejemplo: `levelup-journey.servicebus.windows.net`

2. **Event Hubs** (Topics) creados:
   - `execution.analytics` - Para eventos de ejecución de código
   - `iam.user.registered` - Para eventos de registro de usuarios

3. **Connection String** obtenido:
   - Ir a: Event Hub Namespace → Shared access policies → RootManageSharedAccessKey
   - Copiar "Connection string-primary key"

### En el Proyecto

1. Go 1.19 o superior
2. Dependencias instaladas:
   ```bash
   go mod download
   ```

---

## Configuración de Variables de Entorno

### Variables Obligatorias

```bash
# Bootstrap Server
KAFKA_BOOTSTRAP_SERVERS=levelup-journey.servicebus.windows.net:9093

# Protocolo de Seguridad
KAFKA_SECURITY_PROTOCOL=SASL_SSL

# Mecanismo SASL
KAFKA_SASL_MECHANISM=PLAIN

# Usuario SASL (siempre este valor literal)
KAFKA_SASL_USERNAME=$ConnectionString

# Connection String de Azure Event Hub
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://levelup-journey.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=YOUR-AZURE-KEY-HERE
```

### Variables de Topics y Consumer Groups

```bash
# Topics (nombres de Event Hubs)
KAFKA_TOPIC=execution.analytics
KAFKA_USER_REGISTRATION_TOPIC=iam.user.registered

# Consumer Groups
KAFKA_GROUP_ID=analytics-consumer-group
KAFKA_USER_REGISTRATION_GROUP_ID=user-registration-analytics-group
```

### Variables de Configuración Avanzada

```bash
# Timeouts (en milisegundos)
KAFKA_REQUEST_TIMEOUT_MS=60000
KAFKA_SESSION_TIMEOUT_MS=60000

# Auto-commit de offsets
KAFKA_ENABLE_AUTO_COMMIT=true
```

---

## Diferencias entre Kafka y Azure Event Hub

| Aspecto | Apache Kafka | Azure Event Hub |
|---------|--------------|-----------------|
| **Puerto** | 9092 | 9093 (TLS) |
| **Protocolo** | PLAINTEXT o SASL_SSL | SASL_SSL (obligatorio) |
| **SASL Mechanism** | PLAIN, SCRAM, GSSAPI | PLAIN únicamente |
| **Username** | Variable | Siempre `$ConnectionString` |
| **Password** | Token o password | Connection String completo |
| **Topics** | Topics | Event Hubs |
| **Consumer Groups** | Consumer Groups | Consumer Groups (mismo concepto) |
| **Versión Kafka** | Configurable | Compatible con 1.0+ |

---

## Configuración de Seguridad

### SASL_SSL

Azure Event Hub **requiere** SASL_SSL para todas las conexiones:

```go
// Configuración en el código (ya implementado)
config.Net.TLS.Enable = true
config.Net.TLS.Config = &tls.Config{
    InsecureSkipVerify: false,
    MinVersion:         tls.VersionTLS12,
}

config.Net.SASL.Enable = true
config.Net.SASL.Mechanism = sarama.SASLTypePlaintext  // PLAIN
config.Net.SASL.User = "$ConnectionString"
config.Net.SASL.Password = "Endpoint=sb://..."
config.Net.SASL.Handshake = true
```

### Connection String Format

El connection string debe tener este formato:

```
Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=<policy-name>;SharedAccessKey=<key>
```

**Ejemplo:**
```
Endpoint=sb://levelup-journey.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=YOUR-SHARED-ACCESS-KEY-HERE
```

### Políticas de Acceso

| Política | Permisos | Uso Recomendado |
|----------|----------|-----------------|
| **RootManageSharedAccessKey** | Send, Listen, Manage | Desarrollo/Testing |
| **SendPolicy** | Send | Producers en producción |
| **ListenPolicy** | Listen | Consumers en producción |

> ⚠️ **Producción**: Crear políticas específicas con permisos mínimos necesarios.

---

## Consumer Groups

### ¿Qué son los Consumer Groups?

Los Consumer Groups permiten que múltiples instancias de consumidores lean mensajes de forma distribuida sin duplicar el procesamiento.

### Configuración en el Microservicio

Este microservicio usa dos consumer groups:

1. **analytics-consumer-group**
   - Topic: `execution.analytics`
   - Procesa eventos de ejecución de código

2. **user-registration-analytics-group**
   - Topic: `iam.user.registered`
   - Procesa eventos de registro de usuarios

### Crear Consumer Groups en Azure

1. Ir a: Event Hub → Consumer groups
2. Click en "+ Consumer group"
3. Nombre: `analytics-consumer-group`
4. Repetir para `user-registration-analytics-group`

> 💡 **Nota**: El consumer group `$Default` existe por defecto.

---

## Topics (Event Hubs)

### Crear Event Hubs (Topics)

1. **Ir a Event Hub Namespace**
2. **Crear Event Hub**: `execution.analytics`
   - Partition Count: 2-4 (recomendado)
   - Message Retention: 1-7 días

3. **Crear Event Hub**: `iam.user.registered`
   - Partition Count: 2-4 (recomendado)
   - Message Retention: 1-7 días

### Consideraciones de Particiones

- Más particiones = Mayor paralelismo
- Cada partición puede ser consumida por un consumidor del grupo
- Recomendado: 2-4 particiones para baja/media carga

---

## Timeouts y Performance

### Timeouts Configurados

```bash
# Read/Write timeout (60 segundos)
KAFKA_REQUEST_TIMEOUT_MS=60000

# Session timeout para consumer group (60 segundos)
KAFKA_SESSION_TIMEOUT_MS=60000
```

### Configuraciones de Conexión

El microservicio está configurado con:

- **Dial Timeout**: 30 segundos
- **Read Timeout**: 60 segundos (configurable)
- **Write Timeout**: 30 segundos
- **Metadata Timeout**: 60 segundos
- **Heartbeat Interval**: 3 segundos

### Optimización

Para alta carga:

```bash
# Reducir timeouts (solo si la red es estable)
KAFKA_REQUEST_TIMEOUT_MS=30000
KAFKA_SESSION_TIMEOUT_MS=30000

# Deshabilitar auto-commit para control manual
KAFKA_ENABLE_AUTO_COMMIT=false
```

---

## Troubleshooting

### Error: "Failed to connect to broker"

**Causa**: Configuración incorrecta del bootstrap server o puerto.

**Solución**:
```bash
# Verificar formato correcto
KAFKA_BOOTSTRAP_SERVERS=your-namespace.servicebus.windows.net:9093
```

### Error: "SASL authentication failed"

**Causa**: Connection string incorrecto o username mal configurado.

**Solución**:
```bash
# Username debe ser literal "$ConnectionString"
KAFKA_SASL_USERNAME=$ConnectionString

# Verificar connection string completo
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://...
```

### Error: "context deadline exceeded"

**Causa**: Timeouts muy bajos o problemas de red.

**Solución**:
```bash
# Aumentar timeouts
KAFKA_REQUEST_TIMEOUT_MS=90000
KAFKA_SESSION_TIMEOUT_MS=90000
```

### Error: "Topic not found"

**Causa**: El Event Hub (topic) no existe en Azure.

**Solución**:
1. Ir a Azure Portal
2. Crear Event Hub con el nombre exacto del topic
3. Verificar capitalización (case-sensitive)

### Error: "Consumer group rebalancing too frequently"

**Causa**: Session timeout muy bajo o procesamiento lento.

**Solución**:
```bash
# Aumentar session timeout
KAFKA_SESSION_TIMEOUT_MS=90000
```

### Logs para Debug

El microservicio genera logs detallados:

```
Kafka Configuration:
  Bootstrap Servers: [levelup-journey.servicebus.windows.net:9093]
  Security Protocol: SASL_SSL
  SASL Mechanism: PLAIN
  Group ID: analytics-consumer-group
  Topic: execution.analytics
  Azure Event Hub: Configured ✓
```

---

## Ejemplo de Configuración Completa

### Archivo `.env`

```bash
# Server
SERVER_PORT=8080
SERVER_IP=0.0.0.0

# Database
DB_HOST=iam.postgres.database.azure.com
DB_PORT=5432
DB_USER=levelup
DB_PASSWORD=Journey12
DB_NAME=analytics_db
DB_SSLMODE=require

# Azure Event Hub Configuration
KAFKA_BOOTSTRAP_SERVERS=levelup-journey.servicebus.windows.net:9093
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISM=PLAIN
KAFKA_SASL_USERNAME=$ConnectionString
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://levelup-journey.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=YOUR-SHARED-ACCESS-KEY-HERE

# Topics and Consumer Groups
KAFKA_TOPIC=execution.analytics
KAFKA_USER_REGISTRATION_TOPIC=iam.user.registered
KAFKA_GROUP_ID=analytics-consumer-group
KAFKA_USER_REGISTRATION_GROUP_ID=user-registration-analytics-group

# Timeouts
KAFKA_REQUEST_TIMEOUT_MS=60000
KAFKA_SESSION_TIMEOUT_MS=60000
KAFKA_ENABLE_AUTO_COMMIT=true

# Service Discovery
SERVICE_DISCOVERY_URL=https://discovery.yellowsea-767275f1.westus3.azurecontainerapps.io/eureka/
SERVICE_NAME=analytics-service
SERVICE_DISCOVERY_ENABLED=true
```

### Iniciar el Microservicio

```bash
# Instalar dependencias
go mod download

# Ejecutar
go run main.go
```

### Verificar Conexión

Buscar en los logs:

```
✓ Kafka Configuration:
✓ Azure Event Hub: Configured ✓
✓ Creating consumer group with brokers: [levelup-journey.servicebus.windows.net:9093]
✓ Consumer group created successfully for topic: execution.analytics
✓ Starting Kafka consumer for topic: execution.analytics
✓ Consumer group session setup - MemberID: ... GenerationID: ...
```

---

## Comparación con la Configuración de Java (IAM Service)

### Java (Spring Kafka)

```yaml
spring:
  kafka:
    bootstrap-servers: levelup-journey.servicebus.windows.net:9093
    properties:
      security.protocol: SASL_SSL
      sasl.mechanism: PLAIN
      sasl.jaas.config: org.apache.kafka.common.security.plain.PlainLoginModule required username="$ConnectionString" password="${AZURE_EVENTHUB_CONNECTION_STRING}";
```

### Go (Sarama)

```go
config.Net.SASL.Enable = true
config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
config.Net.SASL.User = "$ConnectionString"
config.Net.SASL.Password = os.Getenv("AZURE_EVENTHUB_CONNECTION_STRING")
```

**Ambas configuraciones son equivalentes** y se conectan al mismo Azure Event Hub usando las mismas credenciales.

---

## Referencias

- [Azure Event Hub - Kafka Protocol](https://docs.microsoft.com/azure/event-hubs/event-hubs-for-kafka-ecosystem-overview)
- [IBM Sarama Documentation](https://pkg.go.dev/github.com/IBM/sarama)
- [Kafka Protocol Guide](https://kafka.apache.org/protocol)

---

## Soporte

Para problemas o preguntas:

1. Revisar la sección [Troubleshooting](#troubleshooting)
2. Verificar logs del microservicio
3. Consultar Azure Portal para estado de Event Hub
4. Verificar conectividad de red (puerto 9093)

---

**Última actualización**: 2024
**Versión del microservicio**: 1.0.0