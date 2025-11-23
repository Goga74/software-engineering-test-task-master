# Database Configuration Management

## Overview

This document describes the refactored database configuration system that separates sensitive credentials from non-sensitive connection parameters, following security best practices.

## Implementation Summary

### What Was Changed

1. **Created Configuration Package** (`internal/config/config.go`)
   - Loads non-sensitive parameters from `config.yaml`
   - Reads sensitive credentials from environment variables
   - Supports environment variable overrides for all parameters
   - Maintains backward compatibility with `POSTGRES_DSN`

2. **Configuration Files**
   - `config.yaml` - Default configuration with Docker-friendly settings
   - `config.example.yaml` - Template for local customization
   - Non-sensitive data (host, port, database name, SSL mode)

3. **Updated Main Entry Point** (`cmd/main.go`)
   - Uses `config.GetDSN()` to build connection string
   - Prioritizes `POSTGRES_DSN` for backward compatibility
   - Falls back to config.yaml + environment variables

4. **Updated Dockerfile**
   - Includes `config.yaml` in the image
   - Uses vendor directory for reliable builds (network-independent)

## Configuration Methods

### Method 1: config.yaml + Environment Variables (Recommended)

This method separates configuration into non-sensitive (config file) and sensitive (environment variables) data.

**Configuration File** (`config.yaml`):
```yaml
database:
  host: postgres
  port: 5432
  name: testdb
  sslmode: disable
```

**Required Environment Variables:**
- `DB_USER` - Database username (required)
- `DB_PASSWORD` - Database password (required)

**Optional Environment Variable Overrides:**
- `DB_HOST` - Override database host
- `DB_PORT` - Override database port
- `DB_NAME` - Override database name
- `DB_SSLMODE` - Override SSL mode

**Example:**
```bash
export DB_USER="postgres"
export DB_PASSWORD="your-secure-password"
export DB_HOST="prod-db.example.com"  # optional override
./main
```

### Method 2: POSTGRES_DSN (Backward Compatible)

Legacy method using a single connection string environment variable.

**Environment Variable:**
- `POSTGRES_DSN` - Complete PostgreSQL connection string

**Example:**
```bash
export POSTGRES_DSN="postgresql://username:password@host:port/database?sslmode=disable"
./main
```

**Note:** When `POSTGRES_DSN` is set, all config.yaml settings are ignored.

## Docker Deployment

### Building the Image
```powershell
# Remove old container (if exists)
docker rm -f my-app

# Build image with configuration
docker build -t software-engineering-app:latest .
```

### Running with New Configuration Method
```powershell
# Run with config.yaml + environment variables
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e DB_USER="postgres" `
  -e DB_PASSWORD="postgres" `
  -e X_API_KEY="dev-api-key-12345" `
  software-engineering-app:latest

# Wait for startup
Start-Sleep -Seconds 3

# Check logs
docker logs my-app

# Test API
curl.exe http://localhost:8080/api/v1/users/ -H "X-API-Key: dev-api-key-12345"
```

### Running with Legacy POSTGRES_DSN Method
```powershell
# Run with POSTGRES_DSN (backward compatible)
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  -e X_API_KEY="dev-api-key-12345" `
  software-engineering-app:latest

# Wait for startup
Start-Sleep -Seconds 3

# Check logs
docker logs my-app

# Test API
curl.exe http://localhost:8080/api/v1/users/ -H "X-API-Key: dev-api-key-12345"
```

### Running with Environment Variable Overrides
```powershell
# Override specific config.yaml values
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e DB_HOST="custom-postgres-host" `
  -e DB_PORT="5433" `
  -e DB_NAME="production_db" `
  -e DB_USER="produser" `
  -e DB_PASSWORD="secure-password" `
  -e X_API_KEY="production-api-key" `
  software-engineering-app:latest
```

## Local Development

### Setup

1. Copy the example configuration:
```bash
cp config.example.yaml config.yaml
```

2. Edit `config.yaml` for your local environment:
```yaml
database:
  host: localhost
  port: 5432
  name: development_db
  sslmode: disable
```

3. Set environment variables:
```bash
export DB_USER="devuser"
export DB_PASSWORD="devpass"
export X_API_KEY="dev-api-key-12345"
```

4. Run the application:
```bash
go run ./cmd/main.go
```

## Configuration Priority

The configuration system follows this priority order (highest to lowest):

1. **POSTGRES_DSN** environment variable (if set, all other config is ignored)
2. **Environment variable overrides** (DB_HOST, DB_PORT, DB_NAME, DB_SSLMODE)
3. **config.yaml** file values
4. **Required credentials** (DB_USER, DB_PASSWORD) must always be provided via environment variables

## Error Handling

### Missing Credentials

If `POSTGRES_DSN` is not set and required credentials are missing:
```
Error: DB_USER environment variable is required when POSTGRES_DSN is not set
Error: DB_PASSWORD environment variable is required when POSTGRES_DSN is not set
```

**Solution:** Set the required environment variables.

### Missing Configuration File

If `config.yaml` is not found:
```
Error: failed to load database configuration: failed to read config file
```

**Solution:** Ensure `config.yaml` exists in the working directory or use `POSTGRES_DSN`.

### Invalid Configuration

If config.yaml has syntax errors:
```
Error: failed to load database configuration: failed to parse config file
```

**Solution:** Validate YAML syntax in `config.yaml`.

## Security Best Practices

1. **Never commit credentials** - Use environment variables for sensitive data
2. **Keep config.yaml in version control** - It contains only non-sensitive parameters
3. **Use strong passwords** - Especially in production environments
4. **Rotate credentials regularly** - Implement credential rotation policy
5. **Use SSL/TLS** - Set `sslmode: require` in production
6. **Restrict database access** - Use firewall rules and network policies

## Kubernetes/Production Deployment

### Using Kubernetes Secrets
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: database-credentials
type: Opaque
stringData:
  DB_USER: postgres
  DB_PASSWORD: your-secure-password

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    database:
      host: postgres-service
      port: 5432
      name: production_db
      sslmode: require

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
      - name: app
        image: software-engineering-app:latest
        env:
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: DB_USER
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: DB_PASSWORD
        volumeMounts:
        - name: config
          mountPath: /root/config.yaml
          subPath: config.yaml
      volumes:
      - name: config
        configMap:
          name: app-config
```

## Files Modified

- **Created**: `internal/config/config.go` - Configuration loader
- **Created**: `config.yaml` - Default configuration
- **Created**: `config.example.yaml` - Configuration template
- **Modified**: `cmd/main.go` - Uses new configuration system
- **Modified**: `Dockerfile` - Includes config.yaml and uses vendor
- **Modified**: `go.mod`, `go.sum` - Added gopkg.in/yaml.v3 dependency
- **Added**: `vendor/` - Vendored dependencies for reliable builds

## Testing

Both configuration methods have been tested successfully:

✅ **Method 1 (config.yaml + env vars)**: Verified working in Docker
✅ **Method 2 (POSTGRES_DSN)**: Backward compatibility verified
✅ **Environment overrides**: DB_HOST, DB_PORT, DB_NAME work correctly
✅ **Error handling**: Missing credentials produce clear error messages

## Benefits

1. **Security**: Separates sensitive credentials from configuration
2. **Flexibility**: Easy to customize per environment without code changes
3. **Maintainability**: Clear separation of concerns
4. **Backward Compatibility**: Existing deployments continue to work
5. **Production Ready**: Follows 12-factor app methodology
6. **Docker Friendly**: Works seamlessly in containerized environments

## Implementation Status

✅ **Completed** - Database configuration has been successfully refactored following security best practices and maintaining backward compatibility.

