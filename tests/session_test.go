package tests

import (
	"testing"
	"time"

	"github.com/claude/blog/internal/model"
)

func TestCreateSession(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "sess_alice", "sess_alice@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	token, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if len(token) < 20 {
		t.Errorf("token looks too short: %q", token)
	}
}

func TestCreateSession_TokensAreUnique(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "sess_bob", "sess_bob@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	t1, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("first CreateSession: %v", err)
	}
	t2, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("second CreateSession: %v", err)
	}
	if t1 == t2 {
		t.Error("expected unique tokens, got duplicates")
	}
}

func TestSessionUser_Valid(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "sess_carol", "sess_carol@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	token, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	found, err := model.SessionUser(db, token)
	if err != nil {
		t.Fatalf("SessionUser: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.ID != u.ID {
		t.Errorf("user ID: got %d, want %d", found.ID, u.ID)
	}
	if found.Username != "sess_carol" {
		t.Errorf("username: got %q, want %q", found.Username, "sess_carol")
	}
}

func TestSessionUser_UnknownToken(t *testing.T) {
	db := openTestDB(t)

	_, err := model.SessionUser(db, "nonexistent-token")
	if err != model.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionUser_ExpiredToken(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "sess_dave", "sess_dave@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Insert a session that already expired.
	expired := time.Now().UTC().Add(-1 * time.Hour)
	_, err = db.Exec(
		`INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		"expired-token", u.ID, expired.Add(-24*time.Hour), expired,
	)
	if err != nil {
		t.Fatalf("insert expired session: %v", err)
	}

	_, err = model.SessionUser(db, "expired-token")
	if err != model.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for expired token, got %v", err)
	}

	// Verify the expired session was deleted.
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE token = 'expired-token'`).Scan(&count)
	if count != 0 {
		t.Error("expected expired session to be deleted, but it still exists")
	}
}

func TestDeleteSession(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "sess_eve", "sess_eve@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	token, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := model.DeleteSession(db, token); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	_, err = model.SessionUser(db, token)
	if err != model.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound after deletion, got %v", err)
	}
}

func TestDeleteSession_Idempotent(t *testing.T) {
	db := openTestDB(t)

	// Deleting a non-existent token should not error.
	if err := model.DeleteSession(db, "ghost-token"); err != nil {
		t.Errorf("DeleteSession on missing token: %v", err)
	}
}

func TestDeleteUserSessions(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "sess_frank", "sess_frank@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	tokens := make([]string, 3)
	for i := range tokens {
		tok, err := model.CreateSession(db, u.ID)
		if err != nil {
			t.Fatalf("CreateSession %d: %v", i, err)
		}
		tokens[i] = tok
	}

	if err := model.DeleteUserSessions(db, u.ID); err != nil {
		t.Fatalf("DeleteUserSessions: %v", err)
	}

	for _, tok := range tokens {
		_, err := model.SessionUser(db, tok)
		if err != model.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound for token %q after DeleteUserSessions, got %v", tok, err)
		}
	}
}

func TestDeleteUserSessions_OtherUsersUnaffected(t *testing.T) {
	db := openTestDB(t)

	u1, err := model.CreateUser(db, "sess_grace", "sess_grace@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser u1: %v", err)
	}
	u2, err := model.CreateUser(db, "sess_hank", "sess_hank@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser u2: %v", err)
	}

	tok1, err := model.CreateSession(db, u1.ID)
	if err != nil {
		t.Fatalf("CreateSession u1: %v", err)
	}
	tok2, err := model.CreateSession(db, u2.ID)
	if err != nil {
		t.Fatalf("CreateSession u2: %v", err)
	}

	if err := model.DeleteUserSessions(db, u1.ID); err != nil {
		t.Fatalf("DeleteUserSessions: %v", err)
	}

	if _, err := model.SessionUser(db, tok1); err != model.ErrSessionNotFound {
		t.Errorf("u1 session should be gone, got %v", err)
	}
	if _, err := model.SessionUser(db, tok2); err != nil {
		t.Errorf("u2 session should still be valid, got %v", err)
	}
}
