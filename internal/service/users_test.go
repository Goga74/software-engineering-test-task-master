package service

import (
	"cruder/internal/model"
	"database/sql"
	"testing"
)

// Mock repository for testing
type mockUserRepository struct {
	users map[string]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) GetAll() ([]model.User, error) {
	var users []model.User
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *mockUserRepository) GetByUsername(username string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepository) GetByID(id int64) (*model.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepository) GetByUUID(uuid string) (*model.User, error) {
	user, exists := m.users[uuid]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (m *mockUserRepository) Create(user *model.User) error {
	// Generate UUID for test
	if user.UUID == "" {
		user.UUID = "test-uuid-" + user.Username
	}
	if user.ID == 0 {
		user.ID = int64(len(m.users) + 1)
	}
	m.users[user.UUID] = user
	return nil
}

func (m *mockUserRepository) Update(uuid string, user *model.User) error {
	if _, exists := m.users[uuid]; !exists {
		return sql.ErrNoRows
	}
	user.UUID = uuid
	m.users[uuid] = user
	return nil
}

func (m *mockUserRepository) Delete(uuid string) error {
	if _, exists := m.users[uuid]; !exists {
		return sql.ErrNoRows
	}
	delete(m.users, uuid)
	return nil
}

// Tests for Create
func TestCreateUser_Success(t *testing.T) {
	// Given: Empty repository
	repo := newMockUserRepository()
	service := NewUserService(repo)
	
	newUser := &model.User{
		Username: "newuser",
		Email:    "newuser@example.com",
		FullName: "New User",
	}
	
	// When: Creating a new user
	err := service.Create(newUser)
	
	// Then: User should be created successfully
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if newUser.UUID == "" {
		t.Error("expected UUID to be set")
	}
	if newUser.ID == 0 {
		t.Error("expected ID to be set")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	// Given: Repository with existing user
	repo := newMockUserRepository()
	service := NewUserService(repo)
	
	existingUser := &model.User{
		UUID:     "existing-uuid",
		Username: "existinguser",
		Email:    "existing@example.com",
		FullName: "Existing User",
	}
	repo.users["existing-uuid"] = existingUser
	
	newUser := &model.User{
		Username: "existinguser", // Same username
		Email:    "new@example.com",
		FullName: "New User",
	}
	
	// When: Trying to create user with duplicate username
	err := service.Create(newUser)
	
	// Then: Should return error
	if err == nil {
		t.Error("expected error for duplicate username")
	}
	if err.Error() != "username already exists" {
		t.Errorf("expected 'username already exists', got %v", err)
	}
}

// Tests for Update
func TestUpdateUser_Success(t *testing.T) {
	// Given: Repository with existing user
	repo := newMockUserRepository()
	service := NewUserService(repo)
	
	existingUser := &model.User{
		ID:       1,
		UUID:     "test-uuid",
		Username: "oldusername",
		Email:    "old@example.com",
		FullName: "Old Name",
	}
	repo.users["test-uuid"] = existingUser
	
	updatedUser := &model.User{
		Username: "newusername",
		Email:    "new@example.com",
		FullName: "New Name",
	}
	
	// When: Updating the user
	err := service.Update("test-uuid", updatedUser)
	
	// Then: User should be updated successfully
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	user, _ := repo.GetByUUID("test-uuid")
	if user.Username != "newusername" {
		t.Errorf("expected username 'newusername', got %s", user.Username)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	// Given: Empty repository
	repo := newMockUserRepository()
	service := NewUserService(repo)
	
	updatedUser := &model.User{
		Username: "newusername",
		Email:    "new@example.com",
		FullName: "New Name",
	}
	
	// When: Trying to update non-existent user
	err := service.Update("non-existent-uuid", updatedUser)
	
	// Then: Should return error
	if err == nil {
		t.Error("expected error for non-existent user")
	}
	if err.Error() != "users not found" {
		t.Errorf("expected 'users not found', got %v", err)
	}
}

// Tests for Delete
func TestDeleteUser_Success(t *testing.T) {
	// Given: Repository with existing user
	repo := newMockUserRepository()
	service := NewUserService(repo)
	
	existingUser := &model.User{
		ID:       1,
		UUID:     "test-uuid",
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
	}
	repo.users["test-uuid"] = existingUser
	
	// When: Deleting the user
	err := service.Delete("test-uuid")
	
	// Then: User should be deleted successfully
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	_, err = repo.GetByUUID("test-uuid")
	if err != sql.ErrNoRows {
		t.Error("expected user to be deleted")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	// Given: Empty repository
	repo := newMockUserRepository()
	service := NewUserService(repo)
	
	// When: Trying to delete non-existent user
	err := service.Delete("non-existent-uuid")
	
	// Then: Should return error
	if err == nil {
		t.Error("expected error for non-existent user")
	}
	if err.Error() != "users not found" {
		t.Errorf("expected 'users not found', got %v", err)
	}
}