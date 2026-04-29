package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

var ErrDuplicateSlug = errors.New("slug already exists in this category")

type Post struct {
	ID          int64
	CategoryID  int64
	AuthorID    int64
	Title       string
	Slug        string
	Body        string
	ExtraData   string // JSON blob, empty when not set
	Status      string // "draft" | "published"
	IsDeleted   bool
	DeletedBy   *int64
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Joined fields
	AuthorUsername string
	CategorySlug   string
	CategoryName   string
	GroupName      string
}

// GenerateSlug converts a title into a URL-safe slug: lowercase, any run of
// non-alphanumeric characters becomes a single hyphen, leading/trailing hyphens
// are trimmed.
func GenerateSlug(title string) string {
	s := strings.ToLower(title)
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

const postSelect = `
	SELECT p.id, p.category_id, p.author_id, p.title, p.slug, p.body,
	       p.extra_data, p.status, p.is_deleted, p.deleted_by,
	       p.published_at, p.created_at, p.updated_at,
	       u.username, c.slug, c.name, c.group_name
	FROM posts p
	JOIN users u ON u.id = p.author_id
	JOIN categories c ON c.id = p.category_id`

// CreatePost inserts a new post. Sets published_at when status is "published".
// Returns ErrDuplicateSlug if the slug already exists in the category.
func CreatePost(db *sql.DB, categoryID, authorID int64, title, slug, body, extraData, status string) (*Post, error) {
	var publishedAt *time.Time
	if status == "published" {
		now := time.Now().UTC()
		publishedAt = &now
	}

	var id int64
	err := db.QueryRow(`
		INSERT INTO posts (category_id, author_id, title, slug, body, extra_data, status, published_at)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?)
		RETURNING id`,
		categoryID, authorID, title, slug, body, extraData, status, publishedAt,
	).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "posts.slug") {
			return nil, ErrDuplicateSlug
		}
		return nil, fmt.Errorf("CreatePost: %w", err)
	}
	return PostByID(db, id)
}

