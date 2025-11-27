# 🔐 Guía de Configuración Segura - Analytics Microservice

## ⚠️ IMPORTANTE: Seguridad de Credenciales

Este proyecto contiene archivos de configuración con **credenciales sensibles**. Sigue estas instrucciones para configurar el microservicio de forma segura.

---

## 📁 Archivos de Configuración

### Archivos SEGUROS (pueden estar en Git)
- ✅ `.env.example` - Template genérico sin credenciales
- ✅ `.env.azure` - Template con placeholders `YOUR-NAMESPACE`, `YOUR-KEY`, etc.

### Archivos PRIVADOS (NO subir a Git)
- 🔒 `.env` - Tu configuración local (en .gitignore)
- 🔒 `.env.production` - Credenciales reales de producción (en .gitignore)
- 🔒 `.env.local` - Configuración de desarrollo (en .gitignore)

---

## 🚀 Configuración Rápida

### Opción 1: Desarrollo Local (sin Azure)

```bash
# 1. Copiar template
cp .env.example .env

# 2. Editar .env y configurar Kafka local
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_SECURITY_PROTOCOL=PLAINTEXT
DB_HOST=localhost

# 3. Iniciar con Docker Compose
docker-compose up -d

# 4. Ejecutar servicio
go run main.go
```

### Opción 2: Azure Event Hub (Producción)

```bash
# 1. Usar el archivo de producción (YA TIENE LAS CREDENCIALES REALES)
cp .env.production .env

# 2. (Opcional) Si el archivo .env.production no existe, crear desde template
cp .env.azure .env

# 3. Editar .env y reemplazar placeholders:
#    - YOUR-NAMESPACE → levelup-journey
#    - YOUR-SHARED-ACCESS-KEY-HERE → (tu key de Azure Portal)
#    - YOUR-SERVER → analytics (corregir si tiene typo)
#    - YOUR-USERNAME → levelup
#    - YOUR-PASSWORD → Journey12

# 4. Ejecutar servicio
go run main.go
```

---

## 🔑 Obtener Credenciales de Azure

### 1. Connection String de Event Hub

