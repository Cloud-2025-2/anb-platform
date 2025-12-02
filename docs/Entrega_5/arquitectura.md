# Arquitectura de la solución anb-platform

## 1. Visión general

La solución `anb-platform` es una plataforma de procesamiento de videos desplegada en AWS que sigue una arquitectura típica de tres capas, con componentes adicionales para colas y almacenamiento de objetos:

- **Capa de red (VPC y subnets)**
  - 1 VPC dedicada (`10.0.0.0/16`) con subredes públicas y privadas.
  - Internet Gateway para exponer los balanceadores de carga.



- **Capa de cómputo (ECS Fargate)**
  - 1 clúster ECS Fargate con 3 servicios:
    - Backend API (Go)
    - Frontend (React + Nginx)
    - Worker de procesamiento de video

- **Capa de datos y mensajería**
  - RDS PostgreSQL para datos transaccionales.
  - S3 como almacenamiento de videos.
  - SQS como cola de trabajos de procesamiento, más una DLQ (dead-letter queue).

- **Observabilidad**
  - CloudWatch Logs para logs de cada servicio ECS.
  - CloudWatch Alarms sobre CPU de backend, frontend y worker.

Todo esto se crea con Terraform en `terraform/main.tf` y los módulos en `terraform/modules/*`, y las imágenes Docker se construyen y suben a ECR con los scripts en `scripts/*.sh`.

## 2. Capa de red (módulo `networking`)

**Archivos clave:** `terraform/modules/networking/main.tf`, `variables.tf`, `outputs.tf`

Se crea:

- **VPC principal**

```hcl
resource "aws_vpc" "main" { cidr_block = var.vpc_cidr }  # 10.0.0.0/16
```

- Con DNS habilitado (`enable_dns_hostnames` y `enable_dns_support`).
- **Internet Gateway** asociado a la VPC (`aws_internet_gateway.main`).
- **Subnets públicas** (`aws_subnet.public`) en 2 AZ (tomadas de `data.aws_availability_zones`), usadas por los ALB.
- **Subnets privadas** (`aws_subnet.private`) también en 2 AZ, pensadas para los servicios internos (ECS, RDS).
- **Route tables y asociaciones**:
  - Route table pública con ruta `0.0.0.0/0` → Internet Gateway.

En el diagrama conceptual:

- Un rectángulo grande: `VPC anb-platform-dev-vpc (10.0.0.0/16)`.
- Dentro, dos columnas de AZ (`us-east-1a`, `us-east-1b`) con:
  - Subnet pública en cada AZ.
  - Subnet privada en cada AZ.

## 3. Seguridad (módulo `security`)

**Archivo:** `terraform/modules/security/main.tf`

Se definen tres Security Groups principales:

- **SG del ALB** (`aws_security_group.alb`)
  - Ingress: HTTP 80 y HTTPS 443 desde `0.0.0.0/0`.
  - Egress: abierto (`0.0.0.0/0`).

- **SG de ECS** (`aws_security_group.ecs`)
  - Permite tráfico desde el SG del ALB hacia los puertos internos de las tareas ECS (backend/frontend).
  - Egress abierto para que las tareas salgan a RDS, S3, SQS, etc.

- **SG de RDS** (`aws_security_group.rds`)
  - Ingress al puerto `5432` (Postgres) desde el SG de ECS.
  - Egress abierto.

En el diagrama:

- Dibujar icono de RDS con un candado: `SG RDS: permite 5432 desde SG ECS`.
- Dibujar `SG ALB` y `SG ECS` entre Internet y el clúster ECS.

## 4. Capa de datos: RDS PostgreSQL (módulo `rds`)

**Archivo:** `terraform/modules/rds/main.tf`

Se crean:

- **Grupo de subredes de BD** (`aws_db_subnet_group.main`)
  - Usa subnets (conceptualmente el “DB Subnet Group” con 2 subnets).

- **Instancia RDS PostgreSQL** (`aws_db_instance.postgres`)
  - Engine: `postgres`.
  - Clase: `db.t3.micro`.
  - Storage: `20 GB gp2`.
  - Usuario, password y DB name vienen de variables (`db_username`, `db_password`, `db_name`).
  - Asociada al SG de RDS y al DB Subnet Group.
  - Backups habilitados (retention 7 días, ventana de mantenimiento).

