# Test task

The purpose of the test task is to check the skills of the candidate in software engineering and DevOps procedures. Please write the code following the clean code best practices.

Fork the repository, do the required changes and push the code to Github.

## Tasks

**1.** Find the reason why `make run` does not launch the application

**2.** Adjust service layer with business validation

Adjust service layer with user existence validation and print "users not found" for GetByUsername and GetByID.

**Note** Validation might require adjustments in repository layer

**3.** Create C/U/D endpoints

Add 3 endpoints of C (create) U (update) D (DELETE) operations

- DELETE /api/v1/users/:uuid - Allows to delete user via UUID
- PATCH /api/v1/users/:uuid - Allows to update user via UUID
- POST /api/v1/users - Allows to create new user

**4.** Write tests for service layer

Implement functional tests for the service layer of the API. Follow the Given When Then principle.

Example:
```go
// Given: A user exists in the database with UUID "123e4567-e89b-12d3-a456-426614174000"
func TestDeleteUser_Success(t *testing.T) {
    // Setup: Insert a user into the test database
    user := User{UUID: "123e4567-e89b-12d3-a456-426614174000", Name: "John Doe"}
    insertTestUser(user)

    // When: Sending a DELETE request to /api/v1/users/123e4567-e89b-12d3-a456-426614174000
    req, _ := http.NewRequest("DELETE", "/api/v1/users/123e4567-e89b-12d3-a456-426614174000", nil)
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    // Then: The response status should be 204 No Content and user should be removed from the database
    if rr.Code != http.StatusNoContent {
        t.Errorf("expected status 204, got %d", rr.Code)
    }
    if userExists("123e4567-e89b-12d3-a456-426614174000") {
        t.Errorf("user was not deleted from the database")
    }
}
```

**5.** Write Dockerfile for the application

Write simple dockerfile for the application. Make the docker container to be as minimal as possible (in size).

**6.** Write simple CI pipeline for verification

Write simple CI pipeline which will verify the code before getting to production

**Recommendation** Golang has builtin validators like `go vet`, `go fmt` and `go test`. Additionally [golang-ci lint](https://golangci-lint.run/) and [gosec](https://github.com/securego/gosec) can be used.

## Bonus points

- Implement application logging in JSON format

Implement application logging in JSON format via service or middleware layer.

**Recommendation**: The easiest way would be building Logger middleware with custom attributes and attach it to Router

Example:
```
2025/09/23 11:27:01 Incoming request: {"timestamp":"2025-09-23T11:27:01.691991902+03:00","http.server.request.duration":1,"http.log.level":"info","http.request.method":"GET","http.response.status_code":200,"http.route":"/api/v1/users/username/:username","http.request.message":"Incoming request:","server.address":"/api/v1/users/username/xyz","http.request.host":"localhost","user_id":"xyz"}  
```

- Add X-API-Key Authentication Middleware
Write simple middleware, which will allow the requests with correct X-API-Key value in the header to go through.
The requests without header must give 401 Unauthorized and with wrong key must give 403 Forbidden.

- Rewrite the database connection string in `cmd/main.go`

The database connection string is hardcoded in `cmd/main.go`. Add dynamic variables assignment for the connection string values.

Preferred way would be config.yaml for non-sensitive data and environment variables for sensitive data.

- Write terraform code to host the application in remote

Write re-usable terraform code to create infrastructure for the application. The application must be publicly available.

- Add kubernetes support

Write kubernetes manifests or helm chart to deploy the application into kubernetes cluster

**Note** Please use remote terraform state

- Write simple CD flow for build and deploy

Write CD pipeline via Github Actions to deploy the changes to remote environment whenever the changes are pushed to main.

## Comments

Comments section for candidate