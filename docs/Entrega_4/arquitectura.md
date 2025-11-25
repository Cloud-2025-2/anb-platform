# Entrega 4 – Escalabilidad en el Backend de ANB Platform

**Curso:** Arquitectura de Software en la Nube  
**Proyecto:** ANB Platform – Backend escalable para procesamiento de videos deportivos  
**Integrantes:** Ivan Avila - David tobon - Ana M. Sánchez

---

## 1. Introducción

En las entregas anteriores del proyecto ANB Platform se construyó un backend funcional para registrar usuarios, autenticar jugadores y entrenadores, cargar videos de entrenamientos y consultar rankings públicos. En la Entrega 3 se realizó un análisis de capacidad del backend con una arquitectura básica (un solo *task* de backend) y se identificaron los límites de rendimiento bajo diferentes patrones de uso.

La Entrega 4 tiene como objetivo **diseñar y desplegar una arquitectura escalable** que permita soportar picos de carga en el backend sin degradar la experiencia de los usuarios. En particular, se busca desacoplar el procesamiento pesado de videos, introducir *auto-scaling* en los servicios críticos y garantizar alta disponibilidad mediante despliegue multi-AZ.

En este documento se describe la arquitectura actualizada del sistema, se explican las decisiones de diseño y se muestra cómo la solución propuesta responde a los escenarios de capacidad definidos en el proyecto.

---

## 2. Contexto del proyecto y requisitos de capacidad

ANB Platform es una plataforma web para gestión y análisis de talento deportivo en baloncesto. Los principales casos de uso son:

- Registro y autenticación de jugadores/entrenadores.
- Carga de videos de jugadas y entrenamientos.
- Procesamiento de videos para extraer métricas técnicas.
- Consulta de videos y rankings públicos.

Desde el punto de vista de calidad, los requisitos relevantes para esta entrega son:

- **Capacidad y rendimiento del backend**
  - Soportar al menos _N_ usuarios concurrentes (medidos con k6) ejecutando flujos de login, listado de videos y consulta de rankings.
  - Mantener un **p95 de latencia de las peticiones HTTP < 500 ms** en los endpoints síncronos (`/api/auth/login`, `/api/videos`, `/api/public/videos`, `/api/public/rankings`).
- **Escalabilidad**
  - Permitir escalar horizontalmente los servicios de frontend, backend y workers de procesamiento de video en función de métricas de carga (CPU, número de peticiones o longitud de la cola).
- **Disponibilidad**
  - Tolerar la caída de una zona de disponibilidad sin dejar el sistema completamente fuera de servicio.
  - Evitar que las operaciones de procesamiento pesado bloqueen el camino crítico de las peticiones del usuario.

Sobre estos requisitos se definieron los escenarios de capacidad y se diseñó la arquitectura que se describe a continuación.

---

## 3. Arquitectura lógica

### 3.1 Componentes principales

A nivel lógico, la solución se compone de los siguientes componentes:

1. **Frontend Web**
   - Aplicación SPA que consume la API REST del backend.
   - Gestiona login, registro, subida de videos (vía API) y visualización de rankings.

2. **API Backend**
   - Servicio HTTP que expone endpoints bajo `/api/...`.
   - Responsabilidades:
     - Autenticación y emisión de *access tokens*.
     - Gestión de usuarios y perfiles.
     - Registro de metadatos de videos y consulta de rankings.
     - Generación de trabajos asíncronos de procesamiento de video y envío a la cola.

3. **Workers de procesamiento de video**
   - Procesos *background* que consumen mensajes de una cola (por ejemplo, Amazon SQS).
   - Se encargan de:
     - Descargar el video desde S3.
     - Ejecutar el pipeline de procesamiento (por ejemplo, servicios de análisis de video).
     - Actualizar los resultados en la base de datos.

4. **Base de datos transaccional (RDS)**
   - Motor relacional (PostgreSQL/MySQL).
   - Almacena usuarios, tokens, metadatos de videos, estadísticas y rankings.

5. **Almacenamiento de objetos (S3)**
   - Guarda los archivos de video originales y, opcionalmente, derivados procesados.

6. **Cola de mensajes (SQS)**
   - Desacopla las peticiones HTTP de los trabajos de procesamiento intensivo.
   - Permite *buffering* y escalamiento independiente de los workers.

### 3.2 Interacciones clave

1. **Login**
   - El usuario envía credenciales desde el frontend al API Backend.
   - El backend valida contra la base de datos y, si es correcto, retorna un token JWT.
2. **Listado de videos y rankings**
   - El frontend llama a los endpoints públicos del backend.
   - El backend lee la información desde la base de datos y responde de forma síncrona.
3. **Carga de video**
   - El usuario inicia la carga desde el frontend.
   - El backend registra los metadatos del video, sube o autoriza la subida a S3 y envía un mensaje a la cola para su procesamiento asíncrono.
   - Un worker recoge el mensaje, procesa el video y actualiza los resultados.

Con esta separación se evita que el procesamiento pesado de video bloquee el tiempo de respuesta de las peticiones que se miden en las pruebas de carga.