// UpdatePost updates a post's editable fields. Sets published_at when the post
// transitions from draft to published for the first time.
// Returns ErrDuplicateSlug if the new slug conflicts with another post in the same category.
func UpdatePost(db *sql.DB, id int64, title, slug, body, extraData, status string) (*Post, error) {
	current, err := PostByID(db, id)
	if err != nil {
		return nil, fmt.Errorf("UpdatePost: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("UpdatePost: post %d not found", id)
	}

	var q string
	var args []any
	if current.Status == "draft" && status == "published" {
		q = `UPDATE posts
		     SET title = ?, slug = ?, body = ?, extra_data = NULLIF(?, ''),
		         status = ?, published_at = CURRENT_TIMESTAMP,
		         updated_at = CURRENT_TIMESTAMP
		     WHERE id = ?`
		args = []any{title, slug, body, extraData, status, id}
	} else {
		q = `UPDATE posts
		     SET title = ?, slug = ?, body = ?, extra_data = NULLIF(?, ''),
		         status = ?, updated_at = CURRENT_TIMESTAMP
		     WHERE id = ?`
		args = []any{title, slug, body, extraData, status, id}
	}

	if _, err := db.Exec(q, args...); err != nil {
		if strings.Contains(err.Error(), "posts.slug") {
			return nil, ErrDuplicateSlug
		}
		return nil, fmt.Errorf("UpdatePost: %w", err)
	}
	return PostByID(db, id)
}

// PostByID returns the post with the given ID, or nil if not found.
func PostByID(db *sql.DB, id int64) (*Post, error) {
	row := db.QueryRow(postSelect+` WHERE p.id = ?`, id)
	p, err := scanPost(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("PostByID: %w", err)
	}
	return p, nil
}

// PostBySlug returns a published, non-deleted post by category and slug.
func PostBySlug(db *sql.DB, categoryID int64, slug string) (*Post, error) {
	row := db.QueryRow(postSelect+`
		WHERE p.category_id = ? AND p.slug = ?
		  AND p.status = 'published' AND p.is_deleted = 0`, categoryID, slug)
	p, err := scanPost(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("PostBySlug: %w", err)
	}
	return p, nil
}

// PostsByCategory returns published, non-deleted posts for a category, newest first.
// page is 1-based. Returns the posts and the total count of matching posts.
func PostsByCategory(db *sql.DB, categoryID int64, page, pageSize int) ([]Post, int, error) {
	var total int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM posts
		WHERE category_id = ? AND status = 'published' AND is_deleted = 0`, categoryID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("PostsByCategory count: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := db.Query(postSelect+`
		WHERE p.category_id = ? AND p.status = 'published' AND p.is_deleted = 0
		ORDER BY p.published_at DESC
		LIMIT ? OFFSET ?`, categoryID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("PostsByCategory: %w", err)
	}
	defer rows.Close()

	posts, err := scanPosts(rows)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

// PostsByAuthor returns published, non-deleted posts by an author, newest first.
// page is 1-based. Returns the posts and the total count of matching posts.
func PostsByAuthor(db *sql.DB, authorID int64, page, pageSize int) ([]Post, int, error) {
	var total int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM posts
		WHERE author_id = ? AND status = 'published' AND is_deleted = 0`, authorID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("PostsByAuthor count: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := db.Query(postSelect+`
		WHERE p.author_id = ? AND p.status = 'published' AND p.is_deleted = 0
		ORDER BY p.published_at DESC
		LIMIT ? OFFSET ?`, authorID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("PostsByAuthor: %w", err)
	}
	defer rows.Close()

	posts, err := scanPosts(rows)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

// PostsByGroup returns published, non-deleted posts across all categories in a group, newest first.
// page is 1-based. Returns the posts and the total count of matching posts.
func PostsByGroup(db *sql.DB, groupName string, page, pageSize int) ([]Post, int, error) {
	var total int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM posts p
		JOIN categories c ON c.id = p.category_id
		WHERE c.group_name = ? AND p.status = 'published' AND p.is_deleted = 0`, groupName,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("PostsByGroup count: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := db.Query(postSelect+`
		WHERE c.group_name = ? AND p.status = 'published' AND p.is_deleted = 0
		ORDER BY p.published_at DESC
		LIMIT ? OFFSET ?`, groupName, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("PostsByGroup: %w", err)
	}
	defer rows.Close()

	posts, err := scanPosts(rows)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

// RecentPosts returns the most recently published, non-deleted posts across all categories.
func RecentPosts(db *sql.DB, limit int) ([]Post, error) {
	rows, err := db.Query(postSelect+`
		WHERE p.status = 'published' AND p.is_deleted = 0
		ORDER BY p.published_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("RecentPosts: %w", err)
	}
	defer rows.Close()
	return scanPosts(rows)
}

// SoftDeletePost marks a post as deleted without removing it from the database.
func SoftDeletePost(db *sql.DB, id, deletedBy int64) error {
	_, err := db.Exec(`
		UPDATE posts SET is_deleted = 1, deleted_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, deletedBy, id)
	if err != nil {
		return fmt.Errorf("SoftDeletePost: %w", err)
	}
	return nil
}

// RestorePost clears the soft-delete flag on a post.
func RestorePost(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		UPDATE posts SET is_deleted = 0, deleted_by = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("RestorePost: %w", err)
	}
	return nil
}

// HardDeletePost permanently removes a post from the database.
func HardDeletePost(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM posts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("HardDeletePost: %w", err)
	}
	return nil
}

func scanPost(s scanner) (*Post, error) {
	var p Post
	var extraData sql.NullString
	var deletedBy sql.NullInt64
	var publishedAt sql.NullTime
	var isDeleted int
	err := s.Scan(
		&p.ID, &p.CategoryID, &p.AuthorID, &p.Title, &p.Slug, &p.Body,
		&extraData, &p.Status, &isDeleted, &deletedBy,
		&publishedAt, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.CategorySlug, &p.CategoryName, &p.GroupName,
	)
	if err != nil {
		return nil, err
	}
	p.IsDeleted = isDeleted == 1
	if deletedBy.Valid {
		p.DeletedBy = &deletedBy.Int64
	}
	if publishedAt.Valid {
		p.PublishedAt = &publishedAt.Time
	}
	if extraData.Valid {
		p.ExtraData = extraData.String
	}
	return &p, nil
}

func scanPosts(rows *sql.Rows) ([]Post, error) {
	var posts []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, *p)
	}
	return posts, rows.Err()
}
