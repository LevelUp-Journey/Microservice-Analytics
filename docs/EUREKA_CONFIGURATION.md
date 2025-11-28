# Configuración de Eureka - IP Pública vs IP de Servidor

## 📌 Problema

Cuando ejecutas el microservicio, necesitas dos configuraciones diferentes de IP:

1. **IP del Servidor (SERVER_IP)**: La interfaz en la que el servidor HTTP escucha
2. **IP de Eureka (EUREKA_INSTANCE_IP)**: La IP que otros servicios usarán para conectarse

---

## 🎯 Escenarios de Configuración

### Escenario 1: Desarrollo Local (Default)

```bash
# El servidor escucha en todas las interfaces
SERVER_IP=0.0.0.0
SERVER_PORT=8080

# Eureka detecta automáticamente tu IP local
EUREKA_INSTANCE_IP=
# Resultado: Registra 192.168.0.56:8080
```

**Funcionamiento:**
- El servidor escucha en `0.0.0.0:8080` (todas las interfaces)
- Eureka detecta automáticamente tu IP local (ej: `192.168.0.56`)
- Otros servicios en tu red local pueden conectarse

---

### Escenario 2: Servidor con IP Pública

```bash
# El servidor escucha en todas las interfaces
SERVER_IP=0.0.0.0
SERVER_PORT=8080

# IP pública para que otros servicios se conecten
EUREKA_INSTANCE_IP=203.0.113.10
```

**Funcionamiento:**
- El servidor escucha en `0.0.0.0:8080` (todas las interfaces)
- Eureka registra `203.0.113.10:8080`
- Otros servicios usan la IP pública para conectarse

**Caso de uso:** Servidor en cloud con IP pública estática

---

### Escenario 3: Azure/AWS con Hostname

```bash
# El servidor escucha en todas las interfaces
SERVER_IP=0.0.0.0
SERVER_PORT=8080

# Hostname público
SERVER_HOSTNAME=analytics.azurecontainerapps.io
EUREKA_INSTANCE_IP=
```

**Funcionamiento:**
- El servidor escucha en `0.0.0.0:8080`
- Eureka registra `analytics.azurecontainerapps.io:8080`
- Otros servicios resuelven el hostname via DNS

**Caso de uso:** Azure Container Apps, AWS ECS, Kubernetes

---

### Escenario 4: IP Pública + Hostname (Prioridad a IP)

```bash
SERVER_IP=0.0.0.0
SERVER_PORT=8080

# Si ambas están configuradas, EUREKA_INSTANCE_IP tiene prioridad
SERVER_HOSTNAME=analytics.azurecontainerapps.io
EUREKA_INSTANCE_IP=20.201.27.6
```

**Funcionamiento:**
- El servidor escucha en `0.0.0.0:8080`
- Eureka registra `20.201.27.6:8080` (la IP tiene prioridad)

---

## 🔍 Orden de Prioridad

El microservicio determina la IP de Eureka en este orden:

```
1. EUREKA_INSTANCE_IP (si está configurada)
   ↓
2. SERVER_HOSTNAME (si EUREKA_INSTANCE_IP está vacía)
   ↓
3. Auto-detección (si ambas están vacías o son 0.0.0.0)
```

---

## 📝 Ejemplos Prácticos

### Ejemplo 1: Desarrollo Local

```bash
# .env
SERVER_IP=0.0.0.0
SERVER_PORT=8080
EUREKA_INSTANCE_IP=
```

**Logs esperados:**
```
Auto-detected IP address for Eureka registration: 192.168.0.56
Successfully registered with Eureka: analytics-service:192.168.0.56:8080
```

---

### Ejemplo 2: Azure Container Apps

```bash
# .env
SERVER_IP=0.0.0.0
SERVER_PORT=8080
SERVER_HOSTNAME=analytics.yellowsea-767275f1.westus3.azurecontainerapps.io
EUREKA_INSTANCE_IP=
```

**Logs esperados:**
```
Using configured IP address for Eureka registration: analytics.yellowsea-767275f1.westus3.azurecontainerapps.io
Successfully registered with Eureka: analytics-service:analytics.yellowsea-767275f1.westus3.azurecontainerapps.io:8080
```

---

### Ejemplo 3: Servidor con IP Pública Fija

```bash
# .env
SERVER_IP=0.0.0.0
SERVER_PORT=8080
EUREKA_INSTANCE_IP=203.0.113.10
```

**Logs esperados:**
```
Using configured IP address for Eureka registration: 203.0.113.10
Successfully registered with Eureka: analytics-service:203.0.113.10:8080
```

---

## 🚀 Configuración para Producción

### Azure Container Apps

1. Obtén tu hostname público:
   ```bash
   az containerapp show \
     --name analytics-service \
     --resource-group tu-resource-group \
     --query properties.configuration.ingress.fqdn -o tsv
   ```

2. Configura en `.env`:
   ```bash
   SERVER_HOSTNAME=analytics.yellowsea-767275f1.westus3.azurecontainerapps.io
   ```

### AWS ECS/EC2

1. Obtén tu IP pública:
   ```bash
   curl -s http://169.254.169.254/latest/meta-data/public-ipv4
   ```

2. Configura en `.env`:
   ```bash
   EUREKA_INSTANCE_IP=203.0.113.10
   ```

### Kubernetes

1. Usa el Service Name interno:
   ```bash
   SERVER_HOSTNAME=analytics-service.default.svc.cluster.local
   ```

2. O configura Ingress y usa el hostname externo:
   ```bash
   SERVER_HOSTNAME=analytics.example.com
   ```

---

## ❓ Preguntas Frecuentes

