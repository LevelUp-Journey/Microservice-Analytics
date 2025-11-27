# Arquitectura del Microservicio de Analytics

## Visión General

El microservicio de Analytics está construido siguiendo los principios de **Domain-Driven Design (DDD)** y **CQRS**, implementado en Go con integración a **Azure Event Hub** (protocolo Kafka) y **PostgreSQL**.

---

## Diagrama de Arquitectura

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         AZURE EVENT HUB (Kafka Protocol)                    │
│  ┌─────────────────────────┐         ┌──────────────────────────────────┐  │
│  │  execution.analytics    │         │  iam.user.registered             │  │
│  │  (Event Hub)            │         │  (Event Hub)                     │  │
│  │                         │         │                                  │  │
│  │  - Partitions: 2-4      │         │  - Partitions: 2-4               │  │
│  │  - Retention: 1-7 days  │         │  - Retention: 1-7 days           │  │
│  └────────────┬────────────┘         └────────────┬─────────────────────┘  │
│               │                                   │                         │
└───────────────┼───────────────────────────────────┼─────────────────────────┘
                │                                   │
                │ SASL_SSL (TLS 1.2+)              │ SASL_SSL (TLS 1.2+)
                │ Username: $ConnectionString       │ Username: $ConnectionString
                │                                   │
                ▼                                   ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                       ANALYTICS MICROSERVICE (Go)                             │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                    INFRASTRUCTURE LAYER                                 │ │
│  │  ┌──────────────────────────────────────────────────────────────────┐  │ │
│  │  │         Kafka Consumers (IBM Sarama)                             │  │ │
│  │  │                                                                  │  │ │
│  │  │  ┌─────────────────────┐    ┌───────────────────────────────┐  │  │ │
│  │  │  │ ExecutionConsumer   │    │ UserRegistrationConsumer      │  │  │ │
│  │  │  │                     │    │                               │  │  │ │
│  │  │  │ - Group ID:         │    │ - Group ID:                   │  │  │ │
│  │  │  │   analytics-        │    │   user-registration-          │  │  │ │
│  │  │  │   consumer-group    │    │   analytics-group             │  │  │ │
│  │  │  │                     │    │                               │  │  │ │
│  │  │  │ - Auto-commit: true │    │ - Auto-commit: true           │  │  │ │
│  │  │  │ - Offset: newest    │    │ - Offset: newest              │  │  │ │
│  │  │  └──────────┬──────────┘    └────────────┬──────────────────┘  │  │ │
│  │  └─────────────┼──────────────────────────────┼─────────────────────┘  │ │
│  │                │                              │                         │ │
│  │                │ JSON Events                  │ JSON Events             │ │
│  │                ▼                              ▼                         │ │
│  │  ┌──────────────────────────────────────────────────────────────────┐  │ │
│  │  │              Event to Domain Conversion                          │  │ │
│  │  │  (Deserialización y validación de Value Objects)                │  │ │
│  │  └──────────────────┬───────────────────────────┬───────────────────┘  │ │
│  └─────────────────────┼───────────────────────────┼──────────────────────┘ │
│                        │                           │                         │
│                        ▼                           ▼                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                     APPLICATION LAYER (CQRS)                            │ │
│  │                                                                         │ │
│  │  ┌────────────────────────────┐      ┌────────────────────────────┐   │ │
│  │  │   Command Services         │      │   Query Services           │   │ │
│  │  │                            │      │                            │   │ │
│  │  │  - ExecutionAnalytics      │      │  - ExecutionAnalytics      │   │ │
│  │  │    CommandService          │      │    QueryService            │   │ │
│  │  │                            │      │                            │   │ │
│  │  │  - UserRegistration        │      │  - UserRegistration        │   │ │
│  │  │    CommandService          │      │    QueryService            │   │ │
│  │  │                            │      │                            │   │ │
│  │  │  - SyncService             │      │  - KPI calculations        │   │ │
│  │  └────────────┬───────────────┘      └────────────┬───────────────┘   │ │
│  └───────────────┼──────────────────────────────────┼─────────────────────┘ │
│                  │                                  │                        │
│                  ▼                                  ▲                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                        DOMAIN LAYER                                     │ │
│  │                                                                         │ │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────────┐   │ │
│  │  │  Value Objects   │  │    Entities      │  │  Aggregate Roots   │   │ │
│  │  │                  │  │                  │  │                    │   │ │
│  │  │  - ExecutionID   │  │  - TestResult    │  │  - Execution       │   │ │
│  │  │  - ChallengeID   │  │                  │  │    Analytics       │   │ │
│  │  │  - StudentID     │  │                  │  │                    │   │ │
│  │  │  - TestID        │  │                  │  │  - UserRegistration│   │ │
│  │  │  - Language      │  │                  │  │    Analytics       │   │ │
│  │  │  - Status        │  │                  │  │                    │   │ │
│  │  │  - UserID        │  │                  │  │  Business Logic:   │   │ │
│  │  │  - ProfileID     │  │                  │  │  - Validations     │   │ │
│  │  └──────────────────┘  └──────────────────┘  │  - Calculations    │   │ │
│  │                                               └────────────────────┘   │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                  │                                           │
│                                  ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │              INFRASTRUCTURE LAYER - PERSISTENCE                         │ │
│  │                                                                         │ │
│  │  ┌────────────────────────────────────────────────────────────────┐    │ │
│  │  │              PostgreSQL Repositories (GORM)                    │    │ │
│  │  │                                                                │    │ │
│  │  │  - ExecutionAnalyticsRepository                               │    │ │
│  │  │  - UserRegistrationAnalyticsRepository                        │    │ │
│  │  └────────────────────────────┬───────────────────────────────────┘    │ │
│  └───────────────────────────────┼────────────────────────────────────────┘ │
│                                  │                                           │
│                                  ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                     INTERFACES LAYER (REST API)                         │ │
│  │                                                                         │ │
│  │  ┌──────────────────────────────────────────────────────────────────┐  │ │
│  │  │              Gin HTTP Controllers                                │  │ │
│  │  │                                                                  │  │ │
│  │  │  - AnalyticsController      (Query endpoints)                   │  │ │
│  │  │  - SyncController            (Manual sync)                      │  │ │
│  │  │  - UserRegistrationController (Query endpoints)                 │  │ │
│  │  │                                                                  │  │ │
│  │  │  Inline transformations (No DTOs)                               │  │ │
│  │  └──────────────────────────────────────────────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                  │                                           │
└──────────────────────────────────┼───────────────────────────────────────────┘
                                   │
                                   │ HTTP/REST
                                   ▼
                    ┌──────────────────────────────┐
                    │   API Clients / Frontend     │
                    │                              │
                    │  - Dashboard                 │
                    │  - Monitoring Tools          │
                    │  - Other Microservices       │
                    └──────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                        AZURE POSTGRESQL DATABASE                            │
