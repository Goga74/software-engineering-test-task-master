# Bonus Task: X-API-Key Authentication Middleware

## Overview
This document describes the implementation of X-API-Key authentication middleware for the application, completed as a bonus task.

## Summary of Changes

### 1. Created Authentication Middleware (`internal/middleware/auth.go`)

The middleware implements secure API key validation with the following logic:

- **Missing X-API-Key header** → Returns HTTP 401 Unauthorized with `{"error": "API key required"}`
- **Invalid X-API-Key value** → Returns HTTP 403 Forbidden with `{"error": "Invalid API key"}`
- **Valid X-API-Key** → Request proceeds normally to the handler

### 2. Updated Router (`internal/handler/router.go:10,17`)

- Modified the `New()` function to accept an `apiKey` parameter
- Applied `APIKeyAuth` middleware to all `/api/v1/users/*` routes
- The middleware works alongside the existing JSON logger middleware

### 3. Updated Main Entry Point (`cmd/main.go:20-26,37`)

- Added API key loading from `X_API_KEY` environment variable
- Implemented fallback to default key `"dev-api-key-12345"` for development/testing
- Added warning message when using default key
- Passes API key to router initialization

## Security Features

- ✅ Secure header-based authentication
- ✅ Clear separation between missing and invalid keys
- ✅ Environment variable configuration for production
- ✅ Proper error responses with appropriate HTTP status codes
- ✅ Request abortion on authentication failure

## Usage Examples

### Setting Custom API Key (Production)
```bash
export X_API_KEY="your-secure-api-key-here"
./main
```

### Using Default Key (Development)
```bash
./main
# Output: Warning: Using default API key. Set X_API_KEY environment variable for production.
# Default key: "dev-api-key-12345"
```

## Testing the Middleware

### Basic Testing (Linux/Mac)

#### 1. Request without X-API-Key header:
```bash
curl -X GET http://localhost:8080/api/v1/users/

# Response: HTTP 401
# {"error":"API key required"}
```

#### 2. Request with invalid X-API-Key:
```bash
curl -X GET http://localhost:8080/api/v1/users/ -H "X-API-Key: wrong-key"

# Response: HTTP 403
# {"error":"Invalid API key"}
```

#### 3. Request with valid X-API-Key:
```bash
curl -X GET http://localhost:8080/api/v1/users/ -H "X-API-Key: dev-api-key-12345"

# Response: HTTP 200
# [user data...]
```

## Docker Testing

### Step 1: Build and Run Application

When running in Docker, pass the API key as an environment variable:
```powershell
# Stop and remove existing container
docker stop my-app
docker rm my-app

# Rebuild image without cache to include latest code
docker build -t software-engineering-app:latest . --no-cache

# Start PostgreSQL (if not already running)
docker start postgres

# Run application with custom API key (production mode)
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  -e X_API_KEY="my-secure-production-key" `
  software-engineering-app:latest

# OR run with default key (development mode)
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  software-engineering-app:latest

# Wait for startup
Start-Sleep -Seconds 3
```

### Step 2: Verify API Key Configuration
```powershell
# Check that X_API_KEY environment variable is set in container
docker exec my-app env | Select-String "API"

# Expected output:
# X_API_KEY=dev-api-key-12345

# Check application startup logs
docker logs my-app | Select-Object -First 15
```

### Step 3: Test Authentication (Windows PowerShell)

**Important**: On Windows, use `curl.exe` (not PowerShell's `curl` alias) for proper header handling.

#### Test 1: Request WITHOUT API key (Expected: 401 Unauthorized)
```powershell
curl.exe -i http://localhost:8080/api/v1/users/

# Expected Response:
# HTTP/1.1 401 Unauthorized
# Content-Type: application/json; charset=utf-8
# {"error":"API key required"}
```

#### Test 2: Request with INVALID API key (Expected: 403 Forbidden)
```powershell
curl.exe -i http://localhost:8080/api/v1/users/ -H "X-API-Key: wrong-key"

# Expected Response:
# HTTP/1.1 403 Forbidden
# Content-Type: application/json; charset=utf-8
# {"error":"Invalid API key"}
```

#### Test 3: Request with VALID API key (Expected: 200 OK)
```powershell
curl.exe -i http://localhost:8080/api/v1/users/ -H "X-API-Key: dev-api-key-12345"

# Expected Response:
# HTTP/1.1 200 OK
# Content-Type: application/json; charset=utf-8
# [{"id":1,"uuid":"...","username":"jdoe",...}, ...]
```

### Step 4: Test Other Endpoints
```powershell
# Test GET by username (with valid key)
curl.exe -i http://localhost:8080/api/v1/users/username/jdoe -H "X-API-Key: dev-api-key-12345"