**Uso desde la app:**

Terraform pasa los valores `db_host`, `db_name`, `db_username`, `db_password`, `db_port` al módulo de ECS como variables de entorno (`POSTGRES_HOST`, `POSTGRES_USER`, etc.) en las definiciones de tarea (`aws_ecs_task_definition.backend` y `worker`).

## 5. Almacenamiento de objetos: S3 (módulo `s3`)

**Archivo:** `terraform/modules/s3/main.tf`

Se crea el bucket:

- `anb-platform-video-storage-${account_id}` para almacenar videos.

Configuración relevante:

- Versioning habilitado (`aws_s3_bucket_versioning`).
- Lifecycle rules que eliminan versiones antiguas después de cierto tiempo (p. ej. 90 días para objetos activos y 30 días para versiones no actuales).

**Uso en backend/worker:**

El código en `backend/internal/storage/local.go` implementa una interfaz `Storage` que usa el SDK de S3 para subir y descargar archivos. Terraform inyecta en las tareas ECS el nombre del bucket (`s3_bucket_name`) como variable de entorno.

## 6. Cola de tareas: SQS (módulo `sqs`)

**Archivo:** `terraform/modules/sqs/main.tf`

Se crean:

- **Cola principal SQS** `anb-platform-video-processing-queue`:
  - `visibility_timeout_seconds = 300`
  - `message_retention_seconds = 86400` (1 día)
  - `receive_wait_time_seconds = 20` (long polling)

- **Dead-Letter Queue** `anb-platform-video-processing-dlq`:
  - retención 14 días.
  - Redrive policy: mensajes que fallen más de 3 veces (`maxReceiveCount = 3`) se mueven a la DLQ.

**Uso en el código:**

- `backend/internal/queue/sqs/producer.go` publica mensajes a la cola cuando se sube un video.
- `backend/internal/queue/sqs/consumer.go` es el consumidor (usado por el worker).
- El worker (`backend/cmd/worker/main.go`) escucha la cola y procesa los mensajes (invocando `internal/processing/video_processor.go`).

## 7. Capa de cómputo: ECS Fargate (módulo `ecs`)

**Archivo:** `terraform/modules/ecs/main.tf`

### 7.1 Clúster

`aws_ecs_cluster.main` llamado `${project_name}-${environment}-cluster`, con Container Insights activo para métricas.

### 7.2 Backend (API)

**Task Definition** `aws_ecs_task_definition.backend`:

