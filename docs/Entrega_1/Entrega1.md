# Entrega 1: Implementación de una API REST escalable con orquestación de tareas asíncronas para el procesamiento de archivos

## Team

### Ivan Avila - 202216280
### Raul Insuasty - 202015512
### Ana María Sánchez Mejía - 202013587
### David Tobón Molina - 202123804

## Quickstart

## Quickstart

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (version 20.10 or higher)
- [Docker Compose](https://docs.docker.com/compose/install/) (version 2.0 or higher)

### Running the Project

1. **Clone the repository**
   ```bash
   git clone https://github.com/Cloud-2025-2/anb-platform.git
   cd anb-platform
   ```

   Or download the source code in a _.zip_ file.

2. **Start all services**
   ```bash
   docker-compose up -d
   ```

   This command will:
   - Build and start the Go API service
   - Start PostgreSQL database
   - Build and start the frontend with Nginx
   - Create necessary networks and volumes

3. **Check service status**
   ```bash
   docker-compose ps
   ```

4. **View logs (optional)**
   ```bash
   # View all services logs
   docker-compose logs -f
   
   # View specific service logs
   docker-compose logs -f api
   docker-compose logs -f db
   docker-compose logs -f frontend
   ```

### Accessing the Application

- **Frontend**: TODO
- **Check API Health**: http://localhost:8080/health

### Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (This will delete all data)
docker-compose down -v
```


## Tech-stack   

TODO

### health-check

```
http://localhost:8000/api/health
```

### OpenAPI Documentation

```
http://localhost:8000/swagger/index.html
```

### API Tests with Postman

```
npm install
npm run test
```

To generate an *.html* test report:

```
npm run test:report
```

Note: `Vote for Video` test takes around 20 seconds to run while newman waits for the video to be processed and made public.