# Test GET by ID (with valid key)
curl.exe -i http://localhost:8080/api/v1/users/id/1 -H "X-API-Key: dev-api-key-12345"

# Test CREATE endpoint (POST) - should fail without key
curl.exe -i -X POST http://localhost:8080/api/v1/users/ `
  -H "Content-Type: application/json" `
  -d '{\"username\":\"testuser\",\"email\":\"test@example.com\",\"full_name\":\"Test User\"}'

# Test CREATE endpoint (POST) - should succeed with valid key
curl.exe -i -X POST http://localhost:8080/api/v1/users/ `
  -H "Content-Type: application/json" `
  -H "X-API-Key: dev-api-key-12345" `
  -d '{\"username\":\"testuser\",\"email\":\"test@example.com\",\"full_name\":\"Test User\"}'
```

### Alternative: Using PowerShell Invoke-WebRequest
```powershell
# Without API key (401)
try {
    Invoke-WebRequest -Uri "http://localhost:8080/api/v1/users/" -Method GET
} catch {
    Write-Host "Status Code:" $_.Exception.Response.StatusCode.value__
    Write-Host "Error:" $_.ErrorDetails.Message
}

# With invalid API key (403)
try {
    $headers = @{ "X-API-Key" = "wrong-key" }
    Invoke-WebRequest -Uri "http://localhost:8080/api/v1/users/" -Headers $headers -Method GET
} catch {
    Write-Host "Status Code:" $_.Exception.Response.StatusCode.value__
    Write-Host "Error:" $_.ErrorDetails.Message
}

# With valid API key (200)
$headers = @{ "X-API-Key" = "dev-api-key-12345" }
$response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/users/" -Headers $headers -Method GET
Write-Host "Status Code:" $response.StatusCode
Write-Host "Content:" $response.Content
```

### Step 5: View Request Logs

The JSON logger middleware logs all requests, including authentication failures:
```powershell
# View all JSON request logs
docker logs my-app | Select-String "Incoming request"

# Expected output includes logs for all requests with different status codes:
# - 401 for requests without API key
# - 403 for requests with invalid API key
# - 200 for successful authenticated requests
```

### Step 6: Cleanup
```powershell
# Stop and remove containers
docker stop my-app postgres
docker rm my-app postgres

# Remove network (optional)
docker network rm app-network
```

## Protected Routes

All the following routes now require X-API-Key authentication:

- `GET /api/v1/users/`
- `GET /api/v1/users/username/:username`
- `GET /api/v1/users/id/:id`
- `POST /api/v1/users/`
- `PATCH /api/v1/users/:uuid`
- `DELETE /api/v1/users/:uuid`

## Security Best Practices

1. **Never commit API keys to version control** - Always use environment variables
2. **Use strong, random API keys in production** - Minimum 32 characters recommended
3. **Rotate API keys regularly** - Implement key rotation policy
4. **Log authentication failures** - Monitor for potential attacks (already logged by JSON middleware)
5. **Use HTTPS in production** - Prevent API key interception

## Files Modified

- **Created**: `internal/middleware/auth.go` - API key authentication middleware implementation
- **Modified**: `internal/handler/router.go` - Added middleware integration and apiKey parameter
- **Modified**: `cmd/main.go` - Added environment variable loading and API key configuration

## Middleware Execution Order

1. **JSON Logger** - Logs all requests (including failed auth attempts)
2. **API Key Auth** - Validates X-API-Key header
3. **Route Handler** - Processes the actual request (if authentication passes)

## Troubleshooting

### Issue: Authentication not working (all requests succeed)

**Solution**: Rebuild Docker image without cache to ensure latest code is included:
```powershell
docker stop my-app
docker rm my-app
docker rmi software-engineering-app:latest
docker build -t software-engineering-app:latest . --no-cache
```

### Issue: curl command not working in PowerShell

**Solution**: Use `curl.exe` instead of `curl` (which is an alias for Invoke-WebRequest):
```powershell
curl.exe -i http://localhost:8080/api/v1/users/ -H "X-API-Key: dev-api-key-12345"
```

### Issue: Cannot see HTTP status codes

**Solution**: Add `-i` flag to curl.exe to include response headers:
```powershell
curl.exe -i http://localhost:8080/api/v1/users/
```

## Implementation Status

✅ **Completed** - The authentication middleware has been successfully implemented, integrated, and tested. The implementation follows Go best practices, is secure, and integrates seamlessly with the existing middleware stack.
