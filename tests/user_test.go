package tests

import (
	"testing"

	"github.com/claude/blog/internal/model"
)

func TestCreateUser_Success(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "alice", "alice@example.com", "hunter2")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if u.Username != "alice" {
		t.Errorf("username: got %q, want %q", u.Username, "alice")
	}
	if u.Email != "alice@example.com" {
		t.Errorf("email: got %q, want %q", u.Email, "alice@example.com")
	}
	if u.Role != "unverified" {
		t.Errorf("role: got %q, want %q", u.Role, "unverified")
	}
	if u.PasswordHash == "hunter2" {
		t.Error("password must be stored as a hash, not plaintext")
	}
	if u.PasswordHash == "" {
		t.Error("password hash must not be empty")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := openTestDB(t)

	if _, err := model.CreateUser(db, "bob", "bob@example.com", "pass"); err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}

	_, err := model.CreateUser(db, "bob", "different@example.com", "pass")
	if err != model.ErrDuplicateUsername {
		t.Errorf("expected ErrDuplicateUsername, got %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db := openTestDB(t)

	if _, err := model.CreateUser(db, "carol", "shared@example.com", "pass"); err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}

	_, err := model.CreateUser(db, "differentname", "shared@example.com", "pass")
	if err != model.ErrDuplicateEmail {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestCheckPassword(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "dave", "dave@example.com", "correct-password")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if !model.CheckPassword(u, "correct-password") {
		t.Error("CheckPassword returned false for correct password")
	}
	if model.CheckPassword(u, "wrong-password") {
		t.Error("CheckPassword returned true for wrong password")
	}
}

func TestUserByID(t *testing.T) {
	db := openTestDB(t)

	created, err := model.CreateUser(db, "eve", "eve@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	found, err := model.UserByID(db, created.ID)
	if err != nil {
		t.Fatalf("UserByID: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Username != "eve" {
		t.Errorf("username: got %q, want %q", found.Username, "eve")
	}
}

func TestUserByID_NotFound(t *testing.T) {
	db := openTestDB(t)

	u, err := model.UserByID(db, 99999)
	if err != nil {
		t.Fatalf("UserByID: %v", err)
	}
	if u != nil {
		t.Errorf("expected nil for missing user, got %+v", u)
	}
}

func TestUserByEmail(t *testing.T) {
	db := openTestDB(t)

	if _, err := model.CreateUser(db, "frank", "frank@example.com", "pass"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := model.UserByEmail(db, "frank@example.com")
	if err != nil {
		t.Fatalf("UserByEmail: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Username != "frank" {
		t.Errorf("username: got %q, want %q", u.Username, "frank")
	}
}

func TestUserByUsername(t *testing.T) {
	db := openTestDB(t)

	if _, err := model.CreateUser(db, "grace", "grace@example.com", "pass"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := model.UserByUsername(db, "grace")
	if err != nil {
		t.Fatalf("UserByUsername: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Email != "grace@example.com" {
		t.Errorf("email: got %q, want %q", u.Email, "grace@example.com")
	}
}

func TestUpdateUserRole(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "hank", "hank@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.Role != "unverified" {
		t.Fatalf("expected initial role 'unverified', got %q", u.Role)
	}

	if err := model.UpdateUserRole(db, u.ID, "verified"); err != nil {
		t.Fatalf("UpdateUserRole: %v", err)
	}

	updated, err := model.UserByID(db, u.ID)
	if err != nil {
		t.Fatalf("UserByID: %v", err)
	}
	if updated.Role != "verified" {
		t.Errorf("role: got %q, want %q", updated.Role, "verified")
	}
}

func TestUpdateLastActive(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "ivy", "ivy@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := model.UpdateLastActive(db, u.ID); err != nil {
		t.Fatalf("UpdateLastActive: %v", err)
	}

	// Verify the update ran without error and the user still exists.
	found, err := model.UserByID(db, u.ID)
	if err != nil {
		t.Fatalf("UserByID after UpdateLastActive: %v", err)
	}
	if found == nil {
		t.Fatal("user not found after UpdateLastActive")
	}
}
