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

## Example Test Output

```bash
$ go test ./internal/service/... -v
=== RUN   TestCreateUser_Success
--- PASS: TestCreateUser_Success (0.00s)
=== RUN   TestCreateUser_DuplicateUsername
--- PASS: TestCreateUser_DuplicateUsername (0.00s)
=== RUN   TestUpdateUser_Success
--- PASS: TestUpdateUser_Success (0.00s)
=== RUN   TestUpdateUser_NotFound
--- PASS: TestUpdateUser_NotFound (0.00s)
=== RUN   TestDeleteUser_Success
--- PASS: TestDeleteUser_Success (0.00s)
=== RUN   TestDeleteUser_NotFound
--- PASS: TestDeleteUser_NotFound (0.00s)
=== RUN   TestGetByUsername_Success
--- PASS: TestGetByUsername_Success (0.00s)
=== RUN   TestGetByUsername_NotFound
--- PASS: TestGetByUsername_NotFound (0.00s)
=== RUN   TestGetByID_Success
--- PASS: TestGetByID_Success (0.00s)
=== RUN   TestGetByID_NotFound
--- PASS: TestGetByID_NotFound (0.00s)
=== RUN   TestGetAll_Success
--- PASS: TestGetAll_Success (0.00s)
=== RUN   TestGetAll_Empty
--- PASS: TestGetAll_Empty (0.00s)
PASS
ok      cruder/internal/service 0.585s
```

## Current Test Coverage

Service layer tests:
- ✅ Create operations (2 tests)
- ✅ Update operations (2 tests)
- ✅ Delete operations (2 tests)
- ✅ GetByUsername (2 tests)
- ✅ GetByID (2 tests)
- ✅ GetAll (2 tests)

**Total: 12 tests** - All tests passing ✅
