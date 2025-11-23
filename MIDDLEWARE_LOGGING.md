# Bonus Task: JSON Format Logging Implementation

## Overview
This document describes the implementation of JSON format logging middleware for the application, completed as a bonus task.

## Summary of Changes

### 1. Created Logger Middleware (`internal/middleware/logger.go`)

The middleware captures all incoming HTTP requests and logs them in JSON format with the following features:

- **Timestamp**: ISO 8601 format (RFC3339Nano)
- **Request Duration**: Calculated in milliseconds
- **Log Level**: Dynamic based on status code (info, warning, error)
- **HTTP Method**: GET, POST, PATCH, DELETE, etc.
- **Response Status Code**: The actual HTTP status code returned
- **Route Pattern**: The Gin route pattern (e.g., `/api/v1/users/username/:username`)
- **Request Path**: The actual request URL path
- **Host**: The request host
- **Route Parameters**: Automatically extracted (username, id, uuid are mapped to user_id)

### 2. Integrated Middleware (`internal/handler/router.go:12`)

The middleware is applied globally to all routes using `router.Use(middleware.JSONLogger())`, ensuring every request is logged.

### 3. Key Implementation Details

- The middleware records the start time before processing the request
- It calculates duration after the request completes
- Log level is determined automatically:
  - `error` for 5xx status codes
  - `warning` for 4xx status codes
  - `info` for 2xx and 3xx status codes
- Route parameters are extracted and included in the JSON output
- The output format matches the requirements exactly

## Example Output

When you make a request to `/api/v1/users/username/xyz`, you'll see:
```
2025/09/23 11:27:01 Incoming request: {"timestamp":"2025-09-23T11:27:01.691991902+03:00","http.server.request.duration":1,"http.log.level":"info","http.request.method":"GET","http.response.status_code":200,"http.route":"/api/v1/users/username/:username","http.request.message":"Incoming request:","server.address":"/api/v1/users/username/xyz","http.request.host":"localhost","user_id":"xyz"}
```

## Log Level Rules

| Status Code Range | Log Level |
|------------------|-----------|
| 200-399          | info      |
| 400-499          | warning   |
| 500-599          | error     |

## Files Modified

- **Created**: `internal/middleware/logger.go` - JSON logging middleware implementation
- **Modified**: `internal/handler/router.go` - Added middleware integration

## Testing

### Production Testing

To test the logging in production:
```bash
# Start the application
make run

# Make a test request
curl http://localhost:8080/api/v1/users/username/jdoe

# Check the logs - you should see JSON formatted output
```

### Local Development Testing with Docker

For local debugging and testing by developers using Docker:

#### 1. Rebuild and start the application
```powershell
# Stop and remove existing container
docker stop my-app
docker rm my-app

# Rebuild the image with latest changes (no cache)
docker build -t software-engineering-app:latest . --no-cache

# Start PostgreSQL (if not already running)
docker start postgres

# Start the application
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  software-engineering-app:latest

# Wait for startup
Start-Sleep -Seconds 3
```

#### 2. Test different scenarios
```powershell
# Successful request (200 - info level)
curl http://localhost:8080/api/v1/users/username/jdoe
curl http://localhost:8080/api/v1/users/username/asmith

# Non-existent user (404 - warning level)
curl http://localhost:8080/api/v1/users/username/nonexistent

# Get user by ID (200 - info level)
curl http://localhost:8080/api/v1/users/id/1
curl http://localhost:8080/api/v1/users/id/2

# Get all users (200 - info level)
curl http://localhost:8080/api/v1/users/

# Invalid ID format (404 - warning level)
curl http://localhost:8080/api/v1/users/id/999
```

#### 3. View logs
```powershell
# View all logs
docker logs my-app

# View only JSON request logs
docker logs my-app | Select-String "Incoming request"

# Follow logs in real-time
docker logs -f my-app
```

#### 4. Expected log output examples

**Successful request (info level):**
```json
2025/11/21 22:47:19 Incoming request: {"http.log.level":"info","http.request.host":"localhost:8080","http.request.message":"Incoming request:","http.request.method":"GET","http.response.status_code":200,"http.route":"/api/v1/users/username/:username","http.server.request.duration":39,"server.address":"/api/v1/users/username/jdoe","timestamp":"2025-11-21T22:47:19.750381787Z","user_id":"jdoe"}
```

**Not found request (warning level):**
```json
2025/11/21 22:50:15 Incoming request: {"http.log.level":"warning","http.request.host":"localhost:8080","http.request.message":"Incoming request:","http.request.method":"GET","http.response.status_code":404,"http.route":"/api/v1/users/username/:username","http.server.request.duration":12,"server.address":"/api/v1/users/username/nonexistent","timestamp":"2025-11-21T22:50:15.123456789Z","user_id":"nonexistent"}
```

#### 5. Cleanup
```powershell
# Stop containers
docker stop my-app postgres

# Remove containers (optional)
docker rm my-app postgres

# Remove network (optional)
docker network rm app-network
```

## Benefits

1. **Structured Logging**: JSON format makes logs easily parsable by log aggregation tools
2. **Comprehensive Information**: Captures all relevant request/response details
3. **Performance Tracking**: Request duration helps identify slow endpoints
4. **Automatic Log Levels**: Status code-based levels help with log filtering
5. **Parameter Extraction**: Route parameters are automatically included for better traceability

## Implementation Status

**Completed** - The middleware has been successfully implemented, integrated, and tested.
