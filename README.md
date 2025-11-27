# Analytics Microservice

Microservicio de análisis de ejecuciones de código implementado con **Go**, **Domain-Driven Design (DDD)**, **Kafka**, y **PostgreSQL**.

## 🏗️ Arquitectura DDD

Este proyecto sigue una arquitectura limpia basada en DDD con las siguientes capas:

```
analytics/
├── domain/                    # Capa de dominio (núcleo del negocio)
│   ├── model/
│   │   ├── valueobjects/     # Value Objects inmutables
│   │   ├── entities/         # Entidades del dominio
│   │   └── aggregates/       # Aggregate Roots
│   ├── repositories/         # Interfaces de repositorios
│   └── services/             # Servicios de dominio (interfaces)
│
├── application/              # Capa de aplicación
│   ├── commandservices/     # Implementación de servicios de comando
│   └── queryservices/       # Implementación de servicios de consulta (CQRS)
│
├── infrastructure/           # Capa de infraestructura
│   ├── persistence/
│   │   └── postgres/
│   │       └── repositories/ # Implementación de repositorios con GORM
│   ├── messaging/
│   │   └── kafka/           # Consumidor de eventos Kafka
│   └── config/              # Configuración (DB, Eureka, etc.)
│
└── interfaces/              # Capa de interfaces
    └── rest/
        └── controllers/     # Controladores REST con Gin (SIN DTOs)
```

## 🚀 Características

- ✅ **DDD Completo**: Value Objects, Entities, Aggregates, Repositories
- ✅ **CQRS**: Separación de comandos y consultas
- ✅ **Kafka Consumer**: Consumo de eventos `execution.analytics`
- ✅ **PostgreSQL**: Persistencia con GORM
- ✅ **RESTful API**: Endpoints para analytics y KPIs
- ✅ **Service Discovery**: Integración con Eureka
- ✅ **Sin DTOs**: Transformación inline según guía DDD
- ✅ **Docker**: Configuración completa con docker-compose

## 📊 Modelo de Dominio

### Value Objects
- `ExecutionID`: Identificador único de ejecución (UUID)
- `ChallengeID`: Identificador de desafío (UUID)
- `StudentID`: Identificador de estudiante (UUID)
- `TestID`: Identificador de test (UUID)
- `ProgrammingLanguage`: Lenguaje de programación (enum)
- `ExecutionStatus`: Estado de ejecución (enum)

### Aggregates
- **ExecutionAnalytics** (Aggregate Root)
  - Contiene toda la información de una ejecución
  - Incluye resultados de tests (`TestResult` entities)
  - Lógica de negocio: cálculo de tasa de éxito, validaciones

## 🔌 Kafka Integration

### Topic: `execution.analytics`

```json
{
  "execution_id": "580cf2d5-aee4-4c9a-ba1e-d13ab879bd5c",
  "challenge_id": "aecd4cf5-ccd2-4b17-af75-755730733bf3",
  "student_id": "0354e9c7-724a-4dc5-91e7-16fe79ae6797",
  "language": "cpp",
  "status": "completed",
  "timestamp": "2025-11-19T17:03:48.398926291-05:00",
  "execution_time_ms": 9056,
  "exit_code": 1,
  "total_tests": 4,
  "passed_tests": 0,
  "failed_tests": 4,
  "success": false,
  "test_results": [...]
}
```

## 🌐 API Endpoints

### Analytics Endpoints

- `GET /api/v1/analytics/execution/:executionId` - Obtener análisis por ID de ejecución
- `GET /api/v1/analytics/student/:studentId` - Obtener ejecuciones por estudiante
- `GET /api/v1/analytics/challenge/:challengeId` - Obtener ejecuciones por desafío
- `GET /api/v1/analytics/date-range` - Obtener ejecuciones por rango de fechas

### KPI Endpoints

- `GET /api/v1/analytics/kpi/student/:studentId` - KPIs de estudiante
- `GET /api/v1/analytics/kpi/challenge/:challengeId` - KPIs de desafío
- `GET /api/v1/analytics/kpi/daily` - Estadísticas diarias
- `GET /api/v1/analytics/kpi/languages` - Estadísticas por lenguaje
- `GET /api/v1/analytics/kpi/top-failed-challenges` - Top desafíos fallidos

### Health Check

- `GET /health` - Estado del servicio
- `GET /info` - Información del servicio