**Método 1: Azure Portal**
1. Ve a [Azure Portal](https://portal.azure.com)
2. Event Hub Namespace → `levelup-journey`
3. Shared access policies → `RootManageSharedAccessKey`
4. Copiar "Connection string-primary key"

**Método 2: Azure CLI**
```bash
az eventhubs namespace authorization-rule keys list \
  --resource-group tu-resource-group \
  --namespace-name levelup-journey \
  --name RootManageSharedAccessKey \
  --query primaryConnectionString -o tsv
```

### 2. Credenciales de Base de Datos

**PostgreSQL de Azure:**
- Host: `analytics.postgres.database.azure.com` (sin doble 's')
- Usuario: `levelup`
- Password: (obtener de Azure Portal o configuración segura)
- Base de datos: `analytics_db`
- SSL Mode: `require` (obligatorio para Azure)

---

## 🛡️ Seguridad: Qué NO Hacer

### ❌ NUNCA hacer esto:
```bash
# NO subir archivos con credenciales reales a GitHub
git add .env.production
git add .env

# NO hardcodear credenciales en el código
password := "Journey12"  // ❌ MAL

# NO compartir connection strings en documentación
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://...;SharedAccessKey=Ym7d+...  // ❌ MAL
```

### ✅ SÍ hacer esto:
```bash
# Usar variables de entorno
password := os.Getenv("DB_PASSWORD")  // ✅ BIEN

# Usar placeholders en documentación
AZURE_EVENTHUB_CONNECTION_STRING=Endpoint=sb://YOUR-NAMESPACE...;SharedAccessKey=YOUR-KEY  // ✅ BIEN

# Mantener .env en .gitignore
echo ".env" >> .gitignore  // ✅ BIEN
```

---

## 📝 Template de Configuración

### Configuración Completa para Azure Event Hub

```bash
# ===================================================
# Server
# ===================================================
SERVER_PORT=8080
SERVER_IP=0.0.0.0

# ===================================================
# Database (Azure PostgreSQL)
# ===================================================
DB_HOST=analytics.postgres.database.azure.com
DB_PORT=5432
DB_USER=levelup
DB_PASSWORD=TU-PASSWORD-AQUI
DB_NAME=analytics_db
DB_SSLMODE=require

# ===================================================
# Azure Event Hub (Kafka Protocol)
# ===================================================
KAFKA_BOOTSTRAP_SERVERS=levelup-journey.servicebus.windows.net:9093
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISM=PLAIN
KAFKA_SASL_USERNAME=$ConnectionString
AZURE_EVENTHUB_CONNECTION_STRING=TU-CONNECTION-STRING-COMPLETO-AQUI

# Topics
KAFKA_TOPIC=execution.analytics
KAFKA_USER_REGISTRATION_TOPIC=iam.user.registered

# Consumer Groups
KAFKA_GROUP_ID=analytics-consumer-group
KAFKA_USER_REGISTRATION_GROUP_ID=user-registration-analytics-group

# Timeouts
KAFKA_REQUEST_TIMEOUT_MS=60000
KAFKA_SESSION_TIMEOUT_MS=60000
KAFKA_ENABLE_AUTO_COMMIT=true

# ===================================================
# Service Discovery (Eureka)
# ===================================================
SERVICE_DISCOVERY_URL=https://discovery.yellowsea-767275f1.westus3.azurecontainerapps.io/eureka/
SERVICE_NAME=analytics-service
SERVICE_DISCOVERY_ENABLED=true
```

---

## 🔍 Verificar Configuración

### Script Automático
```bash
# Ejecutar script de verificación
bash scripts/verify-config.sh

# Salida esperada:
# ✓ All required variables are set correctly!
```

### Verificación Manual
```bash
# Verificar que las variables estén configuradas
echo $KAFKA_BOOTSTRAP_SERVERS
echo $AZURE_EVENTHUB_CONNECTION_STRING
echo $DB_HOST

# Probar conexión
go run main.go

# Buscar en logs:
# ✓ Azure Event Hub: Configured ✓
# ✓ Consumer group created successfully
```

---

## 🐛 Troubleshooting

### Error: "no pg_hba.conf entry... no encryption"

**Problema:** PostgreSQL requiere SSL pero está configurado como `disable`

**Solución:**
```bash
# En .env, cambiar:
DB_SSLMODE=require  # Obligatorio para Azure PostgreSQL
```

### Error: "host=analyticss.postgres.database.azure.com"

**Problema:** Typo en el nombre del host (doble 's')

**Solución:**
```bash
# Corregir a:
DB_HOST=analytics.postgres.database.azure.com
```

### Error: "SASL authentication failed"

**Problema:** Connection string incorrecto

**Solución:**
1. Verificar que `KAFKA_SASL_USERNAME=$ConnectionString` (literal)
2. Verificar que el connection string sea completo desde Azure Portal
3. No debe tener espacios ni saltos de línea

### Error: "Repository rule violations - Azure Event Hub Key"

**Problema:** GitHub bloqueó el push porque detectó credenciales en archivos .md

**Solución:**
1. Remover credenciales reales de archivos de documentación
2. Usar solo placeholders: `YOUR-KEY`, `YOUR-NAMESPACE`
3. Mantener credenciales reales SOLO en `.env` o `.env.production`

---

## 📦 Crear Event Hubs en Azure

Los siguientes Event Hubs deben existir:

### 1. execution.analytics
```bash
az eventhubs eventhub create \
  --resource-group tu-resource-group \
  --namespace-name levelup-journey \
  --name execution.analytics \
  --partition-count 4 \
  --message-retention 1
```

### 2. iam.user.registered
```bash
az eventhubs eventhub create \
  --resource-group tu-resource-group \
  --namespace-name levelup-journey \
  --name iam.user.registered \
  --partition-count 4 \
  --message-retention 1
```

---

## ✅ Checklist de Configuración

Antes de iniciar el servicio:

- [ ] Archivo `.env` creado desde `.env.production` o `.env.azure`
- [ ] Connection String de Azure Event Hub configurado
- [ ] Host de base de datos corregido (sin doble 's')
- [ ] SSL Mode configurado como `require`
- [ ] Event Hubs creados en Azure: `execution.analytics` y `iam.user.registered`
- [ ] Script de verificación ejecutado sin errores
- [ ] Archivo `.env` está en `.gitignore`
- [ ] NO hay credenciales reales en archivos `.md` o código fuente

---

## 🎯 Próximos Pasos

Después de configurar:

1. **Iniciar el servicio**
   ```bash
   go run main.go
   ```

2. **Verificar conexión**
   ```bash
   curl http://localhost:8080/health
   # Respuesta esperada: {"status":"UP"}
   ```

3. **Explorar Swagger UI**
   ```
   http://localhost:8080/swagger/index.html
   ```

4. **Monitorear logs**
   - Buscar: "Azure Event Hub: Configured ✓"
   - Buscar: "Consumer group created successfully"

---

## 📚 Documentación Adicional

- **[QUICKSTART.md](QUICKSTART.md)** - Inicio rápido en 3 pasos
- **[docs/AZURE_EVENT_HUB_CONFIG.md](docs/AZURE_EVENT_HUB_CONFIG.md)** - Guía completa de Azure
- **[README.md](README.md)** - Documentación general del proyecto

---

## 🆘 Soporte

Si encuentras problemas:

1. Revisa los logs del servicio
2. Ejecuta `bash scripts/verify-config.sh`
3. Consulta la sección de Troubleshooting arriba
4. Revisa que todas las credenciales sean correctas

---

**Última actualización:** 2024  
**Versión:** 1.1.0  
**Estado:** ✅ Configuración Segura Implementada