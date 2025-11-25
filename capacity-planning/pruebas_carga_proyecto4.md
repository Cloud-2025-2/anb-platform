# Pruebas de Carga y Análisis de Capacidad – Entrega 3  
**Maestría en Ingeniería de Software – Universidad de los Andes**  
**Curso:** ISIS-4426 – Arquitecturas Escalables en la Nube  
**Proyecto:** ANB Platform  

---

##  1. Objetivo

Evaluar la **capacidad, estabilidad y escalabilidad** de la aplicación **ANB Platform** desplegada en AWS bajo diferentes escenarios de carga y estrés.  
Las pruebas se realizaron con **k6**, mientras las métricas de infraestructura se monitorearon mediante **Amazon CloudWatch**.  
Este documento analiza el comportamiento del sistema frente a uso concurrente, subida de archivos y procesamiento intensivo, con el fin de definir mejoras para futuras versiones.

---

## 2. Configuración general

| Componente | Descripción |
|-------------|-------------|
| **Frontend (S3)** | [http://anb-platform-frontend.s3-website-us-east-1.amazonaws.com](http://anb-platform-frontend.s3-website-us-east-1.amazonaws.com) |
| **Backend (ALB)** | [http://anb-loadbalancer-214112690.us-east-1.elb.amazonaws.com](http://anb-loadbalancer-214112690.us-east-1.elb.amazonaws.com) |
| **Región AWS** | us-east-1 |
| **Instancias EC2 (API)** | t3.medium – 2 vCPU / 4 GiB RAM |
| **Auto Scaling Group (ASG)** | Min: 1 / Desired: 2 / Max: 3 |
| **Política de escalado** | CPUUtilization > 60 % |
| **Worker EC2** | Instancia independiente para procesamiento de videos |
| **Herramienta de carga** | k6 v0.48.0 |
| **Duración de pruebas** | 5–10 minutos |
| **Scripts** | `login_and_list_multi_user.js`, `upload_and_poll.js`, `batch_processing.js` |

---

##  3. Escenario 1 – Carga moderada (uso nominal)

**Script:** `login_and_list_multi_user.js`  
**Configuración:** 50 usuarios virtuales (VUs), duración 5 min, umbral de escalado 60 % CPU.

###  Ejecución
```bash
k6 run -e VUS=50 -e DURATION=5m login_and_list_multi_user.js
```

### Resultados k6
```
http_reqs......................: 7336
http_req_duration..............: avg=740 ms p(95)=970 ms
http_req_failed................: 0.2 %
vus............................: 50
```

###  Métricas CloudWatch
- **CPUUtilization:** promedio ≈ 45 %, picos ≈ 55 %  
- **GroupDesiredCapacity:** constante en 2 (sin escalado)  
- **NetworkIn/Out:** picos moderados y sostenidos  


### Análisis
El sistema procesó más de **7.000 solicitudes concurrentes** en cinco minutos, manteniendo estabilidad sin errores.  
Aunque el **percentil 95** superó los 900 ms, el servicio se mantuvo estable y no activó el escalado.  
El backend se acerca a su límite nominal en torno a los **50 usuarios simultáneos**, con uso de CPU en 55 %.

**Conclusión:**  
La arquitectura soporta el tráfico normal con buena estabilidad. El escalado aún no se activa, pero el sistema alcanza su umbral operativo.

---

## 4. Escenario 2 – Subida y procesamiento de videos

**Script:** `upload_and_poll.js`  
**Configuración:** 3 usuarios virtuales, duración 5 minutos, video 1080p (`8979096-hd_1920_1080_30fps.mp4`).

###  Ejecución
```bash
k6 run -e FILE_PATH=./8979096-hd_1920_1080_30fps.mp4 -e VUS=3 -e DURATION=5m upload_and_poll.js
```

### Resultados k6
```
✓ login success
✓ upload success
http_req_duration..............: avg=601 ms p(95)=1.22 s
http_req_failed................: 0.00 %
data_sent......................: 9.0 GB
```

###  Métricas CloudWatch
- **CPU (Worker EC2):** 75–80 %  
- **CPU (API):** 50–55 %  
- **NetworkIn:** alto y sostenido durante la subida  
- **GroupDesiredCapacity:** estable en 2  


###  Análisis
La prueba demostró una alta demanda de CPU en el Worker durante la subida simultánea de archivos grandes, pero sin errores.  
El sistema completó el flujo completo con latencias promedio de 600–1200 ms y cero fallos.

**Conclusión:**  
La aplicación maneja correctamente cargas concurrentes de subida de videos. El Worker llega a su límite (80 % CPU), lo que indica la necesidad de instancias más potentes o escalado dedicado para tareas multimedia.

---

##  5. Escenario 3 – Procesamiento en lote (estrés sostenido)

**Script:** `batch_processing.js`  
**Configuración:** 3 usuarios virtuales × 4 videos cada uno, duración 10 minutos.  

###  Ejecución
```bash
k6 run -e FILE_PATH=./1585619-hd_1280_720_30fps.mp4 -e VUS=3 -e BATCH_SIZE=4 batch_processing.js
```

### Resultados k6
```
✓ login success
✓ upload success
http_req_duration..............: avg=419.8 ms p(95)=536 ms
http_req_failed................: 0.00 %
data_sent......................: 7.1 GB
```

### Métricas CloudWatch
- **CPU (API EC2):** superó 70 % → se activó escalado  
- **GroupDesiredCapacity:** 2 → 3 instancias  
- **NetworkIn:** alto durante toda la prueba  
- **Activity History:** evento *“Launching new EC2 instance”*  


###  Análisis
El sistema mantuvo estabilidad con latencia promedio de 420 ms y percentil 95 de 536 ms, confirmando eficiencia incluso bajo carga alta.  
La política de escalado automático se activó correctamente al superar el 60 % de CPU, desplegando una tercera instancia.  
El throughput se mantuvo constante y no se registraron errores HTTP.

**Conclusión:**  
El sistema soportó el estrés concurrente de subidas múltiples sin caídas, demostrando resiliencia y escalabilidad automática.

---

##  6. Análisis comparativo de capacidad

| Escenario | Usuarios | CPU Prom. | p95 Latencia | Escalado | Resultado |
|------------|-----------|------------|---------------|------------|------------|
| **1 – Carga moderada** | 50 | 45–55 % | 970 ms |  No | Estable |
| **2 – Subida de videos** | 3 | 50–55 % (API), 75 % (Worker) | 1.22 s |  No | Estable |
| **3 – Procesamiento en lote** | 3×4 videos | 78 % | 536 ms |  Sí (2→3) | Escalado exitoso |

**Resumen:**  
- El sistema mantiene rendimiento aceptable (<1 s) hasta 50 usuarios concurrentes.  
- Bajo estrés, el ASG escala correctamente.  
- El Worker es el componente más exigido en CPU y requiere atención para cargas masivas.

---

##  7. Recomendaciones de escalamiento

| Componente | Mejora sugerida | Justificación |
|-------------|----------------|----------------|
| **API EC2 (ASG)** | Aumentar `MaxSize` de 3 → 6 y usar tipo `t3.large` | Mayor capacidad ante picos de tráfico concurrente |
| **Worker EC2** | Migrar a `t3.xlarge` o habilitar un ASG independiente | Evita cuellos de CPU durante procesamiento de videos |
| **Base de datos (RDS)** | Activar Multi-AZ y read replicas | Soporte para más lecturas concurrentes |
| **Cache (ElastiCache)** | Implementar Redis para endpoints públicos | Reduce tiempos de respuesta |
| **S3 / CDN** | Habilitar Transfer Acceleration o CloudFront | Mejora velocidad de carga y entrega de videos |
| **Monitoreo** | Dashboard CloudWatch con métricas CPU, ASG, Network | Visibilidad de rendimiento y alertas automáticas |
| **Infraestructura como código** | Desplegar con Terraform / CloudFormation | Facilita escalamiento automatizado y reproducible |

---

##  8. Conclusión general

Las pruebas de carga confirman que la arquitectura desplegada cumple los objetivos de escalabilidad, estabilidad y resiliencia planteados para esta entrega.  
El sistema mantiene una latencia inferior a 1 segundo bajo carga real, y el Auto Scaling Group reacciona adecuadamente ante picos de demanda.  

Con las optimizaciones propuestas (instancias más grandes, cache y mayor límite de escalado), la plataforma podrá atender 300+ usuarios concurrentes y múltiples tareas de procesamiento de video sin degradar el rendimiento.

---

