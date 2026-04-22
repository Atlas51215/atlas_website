package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userContextKey).(*User)
	return u
}

func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// Sentinel errors returned when a UNIQUE constraint is violated.
var (
	ErrDuplicateUsername = errors.New("username already taken")
	ErrDuplicateEmail    = errors.New("email already registered")
)

type User struct {
	ID                 int64
	Username           string
	Email              string
	PasswordHash       string
	Role               string // "curator" | "moderator" | "verified" | "unverified"
	Bio                string
	Theme              string // "dark" | "light"
	EmailNotifyPosts   bool
	EmailNotifyReplies bool
	IsBanned           bool
	VerifiedAt         *time.Time
	LastActiveAt       time.Time
	CreatedAt          time.Time
}

// CreateUser inserts a new user with a bcrypt-hashed password and role "unverified".
// Returns ErrDuplicateUsername or ErrDuplicateEmail on constraint violations.
func CreateUser(db *sql.DB, username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var id int64
	err = db.QueryRow(`
		INSERT INTO users (username, email, password_hash, role)
		VALUES (?, ?, ?, 'unverified')
		RETURNING id`,
		username, email, string(hash),
	).Scan(&id)
	if err != nil {
		if msg := err.Error(); strings.Contains(msg, "users.username") {
			return nil, ErrDuplicateUsername
		} else if strings.Contains(msg, "users.email") {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("CreateUser: %w", err)
	}

	return UserByID(db, id)
}

// UserByID returns the user with the given ID, or nil if not found.
func UserByID(db *sql.DB, id int64) (*User, error) {
	row := db.QueryRow(`
		SELECT id, username, email, password_hash, role, bio, theme,
		       email_notify_posts, email_notify_replies, is_banned,
		       verified_at, last_active_at, created_at
		FROM users WHERE id = ?`, id)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("UserByID: %w", err)
	}
	return u, nil
}

// UserByEmail returns the user with the given email, or nil if not found.
func UserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow(`
		SELECT id, username, email, password_hash, role, bio, theme,
		       email_notify_posts, email_notify_replies, is_banned,
		       verified_at, last_active_at, created_at
		FROM users WHERE email = ?`, email)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("UserByEmail: %w", err)
	}
	return u, nil
}

// UserByUsername returns the user with the given username, or nil if not found.
func UserByUsername(db *sql.DB, username string) (*User, error) {
	row := db.QueryRow(`
		SELECT id, username, email, password_hash, role, bio, theme,
		       email_notify_posts, email_notify_replies, is_banned,
		       verified_at, last_active_at, created_at
		FROM users WHERE username = ?`, username)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("UserByUsername: %w", err)
	}
	return u, nil
}

// CheckPassword returns true if password matches the user's stored hash.
func CheckPassword(u *User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// UpdateUserRole changes the role of the user with the given ID.
func UpdateUserRole(db *sql.DB, id int64, role string) error {
	_, err := db.Exec(`UPDATE users SET role = ? WHERE id = ?`, role, id)
	if err != nil {
		return fmt.Errorf("UpdateUserRole: %w", err)
	}
	return nil
}

// UpdateLastActive sets last_active_at to now for the given user ID.
func UpdateLastActive(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE users SET last_active_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("UpdateLastActive: %w", err)
	}
	return nil
}

func scanUser(row *sql.Row) (*User, error) {
	var u User
	var verifiedAt sql.NullTime
	var notifyPosts, notifyReplies, isBanned int
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role,
		&u.Bio, &u.Theme,
		&notifyPosts, &notifyReplies, &isBanned,
		&verifiedAt, &u.LastActiveAt, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.EmailNotifyPosts = notifyPosts == 1
	u.EmailNotifyReplies = notifyReplies == 1
	u.IsBanned = isBanned == 1
	if verifiedAt.Valid {
		u.VerifiedAt = &verifiedAt.Time
	}
	return &u, nil
}
