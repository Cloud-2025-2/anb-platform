#!/bin/bash


# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW} Necesitas las credenciales de AWS Academy Lab${NC}"
echo "   1. Ve a AWS Academy"
echo "   2. Click en 'AWS Details'"
echo "   3. Copia las credenciales"
echo ""

# Pedir credenciales
read -p "AWS Access Key ID: " AWS_ACCESS_KEY
echo ""
read -sp "AWS Secret Access Key: " AWS_SECRET_KEY
echo ""
read -sp "AWS Session Token: " AWS_SESSION_TOKEN
echo ""
echo ""

# Exportar variables
export AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$AWS_SECRET_KEY"
export AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN"
export AWS_DEFAULT_REGION="us-east-1"

# Guardar en archivo temporal (NO commitear)
cat > ~/.aws-temp-credentials << CREDS
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_KEY}"
export AWS_SESSION_TOKEN="${AWS_SESSION_TOKEN}"
export AWS_DEFAULT_REGION="us-east-1"
CREDS

echo -e "${GREEN} Credenciales configuradas${NC}"
echo ""

# Verificar conexión
echo " Verificando conexión a AWS..."
if aws sts get-caller-identity > /dev/null 2>&1; then
    echo -e "${GREEN}Conexión exitosa a AWS${NC}"
    echo ""
    aws sts get-caller-identity
    echo ""
    echo -e "${GREEN}Account ID:${NC} $(aws sts get-caller-identity --query Account --output text)"
    echo -e "${GREEN}Region:${NC} $AWS_DEFAULT_REGION"
    echo ""
    echo -e "${YELLOW} Para usar estas credenciales en otras terminales, ejecuta:${NC}"
    echo "   source ~/.aws-temp-credentials"
else
    echo -e "${RED} Error de conexión a AWS${NC}"
    echo "   Verifica que las credenciales sean correctas"
    exit 1
fi