│                                                                             │
│  Tables:                                                                    │
│  ┌─────────────────────────────┐    ┌────────────────────────────────┐    │
│  │  execution_analytics        │    │  user_registration_analytics   │    │
│  │                             │    │                                │    │
│  │  - id (PK)                  │    │  - id (PK)                     │    │
│  │  - execution_id (UUID)      │    │  - user_id (UUID)              │    │
│  │  - challenge_id (UUID)      │    │  - profile_id                  │    │
│  │  - student_id (UUID)        │    │  - username                    │    │
│  │  - language                 │    │  - profile_url                 │    │
│  │  - status                   │    │  - registered_at               │    │
│  │  - timestamp                │    │  - created_at                  │    │
│  │  - execution_time_ms        │    │  - updated_at                  │    │
│  │  - success                  │    └────────────────────────────────┘    │
│  │  - test_results (JSONB)     │                                          │
│  │  - created_at               │                                          │
│  │  - updated_at               │                                          │
│  └─────────────────────────────┘                                          │
│                                                                             │
│  Indexes:                                                                   │
│  - idx_execution_analytics_student_id                                      │
│  - idx_execution_analytics_challenge_id                                    │
│  - idx_execution_analytics_timestamp                                       │
│  - idx_user_registration_analytics_user_id                                 │
│  - idx_user_registration_analytics_registered_at                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         EUREKA SERVICE DISCOVERY                            │
│                                                                             │
│  Service Registration:                                                      │
│  - Name: analytics-service                                                  │
│  - Instance: analytics-service:8080:random-value                           │
│  - Health Check: /health                                                    │
│  - Heartbeat: Every 10 seconds                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Flujo de Datos