## ⚙️ Configuración

### Variables de Entorno (.env)

```bash
# Server
SERVER_PORT=8080
SERVER_IP=127.0.0.1

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=analytics_db
DB_SSLMODE=disable

# Kafka
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_GROUP_ID=analytics-consumer-group
KAFKA_TOPIC=execution.analytics
KAFKA_SECURITY_PROTOCOL=PLAINTEXT

# Eureka Service Discovery
SERVICE_DISCOVERY_URL=http://127.0.0.1:8761/eureka/
SERVICE_NAME=analytics-service
SERVICE_DISCOVERY_ENABLED=true
```

## 🐳 Docker

### Iniciar con Docker Compose

```bash
docker-compose up -d
```

Esto iniciará:
- PostgreSQL (puerto 5432)
- Zookeeper (puerto 2181)
- Kafka (puerto 9092)
- Analytics Service (puerto 8080)

## 🛠️ Desarrollo Local

### Prerrequisitos

- Go 1.23+
- PostgreSQL 15+
- Kafka 3.x+

### Instalación

```bash
# 1. Clonar repositorio
git clone <repo-url>
cd Microservice-Analytics

# 2. Instalar dependencias
go mod download

# 3. Configurar variables de entorno
cp .env.example .env

# 4. Ejecutar migraciones (automáticas con GORM)
# Las tablas se crean automáticamente al iniciar

# 5. Iniciar servicio
go run main.go
```

## 📦 Dependencias Principales

```go
require (
    github.com/IBM/sarama v1.42.1          // Kafka client
    github.com/gin-gonic/gin v1.9.1        // HTTP framework
    github.com/google/uuid v1.6.0          // UUID generation
    github.com/joho/godotenv v1.5.1        // Environment variables
    gorm.io/driver/postgres v1.5.4         // PostgreSQL driver
    gorm.io/gorm v1.25.5                   // ORM
)
```

## 📝 Ejemplos de Uso

### Obtener KPIs de un Estudiante

```bash
curl http://localhost:8080/api/v1/analytics/kpi/student/0354e9c7-724a-4dc5-91e7-16fe79ae6797
```

Respuesta:
```json
{
  "student_id": "0354e9c7-724a-4dc5-91e7-16fe79ae6797",
  "total_executions": 42,
  "success_rate": 78.5
}
```

### Obtener Estadísticas Diarias

```bash
curl "http://localhost:8080/api/v1/analytics/kpi/daily?startDate=2024-01-01T00:00:00Z&endDate=2024-01-07T23:59:59Z"
```

Respuesta:
```json
[
  {
    "date": "2024-01-01",
    "total_executions": 150,
    "successful_executions": 120,
    "failed_executions": 30,
    "success_rate": 80.0,
    "avg_execution_time_ms": 2500.5
  }
]
```

## 🎯 Principios DDD Aplicados

### 1. No DTOs
✅ Las respuestas se transforman inline en los controladores usando `gin.H`
✅ Los aggregates se convierten directamente a JSON sin capa intermedia

### 2. No Mappers
✅ Las transformaciones se realizan con funciones simples inline
✅ Conversión directa de modelos de dominio a respuestas

### 3. Value Objects Inmutables
✅ Todos los value objects validan en su constructor
✅ Encapsulación de reglas de negocio

### 4. Aggregate Root
✅ `ExecutionAnalytics` es el aggregate root
✅ Contiene lógica de negocio (CalculateSuccessRate, validaciones)

### 5. CQRS
✅ Servicios de comando y consulta separados
✅ Optimización de lecturas en query services

## 🔒 Seguridad

- Validación de UUIDs en value objects
- Validación de fechas en queries
- Manejo de errores apropiado
- Límites en paginación

## 📚 Documentación API

La documentación Swagger estará disponible en:
```
http://localhost:8080/swagger/index.html
```

## 🧪 Testing

```bash
# Ejecutar tests unitarios
go test ./...

# Tests con coverage
go test -cover ./...
```

## 📄 Licencia

MIT

## 👥 Contribuidores

- Tu Nombre - Desarrollo inicial

A simple analytics microservice built with Go Fiber and Swagger.

## Running the Application

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Generate Swagger docs:
   ```bash
   go run github.com/swaggo/swag/cmd/swag@latest init
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

4. Access the API:
   - API: http://localhost:3000/
   - Swagger UI: http://localhost:3000/swagger/index.html