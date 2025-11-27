# Guía Rápida de Inicio - Analytics Microservice con Azure Event Hub

Esta guía te ayudará a configurar y ejecutar el microservicio de Analytics conectado a Azure Event Hub en menos de 5 minutos.

---

## 📋 Prerequisitos

- Go 1.23 o superior instalado
- Acceso a Azure Portal con Event Hub configurado
- PostgreSQL (local o Azure)

---

## 🚀 Inicio Rápido (3 pasos)

### Paso 1: Configurar Variables de Entorno

Copia el archivo de configuración de Azure:

```bash
# Windows (PowerShell)
Copy-Item .env.azure .env

# Windows (CMD)
copy .env.azure .env

# Linux/Mac
cp .env.azure .env
```

Edita el archivo `.env` y actualiza estas variables críticas:

```bash
# Connection String de Azure Event Hub
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://TU-NAMESPACE.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=TU-KEY

# Bootstrap Server (reemplaza TU-NAMESPACE)
KAFKA_BOOTSTRAP_SERVERS=TU-NAMESPACE.servicebus.windows.net:9093

# Base de datos
DB_HOST=tu-servidor.postgres.database.azure.com
DB_USER=tu-usuario
DB_PASSWORD=tu-password
DB_NAME=analytics_db
```

### Paso 2: Verificar Configuración

```bash
# Windows (Git Bash)
bash scripts/verify-config.sh

# Si todo está correcto verás:
# ✓ All required variables are set correctly!
```

### Paso 3: Iniciar el Servicio

```bash
# Descargar dependencias (primera vez)
go mod download

# Ejecutar el servicio
go run main.go
```

**¡Listo!** El servicio estará disponible en `http://localhost:8080`

---

## 🔍 Verificar que Funciona

### 1. Health Check

```bash
curl http://localhost:8080/health
```

**Respuesta esperada:**
```json
{"status":"UP"}
```

### 2. Swagger UI

Abre en tu navegador:
```
http://localhost:8080/swagger/index.html
```

### 3. Logs del Servicio

Busca estas líneas en la consola:

```
✓ Kafka Configuration:
  Bootstrap Servers: [YOUR-NAMESPACE.servicebus.windows.net:9093]
  Security Protocol: SASL_SSL
  Azure Event Hub: Configured ✓

✓ Consumer group created successfully for topic: execution.analytics
✓ Consumer group created successfully for topic: iam.user.registered
✓ Starting HTTP server on 0.0.0.0:8080
```

---

## 🎯 Obtener Connection String de Azure

### Método 1: Azure Portal (GUI)