### 1. Consumo de Eventos de Ejecución

```
IAM/Code Runner Service
        │
        │ Produces event
        ▼
Azure Event Hub (execution.analytics)
        │
        │ SASL_SSL connection
        │ Consumer Group: analytics-consumer-group
        ▼
ExecutionConsumer (Sarama)
        │
        │ Deserialize JSON
        ▼
Event to Domain Conversion
        │
        │ Create Value Objects
        │ Create Aggregate
        ▼
ExecutionAnalyticsCommandService
        │
        │ Validate & Store
        ▼
ExecutionAnalyticsRepository (GORM)
        │
        ▼
PostgreSQL Database
```

### 2. Consumo de Eventos de Registro de Usuarios

```
IAM Service
        │
        │ Produces event
        ▼
Azure Event Hub (iam.user.registered)
        │
        │ SASL_SSL connection
        │ Consumer Group: user-registration-analytics-group
        ▼
UserRegistrationConsumer (Sarama)
        │
        │ Deserialize JSON
        ▼
Event to Domain Conversion
        │
        │ Create Value Objects
        │ Create Aggregate
        ▼
UserRegistrationCommandService
        │
        │ Validate & Store
        ▼
UserRegistrationAnalyticsRepository (GORM)
        │
        ▼
PostgreSQL Database
```

### 3. Consulta de Analytics (REST API)

```
Client (Frontend/API)
        │
        │ HTTP GET Request
        ▼
Gin Router
        │
        ▼
AnalyticsController
        │
        │ Call query service
        ▼
ExecutionAnalyticsQueryService
        │
        │ Query database
        ▼
ExecutionAnalyticsRepository (GORM)
        │
        │ Fetch data
        ▼
PostgreSQL Database
        │
        │ Return aggregates
        ▼
Controller
        │
        │ Transform inline (No DTOs)
        │ Convert to gin.H
        ▼
JSON Response
```

---

## Componentes Principales

### 1. **Azure Event Hub Integration**

**Configuración:**
- **Protocol**: Kafka (SASL_SSL)
- **Port**: 9093 (TLS)
- **Authentication**: SASL/PLAIN with Connection String
- **Client Library**: IBM Sarama v1.42.1

**Event Hubs (Topics):**
- `execution.analytics` - Eventos de ejecución de código
- `iam.user.registered` - Eventos de registro de usuarios

**Consumer Groups:**
- `analytics-consumer-group` - Para ejecuciones
- `user-registration-analytics-group` - Para usuarios

### 2. **Domain Layer (DDD)**

**Value Objects:**
- Inmutables
- Validación en constructor
- Encapsulación de reglas de negocio

**Entities:**
- `TestResult` - Resultado de un test individual

**Aggregates:**
- `ExecutionAnalytics` - Aggregate Root para ejecuciones
- `UserRegistrationAnalytics` - Aggregate Root para registros

### 3. **Application Layer (CQRS)**

**Command Services:**
- Procesan eventos de Kafka
- Crean/actualizan aggregates
- Persisten en base de datos

**Query Services:**
- Consultas optimizadas
- Cálculo de KPIs
- Agregaciones y estadísticas

### 4. **Infrastructure Layer**

**Messaging:**
- Consumers de Kafka con Sarama
- Configuración SASL_SSL para Azure
- Auto-commit de offsets
- Retry logic y error handling

**Persistence:**
- Repositories con GORM
- PostgreSQL como store
- Migraciones automáticas

### 5. **Interfaces Layer**

**REST API (Gin):**
- Analytics endpoints
- KPI endpoints
- Health check endpoints
- Swagger documentation

---

## Patrones Implementados

### 1. **Domain-Driven Design (DDD)**
- ✅ Ubiquitous Language
- ✅ Bounded Context
- ✅ Aggregate Roots
- ✅ Value Objects
- ✅ Domain Services
- ✅ Repository Pattern

### 2. **CQRS (Command Query Responsibility Segregation)**
- ✅ Separación de comandos y consultas
- ✅ Optimización de lecturas
- ✅ Modelos de lectura específicos

