package handler

import (
	"bytes"
	"cruder/internal/controller"
	"cruder/internal/model"
	"cruder/internal/repository"
	"cruder/internal/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var (
	testDB     *sql.DB
	testRouter *gin.Engine
	apiKey     = "test-api-key-12345"
)

// TestMain sets up test database and runs all tests
func TestMain(m *testing.M) {
	// Setup: Initialize test database connection
	var err error
	testDB, err = setupTestDB()
	if err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		fmt.Println("Skipping integration tests. Set TEST_DATABASE_URL to run them.")
		os.Exit(0) // Skip tests instead of failing
	}

	// Run migrations
	if err := runMigrations(testDB); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		testDB.Close()
		os.Exit(1)
	}

	// Setup router with test API key
	testRouter = setupTestRouter(testDB, apiKey)

	// Run tests
	code := m.Run()

	// Cleanup: Close database connection
	testDB.Close()

	os.Exit(code)
}

// setupTestDB connects to test database using TEST_DATABASE_URL environment variable
func setupTestDB() (*sql.DB, error) {
	// Get test database URL from environment
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		return nil, fmt.Errorf("TEST_DATABASE_URL environment variable not set")
	}

	// Connect to test database
	db, err := sql.Open("postgres", testDatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	return db, nil
}

// runMigrations executes database migrations for testing
func runMigrations(db *sql.DB) error {
	// Create users table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			full_name VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL
		);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

// setupTestRouter creates a test router with all handlers
func setupTestRouter(db *sql.DB, apiKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add simple API key middleware for testing
	router.Use(func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}
		if key != apiKey {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}
		c.Next()
	})

	// Setup dependencies
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	ctrl := controller.NewUserController(svc)

	// Setup routes
	v1 := router.Group("/api/v1")
	{
		userGroup := v1.Group("/users")
		{
			userGroup.GET("/", ctrl.GetAllUsers)
			userGroup.GET("/username/:username", ctrl.GetUserByUsername)
			userGroup.GET("/id/:id", ctrl.GetUserByID)
			userGroup.POST("/", ctrl.CreateUser)
			userGroup.PATCH("/:uuid", ctrl.UpdateUser)
			userGroup.DELETE("/:uuid", ctrl.DeleteUser)
		}
	}

	return router
}

// clearDatabase removes all test data between tests
func clearDatabase(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clear database: %v", err)
	}
}

// Helper Functions

// insertTestUser inserts a user into test database and returns the created user with generated UUID
func insertTestUser(t *testing.T, user *model.User) *model.User {
	t.Helper()
	query := `
		INSERT INTO users (username, email, full_name)
		VALUES ($1, $2, $3)
		RETURNING id, uuid
	`
	err := testDB.QueryRow(query, user.Username, user.Email, user.FullName).Scan(&user.ID, &user.UUID)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	return user
}

// userExists checks if a user exists in the database by UUID
func userExists(t *testing.T, uuid string) bool {
	t.Helper()
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM users WHERE uuid = $1", uuid).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check if user exists: %v", err)
	}
	return count > 0
}

// getUserByUUID retrieves a user from the database by UUID
func getUserByUUID(t *testing.T, uuid string) *model.User {
	t.Helper()
	var user model.User
	query := "SELECT id, uuid, username, email, full_name FROM users WHERE uuid = $1"
	err := testDB.QueryRow(query, uuid).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.FullName)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		t.Fatalf("Failed to get user by UUID: %v", err)
	}
	return &user
}

// makeRequest helper to create HTTP request with API key
func makeRequest(t *testing.T, method, url string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	return rr
}

// Test Cases for GET /api/v1/users/ - Get All Users

