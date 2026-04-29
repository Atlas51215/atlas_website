package model

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

var ErrSessionNotFound = errors.New("session not found or expired")

// CreateSession generates a secure random token, stores a session for userID
// that expires in 30 days, and returns the token.
func CreateSession(db *sql.DB, userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	now := time.Now().UTC()
	expiresAt := now.Add(30 * 24 * time.Hour)

	_, err := db.Exec(
		`INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		token, userID, now, expiresAt,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

// SessionUser looks up the user for a valid, non-expired session token.
// If the token is expired it is deleted before returning ErrSessionNotFound.
func SessionUser(db *sql.DB, token string) (*User, error) {
	now := time.Now().UTC()

	row := db.QueryRow(`
		SELECT u.id, u.username, u.email, u.password_hash, u.role, u.bio, u.theme,
		       u.email_notify_posts, u.email_notify_replies, u.is_banned,
		       u.verified_at, u.last_active_at, u.created_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > ?
	`, token, now)

	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		// Clean up the token if it exists but is expired.
		db.Exec(`DELETE FROM sessions WHERE token = ? AND expires_at <= ?`, token, now) //nolint:errcheck
		return nil, ErrSessionNotFound
	}
	return user, err
}

// DeleteSession removes a single session (logout).
func DeleteSession(db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// DeleteUserSessions removes all sessions for a user (log out everywhere).
func DeleteUserSessions(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}
