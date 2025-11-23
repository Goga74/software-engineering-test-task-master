# Testing Guide

This guide provides comprehensive information about testing in this Go web application, including both unit tests and integration tests.

## Table of Contents

- [Overview](#overview)
- [Test Types](#test-types)
- [Prerequisites](#prerequisites)
- [Running Tests](#running-tests)
- [Integration Tests](#integration-tests)
- [Unit Tests](#unit-tests)
- [Test Structure](#test-structure)
- [Writing New Tests](#writing-new-tests)
- [Troubleshooting](#troubleshooting)
- [CI/CD Integration](#cicd-integration)

## Overview

This project includes two types of tests:

1. **Unit Tests** - Test individual components in isolation using mocks
2. **Integration Tests** - Test the full stack with a real PostgreSQL database

All tests follow the **Given-When-Then** principle for clarity and maintainability.

## Test Types

### Unit Tests

- **Location**: `internal/service/users_test.go`
- **Purpose**: Test service layer business logic in isolation
- **Dependencies**: Uses mock repository (no database required)
- **Fast**: Run in milliseconds

### Integration Tests

- **Location**: `internal/handler/handler_integration_test.go`
- **Purpose**: Test full HTTP request → Controller → Service → Repository → Database flow
- **Dependencies**: Requires real PostgreSQL database
- **Comprehensive**: Tests complete system behavior

## Prerequisites

### For Unit Tests

No additional setup required. Just run:

```bash
go test ./internal/service/
```

### For Integration Tests

**Required**:
- PostgreSQL database (version 12+)
- `TEST_DATABASE_URL` environment variable

**Option A: Use Docker PostgreSQL (Recommended)**

```bash
# Start PostgreSQL container
docker run -d \
  --name test-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:16-alpine

# Verify connection
docker exec test-postgres pg_isready

# Set environment variable
export TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5432/testdb?sslmode=disable"
```

**Option B: Use docker-compose**

```bash
# Start services
docker-compose up -d postgres

# Set environment variable
export TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5432/testdb?sslmode=disable"
```

**Option C: Use existing PostgreSQL**

```bash
# Create test database
createdb testdb

# Set environment variable
export TEST_DATABASE_URL="postgresql://your_user:your_password@localhost:5432/testdb?sslmode=disable"
```

## Windows Setup

### Prerequisites for Windows

**Required Software:**
- Go 1.25+ installed
- Docker Desktop for Windows running
- PowerShell or Git Bash

### Quick Start for Windows

#### Step 1: Start PostgreSQL Test Database
```powershell
# Start PostgreSQL container on port 5433 (to avoid conflicts with existing instances)
docker run -d `
  --name test-postgres `
  -e POSTGRES_PASSWORD=postgres `
  -e POSTGRES_DB=testdb `
  -p 5433:5432 `
  postgres:16-alpine

# Wait for PostgreSQL to start (5-10 seconds)
Start-Sleep -Seconds 5

# Verify container is running
docker ps | Select-String "test-postgres"

# Verify PostgreSQL is ready
docker exec test-postgres pg_isready
```

#### Step 2: Set Environment Variable

**PowerShell:**
```powershell
# Set for current session
$env:TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5433/testdb?sslmode=disable"

# Verify
echo $env:TEST_DATABASE_URL
```

**Git Bash:**
```bash
# Set for current session
export TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5433/testdb?sslmode=disable"

# Verify
echo $TEST_DATABASE_URL
```

**Command Prompt:**
```cmd
REM Set for current session
set TEST_DATABASE_URL=postgresql://postgres:postgres@localhost:5433/testdb?sslmode=disable

REM Verify
echo %TEST_DATABASE_URL%
```

#### Step 3: Fix GOOS Environment Variable

**Important:** If you've been building Docker images, your `GOOS` may be set to `linux`. Reset it:
```powershell
# Check current setting
go env GOOS

# If it shows "linux", fix it:
go env -w GOOS=windows
go env -w GOARCH=amd64

# Verify
go env GOOS
# Should output: windows

# Clean build cache
go clean -cache
go clean -testcache
```

#### Step 4: Run Tests
```powershell
# Run integration tests only
go test -v ./internal/handler/

# Run all tests (unit + integration)
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Expected Output
```
=== RUN   TestGetAllUsers_Success
--- PASS: TestGetAllUsers_Success (0.01s)
=== RUN   TestGetAllUsers_Empty
--- PASS: TestGetAllUsers_Empty (0.00s)
...
=== RUN   TestAPIKeyAuthentication_InvalidKey
--- PASS: TestAPIKeyAuthentication_InvalidKey (0.00s)
PASS
ok      cruder/internal/handler 1.975s
```

### Cleanup After Testing
```powershell
# Stop and remove test database container
docker stop test-postgres
docker rm test-postgres

# Optional: Remove test database image
docker rmi postgres:16-alpine
```

### Common Windows Issues and Solutions

#### Issue 1: Port 5432 Already in Use

**Problem:** Another PostgreSQL instance is running on default port 5432

**Solution:** Use alternative port (5433 as shown above)
```powershell
# Check what's using port 5432
netstat -ano | Select-String "5432"

# Use port 5433 for tests instead
docker run -d --name test-postgres -p 5433:5432 ...
$env:TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5433/testdb?sslmode=disable"
```

#### Issue 2: "%1 is not a valid Win32 application"

**Problem:** `GOOS` environment variable set to `linux`

**Solution:**
```powershell
go env -w GOOS=windows
go env -w GOARCH=amd64
go clean -cache
go clean -testcache
go test -v ./internal/handler/
```

#### Issue 3: Docker Not Running

**Problem:** Docker Desktop is not started

**Solution:**
```powershell
# Start Docker Desktop from Start Menu
# Wait until Docker icon in system tray shows "Docker Desktop is running"

# Verify Docker is running
docker version
```

#### Issue 4: Connection Refused

**Problem:** PostgreSQL container not fully started

**Solution:**
```powershell
# Wait longer for PostgreSQL to start
Start-Sleep -Seconds 10

# Check container logs
docker logs test-postgres

# Verify PostgreSQL is ready
docker exec test-postgres pg_isready -U postgres
```

#### Issue 5: Tests Skip with "TEST_DATABASE_URL not set"

**Problem:** Environment variable not set or not visible to Go test

**Solution:**
```powershell
# Set in same PowerShell session where you run tests
$env:TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5433/testdb?sslmode=disable"

# Run tests in SAME session
go test -v ./internal/handler/

# Do NOT close PowerShell between setting variable and running tests
```

### Windows Development Workflow
```powershell
# 1. Start your development session
docker start test-postgres  # If already created
# OR
docker run -d --name test-postgres -p 5433:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=testdb postgres:16-alpine

# 2. Set environment variable (do this EVERY time you open new PowerShell)
$env:TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5433/testdb?sslmode=disable"

# 3. During development - run unit tests (fast)
go test ./internal/service/

# 4. Before commit - run integration tests
go test -v ./internal/handler/

# 5. Before push - run everything
go test -v ./...

# 6. End of day - stop container (keeps data)
docker stop test-postgres

# 7. Next day - restart container
docker start test-postgres
```

### PowerShell Script for Easy Testing

Create a file `test-integration.ps1`:
```powershell
# test-integration.ps1
# Helper script to run integration tests on Windows

# Ensure Docker is running
if (-not (docker info 2>$null)) {
    Write-Error "Docker is not running. Please start Docker Desktop."
    exit 1
}

# Start test database if not running
$containerExists = docker ps -a --format "{{.Names}}" | Select-String "test-postgres"
if (-not $containerExists) {
    Write-Host "Creating test-postgres container..."
    docker run -d --name test-postgres `
        -e POSTGRES_PASSWORD=postgres `
        -e POSTGRES_DB=testdb `
        -p 5433:5432 `
        postgres:16-alpine
    Start-Sleep -Seconds 5
} else {
    $containerRunning = docker ps --format "{{.Names}}" | Select-String "test-postgres"
    if (-not $containerRunning) {
        Write-Host "Starting test-postgres container..."
        docker start test-postgres
        Start-Sleep -Seconds 3
    }
}

# Verify PostgreSQL is ready
docker exec test-postgres pg_isready -U postgres
if ($LASTEXITCODE -ne 0) {
    Write-Error "PostgreSQL is not ready"
    exit 1
}

# Set environment variable
$env:TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5433/testdb?sslmode=disable"

# Fix GOOS if needed
$currentGOOS = go env GOOS
if ($currentGOOS -ne "windows") {
    Write-Host "Fixing GOOS setting..."
    go env -w GOOS=windows
    go env -w GOARCH=amd64
    go clean -cache
    go clean -testcache
}

# Run tests
Write-Host "Running integration tests..."
go test -v ./internal/handler/

Write-Host "`nTo run all tests: go test -v ./..."
```

Usage:
```powershell
# Make script executable and run
.\test-integration.ps1
```


## Running Tests

### Run All Tests

```bash
# Run all tests in the project
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Unit Tests Only

```bash
# Run service layer unit tests
go test ./internal/service/

# Run with verbose output
go test -v ./internal/service/

# Run specific test
go test -v ./internal/service/ -run TestCreateUser_Success
```

### Run Integration Tests Only

```bash
# Set test database URL
export TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5432/testdb?sslmode=disable"

# Run integration tests
go test -v ./internal/handler/

# Run specific integration test
go test -v ./internal/handler/ -run TestGetAllUsers_Success

# Run with timeout (useful for slow database connections)
go test -v -timeout 30s ./internal/handler/
```

### Run Tests in Parallel

```bash
# Run tests in parallel (default is GOMAXPROCS)
go test -v -parallel 4 ./...

# Note: Integration tests clear database between tests,
# so parallel execution may cause issues. Use with caution.
```

## Integration Tests

### Test Database Setup

Integration tests automatically:

1. Connect to test database using `TEST_DATABASE_URL`
2. Run migrations to create `users` table
3. Execute tests
4. Clean up data between tests

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TEST_DATABASE_URL` | Yes | - | PostgreSQL connection string |

**Example**:
```bash
TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5432/testdb?sslmode=disable"
```

### Test Coverage

Integration tests cover all HTTP endpoints:

#### GET Endpoints
- ✅ `GET /api/v1/users/` - Get all users (success, empty)
- ✅ `GET /api/v1/users/username/:username` - Get by username (success, not found)
- ✅ `GET /api/v1/users/id/:id` - Get by ID (success, not found, invalid ID)

#### POST Endpoints
- ✅ `POST /api/v1/users/` - Create user (success, invalid data, duplicate username)

#### PATCH Endpoints
- ✅ `PATCH /api/v1/users/:uuid` - Update user (success, not found, invalid data)

#### DELETE Endpoints
- ✅ `DELETE /api/v1/users/:uuid` - Delete user (success, not found)

#### Authentication
- ✅ API Key authentication (missing key, invalid key)

### Test Data Management

Tests use helper functions to manage test data:

- `clearDatabase(t)` - Removes all test data (called before each test)
- `insertTestUser(t, user)` - Inserts a test user
- `userExists(t, uuid)` - Checks if user exists
- `getUserByUUID(t, uuid)` - Retrieves user by UUID
- `makeRequest(t, method, url, body)` - Makes HTTP request with API key

### Example Test Structure

```go
func TestDeleteUser_Success(t *testing.T) {
    // Given: A user exists in the database
    clearDatabase(t)
    user := &model.User{
        Username: "userToDelete",
        Email:    "delete@example.com",
        FullName: "Delete Me",
    }
    insertTestUser(t, user)

    // When: Sending a DELETE request
    url := fmt.Sprintf("/api/v1/users/%s", user.UUID)
    rr := makeRequest(t, "DELETE", url, nil)

    // Then: Response should be 204 and user should be deleted
    if rr.Code != http.StatusNoContent {
        t.Errorf("expected status 204, got %d", rr.Code)
    }
    if userExists(t, user.UUID) {
        t.Error("user was not deleted from database")
    }
}
```

## Unit Tests

### Test Coverage

Unit tests in `internal/service/users_test.go` cover:

- ✅ `Create()` - Create user (success, duplicate username)
- ✅ `Update()` - Update user (success, not found, duplicate username)
- ✅ `Delete()` - Delete user (success, not found)
- ✅ `GetByUsername()` - Get by username (success, not found)
- ✅ `GetByID()` - Get by ID (success, not found)
- ✅ `GetAll()` - Get all users (success, empty)

### Mock Repository

Unit tests use `mockUserRepository` that implements the repository interface with in-memory storage. No database connection required.

## Test Structure

### Project Test Layout

```
software-engineering-test-task-master/
├── internal/
│   ├── handler/
│   │   └── handler_integration_test.go  # Integration tests (HTTP → DB)
│   ├── service/
│   │   └── users_test.go                # Unit tests (service logic)
│   ├── repository/
│   │   └── users.go                     # Repository implementation
│   ├── controller/
│   │   └── users.go                     # HTTP controllers
│   └── model/
│       └── users.go                     # Data models
└── TESTING.md                           # This file
```

### Test Naming Convention

Tests follow the pattern: `Test{Function}_{Scenario}`

Examples:
- `TestGetAllUsers_Success` - Get all users successfully
- `TestGetAllUsers_Empty` - Get all users when database is empty
- `TestCreateUser_DuplicateUsername` - Create user with duplicate username

### Given-When-Then Structure

All tests follow this pattern for clarity:

```go
func TestExample(t *testing.T) {
    // Given: Setup initial state and preconditions

    // When: Execute the action being tested

    // Then: Verify the expected outcome
}
```

## Writing New Tests

### Adding Unit Tests

1. Create test function in `internal/service/users_test.go`
2. Use `mockUserRepository` for dependencies
3. Follow Given-When-Then structure
4. Test both success and error cases

Example:
```go
func TestNewFeature_Success(t *testing.T) {
    // Given: Setup mock repository
    repo := newMockUserRepository()
    service := NewUserService(repo)

    // When: Execute action
    result, err := service.NewFeature()

    // Then: Verify outcome
    if err != nil {
        t.Errorf("expected no error, got %v", err)
    }
    // Additional assertions...
}
```

### Adding Integration Tests

1. Add test function in `internal/handler/handler_integration_test.go`
2. Clear database before test: `clearDatabase(t)`
3. Set up test data using helper functions
4. Make HTTP request using `makeRequest()`
5. Verify HTTP response AND database state
6. Follow Given-When-Then structure

Example:
```go
func TestNewEndpoint_Success(t *testing.T) {
    // Given: Setup test data
    clearDatabase(t)
    // Insert test data...

    // When: Make HTTP request
    rr := makeRequest(t, "GET", "/api/v1/new-endpoint", nil)

    // Then: Verify response and database state
    if rr.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", rr.Code)
    }
    // Verify database state...
}
```

## Troubleshooting

### Integration Tests Skipped

**Problem**: Tests exit with message "Skipping integration tests. Set TEST_DATABASE_URL to run them."

**Solution**:
```bash
export TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5432/testdb?sslmode=disable"
go test -v ./internal/handler/
```

### Database Connection Failed

**Problem**: `Failed to connect to test database`

**Solutions**:

1. **Check PostgreSQL is running**:
   ```bash
   # For Docker
   docker ps | grep postgres

   # For local PostgreSQL
   pg_isready
   ```

2. **Verify connection string**:
   ```bash
   # Test connection manually
   psql "postgresql://postgres:postgres@localhost:5432/testdb"
   ```

3. **Check firewall/ports**:
   ```bash
   # Verify port 5432 is accessible
   telnet localhost 5432
   ```

### Migration Errors

**Problem**: `Failed to run migrations`

**Solutions**:

1. **Check database permissions**:
   ```sql
   -- Connect as admin and grant permissions
   GRANT ALL PRIVILEGES ON DATABASE testdb TO postgres;
   ```

2. **Manually create table**:
   ```sql
   CREATE TABLE IF NOT EXISTS users (
       id SERIAL PRIMARY KEY,
       username VARCHAR(50) UNIQUE NOT NULL,
       email VARCHAR(100) UNIQUE NOT NULL,
       full_name VARCHAR(100),
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL
   );
   ```

### Test Failures

**Problem**: Tests fail with unexpected errors

**Solutions**:

1. **Clear test database**:
   ```sql
   -- Connect to test database
   \c testdb

   -- Drop and recreate users table
   DROP TABLE IF EXISTS users;

   -- Run migrations again
   ```

2. **Check for test isolation issues**:
   - Ensure `clearDatabase(t)` is called at start of each test
   - Verify no concurrent test execution interfering

3. **Increase timeout**:
   ```bash
   go test -v -timeout 60s ./internal/handler/
   ```

### Tests Hang or Timeout

**Problem**: Tests don't complete

**Solutions**:

1. **Check database connection pool**:
   ```bash
   # Check active connections
   docker exec test-postgres psql -U postgres -c "SELECT * FROM pg_stat_activity;"
   ```

2. **Kill hanging connections**:
   ```bash
   docker restart test-postgres
   ```

3. **Use shorter timeout**:
   ```bash
   go test -v -timeout 10s ./internal/handler/
   ```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run Unit Tests
        run: go test -v ./internal/service/

      - name: Run Integration Tests
        env:
          TEST_DATABASE_URL: postgresql://postgres:postgres@localhost:5432/testdb?sslmode=disable
        run: go test -v ./internal/handler/

      - name: Run All Tests with Coverage
        env:
          TEST_DATABASE_URL: postgresql://postgres:postgres@localhost:5432/testdb?sslmode=disable
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload Coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage
          path: coverage.html
```

### GitLab CI Example

```yaml
test:
  image: golang:1.25-alpine

  services:
    - postgres:16-alpine

  variables:
    POSTGRES_DB: testdb
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
    TEST_DATABASE_URL: postgresql://postgres:postgres@postgres:5432/testdb?sslmode=disable

  before_script:
    - apk add --no-cache postgresql-client

  script:
    - go test -v ./internal/service/
    - go test -v ./internal/handler/
    - go test -coverprofile=coverage.out ./...

  artifacts:
    paths:
      - coverage.out
```

## Best Practices

### Do's ✅

- ✅ Always use `clearDatabase(t)` before each integration test
- ✅ Follow Given-When-Then structure with clear comments
- ✅ Verify both HTTP responses AND database state
- ✅ Use helper functions for common operations
- ✅ Test both success and error cases
- ✅ Use meaningful test names that describe the scenario
- ✅ Keep tests independent and isolated

### Don'ts ❌

- ❌ Don't rely on test execution order
- ❌ Don't leave test data in database after tests
- ❌ Don't use production database for testing
- ❌ Don't skip error checking in tests
- ❌ Don't test implementation details, test behavior
- ❌ Don't create brittle tests that break on minor changes

## Performance Tips

### Speed Up Test Execution

1. **Run unit tests more frequently** (they're fast)
   ```bash
   go test ./internal/service/
   ```

2. **Use test caching**
   ```bash
   go test -count=1 ./...  # Disable cache
   go test ./...           # Use cache
   ```

3. **Focus on specific tests during development**
   ```bash
   go test -run TestCreateUser ./internal/handler/
   ```

4. **Use shorter timeouts for faster feedback**
   ```bash
   go test -timeout 5s ./...
   ```

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [httptest Package](https://golang.org/pkg/net/http/httptest/)
- [Given-When-Then Pattern](https://martinfowler.com/bliki/GivenWhenThen.html)

---

**Last Updated**: 2025-11-23
**Go Version**: 1.25+
**PostgreSQL Version**: 12+
