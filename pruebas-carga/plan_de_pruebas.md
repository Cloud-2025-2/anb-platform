# Plan de Pruebas de Carga y Análisis de Capacidad

## 1. Objetivos
- Medir la **capacidad de procesamiento (throughput)**, **tiempos de respuesta** y **utilización de recursos** del sistema `anb-platform` bajo diferentes niveles de concurrencia.  
- Validar **rutas críticas** del usuario (capa web) y un **flujo batch** (procesamiento asíncrono vía broker).  
- Establecer **criterios de aceptación (SLO/SLI)** y documentación reproducible de resultados, hallazgos y mejoras.

---

## 2. Alcance
- **Incluye**: API (Go + Gin), frontend (Vite/React), worker de procesamiento (Kafka/asíncrono), DB (PostgreSQL), Nginx (reverse proxy) y almacenamiento.  
- **Excluye**: Pruebas funcionales exhaustivas, seguridad ofensiva y pruebas E2E de UI visual.

---

## 3. Entorno de Pruebas

### 3.1 Topología de ambientes
- **SUT (System Under Test)**: despliegue Docker Compose en VM Linux.  
- **Generador de carga**: VM separada para ejecutar k6.

### 3.2 Infraestructura
| Rol | Tipo | vCPU | RAM | Disco | Uso |
|---|---|---:|---:|---:|---|
| SUT | VM (Docker Compose) | 2 | 4–8 GB | 50–100 GB | API+DB+Kafka+Worker+Frontend |
| Carga | VM | 2 | 8 GB | 20 GB | k6 runner |

### 3.3 Monitoreo
- **Prometheus + Grafana**: CPU, MEM, disco, red.  
- **Postgres Exporter**: conexiones, locks, I/O.  
- **Kafka Exporter**: throughput, lag, reintentos, DLQ.  

---

## 4. Herramientas
- **k6**: pruebas de carga principales.  
- **ab (ApacheBench)**: pruebas rápidas de humo.  
- **Prometheus/Grafana**: observabilidad y métricas.  
- **Postman/Newman**: preparación de datos de prueba.

---

## 5. Criterios de Aceptación (SLO)
| Métrica | Umbral |
|---|---|
| **p95 tiempo de respuesta** | ≤ 800 ms |
| **p99 tiempo de respuesta** | ≤ 1500 ms |
| **Errores** | ≤ 1% (HTTP 5xx/4xx inesperados) |
| **CPU** | ≤ 80% sostenido |
| **Kafka lag** | < 100 mensajes en steady state |
| **Reintentos/DLQ** | ≤ 3% / DLQ = 0 |

---

## 6. Escenarios de Prueba

### Escenario A — Autenticación y navegación
1. `POST /api/auth/login`  
2. `GET /api/videos`  
3. `GET /api/users/me`

### Escenario B — Carga y publicación de video
1. `POST /api/auth/login`  
2. `POST /api/videos` (multipart 50–100MB)  
3. `GET /api/videos/my` (estado: UPLOADED → PROCESSING → READY)

### Escenario C — Procesamiento batch
- Generar uploads → encolar en Kafka → worker procesa → actualizar DB → estado READY.

---

## 7. Estrategia
1. **Humo**: 5–10 usuarios, 2–3 min.  
2. **Carga progresiva**: 10 → 25 → 50 → 100 usuarios.  
3. **Estrés**: aumentar hasta degradación clara.  
4. **Soak**: 60 min a carga media para fugas de memoria.

---

## 8. Topología (Mermaid)

```
mermaid
flowchart LR
  subgraph Client[Generador de carga (k6)]
  end
  subgraph SUT[VM SUT - Docker]
    Nginx --> API[API Go/Gin]
    API --> DB[(PostgreSQL)]
    API --> Kafka[(Kafka Broker)]
    Worker[Worker/Consumer] --> Kafka
    Worker --> Storage[(Almacenamiento)]
  end
  Client -->|HTTPs| Nginx´´´