- `network_mode = "awsvpc"` (cada task tiene su propia ENI).
- Fargate, `cpu = 512`, `memory = 1024`.
- Usa `var.backend_image_uri` (imagen en ECR construida con `scripts/build-and-push-images.sh`).
- Expone `containerPort 8000`.
- Variables de entorno:
  - `POSTGRES_HOST`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_PORT`.
  - `AWS_REGION`, `S3_BUCKET_NAME`, `SQS_QUEUE_URL`.
  - Variables de autenticación (`JWT_SECRET`, `JWT_EXPIRE_MINUTES`, etc.).
- Logs enviados a CloudWatch (`awslogs`).

**Service** `aws_ecs_service.backend`:

- Tipo Fargate (`awsvpc`).
- Corre en subnets privadas (`private_subnet_ids`).
- Asociado al SG de ECS.
- Registrado como target de un ALB backend (target group HTTP 80).

### 7.3 Frontend

**Task Definition** `aws_ecs_task_definition.frontend`:

- Imagen construida a partir de `frontend/Dockerfile` y subida a ECR como `anb-frontend`.
- Nginx sirve una SPA de React.
- Recibe por env var la URL del backend (`backend_alb_dns`) para hacer las llamadas API desde el navegador.

**Service** `aws_ecs_service.frontend`:

- Fargate/`awsvpc`.
- Conectado al ALB frontend, que es público.

### 7.4 Worker

**Task Definition** `aws_ecs_task_definition.worker`:

- Imagen `anb-video-processor` (construida desde `backend/Dockerfile.worker`).
- Variables de entorno para conectarse a SQS, Postgres y S3.
- No expone puertos externos.

**Service** `aws_ecs_service.worker`:

- Fargate en subnets privadas.
- Sin ALB; servicio interno que solo consume SQS.

### 7.5 Auto-escalado

En el módulo ECS hay recursos `aws_appautoscaling_target` y `aws_appautoscaling_policy` para:

- Escalar backend, frontend y worker entre un `min` y `max` de tareas (variables `backend_min_capacity`, `backend_max_capacity`, etc.).
- Política de escalado objetivo basada en CPU promedio (`ECSServiceAverageCPUUtilization`, `target_value = var.cpu_target_value`, por defecto 70%).

## 8. Balanceadores de carga: módulo `alb`

**Archivo:** `terraform/modules/alb/main.tf`

- **Frontend ALB** (`aws_lb.frontend`): Application Load Balancer público (`internal = false`), usa `alb_security_group_id` y las subnets públicas.
- **Listener** HTTP 80 (`aws_lb_listener.frontend`) → target group del frontend (contenedor del frontend).
- **Backend ALB** (`aws_lb.backend`): Application Load Balancer que maneja solicitudes HTTP 80 hacia el servicio backend.

Su DNS (`backend_alb_dns`) se pasa al frontend como variable de entorno para que la SPA apunte al backend correcto.

## 9. Observabilidad: CloudWatch

**Archivo:** `terraform/modules/cloudwatch/main.tf`

Se crean 3 CloudWatch Alarms:

- `backend_cpu_high`
- `frontend_cpu_high`
- `worker_cpu_high`

Todas consultan la métrica `AWS/ECS:CPUUtilization` filtrada por `ClusterName = module.ecs.cluster_name` y `ServiceName = <service_name>`.

Por ahora las `alarm_actions` están vacías (no disparan SNS), pero sirven para monitoreo y dashboard.

Además, en `ecs/main.tf` cada `task_definition` está configurada con `logConfiguration -> awslogs`, enviando logs a grupos como:

- `/ecs/anb-platform-dev-backend`
- `/ecs/anb-platform-dev-frontend`
- `/ecs/anb-platform-dev-worker`

## 10. Scripts de construcción y despliegue

En `scripts/` tienes:

- `build-and-push-images.sh` — Construye las imágenes Docker de backend, frontend y worker; crea repositorios en ECR (`anb-backend`, `anb-frontend`, `anb-video-processor`) si no existen; hace `docker build` y `docker push` a ECR.

Las URIs resultantes se usan luego como `backend_image_uri`, `frontend_image_uri`, `worker_image_uri` en Terraform.

Otros scripts (`kafka_instance_user_data.sh`, `webserver_user_data.sh`, `worker_user_data.sh`) son user-data para un despliegue alternativo basado en EC2 y Kafka; la arquitectura oficial del laboratorio usa ECS + SQS + S3 + RDS, por lo que estos scripts pueden mencionarse como variantes históricas o alternativas.

## 11. Flujo típico de la aplicación

1. El usuario entra desde Internet al Frontend ALB (HTTP/HTTPS).
2. El ALB enruta al servicio ECS de Frontend, que devuelve la SPA de React.
3. Desde el navegador, la SPA hace peticiones HTTP al Backend ALB (URL configurada por env var `backend_alb_dns`).
4. El servicio ECS de Backend recibe las peticiones:
   - Autentica usuarios (JWT).
   - Lee/escribe en RDS PostgreSQL para usuarios, videos, votos, etc.
5. Cuando el usuario sube un video:
   - Lo guarda en S3 (bucket `video_storage`).
   - Publica un mensaje en SQS `video-processing-queue` describiendo el trabajo.
6. El servicio ECS Worker (video-processor):
   - Lee mensajes de SQS.
   - Descarga el video desde S3.
   - Lo procesa (transcodificación / generación de thumbnails, etc.).
   - Vuelve a subir los resultados a S3 y actualiza el estado en PostgreSQL.
7. Si un mensaje falla repetidamente, SQS lo mueve a la DLQ, evitando bloqueos.

Todo el tráfico y errores quedan registrados en CloudWatch Logs, y las métricas de CPU disparan alarmas en CloudWatch Alarms si superan el umbral.

---

Si quieres, puedo también:

- Añadir un diagrama ASCII o mermaid para visualizar la VPC/ALB/ECS/RDS/S3/SQS.
- Insertar fragmentos específicos de Terraform del repo (`terraform/modules/*`) para referencia directa.
- Añadir instrucciones rápidas para desplegar (comandos Terraform y scripts de build).

Indícame qué prefieres y lo añado.
