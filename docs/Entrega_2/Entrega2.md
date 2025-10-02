# Entrega 2 Despliegue básico en la nube migración de una aplicación web en la nube pública

## Team

* Ivan Avila - i.avilag@gmail.com
* Ana M. Sánchez - am.sanchezm1@uniandes.edu.co
* David Tobón Molina - d.tobonm2@uniandes.edu.co

## Resumen ejecutivo

Se desplegó la plataforma ANB en AWS con separación de responsabilidades en tres instancias EC2 (Web, Worker, File Server/NFS) y una base de datos administrada en Amazon RDS (PostgreSQL). La red reside en una VPC dedicada con subred pública para EC2 y subredes privadas para la BD mediante un DB Subnet Group. El almacenamiento compartido se implementó con NFS en una EC2 (por lineamiento no se usa EFS). Se aplicó endurecimiento de Security Groups y se documentan flujos, puertos y evidencias.


## Topología de red (AWS)

-VPC: anb-vpc-vpc — vpc-0fdbfd1b3f0b091d6 (CIDR 10.0.0.0/16)

-Subred pública: subnet-0f8c9b58f0c1558dc (10.0.1.0/24 · us-east-1a)

-DB Subnet Group (privadas): anb-db-subnet-group (AZ us-east-1b)

-Subnets: subnet-0f64f40a8b3719f2e, subnet-0abc25dcfb3c3b440

-Internet Gateway (IGW): igw-01b5b241146d0aae3


### Tabla de recursos

| Recurso | Detalle | ID/Nombre |
|---|---|---|
| VPC | 10.0.0.0/16 | vpc-0fdbfd1b3f0b091d6 |
| Subred pública | 10.0.1.0/24 (us-east-1a) | subnet-0f8c9b58f0c1558dc |
| DB Subnet Group | privadas (us-east-1b) | anb-db-subnet-group |
| DB Subnets | privadas | subnet-0f64f40a8b3719f2e, subnet-0abc25dcfb3c3b440 |
| IGW | Internet Gateway | igw-01b5b241146d0aae3 |
| EC2 Web | t3.small (AL2023) | i-0dd43473abb183009 — 10.0.1.251 / 13.222.125.199 |
| EC2 NFS | t3.small (AL2023) | i-0c701f7f0a3cef980 — 10.0.1.120 / 50.16.62.134 |
| EC2 Worker | t3.small (AL2023) | i-0e6235950d0553984 — 10.0.1.197 / 3.89.60.26 |
| RDS PostgreSQL | db.t3.small (us-east-1b) | anbdb |
| Endpoint RDS | puerto 5432 | anbdb.c1om664e4sm8.us-east-1.rds.amazonaws.com |
| AMI EC2 | Amazon Linux 2023 | ami-08982f1c5bf93d976 |



### Ruteo


- **RT pública** `rtb-<public>` (asociada a `subnet-0f8c9b58f0c1558dc`)
  - `10.0.0.0/16 → local`
  - `0.0.0.0/0 → igw-<id>`
- **RT privadas (DB)** `rtb-<private>` (asociadas al **DB Subnet Group**)
  - `10.0.0.0/16 → local`
  - _(sin salida a Internet; si se requiere, usar NAT)_

---

## Componentes lógicos y responsabilidades
- **Web (EC2):** expone HTTP/HTTPS, monta `/mnt/shared` vía NFS y conecta a RDS.
- **Worker (EC2):** procesa tareas asíncronas, monta `/mnt/shared` y conecta a RDS.
- **File Server NFS (EC2):** exporta `/srv/files` por NFSv4 a Web/Worker.
- **RDS (PostgreSQL):** persistencia relacional en subred privada; acceso limitado por SG.

---

## Seguridad (Security Groups y puertos)

