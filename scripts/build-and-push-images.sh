#!/bin/bash

set -e



# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Verificar AWS credentials
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo -e "${RED} No hay credenciales de AWS configuradas${NC}"
    echo "   Ejecuta: ./scripts/configure-aws.sh"
    exit 1
fi

# Variables
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
REGION="us-east-1"
PROJECT_NAME="anb-platform"

echo -e "${GREEN}Account ID:${NC} $ACCOUNT_ID"
echo -e "${GREEN}Region:${NC} $REGION"
echo ""

# Crear repositorios ECR si no existen
echo -e "${BLUE} Creando repositorios en ECR...${NC}"

for REPO in anb-backend anb-frontend anb-video-processor; do
    if aws ecr describe-repositories --repository-names $REPO --region $REGION > /dev/null 2>&1; then
        echo "   Repositorio $REPO ya existe"
    else
        echo "   Creando repositorio $REPO..."
        aws ecr create-repository \
            --repository-name $REPO \
            --region $REGION \
            --image-scanning-configuration scanOnPush=true \
            --encryption-configuration encryptionType=AES256 > /dev/null
        echo "   Repositorio $REPO creado"
    fi
done

echo ""

# Login a ECR
echo -e "${BLUE} Autenticando con ECR...${NC}"
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com
echo ""

# Construir y subir Backend
echo -e "${BLUE}  Construyendo imagen Backend...${NC}"
cd backend
docker build -t anb-backend:latest .
docker tag anb-backend:latest ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-backend:latest
echo -e "${BLUE}  Subiendo imagen Backend a ECR...${NC}"
docker push ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-backend:latest
echo -e "${GREEN} Backend subido${NC}"
cd ..
echo ""

# Construir y subir Frontend
echo -e "${BLUE} Construyendo imagen Frontend...${NC}"
cd frontend
docker build -t anb-frontend:latest .
docker tag anb-frontend:latest ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-frontend:latest
echo -e "${BLUE}  Subiendo imagen Frontend a ECR...${NC}"
docker push ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-frontend:latest
echo -e "${GREEN} Frontend subido${NC}"
cd ..
echo ""

# Construir y subir Worker
echo -e "${BLUE}  Construyendo imagen Worker...${NC}"
cd backend
docker build -t anb-video-processor:latest -f Dockerfile.worker .
docker tag anb-video-processor:latest ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-video-processor:latest
echo -e "${BLUE}  Subiendo imagen Worker a ECR...${NC}"
docker push ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-video-processor:latest
echo -e "${GREEN} Worker subido${NC}"
cd ..
echo ""

# Actualizar terraform.tfvars automáticamente
echo -e "${BLUE} Actualizando terraform.tfvars...${NC}"

cat > terraform/terraform.tfvars << TFVARS
# AWS Configuration
aws_region   = "us-east-1"
project_name = "anb-platform"
environment  = "dev"

# VPC Configuration
vpc_cidr = "10.0.0.0/16"

# Database Configuration
db_username = "postgres"
db_password = "MySecurePassword123!"  # CAMBIAR
db_name     = "anb_platform"

# ECR Image URIs (auto-generado)
backend_image_uri  = "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-backend:latest"
frontend_image_uri = "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-frontend:latest"
worker_image_uri   = "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/anb-video-processor:latest"

# Auto Scaling Configuration
backend_min_capacity  = 1
backend_max_capacity  = 3
frontend_min_capacity = 1
frontend_max_capacity = 3
worker_min_capacity   = 1
worker_max_capacity   = 3
cpu_target_value      = 70
TFVARS

echo -e "${GREEN} terraform.tfvars actualizado${NC}"
echo ""


echo "   1. Edita terraform/terraform.tfvars y cambia db_password"
echo "   2. cd terraform"
echo "   3. terraform init"
echo "   4. terraform plan"
echo "   5. terraform apply"
