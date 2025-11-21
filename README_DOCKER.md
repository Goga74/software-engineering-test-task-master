# Docker Setup Guide

## Overview
This guide explains how to build and run the application using Docker with an optimized multi-stage build process.

## Docker Image Optimization

The Dockerfile uses multi-stage build to minimize the final image size:
- **Build stage**: Uses \golang:1.25-alpine\ to compile the application
- **Run stage**: Uses minimal \alpine:latest\ base image
- **Optimization flags**: \-ldflags="-s -w"\ removes debug symbols and reduces binary size by ~30%

### Final Image Breakdown:
- Alpine base: ~9 MB
- CA certificates: ~1.5 MB
- Go binary (optimized): ~20.5 MB
- Migrations: ~20 KB
- **Total: ~31 MB**

## Building the Docker Image

### 1. Build the optimized image

\\\powershell
# Build the image with optimization flags
docker build -t software-engineering-app:latest .
\\\

### 2. Verify the image size

\\\powershell
# Check image size
docker images software-engineering-app:latest

# View detailed layer information
docker history software-engineering-app:latest
\\\

Expected output:
\\\
IMAGE          CREATED          SIZE
<image_id>     X minutes ago    ~31MB
\\\

## Running the Application

### 1. Create Docker network

\\\powershell
docker network create app-network
\\\

### 2. Start PostgreSQL container

\\\powershell
docker run -d --name postgres --network app-network \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:16

# Wait for PostgreSQL to start
Start-Sleep -Seconds 10
\\\

### 3. Apply database migrations

\\\powershell
# Create users table
docker exec -i postgres psql -U postgres -d testdb -c "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username VARCHAR(50) UNIQUE NOT NULL, email VARCHAR(100) UNIQUE NOT NULL, full_name VARCHAR(100), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"

# Insert test data
docker exec -i postgres psql -U postgres -d testdb -c "INSERT INTO users (username, email, full_name) VALUES ('jdoe', 'jdoe@example.com', 'John Doe'), ('asmith', 'asmith@example.com', 'Alice Smith'), ('bjones', 'bjones@example.com', 'Bob Jones');"

# Add UUID column
docker exec -i postgres psql -U postgres -d testdb -c "ALTER TABLE users ADD COLUMN IF NOT EXISTS uuid UUID DEFAULT gen_random_uuid() UNIQUE;"

# Update existing rows with UUIDs
docker exec -i postgres psql -U postgres -d testdb -c "UPDATE users SET uuid = gen_random_uuid() WHERE uuid IS NULL;"

# Set UUID column as NOT NULL
docker exec -i postgres psql -U postgres -d testdb -c "ALTER TABLE users ALTER COLUMN uuid SET NOT NULL;"
\\\

### 4. Verify migrations

\\\powershell
# Check table structure
docker exec -it postgres psql -U postgres -d testdb -c "\d users"

# View data
docker exec -it postgres psql -U postgres -d testdb -c "SELECT * FROM users;"
\\\

### 5. Start application container

\\\powershell
docker run -d -p 8080:8080 --name my-app --network app-network \
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" \
  software-engineering-app:latest
\\\

### 6. Verify application is running

\\\powershell
# Check logs
docker logs my-app

# Test API endpoint
curl http://localhost:8080/api/v1/users/username/jdoe

# Or open in browser
start http://localhost:8080/api/v1/users/username/jdoe
\\\

Expected response:
\\\json
{
  "id": 1,
  "uuid": "31447f86-79a5-4cf1-9be2-dea783641690",
  "username": "jdoe",
  "email": "jdoe@example.com",
  "full_name": "John Doe"
}
\\\

## Container Management

### Stop containers

\\\powershell
docker stop my-app postgres
\\\

### Start containers

\\\powershell
docker start postgres
Start-Sleep -Seconds 5
docker start my-app
\\\

### Remove containers

\\\powershell
docker rm -f my-app postgres
\\\

### View running containers

\\\powershell
docker ps
\\\

### View all containers (including stopped)

\\\powershell
docker ps -a
\\\

## Troubleshooting

### View application logs

\\\powershell
docker logs my-app
docker logs -f my-app  # Follow logs in real-time
\\\

### View PostgreSQL logs

\\\powershell
docker logs postgres
\\\

### Connect to PostgreSQL directly

\\\powershell
docker exec -it postgres psql -U postgres -d testdb
\\\

### Rebuild image after code changes

\\\powershell
# Stop and remove old container
docker stop my-app
docker rm my-app

# Remove old image
docker rmi software-engineering-app:latest

# Rebuild
docker build -t software-engineering-app:latest .

# Run new container
docker run -d -p 8080:8080 --name my-app --network app-network \
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" \
  software-engineering-app:latest
\\\

## Environment Variables

The application requires the following environment variable:

- \POSTGRES_DSN\: PostgreSQL connection string
  - Format: \postgresql://username:password@host:port/database?sslmode=disable\
  - Example: \postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable\

## Network Configuration

The application and database communicate through a Docker network named \pp-network\. This allows containers to reference each other by container name instead of IP addresses.

## Port Mapping

- Application: \8080:8080\ (host:container)
- PostgreSQL: \5432:5432\ (host:container)

## Image Size Comparison

- Without optimization (\-ldflags\): ~42 MB
- With optimization (\-ldflags="-s -w"\): ~31 MB
- **Size reduction: ~26%**

## Production Considerations

1. **Use specific image tags** instead of \latest\ for reproducibility
2. **Store secrets** in environment variables or secret management systems (not in Dockerfile)
3. **Use Docker Compose** for multi-container orchestration
4. **Implement health checks** in Dockerfile
5. **Use volume mounts** for PostgreSQL data persistence
6. **Consider using distroless images** for even smaller size and better security