---

## 4. Arquitectura de despliegue en AWS

La **Figura 1** (![alt text](<DiagramaEntrega4(1) (1).jpg>)) muestra la arquitectura de despliegue propuesta para la Entrega 4.

### 4.1 VPC y zonas de disponibilidad

- Se define una **VPC** única que contiene:
  - Dos **Availability Zones (AZ1 y AZ2)** para lograr alta disponibilidad.
  - En cada AZ:
    - Una **public subnet** que expone los *load balancers*.
    - Una o más **private subnets** donde corren las tareas de frontend, backend y worker, así como la base de datos.

Esta división permite aislar los servicios de aplicación y la base de datos del acceso directo desde Internet.

### 4.2 Balanceadores de carga

1. **Frontend ALB (público)**
   - Se despliega un Application Load Balancer accesible desde Internet.
   - Recibe el tráfico HTTP/HTTPS de los navegadores y lo enruta a las tareas de frontend en las *private subnets*.
   - Incluye *health checks* para retirar instancias no saludables del balanceo.

2. **Backend ALB (interno)**
   - Es un ALB interno (no expuesto directamente a Internet).
   - Solo acepta tráfico proveniente de las tareas de frontend.
   - Reparte las peticiones hacia las tareas del API Backend.

Con esta doble capa de ALBs se separa claramente la capa de presentación de la capa de negocio, permitiendo políticas de seguridad y escalamiento diferenciadas.

### 4.3 Auto Scaling de tareas de aplicación

En cada AZ se define un **Auto Scaling Group (o servicio ECS escalable)** con tres tipos de tareas:

- **Task Frontend**
  - Ejecuta la aplicación web.
  - Se escala principalmente por métricas de uso de CPU, memoria y/o número de peticiones en el Frontend ALB.

- **Task Backend**
  - Ejecuta el API REST principal.
  - Se escala por:
    - Utilización de CPU/memoria.
    - Número de peticiones por segundo o latencia del Backend ALB.

- **Task Worker**
  - Consume mensajes de la cola SQS y procesa videos.
  - Se escala de forma independiente usando métricas relacionadas con la cola:
    - Cantidad de mensajes en cola.
    - Edad máxima de los mensajes.

Esta organización permite **escalar de forma independiente** la cantidad de recursos dedicados a servir peticiones HTTP y a procesar videos, optimizando costos.

### 4.4 Base de datos y redundancia

- Se utiliza **Amazon RDS en configuración Multi-AZ**:
  - Una instancia primaria en una AZ y una réplica en la otra.
  - En caso de fallo de una AZ o de la instancia primaria, RDS puede hacer *failover* transparente a la réplica.
- Las instancias de backend se conectan a RDS a través de un endpoint interno accesible solo desde las *private subnets*.

### 4.5 Otros servicios de infraestructura

- **Amazon S3**: bucket compartido para carga y almacenamiento de videos.
- **Sistema de colas (SQS)**: cola principal `video-processing-queue` usada por el backend (productor) y los workers (consumidores).
- **CloudWatch / CloudWatch Logs**:
  - Métricas de CPU, memoria, número de peticiones y errores HTTP.
  - Alarmas que disparan políticas de auto-scaling.
- **Logs de aplicación** centralizados (por ejemplo, en CloudWatch Logs) para depuración durante las pruebas de carga.

---

## 5. Estrategias de escalabilidad y relación con los escenarios

### 5.1 Separación de camino crítico y procesamiento pesado

En la versión original del sistema el servidor backend se encargaba tanto de responder a las peticiones HTTP como de procesar los videos, generando cuellos de botella cuando aumentaba el número de usuarios.

En esta entrega:

- El **camino crítico** de las peticiones medidas en k6 (login, listado de videos, rankings) solo incluye:
  - Frontend ALB → Task Frontend → Backend ALB → Task Backend → RDS.
- El **procesamiento de video** se ejecuta fuera del camino crítico:
  - Task Backend → cola SQS → Task Worker → S3/RDS.

Esto permite mantener baja latencia en las peticiones síncronas aun cuando haya muchos videos en proceso.

### 5.2 Escalamiento horizontal del backend

- Tanto las tareas de **frontend** como de **backend** se despliegan en **múltiples instancias en ambas AZ**.
- El número de tareas mínimas y máximas se configura de acuerdo con el análisis de capacidad de la Entrega 3.
- Se definen políticas de Auto Scaling basadas en:
  - CPU promedio > X% durante Y minutos.
  - Latencia p95 del ALB > 500 ms durante Y minutos.
- Ante un pico de carga, se crean nuevas tareas de backend que se registran automáticamente en el Backend ALB.

### 5.3 Escalabilidad de workers de procesamiento

- Los **Task Worker** se escalan a partir de métricas de la cola:
  - Cuando la cola supera un umbral de mensajes pendientes o la edad de los mensajes crece, se incrementa el número de workers.
  - Cuando la cola está casi vacía durante cierto tiempo, el número de workers se reduce para ahorrar costos.
- Con esto se desacopla el crecimiento de la carga de procesamiento respecto al tráfico HTTP.