1. Ve a [Azure Portal](https://portal.azure.com)
2. Navega a tu **Event Hub Namespace**
3. En el menú izquierdo: **Shared access policies**
4. Click en **RootManageSharedAccessKey**
5. Copia el valor de **Connection string-primary key**

### Método 2: Azure CLI

```bash
az eventhubs namespace authorization-rule keys list \
  --resource-group tu-resource-group \
  --namespace-name tu-namespace \
  --name RootManageSharedAccessKey \
  --query primaryConnectionString \
  --output tsv
```

---

## 📦 Crear Event Hubs (Topics) en Azure

Debes crear estos Event Hubs en tu namespace:

### Via Azure Portal:

1. Ve a tu **Event Hub Namespace**
2. Click en **+ Event Hub**
3. Nombre: `execution.analytics`
   - Particiones: 2-4
   - Retention: 1 día
4. Repite para: `iam.user.registered`

### Via Azure CLI:

```bash
# Crear Event Hub para ejecuciones
az eventhubs eventhub create \
  --resource-group tu-resource-group \
  --namespace-name tu-namespace \
  --name execution.analytics \
  --partition-count 4 \
  --message-retention 1

# Crear Event Hub para registros de usuarios
az eventhubs eventhub create \
  --resource-group tu-resource-group \
  --namespace-name tu-namespace \
  --name iam.user.registered \
  --partition-count 4 \
  --message-retention 1
```

---

## 🔧 Configuración Mínima vs Completa

### Configuración Mínima (solo variables críticas)

```bash
# Azure Event Hub
KAFKA_BOOTSTRAP_SERVERS=YOUR-NAMESPACE.servicebus.windows.net:9093
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_USERNAME=$ConnectionString
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://YOUR-NAMESPACE.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=YOUR-KEY

# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=analytics_db

# Server
SERVER_PORT=8080
```

### Configuración Completa (producción)

Usa el archivo `.env.azure` que incluye:
- Timeouts optimizados
- Consumer groups configurados
- Service discovery (Eureka)
- SSL para base de datos
- Todas las opciones de Kafka

---

## 🐛 Troubleshooting Rápido

### Error: "Failed to connect to broker"

**Causa:** Bootstrap server incorrecto o puerto equivocado.

**Solución:**
```bash
# Verificar formato correcto (puerto 9093 para Azure)
KAFKA_BOOTSTRAP_SERVERS=tu-namespace.servicebus.windows.net:9093
```

### Error: "SASL authentication failed"

**Causa:** Connection string incorrecto o username mal configurado.

**Solución:**
```bash
# Username debe ser literal "$ConnectionString"
KAFKA_SASL_USERNAME=$ConnectionString

# Verificar connection string completo
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://YOUR-NAMESPACE.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=YOUR-KEY
```

### Error: "Topic not found"

**Causa:** El Event Hub no existe en Azure.

**Solución:**
1. Crear Event Hubs en Azure Portal
2. Verificar nombres exactos (case-sensitive):
   - `execution.analytics`
   - `iam.user.registered`

### Error: "Database connection failed"

**Causa:** Credenciales de base de datos incorrectas.

**Solución:**
```bash
# Para Azure PostgreSQL, usar SSL
DB_SSLMODE=require

# Verificar host completo
DB_HOST=servidor.postgres.database.azure.com
```

---

## 📊 Endpoints Disponibles

### Analytics Endpoints

```bash
# Obtener análisis por ejecución
GET /api/v1/analytics/execution/{executionId}

# Obtener ejecuciones por estudiante
GET /api/v1/analytics/student/{studentId}

# Obtener ejecuciones por desafío
GET /api/v1/analytics/challenge/{challengeId}

# Obtener ejecuciones por rango de fechas
GET /api/v1/analytics/date-range?startDate=2024-01-01T00:00:00Z&endDate=2024-01-31T23:59:59Z
```

### KPI Endpoints

```bash
# KPIs de estudiante
GET /api/v1/analytics/kpi/student/{studentId}

# KPIs de desafío
GET /api/v1/analytics/kpi/challenge/{challengeId}

# Estadísticas diarias
GET /api/v1/analytics/kpi/daily?startDate=2024-01-01T00:00:00Z&endDate=2024-01-07T23:59:59Z

# Estadísticas por lenguaje
GET /api/v1/analytics/kpi/languages?startDate=2024-01-01T00:00:00Z&endDate=2024-01-31T23:59:59Z

# Top desafíos fallidos
GET /api/v1/analytics/kpi/top-failed-challenges?limit=10
```

### User Registration Endpoints

```bash
# Obtener registros por usuario
GET /api/v1/user-registration/{userId}

# Obtener todos los registros
GET /api/v1/user-registration

# KPIs de registros por día
GET /api/v1/user-registration/kpi/daily?startDate=2024-01-01T00:00:00Z&endDate=2024-01-31T23:59:59Z

# Total de registros
GET /api/v1/user-registration/kpi/total
```

---

## 🔄 Desarrollo Local (sin Azure)

Si quieres desarrollar localmente sin Azure Event Hub:

### 1. Usar Kafka local con Docker

```bash
docker-compose up -d
```

### 2. Configurar para Kafka local

```bash
# .env para desarrollo local
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_SECURITY_PROTOCOL=PLAINTEXT
KAFKA_TOPIC=execution.analytics
KAFKA_USER_REGISTRATION_TOPIC=iam.user.registered

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=analytics_db
DB_SSLMODE=disable

SERVER_PORT=8080
SERVER_IP=127.0.0.1
```

---

## 📚 Documentación Completa

Para información detallada, consulta:

- **[AZURE_EVENT_HUB_CONFIG.md](docs/AZURE_EVENT_HUB_CONFIG.md)** - Configuración completa de Azure Event Hub
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Arquitectura y flujo de datos
- **[README.md](README.md)** - Documentación general del proyecto
- **[CHANGELOG.md](CHANGELOG.md)** - Historial de cambios

---

## ✅ Checklist de Configuración

Antes de ejecutar el servicio, verifica:

- [ ] Go 1.23+ instalado
- [ ] Event Hub Namespace creado en Azure
- [ ] Event Hubs creados: `execution.analytics` y `iam.user.registered`
- [ ] Connection String obtenido de Azure
- [ ] Archivo `.env` configurado con tus credenciales
- [ ] Base de datos PostgreSQL accesible
- [ ] Puerto 8080 disponible
- [ ] Script de verificación ejecutado sin errores

---

## 🎓 Próximos Pasos

Después de que el servicio esté corriendo:

1. **Producir eventos de prueba** al Event Hub desde otro servicio
2. **Verificar consumo** en los logs del microservicio
3. **Consultar analytics** a través de los endpoints REST
4. **Explorar Swagger UI** para ver todos los endpoints disponibles
5. **Monitorear** en Azure Portal el consumo de mensajes

---

## 💡 Tips de Producción

### Seguridad

- ✅ Usar Azure Key Vault para credenciales
- ✅ Crear políticas de acceso separadas (Send/Listen)
- ✅ No usar RootManageSharedAccessKey en producción
- ✅ Habilitar SSL en PostgreSQL

### Performance

- ✅ Ajustar número de particiones según carga
- ✅ Configurar múltiples instancias del servicio
- ✅ Usar consumer groups diferentes por instancia
- ✅ Monitorear latencia y throughput

### Monitoring

- ✅ Configurar Azure Monitor
- ✅ Logs estructurados con niveles apropiados
- ✅ Health checks periódicos
- ✅ Alertas para errores críticos

---

## 🆘 Soporte

¿Problemas? Revisa en orden:

1. **Logs del servicio** - Busca errores en la consola
2. **Script de verificación** - `bash scripts/verify-config.sh`
3. **Troubleshooting** - Sección anterior o en `docs/AZURE_EVENT_HUB_CONFIG.md`
4. **Azure Portal** - Verifica estado de Event Hub y base de datos
5. **Conectividad** - Verifica firewall y reglas de red

---

**Tiempo estimado de configuración:** 5-10 minutos  
**Última actualización:** 2024  
**Versión:** 1.1.0

¡Feliz desarrollo! 🚀