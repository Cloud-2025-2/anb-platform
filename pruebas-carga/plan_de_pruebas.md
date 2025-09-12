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
  Client -->|HTTPs| Nginx

---
## 9. Flujo de Proceso (Upload → Procesamiento → Publicación)

```
mermaid
sequenceDiagram
  participant U as Usuario
  participant API as API
  participant K as Kafka
  participant W as Worker
  participant ST as Storage
  participant DB as Postgres

  U->>API: POST /api/videos
  API->>ST: Guardar archivo
  API->>K: Encolar job
  K-->>W: Mensaje de proceso
  W->>ST: Leer video, procesar
  W->>DB: Actualizar estado
  U->>API: GET /api/videos/my

---
## 10. Datos de Prueba
Usuario de prueba con rol estándar.

Archivos de video: 50MB, 75MB, 100MB.

Limpieza de datos y archivos al finalizar.

## 11. Métricas a Recopilar
Web/API: latencia (p50, p95, p99), errores, throughput.

Sistema: CPU, RAM, disco, red.

Kafka: lag, throughput, reintentos, DLQ.

Batch: latencia de tarea (enqueue→done), reintentos.

## 12. Scripts Base (k6)
login_and_list.js
js
Copy code
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 10 },
    { duration: '2m', target: 25 },
    { duration: '2m', target: 50 },
    { duration: '2m', target: 0  },
  ],
  thresholds: {
    http_req_duration: ['p(95)<800', 'p(99)<1500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE = __ENV.BASE_URL || 'https://<tu-dominio>';
const USER = __ENV.USER_EMAIL || 'user@test.com';
const PASS = __ENV.USER_PASS  || 'secret';

export default function () {
  const resLogin = http.post(`${BASE}/api/auth/login`, JSON.stringify({ email: USER, password: PASS }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(resLogin, { 'login 200': (r) => r.status === 200 });
  const token = resLogin.json('token');

  const resList = http.get(`${BASE}/api/videos`, { headers: { Authorization: `Bearer ${token}` } });
  check(resList, { 'list 200': (r) => r.status === 200 });

  sleep(1);
}
upload_and_poll.js
js
Copy code
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 10 },
    { duration: '3m', target: 25 },
    { duration: '3m', target: 50 },
    { duration: '1m', target: 0  },
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
  },
};

const BASE = __ENV.BASE_URL || 'https://<tu-dominio>';
const USER = __ENV.USER_EMAIL || 'user@test.com';
const PASS = __ENV.USER_PASS  || 'secret';
const FILE = __ENV.FILE_PATH  || '/data/video_50mb.mp4';

export default function () {
  const login = http.post(`${BASE}/api/auth/login`, JSON.stringify({ email: USER, password: PASS }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(login, { 'login 200': (r) => r.status === 200 });
  const token = login.json('token');

  const form = {
    video_file: http.file(open(FILE, 'b'), 'video_50mb.mp4', 'video/mp4'),
    title: `k6-upload-${__VU}-${Date.now()}`
  };
  const up = http.post(`${BASE}/api/videos`, form, { headers: { Authorization: `Bearer ${token}` } });
  check(up, { 'upload 201/200': (r) => r.status === 201 || r.status === 200 });

  for (let i = 0; i < 30; i++) {
    const my = http.get(`${BASE}/api/videos/my`, { headers: { Authorization: `Bearer ${token}` } });
    if (my.status === 200 && JSON.stringify(my.body).includes('READY')) break;
    sleep(2);
  }
}
## 13. Ejecución


# Escenario A
´´´BASE_URL=https://<tu-dominio> USER_EMAIL=user@test.com USER_PASS=secret \
k6 run scripts/login_and_list.js'''

# Escenario B
´´BASE_URL=https://<tu-dominio> USER_EMAIL=user@test.com USER_PASS=secret \
FILE_PATH=/data/video_50mb.mp4 k6 run scripts/upload_and_poll.js´´´

## 14. Resultados y Evidencia

Tabla resumen

|Escenario	|Usuarios	|Duración	|p95 (ms)	|p99 (ms)	|Throughput (req/s)	|Errores (%)	|CPU (%)	|MEM (GB)	|Kafka lag|
|A	|10→25→50	|8 min	|—|	—|	—|	—|	—|	—|	—|
|B	|10→25→50	|8 min	|—|	|—	|—	|—	|—	|—	|—|
|C	|N/A	|10 min	|N/A	|N/A	tareas/min|	—	|—	|-	|—|

## 15. Interpretación
Capacidad actual: soporta hasta X usuarios con p95 < 800 ms.

Cuellos de botella: CPU, I/O o Kafka lag según carga.

Errores: registrar causas (timeout, límite de tamaño, etc.).

Batch: latencia promedio de tarea Y s, sin DLQ.

## 16. Plan de Mejora
Ajustar Nginx (client_max_body_size, timeouts, gzip).

Optimizar DB pooling y workers.

Configurar alertas de Kafka (lag, DLQ).

Escalar vCPU/memoria si necesario.

Implementar observabilidad adicional (tracing, métricas de negocio).

## 17. Riesgos
Latencia de red puede sesgar resultados.

Archivos de prueba deben ser no sensibles.

VM de carga debe tener recursos suficientes.

## 18. Anexos
Scripts k6 (/pruebas-carga/scripts/).

Colección Postman (para setup).

Dashboards exportados (Grafana).

Evidencias de ejecución (JSON/CSV y screenshots).

Guía de despliegue reproducible del entorno.

yaml
Copy code