### 5.4 Escalabilidad de la base de datos

- RDS en Multi-AZ proporciona capacidad de cómputo y almacenamiento predeterminados.
- En caso de necesitar mayor capacidad:
  - Se puede escalar verticalmente el *instance class* (más CPU/memoria).
  - A futuro se podría considerar **read replicas** para distribuir carga de lectura si los escenarios lo requieren.

---

## 6. Disponibilidad y tolerancia a fallos

La arquitectura incorpora varias tácticas para mejorar disponibilidad:

1. **Redundancia por zonas (Multi-AZ)**
   - Servicios de frontend, backend y workers se despliegan en al menos dos AZ.  
   - Si una AZ falla, el ALB puede seguir enviando tráfico a las instancias sanas de la otra.

2. **Health checks y reemplazo automático**
   - Los ALBs ejecutan *health checks* periódicos.
   - Las instancias que fallan son removidas del balanceo y el Auto Scaling Group las reemplaza automáticamente.

3. **RDS Multi-AZ**
   - Mantiene una réplica síncrona en otra AZ.
   - El *failover* reduce el tiempo de indisponibilidad de la base de datos.

4. **Cola de mensajes como *buffer***
   - Si los workers se caen o se reduce temporalmente su capacidad, la cola retiene los mensajes hasta que se restablezca el servicio.
   - Esto evita pérdida de trabajos de procesamiento.

5. **Desacoplamiento por capas**
   - La separación entre frontend, backend y workers reduce la posibilidad de que un fallo en una capa tumbe el resto del sistema.

---

## 7. Trazabilidad con los escenarios de calidad

A continuación se resume cómo la arquitectura responde a los escenarios de capacidad definidos:

- **Escenario 1 – Múltiples usuarios listan videos y rankings en hora pico.**  
  - Los ALBs distribuyen las peticiones entre varias instancias de backend.  
  - El auto-scaling aumenta el número de tareas cuando la carga crece, manteniendo el p95 de latencia por debajo de 500 ms.

- **Escenario 2 – Pico de carga en procesamiento de videos, manteniendo capacidad de respuesta en login y consultas.**  
  - El procesamiento intensivo se delega a workers asíncronos a través de SQS.  
  - El backend solo encola trabajos y responde rápido al usuario, por lo que las pruebas k6 sobre `/login` y `/videos` no se ven afectadas.

- **Escenario 3 – Falla parcial de infraestructura (una AZ).**  
  - Gracias al despliegue Multi-AZ de ALBs, tareas de aplicación y RDS, el sistema conserva capacidad de respuesta usando la AZ restante.  
  - La capacidad efectiva se reduce pero se mantiene la disponibilidad del servicio.

Estos puntos se complementarán con los resultados concretos de las pruebas de carga en el documento separado de “Escenarios y resultados de pruebas de estrés”.

---

## 8. Riesgos, limitaciones y trabajo futuro

Aunque la arquitectura actual mejora significativamente la capacidad y disponibilidad del backend, aún existen riesgos y líneas de trabajo futuro:

- **Cuellos de botella en la base de datos.**  
  Si la carga de lectura/escritura crece por encima de lo previsto, RDS puede convertirse en el nuevo límite. Como trabajo futuro se propone:
  - Introducir *caching* de lecturas frecuentes (por ejemplo, Redis para rankings).
  - Evaluar la creación de *read replicas* o particionamiento lógico.

- **Configuración fina de políticas de auto-scaling.**  
  Los umbrales configurados pueden no ser óptimos para todos los patrones de uso. Será necesario ajustar:
  - Métricas objetivo (CPU, latencia, longitud de cola).
  - Pasos mínimos/máximos de cambio en el número de instancias.

- **Observabilidad.**  
  Aunque se utilizan logs y métricas básicas, sería deseable:
  - Agregar *dashboards* dedicados a capacidad (p95, tasa de errores, consumo de cola).
  - Incluir *tracing* distribuido para entender mejor los tiempos de respuesta extremo a extremo.

- **Costos.**  
  El uso de múltiples AZ, ALBs y Auto Scaling Groups aumenta el costo operativo. Se requiere un monitoreo continuo del gasto en AWS y posibles ajustes de tamaño de instancias y límites de escalamiento.

---

## 9. Conclusiones

La arquitectura de la Entrega 4 transforma el backend de ANB Platform en un sistema **elástico y altamente disponible**, preparado para soportar picos de carga sin afectar significativamente la experiencia de los usuarios.

Los cambios principales frente a la entrega anterior son:

- Introducción de **ALBs separados para frontend y backend**.
- Despliegue multi-AZ con **Auto Scaling Groups** de frontend, backend y workers.
- Desacoplamiento del procesamiento de video mediante **colas de mensajes** y workers asíncronos.
- Uso de **RDS Multi-AZ** como repositorio transaccional resiliente.

En conjunto, estas decisiones arquitectónicas alinean el proyecto con buenas prácticas de aplicaciones nativas de la nube y sientan las bases para los experimentos de carga y escalabilidad documentados en el informe de pruebas de estrés.