### 3. **Event-Driven Architecture**
- ✅ Consumo asíncrono de eventos
- ✅ Desacoplamiento entre servicios
- ✅ Event sourcing básico

### 4. **Clean Architecture**
- ✅ Capas bien definidas
- ✅ Dependencias hacia el dominio
- ✅ Independencia de frameworks

---

## Características de Azure Event Hub

### Security
```
┌─────────────────────────────────────────┐
│  TLS 1.2+ Encryption                    │
│  ├── Certificate Validation             │
│  ├── Secure Channel                     │
│  └── No InsecureSkipVerify              │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  SASL/PLAIN Authentication              │
│  ├── Username: $ConnectionString        │
│  ├── Password: Full Connection String   │
│  └── Handshake: Enabled                 │
└─────────────────────────────────────────┘
```

### Performance & Reliability
- **Partitioning**: 2-4 particiones por Event Hub
- **Consumer Groups**: Procesamiento paralelo
- **Auto-commit**: Offsets gestionados automáticamente
- **Retry Logic**: Reintentos en caso de error
- **Timeouts**: 60 segundos (adaptado a Azure)
- **Heartbeat**: 3 segundos para mantener sesión

### Monitoring
- Logs detallados de conexión
- Tracking de offsets
- Error handling con logging
- Health checks del consumer

---

## Escalabilidad

### Horizontal Scaling
```
┌──────────────────────────────────────────────────────────┐
│  Azure Event Hub (4 partitions)                          │
└─────┬──────────┬──────────┬──────────┬──────────────────┘
      │          │          │          │
      │ P0       │ P1       │ P2       │ P3
      │          │          │          │
      ▼          ▼          ▼          ▼
   ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐
   │ C1   │  │ C2   │  │ C3   │  │ C4   │
   └──────┘  └──────┘  └──────┘  └──────┘
   Instance  Instance  Instance  Instance
      1         2         3         4
```

Cada instancia del microservicio puede consumir de una partición diferente del mismo Consumer Group.

### Vertical Scaling
- Aumento de recursos de compute
- Optimización de queries SQL
- Índices en PostgreSQL
- Caché de resultados (futuro)

---

## Seguridad

### 1. **Conexión a Azure Event Hub**
- ✅ TLS 1.2+ obligatorio
- ✅ SASL/PLAIN authentication
- ✅ Connection String encriptado
- ✅ No credenciales en código

### 2. **Base de Datos**
- ✅ SSL required en Azure PostgreSQL
- ✅ Credenciales en variables de entorno
- ✅ No SQL injection (prepared statements de GORM)

### 3. **API REST**
- ✅ Validación de UUIDs
- ✅ Validación de fechas
- ✅ CORS configurado
- ✅ Rate limiting (por implementar)

---

## Deployment

### Environments

**Development:**
- Local Kafka sin SSL
- PostgreSQL local
- Mock consumers

**Staging:**
- Azure Event Hub
- Azure PostgreSQL
- Eureka Service Discovery

**Production:**
- Azure Event Hub (SASL_SSL)
- Azure PostgreSQL (SSL required)
- Eureka Service Discovery
- Multiple instances
- Health checks
- Monitoring con Azure Monitor

---

## Tecnologías Utilizadas

| Componente | Tecnología | Versión |
|------------|-----------|---------|
| **Lenguaje** | Go | 1.23+ |
| **HTTP Framework** | Gin | 1.10.1 |
| **Kafka Client** | IBM Sarama | 1.42.1 |
| **ORM** | GORM | 1.25.5 |
| **Database** | PostgreSQL | 15+ |
| **Message Broker** | Azure Event Hub | Kafka 1.0+ compatible |
| **Service Discovery** | Netflix Eureka | - |
| **Documentation** | Swagger/OpenAPI | - |
| **Containerization** | Docker | - |

---

## Referencias

- [Azure Event Hub Documentation](https://docs.microsoft.com/azure/event-hubs/)
- [IBM Sarama Client](https://github.com/IBM/sarama)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [CQRS Pattern](https://martinfowler.com/bliki/CQRS.html)

---

**Última actualización**: 2024
**Versión**: 1.0.0