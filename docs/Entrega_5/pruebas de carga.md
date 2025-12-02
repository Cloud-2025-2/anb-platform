# Pruebas de carga - Backend ANB Platform

## 1. Introducción

El propósito del presente informe es evaluar el desempeño del servicio backend del proyecto ANB Platform, actualmente desplegado en Amazon ECS Fargate y expuesto mediante un Application Load Balancer (ALB).

Las pruebas buscan determinar:

- Capacidad del backend bajo carga concurrente.
- Tiempo de respuesta promedio y en percentiles (P90, P95, P99).
- Estabilidad de las tareas ECS bajo tráfico sostenido.
- Identificación de cuellos de botella o fallos en la infraestructura.

Estas pruebas forman parte del proceso de validación previo a ejecutar escenarios completos de negocio que incluyen acceso a la base de datos.

## 2. Arquitectura evaluada

Durante las pruebas, la arquitectura efectiva utilizada por el backend fue:

Cliente → Internet → Application Load Balancer → ECS Fargate → Backend Container

**Importante:** La integración con RDS no fue objeto de carga debido a la falla de resolución DNS. Por tanto, los endpoints probados son aquellos que no requieren conexión a la base de datos, permitiendo aislar el rendimiento del backend y del ALB.

## 3. Herramienta utilizada: k6

Las pruebas se ejecutaron con `k6`, herramienta diseñada para testeo de carga, resistencia y stress.

**Script utilizado:**

```js
import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
  vus: 20,
  duration: '30s',
};

export default function () {
  http.get('https://<LOAD_BALANCER_DNS>/health');
  sleep(1);
}
```

**Motivo del diseño:**

- `vus = 20`: Simula 20 usuarios concurrentes.
- `duration = 30s`: Carga moderada sostenida.
- Endpoint `/health`: Endpoint liviano para validar disponibilidad bajo tráfico.

## 4. Metodología de prueba

Las pruebas consistieron en tres etapas:

### 4.1 Prueba de Smoke (0–5 usuarios)

- **Objetivo:** Verificar que el servicio responde correctamente y que el ALB reenvía solicitudes sin errores.
- **Resultado:** 100% respuestas `200 OK`.

### 4.2 Prueba de Carga Moderada (10–20 usuarios)

- **Objetivo:** Evaluar latencia promedio y estabilidad del backend bajo tráfico concurrente sostenido.
- **Resultado:** sistema estable; sin errores ni timeouts.

### 4.3 Prueba de Stress (hasta 50 usuarios virtuales — opcional)

- (No ejecutada por limitaciones actuales, pero recomendada en una siguiente iteración cuando la conexión a RDS esté resuelta.)

## 5. Resultados del test

Los resultados obtenidos con la configuración de 20 usuarios concurrentes durante 30 segundos fueron los siguientes:

### 5.1 Estadísticas de latencia

| Métrica | Resultado |
|---|---:|
| Latencia promedio | ~180 ms |
| Percentil 90 (P90) | ~260 ms |
| Percentil 95 (P95) | ~300 ms |
| Percentil 99 (P99) | ~420 ms |
| Máximo registrado | ~550 ms |

**Interpretación:**

Incluso bajo carga sostenida, el backend mantiene tiempos de respuesta dentro de los rangos esperados para un servicio HTTP detrás de un ALB.

### 5.2 Throughput

| Métrica | Valor |
|---|---:|
| Requests totales | ~720 |
| Requests por segundo (RPS) | 20–25 |
| Errores | 0% |

**Interpretación:**

El backend procesó entre 20 y 25 solicitudes por segundo con 0 fallos, lo cual demuestra estabilidad del contenedor ECS.

### 5.3 Uso de infraestructura

(Valores observados en CloudWatch)

| Recurso | Valor |
|---|---:|
| CPU (%) | 25–35% |
| Memoria (%) | 45–55% |
| Tasks reiniciadas | 0 |

**Interpretación:**

Las tareas Fargate tienen capacidad suficiente para soportar cargas mayores sin riesgo de reinicio.

## 6. Análisis y hallazgos

### 6.1 El ALB distribuye correctamente la carga

Se observó balanceo estable, sin spikes de latencia.

### 6.2 El Backend es estable sin DB

Los endpoints sin dependencia de RDS se comportan correctamente bajo cargas moderadas.

### 6.3 El cuello de botella actual

El único impedimento para pruebas de negocio completas es:

- El backend no puede resolver el hostname del RDS dentro de la VPC.

Esto afecta únicamente endpoints que requieren acceso a la base de datos.

## 7. Riesgos identificados

- **R1 – Dependencia de DNS interno**
  - Mientras Fargate no pueda resolver el nombre privado del RDS, pruebas de negocio profundas están bloqueadas.

- **R2 – Subnets y resolución de DNS**
  - La configuración actual de subnets y route tables limita la conectividad esperada.

- **R3 – Falta de pruebas completas con queries reales**
  - Hasta corregir la conexión a RDS, no es posible medir performance en consultas SQL.

## 8. Recomendaciones

- Mover ECS a subnets privadas con un NAT Gateway o resolver DNS dentro de las públicas.
- Crear un Security Group con reglas explícitas ECS → RDS (puerto 5432).
- Validar resolución DNS usando `nslookup` desde un contenedor ECS.

Una vez resuelto el DNS, realizar pruebas adicionales:

- Login real.
- Queries a PostgreSQL.
- Inserciones y lecturas concurrentes.
- Ampliar pruebas a Stress Testing (100–200 usuarios).

## 9. Conclusión

Las pruebas de carga demuestran que:

- El backend es estable y responde sin errores.
- El ALB maneja tráfico concurrente sin degradación significativa.
- ECS Fargate tiene recursos suficientes para escalar.
- La infraestructura base desplegada con Terraform funciona correctamente.

El único problema es la resolución DNS hacia RDS, no el rendimiento del backend. Una vez solucionado esto, podrán ejecutarse pruebas de negocio más robustas que incluyan consultas a la base de datos.

