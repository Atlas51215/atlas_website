package model

import (
	"database/sql"
	"fmt"
)

// Follow creates a follow relationship from followerID to curatorID.
// Does nothing if the relationship already exists.
func Follow(db *sql.DB, followerID, curatorID int64) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO follows (follower_id, curator_id)
		VALUES (?, ?)`, followerID, curatorID)
	if err != nil {
		return fmt.Errorf("Follow: %w", err)
	}
	return nil
}

// Unfollow removes the follow relationship from followerID to curatorID.
func Unfollow(db *sql.DB, followerID, curatorID int64) error {
	_, err := db.Exec(`
		DELETE FROM follows WHERE follower_id = ? AND curator_id = ?`,
		followerID, curatorID)
	if err != nil {
		return fmt.Errorf("Unfollow: %w", err)
	}
	return nil
}

// IsFollowing returns true if followerID follows curatorID.
func IsFollowing(db *sql.DB, followerID, curatorID int64) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM follows WHERE follower_id = ? AND curator_id = ?`,
		followerID, curatorID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("IsFollowing: %w", err)
	}
	return count > 0, nil
}

// FollowedCurators returns the users that followerID follows, most recently followed first.
func FollowedCurators(db *sql.DB, followerID int64) ([]User, error) {
	rows, err := db.Query(`
		SELECT u.id, u.username, u.email, u.password_hash, u.role, u.bio,
		       u.theme, u.email_notify_posts, u.email_notify_replies,
		       u.is_banned, u.verified_at, u.last_active_at, u.created_at
		FROM users u
		JOIN follows f ON f.curator_id = u.id
		WHERE f.follower_id = ?
		ORDER BY f.created_at DESC`, followerID)
	if err != nil {
		return nil, fmt.Errorf("FollowedCurators: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var verifiedAt sql.NullTime
		var notifyPosts, notifyReplies, isBanned int
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role,
			&u.Bio, &u.Theme,
			&notifyPosts, &notifyReplies, &isBanned,
			&verifiedAt, &u.LastActiveAt, &u.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("FollowedCurators scan: %w", err)
		}
		u.EmailNotifyPosts = notifyPosts == 1
		u.EmailNotifyReplies = notifyReplies == 1
		u.IsBanned = isBanned == 1
		if verifiedAt.Valid {
			u.VerifiedAt = &verifiedAt.Time
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
