# Arquitectura de la Aplicación - Entrega 4
## ANB Platform - Escalabilidad en Backend y Workers

### Información del Equipo
- **Nombre del proyecto:** ANB Platform
- **Fecha:** Noviembre 2025
- **Curso:** Desarrollo de Software en la Nube

---

## 1. Descripción General

Sistema de procesamiento de videos desplegado en AWS con arquitectura escalable y distribuida. La aplicación permite a los usuarios subir videos que son procesados automáticamente agregando marcas de agua.

---

## 2. Arquitectura Implementada

### 2.1 Componentes Principales

#### **Capa de Presentación (Frontend)**
- **Tecnología:** React + Vite + Nginx
- **Despliegue:** Amazon ECS Fargate
- **Auto Scaling:** 1-3 instancias basado en CPU (threshold: 70%)
- **Load Balancer:** Application Load Balancer
- **Zona:** us-east-1 (Multi-AZ: 1a, 1b)
- **Health Check:** `/` (HTTP 200)

#### **Capa de Aplicación (Backend API)**
- **Tecnología:** Go (Gin Framework)
- **Despliegue:** Amazon ECS Fargate
- **Auto Scaling:** 1-3 instancias basado en CPU (threshold: 70%)
- **Load Balancer:** Application Load Balancer
- **Zona:** us-east-1 (Multi-AZ: 1a, 1b)
- **Puerto:** 8000
- **Health Check:** `/api/health`
- **Endpoints principales:**
  - POST /api/auth/signup
  - POST /api/auth/login
  - POST /api/videos/upload
  - GET /api/videos
  - GET /api/public/videos

#### **Capa de Procesamiento (Workers)**
- **Tecnología:** Go + FFmpeg
- **Despliegue:** Amazon ECS Fargate
- **Auto Scaling:** 1-3 instancias basado en CPU (threshold: 70%)
- **Función:** Procesar videos agregando marca de agua
- **Consumidor:** Amazon SQS

#### **Base de Datos**
- **Servicio:** Amazon RDS PostgreSQL 15.4
- **Instancia:** db.t3.micro
- **Almacenamiento:** 20 GB gp2
- **Acceso:** Privado (solo desde ECS)

#### **Almacenamiento de Objetos**
- **Servicio:** Amazon S3
- **Bucket:** anb-video-storage-{account-id}
- **Contenido:** Videos originales y procesados
- **CORS:** Configurado para uploads desde frontend

#### **Sistema de Mensajería**
- **Servicio:** Amazon SQS
- **Cola:** anb-video-processing-queue
- **Visibility Timeout:** 300 segundos
- **Retention Period:** 24 horas
- **Función:** Comunicación asíncrona backend → workers

#### **Monitoreo**
- **Servicio:** Amazon CloudWatch
- **Log Groups:**
  - /ecs/anb-backend
  - /ecs/anb-frontend
  - /ecs/anb-worker
- **Alarmas:** CPU > 80% para backend y workers

---

## 3. Diagrama de Arquitectura
```
                             Internet
                                |
                    ┌───────────┴───────────┐
                    │                       │
            ┌───────▼────────┐      ┌──────▼───────┐
            │  Frontend ALB  │      │  Backend ALB │
            │   (Port 80)    │      │  (Port 80)   │
            └───────┬────────┘      └──────┬───────┘
                    │                      │
        ┌───────────┴──────────┐  ┌────────┴─────────┐
        │                      │  │                  │
    ┌───▼────┐           ┌────▼────┐          ┌─────▼────┐
    │Frontend│           │Frontend │          │ Backend  │
    │ECS Task│           │ECS Task │          │ ECS Task │
    │  (1-3) │           │  (1-3)  │          │  (1-3)   │
    └────────┘           └─────────┘          └─────┬────┘
         │                                          │
         │  (Nginx proxy)                           │ Publish
         └──────────────────────────────────────────┤
                                                    │
                                              ┌─────▼─────┐
                                              │    SQS    │
                                              │   Queue   │
                                              └─────┬─────┘
                                                    │
                                                    │ Consume
                                              ┌─────▼─────┐
                                              │  Worker   │
                                              │ ECS Task  │
                                              │   (1-3)   │
                                              └─────┬─────┘
                                                    │
                        ┌───────────────────────────┼──────────────┐
                        │                           │              │
                   ┌────▼────┐                ┌────▼────┐    ┌───▼────┐
                   │   RDS   │                │   S3    │    │CloudW. │
                   │Postgres │                │ Bucket  │    │  Logs  │
                   └─────────┘                └─────────┘    └────────┘
```