func TestGetAllUsers_Success(t *testing.T) {
	// Given: Multiple users exist in the database
	clearDatabase(t)

	user1 := &model.User{
		Username: "user1",
		Email:    "user1@example.com",
		FullName: "User One",
	}
	user2 := &model.User{
		Username: "user2",
		Email:    "user2@example.com",
		FullName: "User Two",
	}
	insertTestUser(t, user1)
	insertTestUser(t, user2)

	// When: Sending a GET request to /api/v1/users/
	rr := makeRequest(t, "GET", "/api/v1/users/", nil)

	// Then: The response status should be 200 OK and return JSON array with users
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var users []model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestGetAllUsers_Empty(t *testing.T) {
	// Given: No users exist in the database
	clearDatabase(t)

	// When: Sending a GET request to /api/v1/users/
	rr := makeRequest(t, "GET", "/api/v1/users/", nil)

	// Then: The response status should be 200 OK and return empty JSON array
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var users []model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

// Test Cases for GET /api/v1/users/username/:username - Get by Username

func TestGetUserByUsername_Success(t *testing.T) {
	// Given: A user exists in the database with username "johndoe"
	clearDatabase(t)

	user := &model.User{
		Username: "johndoe",
		Email:    "johndoe@example.com",
		FullName: "John Doe",
	}
	insertTestUser(t, user)

	// When: Sending a GET request to /api/v1/users/username/johndoe
	rr := makeRequest(t, "GET", "/api/v1/users/username/johndoe", nil)

	// Then: The response status should be 200 OK and return the user JSON
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var returnedUser model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &returnedUser); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if returnedUser.Username != "johndoe" {
		t.Errorf("expected username 'johndoe', got '%s'", returnedUser.Username)
	}
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	// Given: No user exists in the database with username "nonexistent"
	clearDatabase(t)

	// When: Sending a GET request to /api/v1/users/username/nonexistent
	rr := makeRequest(t, "GET", "/api/v1/users/username/nonexistent", nil)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

// Test Cases for GET /api/v1/users/id/:id - Get by ID

func TestGetUserByID_Success(t *testing.T) {
	// Given: A user exists in the database with a specific ID
	clearDatabase(t)

	user := &model.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		FullName: "Test User",
	}
	insertTestUser(t, user)

	// When: Sending a GET request to /api/v1/users/id/{id}
	url := fmt.Sprintf("/api/v1/users/id/%d", user.ID)
	rr := makeRequest(t, "GET", url, nil)

	// Then: The response status should be 200 OK and return the user JSON
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var returnedUser model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &returnedUser); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if returnedUser.ID != user.ID {
		t.Errorf("expected ID %d, got %d", user.ID, returnedUser.ID)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	// Given: No user exists in the database with ID 999999
	clearDatabase(t)

	// When: Sending a GET request to /api/v1/users/id/999999
	rr := makeRequest(t, "GET", "/api/v1/users/id/999999", nil)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestGetUserByID_InvalidID(t *testing.T) {
	// Given: An invalid ID parameter "invalid"
	clearDatabase(t)

	// When: Sending a GET request to /api/v1/users/id/invalid
	rr := makeRequest(t, "GET", "/api/v1/users/id/invalid", nil)

	// Then: The response status should be 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// Test Cases for POST /api/v1/users/ - Create User

func TestCreateUser_Success(t *testing.T) {
	// Given: Valid user data for creation
	clearDatabase(t)

	newUser := map[string]string{
		"username":  "newuser",
		"email":     "newuser@example.com",
		"full_name": "New User",
	}

	// When: Sending a POST request to /api/v1/users/ with valid data
	rr := makeRequest(t, "POST", "/api/v1/users/", newUser)

	// Then: The response status should be 201 Created and user should be created in database with UUID
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var createdUser model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if createdUser.UUID == "" {
		t.Error("expected UUID to be generated")
	}

	if createdUser.Username != "newuser" {
		t.Errorf("expected username 'newuser', got '%s'", createdUser.Username)
	}

	// Verify user exists in database
	if !userExists(t, createdUser.UUID) {
		t.Error("user was not created in database")
	}
}

func TestCreateUser_InvalidData(t *testing.T) {
	// Given: Invalid user data (missing required fields)
	clearDatabase(t)

	invalidUser := map[string]string{
		"username": "onlyusername",
		// Missing email
	}

	// When: Sending a POST request to /api/v1/users/ with invalid data
	rr := makeRequest(t, "POST", "/api/v1/users/", invalidUser)

	// Then: The response status should be 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	// Given: A user with username "existinguser" already exists
	clearDatabase(t)

	existingUser := &model.User{
		Username: "existinguser",
		Email:    "existing@example.com",
		FullName: "Existing User",
	}
	insertTestUser(t, existingUser)

	duplicateUser := map[string]string{
		"username":  "existinguser",
		"email":     "another@example.com",
		"full_name": "Another User",
	}

	// When: Sending a POST request to /api/v1/users/ with duplicate username
	rr := makeRequest(t, "POST", "/api/v1/users/", duplicateUser)

	// Then: The response status should be 409 Conflict
	if rr.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rr.Code)
	}
}

// Test Cases for PATCH /api/v1/users/:uuid - Update User

func TestUpdateUser_Success(t *testing.T) {
	// Given: A user exists in the database
	clearDatabase(t)

	user := &model.User{
		Username: "oldusername",
		Email:    "old@example.com",
		FullName: "Old Name",
	}
	insertTestUser(t, user)

	updatedData := map[string]string{
		"username":  "newusername",
		"email":     "new@example.com",
		"full_name": "New Name",
	}

	// When: Sending a PATCH request to /api/v1/users/{uuid} with valid data
	url := fmt.Sprintf("/api/v1/users/%s", user.UUID)
	rr := makeRequest(t, "PATCH", url, updatedData)

	// Then: The response status should be 200 OK and user should be updated in database
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rr.Code, rr.Body.String())
	}

	// Verify user was updated in database
	updatedUser := getUserByUUID(t, user.UUID)
	if updatedUser == nil {
		t.Fatal("user not found after update")
	}

	if updatedUser.Username != "newusername" {
		t.Errorf("expected username 'newusername', got '%s'", updatedUser.Username)
	}
	if updatedUser.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", updatedUser.Email)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	// Given: No user exists with UUID "00000000-0000-0000-0000-000000000000"
	clearDatabase(t)

	updatedData := map[string]string{
		"username":  "newusername",
		"email":     "new@example.com",
		"full_name": "New Name",
	}

	// When: Sending a PATCH request to /api/v1/users/00000000-0000-0000-0000-000000000000
	rr := makeRequest(t, "PATCH", "/api/v1/users/00000000-0000-0000-0000-000000000000", updatedData)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestUpdateUser_InvalidData(t *testing.T) {
	// Given: A user exists in the database
	clearDatabase(t)

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
	}
	insertTestUser(t, user)

	// Invalid JSON data
	invalidData := "not valid json"

	// When: Sending a PATCH request with invalid JSON data
	url := fmt.Sprintf("/api/v1/users/%s", user.UUID)
	req, _ := http.NewRequest("PATCH", url, bytes.NewBufferString(invalidData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// Then: The response status should be 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// Test Cases for DELETE /api/v1/users/:uuid - Delete User

func TestDeleteUser_Success(t *testing.T) {
	// Given: A user exists in the database
	clearDatabase(t)

	user := &model.User{
		Username: "userToDelete",
		Email:    "delete@example.com",
		FullName: "Delete Me",
	}
	insertTestUser(t, user)

	// When: Sending a DELETE request to /api/v1/users/{uuid}
	url := fmt.Sprintf("/api/v1/users/%s", user.UUID)
	rr := makeRequest(t, "DELETE", url, nil)

	// Then: The response status should be 204 No Content and user should be removed from database
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rr.Code)
	}

	// Verify user was deleted from database
	if userExists(t, user.UUID) {
		t.Error("user was not deleted from database")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	// Given: No user exists with UUID "00000000-0000-0000-0000-000000000000"
	clearDatabase(t)

	// When: Sending a DELETE request to /api/v1/users/00000000-0000-0000-0000-000000000000
	rr := makeRequest(t, "DELETE", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

// Test Cases for API Key Authentication

func TestAPIKeyAuthentication_MissingKey(t *testing.T) {
	// Given: A request without X-API-Key header
	clearDatabase(t)

	req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
	// Intentionally not setting X-API-Key header

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// When: Sending the request
	// Then: The response status should be 401 Unauthorized
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAPIKeyAuthentication_InvalidKey(t *testing.T) {
	// Given: A request with invalid X-API-Key header
	clearDatabase(t)

	req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
	req.Header.Set("X-API-Key", "invalid-key")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// When: Sending the request
	// Then: The response status should be 403 Forbidden
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}
