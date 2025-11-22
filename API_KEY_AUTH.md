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

### 1. Request without X-API-Key header:
```bash
curl -X GET http://localhost:8080/api/v1/users/

# Response: HTTP 401
# {"error":"API key required"}
```

### 2. Request with invalid X-API-Key:
```bash
curl -X GET http://localhost:8080/api/v1/users/ -H "X-API-Key: wrong-key"

# Response: HTTP 403
# {"error":"Invalid API key"}
```

### 3. Request with valid X-API-Key:
```bash
curl -X GET http://localhost:8080/api/v1/users/ -H "X-API-Key: dev-api-key-12345"

# Response: HTTP 200
# [user data...]
```

## Protected Routes

All the following routes now require X-API-Key authentication:

- `GET /api/v1/users/`
- `GET /api/v1/users/username/:username`
- `GET /api/v1/users/id/:id`
- `POST /api/v1/users/`
- `PATCH /api/v1/users/:uuid`
- `DELETE /api/v1/users/:uuid`

## Docker Testing

When running in Docker, pass the API key as an environment variable:
```powershell
# Using custom API key
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  -e X_API_KEY="my-secure-production-key" `
  software-engineering-app:latest

# Using default key (development)
docker run -d -p 8080:8080 --name my-app --network app-network `
  -e POSTGRES_DSN="postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  software-engineering-app:latest
```

### Testing with Docker:
```powershell
# Without API key (should fail with 401)
curl http://localhost:8080/api/v1/users/

# With invalid API key (should fail with 403)
curl http://localhost:8080/api/v1/users/ -H "X-API-Key: wrong-key"

# With valid API key (should succeed)
curl http://localhost:8080/api/v1/users/ -H "X-API-Key: dev-api-key-12345"

# Or using Invoke-WebRequest in PowerShell
$headers = @{ "X-API-Key" = "dev-api-key-12345" }
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/users/" -Headers $headers
```

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

## Implementation Status

✅ **Completed** - The authentication middleware has been successfully implemented, integrated, and tested. The implementation follows Go best practices, is secure, and integrates seamlessly with the existing middleware stack.
