# Running Tests

## Run All Tests
```bash
go test ./...
```

## Run Tests with Verbose Output
```bash
go test ./... -v
```

## Run Service Layer Tests Only
```bash
go test ./internal/service/... -v
```

## Run Tests with Coverage
```bash
go test ./internal/service/... -cover
```

## Generate Coverage Report
```bash
# Generate coverage profile
go test ./internal/service/... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser (HTML)
go tool cover -html=coverage.out
```

## Run Specific Test
```bash
go test ./internal/service/... -run TestCreateUser_Success -v
```

## Run Tests Matching Pattern
```bash
# Run all Create tests
go test ./internal/service/... -run TestCreate -v

# Run all Delete tests
go test ./internal/service/... -run TestDelete -v
```

## Test Structure

Tests follow the **Given-When-Then** pattern:
```go
func TestCreateUser_Success(t *testing.T) {
    // Given: Setup initial state
    repo := newMockUserRepository()
    service := NewUserService(repo)
    
    // When: Perform action
    err := service.Create(newUser)
    
    // Then: Verify results
    if err != nil {
        t.Errorf("expected no error, got %v", err)
    }
}
```

## Current Test Coverage

Service layer tests:
- ✅ Create operations (2 tests)
- ✅ Update operations (2 tests)
- ✅ Delete operations (2 tests)
- ✅ GetByUsername (2 tests)
- ✅ GetByID (2 tests)
- ✅ GetByUUID (2 tests)

**Total: 12 tests**