### ¿Por qué SERVER_IP debe ser 0.0.0.0?

Para que el servidor escuche en **todas las interfaces de red**, permitiendo conexiones desde:
- Localhost (127.0.0.1)
- Red local (192.168.x.x)
- IP pública (si aplica)

**Alternativas:**
- `127.0.0.1` - Solo localhost (no recomendado en producción)
- `192.168.0.56` - Solo esa IP específica

---

### ¿Cuándo usar EUREKA_INSTANCE_IP vs SERVER_HOSTNAME?

**Usa EUREKA_INSTANCE_IP cuando:**
- Tienes una IP pública fija
- Quieres especificar explícitamente la IP
- No tienes un hostname o DNS configurado

**Usa SERVER_HOSTNAME cuando:**
- Usas servicios cloud con hostnames dinámicos (Azure, AWS)
- Tienes un dominio personalizado
- Usas load balancers o API gateways

---

### ¿Qué pasa si no configuro ninguna?

El microservicio detectará automáticamente tu IP local usando:
```go
conn, _ := net.Dial("udp", "8.8.8.8:80")
localAddr := conn.LocalAddr().(*net.UDPAddr)
ip := localAddr.IP.String()
```

Esto funciona bien para desarrollo local, pero **no es recomendado para producción**.

---

### ¿Puedo usar localhost (127.0.0.1)?

**No recomendado**. Si registras `127.0.0.1` en Eureka, otros servicios intentarán conectarse a su propio localhost, no al tuyo.

---

## 🛠️ Troubleshooting

### Error: "Connection refused" desde otros servicios

**Problema:** Otros servicios no pueden conectarse a tu microservicio.

**Posibles causas:**

1. **IP incorrecta en Eureka:**
   ```bash
   # Verifica qué IP se registró
   curl http://eureka-server:8761/eureka/apps/analytics-service
   ```

2. **Firewall bloqueando:**
   ```bash
   # Verifica que el puerto esté abierto
   telnet <EUREKA_INSTANCE_IP> 8080
   ```

3. **Servidor escuchando en IP incorrecta:**
   ```bash
   # Verifica en qué IP escucha
   netstat -tuln | grep 8080
   # Debería mostrar 0.0.0.0:8080
   ```

**Solución:**
```bash
# Asegúrate de que:
SERVER_IP=0.0.0.0  # No 127.0.0.1
EUREKA_INSTANCE_IP=<tu-ip-publica>  # No 127.0.0.1 ni 192.168.x.x
```

---

### Verificar registro en Eureka

```bash
# Ver todos los servicios registrados
curl http://eureka-server:8761/eureka/apps

# Ver solo analytics-service
curl http://eureka-server:8761/eureka/apps/analytics-service

# Verificar IP registrada
curl http://eureka-server:8761/eureka/apps/analytics-service | grep ipAddr
```

---

## 📊 Tabla Comparativa

| Configuración | SERVER_IP | EUREKA_INSTANCE_IP | SERVER_HOSTNAME | Caso de Uso |
|---------------|-----------|-------------------|-----------------|-------------|
| Desarrollo Local | 0.0.0.0 | (vacío) | (vacío) | Testing local |
| IP Pública Fija | 0.0.0.0 | 203.0.113.10 | (vacío) | VPS, EC2 |
| Azure Container Apps | 0.0.0.0 | (vacío) | *.azurecontainerapps.io | Azure Cloud |
| AWS ECS | 0.0.0.0 | 203.0.113.10 | api.example.com | AWS Cloud |
| Kubernetes | 0.0.0.0 | (vacío) | svc.cluster.local | K8s interno |
| Docker Compose | 0.0.0.0 | (vacío) | analytics-service | Docker red |

---

## ✅ Checklist de Configuración

Antes de desplegar en producción:

- [ ] `SERVER_IP=0.0.0.0` (para escuchar en todas las interfaces)
- [ ] `SERVER_PORT` configurado (default: 8080)
- [ ] `EUREKA_INSTANCE_IP` o `SERVER_HOSTNAME` configurado con IP/hostname público
- [ ] Firewall permite conexiones al puerto configurado
- [ ] Eureka puede resolver el hostname (si usas SERVER_HOSTNAME)
- [ ] Otros servicios pueden alcanzar la IP/hostname configurado
- [ ] Logs muestran: "Using configured IP address for Eureka registration"
- [ ] Registro exitoso en Eureka verificado

---

## 📚 Referencias

- [Netflix Eureka Documentation](https://github.com/Netflix/eureka/wiki)
- [Spring Cloud Netflix Eureka](https://cloud.spring.io/spring-cloud-netflix/reference/html/)
- [Eureka REST API](https://github.com/Netflix/eureka/wiki/Eureka-REST-operations)

---

**Última actualización:** 2024  
**Versión:** 1.1.0  
**Estado:** Configuración Completa

---

## 🎯 Resumen Rápido

```bash
# Desarrollo Local
SERVER_IP=0.0.0.0
EUREKA_INSTANCE_IP=
# ✓ Auto-detecta IP local

# Producción con IP Pública
SERVER_IP=0.0.0.0
EUREKA_INSTANCE_IP=<tu-ip-publica>
# ✓ Usa IP específica

# Producción con Hostname
SERVER_IP=0.0.0.0
SERVER_HOSTNAME=<tu-hostname.com>
# ✓ Usa hostname público
```

**Regla de oro:** Siempre usa `SERVER_IP=0.0.0.0` para que el servidor escuche en todas las interfaces. Configura `EUREKA_INSTANCE_IP` o `SERVER_HOSTNAME` con la dirección que otros servicios deben usar para conectarse.