---

## 4. Estrategia de Auto Scaling

### 4.1 Frontend
- **Métrica:** CPUUtilization
- **Target Value:** 70%
- **Min Capacity:** 1
- **Max Capacity:** 3
- **Scale Out Cooldown:** 60s
- **Scale In Cooldown:** 60s

### 4.2 Backend
- **Métrica:** CPUUtilization
- **Target Value:** 70%
- **Min Capacity:** 1
- **Max Capacity:** 3
- **Scale Out Cooldown:** 60s
- **Scale In Cooldown:** 60s

### 4.3 Workers
- **Métrica:** CPUUtilization
- **Target Value:** 70%
- **Min Capacity:** 1
- **Max Capacity:** 3
- **Scale Out Cooldown:** 300s
- **Scale In Cooldown:** 300s

---

## 5. Alta Disponibilidad

### 5.1 Multi-AZ Deployment
- Todas las tareas ECS se despliegan en 2 zonas de disponibilidad:
  - us-east-1a
  - us-east-1b
- Los Load Balancers distribuyen tráfico entre ambas zonas
- Auto Scaling mantiene mínimo 1 instancia activa

### 5.2 Health Checks
- **Frontend:** GET `/` cada 30s
- **Backend:** GET `/api/health` cada 30s
- **Threshold:** 2 checks consecutivos fallidos para marcar unhealthy

---

## 6. Seguridad

### 6.1 Network Security
- **Security Groups configurados:**
  - ALB SG: Permite HTTP (80) desde Internet
  - ECS SG: Permite tráfico solo desde ALB
  - RDS SG: Permite PostgreSQL (5432) solo desde ECS

### 6.2 Acceso a Datos
- RDS: No públicamente accesible
- S3: Acceso mediante IAM roles
- Credenciales: Gestionadas mediante variables de entorno

---

## 7. Flujo de Procesamiento de Videos

1. Usuario sube video a través del frontend
2. Frontend envía video al backend via POST /api/videos/upload
3. Backend guarda video en S3 y metadata en RDS
4. Backend publica mensaje en SQS con información del video
5. Worker consume mensaje de SQS
6. Worker descarga video de S3, lo procesa (agrega marca de agua)
7. Worker sube video procesado a S3
8. Worker actualiza estado en RDS a "published"
9. Usuario puede visualizar video procesado

---

## 8. Cambios Respecto a Entrega 3

### 8.1 Migraciones
- Migración de Kafka a Amazon SQS
- Implementación de Auto Scaling en web y workers
- Despliegue Multi-AZ
- Configuración de alarmas CloudWatch

### 8.2 Mejoras
- Frontend con proxy a backend (solución de CORS)
- Workers completamente desacoplados vía SQS
- Health checks configurados correctamente
- Load Balancers en ambas capas

---

## 9. Costos Estimados

| Recurso | Tipo | Costo Mensual Aprox. |
|---------|------|---------------------|
| ECS Fargate (6 tasks) | 0.5 vCPU, 1GB RAM | ~$30 |
| RDS db.t3.micro | PostgreSQL | ~$15 |
| Application Load Balancer (2) | - | ~$32 |
| S3 | 10 GB storage | ~$0.23 |
| SQS | 1M requests | ~$0.40 |
| CloudWatch Logs | 5 GB | ~$2.50 |
| **Total** | | **~$80/mes** |

---

## 10. Escalabilidad Futura

### 10.1 Recomendaciones
- Implementar cache con Amazon ElastiCache (Redis)
- Usar CloudFront para distribución de videos
- Implementar auto scaling basado en métricas personalizadas (ej: longitud de cola SQS)
- Migrar a contenedores más grandes (t3.medium) para mayor throughput
- Implementar multi-región para redundancia global

### 10.2 Límites Actuales
- Max 3 instancias por servicio (limitación AWS Academy)
- Max 9 instancias EC2 totales (limitación AWS Academy)
- Max 32 vCPUs totales (limitación AWS Academy)

---

## 11. Conclusiones

La arquitectura implementada cumple con todos los requisitos de escalabilidad y alta disponibilidad:
- Auto Scaling en todas las capas
- Load Balancing para distribución de carga
- Sistema de mensajería desacoplado (SQS)
- Multi-AZ para alta disponibilidad
- Monitoreo y alarmas configuradas

El sistema es capaz de escalar automáticamente según demanda y mantener disponibilidad ante fallos de zona.