| SG | Regla Inbound | Origen | Comentario |
|---|---|---|---|
| **SG-web** `sg-025c9d0efa425de88` | 80/TCP | 0.0.0.0/0 | HTTP público |
|  | 443/TCP | 0.0.0.0/0 | HTTPS público |
|  | 22/TCP | TU_IP/32 | SSH admin (recomendado) |
| **SG-worker** `sg-0052549e294bd7fed` | 22/TCP | TU_IP/32 | SSH admin |
| **SG-nfs** `sg-08c955a1ff18e0abb` | 2049/TCP | SG-web, SG-worker | NFSv4 |
| **SG-rds** `sg-0dda5df77faf6144f` | 5432/TCP | SG-web, SG-worker | PostgreSQL |

- **Outbound:** All traffic _(o mínimo: Web/Worker → 2049 a SG-nfs y 5432 a SG-rds)._  
- **Endurecimiento aplicado:** eliminar puertos no usados; SSH siempre restringido a IP de admin.

---

## Almacenamiento compartido (NFS sobre EC2)

**Servidor (anb-nfs — 10.0.1.120):**
```bash
sudo dnf -y install nfs-utils
sudo mkdir -p /srv/files && sudo chown ec2-user:ec2-user /srv/files
echo "/srv/files 10.0.1.0/24(rw,sync,no_root_squash,no_subtree_check)" | sudo tee -a /etc/exports
sudo systemctl enable --now nfs-server
sudo exportfs -rav


---

### Clientes (anb-web y anb-worker)
```bash
sudo dnf -y install nfs-utils
sudo mkdir -p /mnt/shared
echo "10.0.1.120:/srv/files /mnt/shared nfs4 defaults 0 0" | sudo tee -a /etc/fstab
sudo mount -a
```

---

### Smoke
```bash
# En Web/Users/hernandosanchez/Downloads/ANB_AWS_Despliegue.md
echo ok | sudo tee /mnt/shared/test.txt

# En Worker
cat /mnt/shared/test.txt   # (debe existir)
```

---

## Base de datos (RDS PostgreSQL)

**Instancia:** `anbdb` (db.t3.small, AZ us-east-1b)  
**Endpoint:** `anbdb.c1om664e4sm8.us-east-1.rds.amazonaws.com:5432`  
**Acceso:** solo desde **SG-web** y **SG-worker**.

#### Prueba de conectividad (desde EC2)
```bash
sudo dnf -y install postgresql15
psql "host=anbdb.c1om664e4sm8.us-east-1.rds.amazonaws.com port=5432 dbname=<DB> user=<USER> password=<PASS> sslmode=require" -c '\conninfo'
```

#### Variables de entorno (apps)
```ini
DB_HOST=anbdb.c1om664e4sm8.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_NAME=<tu_db>
DB_USER=<tu_user>
DB_PASSWORD=<tu_pass>
FILES_DIR=/mnt/shared
```

## Aprovisionamiento (user-data de referencia)

**Web/Worker (montaje NFS + cliente psql):**
```bash
#!/bin/bash
set -euo pipefail
dnf -y update
dnf -y install nfs-utils postgresql15
mkdir -p /mnt/shared
echo "10.0.1.120:/srv/files /mnt/shared nfs4 defaults 0 0" >> /etc/fstab
mount -a
# aquí: instalar runtime/app (nginx/node/python/go) y cargar .env
```

## Observabilidad, operación y costos

- **Métricas:** CPU, RAM, red y disco por instancia (Monitoring de EC2) + Performance Insights en RDS si aplica.  
- **Logs:** sistema y aplicación (ideal: CloudWatch Logs).  
- **Backups:** snapshots automáticos de RDS (retención corta para dev).  
- **Costos:** detener EC2 fuera de uso y eliminar RDS al finalizar para evitar cargos.


## Riesgos y mitigaciones

| Riesgo                | Impacto        | Mitigación                                                                 |
|-----------------------|----------------|-----------------------------------------------------------------------------|
| Exposición de puertos | Seguridad      | SG de mínimo privilegio; SSH solo desde IP admin                           |
| Cuello de botella NFS | Rendimiento    | Monitorear p95/p99; evaluar gp3 con más IOPS o separar server              |
| Single-AZ             | Disponibilidad | Aceptado por lineamiento; documentado como trade-off                       |
| Credenciales en host  | Seguridad      | Variables de entorno/secret store; rotación post-demo                      |



