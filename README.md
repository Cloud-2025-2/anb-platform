## Integrantes
* Ivan Avila - i.avilag@gmail.com
* Ana M. Sánchez - am.sanchezm1@uniandes.edu.co
* David Tobón Molina - d.tobonm2@uniandes.edu.co

# Entregas

## Entrega 1

* [Entrega1.md](docs/Entrega_1/Entrega1.md)

* [Video de sustentación](https://youtu.be/R29sdc5Pr-8)


## Entrega 2

* [Entrega2.md](docs/Entrega_2/Entrega2.md)

* Video de sustentación: TBD


# Deployment

## Local Development
See [Entrega2.md](docs/Entrega_2/Entrega2.md) for local Docker Compose setup.

## AWS Production Deployment
For deploying to AWS with 3 EC2 instances (Webserver, Workers, NFS):

### Automated Deployment (Recommended)
* **User Data Deployment**: [USER_DATA_DEPLOYMENT_GUIDE.md](USER_DATA_DEPLOYMENT_GUIDE.md) - Fully automated, zero-config setup
  - Just paste scripts in EC2 user data and launch
  - Instances configure themselves automatically
  - ~20 minutes to fully operational

### Manual Deployment
* **Quick Start**: [AWS_DEPLOYMENT_QUICK_START.md](AWS_DEPLOYMENT_QUICK_START.md)
* **Full Guide**: [docs/AWS_DEPLOYMENT.md](docs/AWS_DEPLOYMENT.md)

### Architecture
- **EC2 #1**: Nginx, Backend API, Frontend, PostgreSQL, Redis, Kafka
- **EC2 #2**: Video processing workers (scalable)
- **EC2 #3**: NFS shared storage

# SonarQube
https://sonarcloud.io/project/overview?id=Cloud-2025-2_anb-platform
