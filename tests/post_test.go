package tests

import (
	"database/sql"
	"testing"

	"github.com/claude/blog/internal/model"
)

// makeUser creates a test user with a unique username and email.
func makeUser(t *testing.T, db *sql.DB, username string) *model.User {
	t.Helper()
	u, err := model.CreateUser(db, username, username+"@example.com", "testpassword")
	if err != nil {
		t.Fatalf("makeUser(%q): %v", username, err)
	}
	return u
}

// makePost creates a test post with sensible defaults.
func makePost(t *testing.T, db *sql.DB, catID, authorID int64, slug, status string) *model.Post {
	t.Helper()
	p, err := model.CreatePost(db, catID, authorID, "Test Post", slug, "body text", "", status)
	if err != nil {
		t.Fatalf("makePost(%q): %v", slug, err)
	}
	return p
}

// blogCatID returns the ID of the blog/general category seeded by migrations.
func blogCatID(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	cat, err := model.CategoryBySlugAndGroup(db, "general", "blog")
	if err != nil || cat == nil {
		t.Fatal("could not find blog/general category")
	}
	return cat.ID
}

// --- GenerateSlug ---

func TestGenerateSlug_Basic(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello World", "hello-world"},
		{"  leading spaces  ", "leading-spaces"},
		{"Special!@# Chars", "special-chars"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"already-fine", "already-fine"},
		{"ALL CAPS", "all-caps"},
		{"123 Numbers", "123-numbers"},
		{"", ""},
	}
	for _, c := range cases {
		got := model.GenerateSlug(c.in)
		if got != c.want {
			t.Errorf("GenerateSlug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- CreatePost ---

func TestCreatePost_Success(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author1")
	catID := blogCatID(t, db)

	p, err := model.CreatePost(db, catID, u.ID, "My Post", "my-post", "body", "", "draft")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if p.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if p.Title != "My Post" {
		t.Errorf("title: got %q, want %q", p.Title, "My Post")
	}
	if p.AuthorUsername != "author1" {
		t.Errorf("author username: got %q, want %q", p.AuthorUsername, "author1")
	}
	if p.Status != "draft" {
		t.Errorf("status: got %q, want %q", p.Status, "draft")
	}
	if p.PublishedAt != nil {
		t.Error("draft post should have nil PublishedAt")
	}
}

func TestCreatePost_Published_SetsPublishedAt(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author2")
	catID := blogCatID(t, db)

	p, err := model.CreatePost(db, catID, u.ID, "Published Post", "published-post", "body", "", "published")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if p.PublishedAt == nil {
		t.Error("published post should have a non-nil PublishedAt")
	}
}

func TestCreatePost_DuplicateSlug(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author3")
	catID := blogCatID(t, db)

	if _, err := model.CreatePost(db, catID, u.ID, "Post One", "dup-slug", "body", "", "draft"); err != nil {
		t.Fatalf("first CreatePost: %v", err)
	}

	_, err := model.CreatePost(db, catID, u.ID, "Post Two", "dup-slug", "body", "", "draft")
	if err != model.ErrDuplicateSlug {
		t.Errorf("expected ErrDuplicateSlug, got %v", err)
	}
}

func TestCreatePost_ExtraData(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author4")

	cat, _ := model.CategoryBySlugAndGroup(db, "movies", "reviews")

	extra := `{"rating":8.5,"director":"Kubrick","release_year":1968}`
	p, err := model.CreatePost(db, cat.ID, u.ID, "2001", "2001", "great film", extra, "draft")
	if err != nil {
		t.Fatalf("CreatePost with extra_data: %v", err)
	}
	if p.ExtraData != extra {
		t.Errorf("extra_data: got %q, want %q", p.ExtraData, extra)
	}
}

// --- PostByID ---

func TestPostByID(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author5")
	catID := blogCatID(t, db)

	created := makePost(t, db, catID, u.ID, "by-id-slug", "draft")

	found, err := model.PostByID(db, created.ID)
	if err != nil {
		t.Fatalf("PostByID: %v", err)
	}
	if found == nil {
		t.Fatal("expected post, got nil")
	}
	if found.Title != "Test Post" {
		t.Errorf("title: got %q, want %q", found.Title, "Test Post")
	}
}

func TestPostByID_NotFound(t *testing.T) {
	db := openTestDB(t)

	p, err := model.PostByID(db, 99999)
	if err != nil {
		t.Fatalf("PostByID: %v", err)
	}
	if p != nil {
		t.Errorf("expected nil for missing post, got %+v", p)
	}
}

// --- PostBySlug ---

func TestPostBySlug_Found(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author6")
	catID := blogCatID(t, db)

	makePost(t, db, catID, u.ID, "slug-test", "published")

	p, err := model.PostBySlug(db, catID, "slug-test")
	if err != nil {
		t.Fatalf("PostBySlug: %v", err)
	}
	if p == nil {
		t.Fatal("expected post, got nil")
	}
	if p.Slug != "slug-test" {
		t.Errorf("slug: got %q, want %q", p.Slug, "slug-test")
	}
}

func TestPostBySlug_ExcludesDraft(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author7")
	catID := blogCatID(t, db)

	makePost(t, db, catID, u.ID, "draft-slug", "draft")

	p, err := model.PostBySlug(db, catID, "draft-slug")
	if err != nil {
		t.Fatalf("PostBySlug: %v", err)
	}
	if p != nil {
		t.Error("draft post should not be returned by PostBySlug")
	}
}

func TestPostBySlug_ExcludesDeleted(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author8")
	catID := blogCatID(t, db)

	p := makePost(t, db, catID, u.ID, "deleted-slug", "published")
	if err := model.SoftDeletePost(db, p.ID, u.ID); err != nil {
		t.Fatalf("SoftDeletePost: %v", err)
	}

	found, err := model.PostBySlug(db, catID, "deleted-slug")
	if err != nil {
		t.Fatalf("PostBySlug: %v", err)
	}
	if found != nil {
		t.Error("soft-deleted post should not be returned by PostBySlug")
	}
}

// --- UpdatePost ---

func TestUpdatePost_UpdatesFields(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author9")
	catID := blogCatID(t, db)

	p := makePost(t, db, catID, u.ID, "update-me", "draft")

	updated, err := model.UpdatePost(db, p.ID, "New Title", "new-slug", "new body", "", "draft")
	if err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("title: got %q, want %q", updated.Title, "New Title")
	}
	if updated.Slug != "new-slug" {
		t.Errorf("slug: got %q, want %q", updated.Slug, "new-slug")
	}
	if updated.PublishedAt != nil {
		t.Error("draft→draft update should not set PublishedAt")
	}
}

func TestUpdatePost_DraftToPublished_SetsPublishedAt(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author10")
	catID := blogCatID(t, db)

	p := makePost(t, db, catID, u.ID, "promote-me", "draft")
	if p.PublishedAt != nil {
		t.Fatal("draft post should not have PublishedAt before promotion")
	}

	updated, err := model.UpdatePost(db, p.ID, p.Title, p.Slug, p.Body, "", "published")
	if err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}
	if updated.PublishedAt == nil {
		t.Error("draft→published should set PublishedAt")
	}
}

func TestUpdatePost_PublishedToPublished_KeepsPublishedAt(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author11")
	catID := blogCatID(t, db)

	p := makePost(t, db, catID, u.ID, "already-pub", "published")
	original := *p.PublishedAt

	updated, err := model.UpdatePost(db, p.ID, "Changed Title", p.Slug, p.Body, "", "published")
	if err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}
	if updated.PublishedAt == nil {
		t.Fatal("published_at should still be set")
	}
	if !updated.PublishedAt.Equal(original) {
		t.Error("published→published update should not change PublishedAt")
	}
}

// --- PostsByCategory ---

func TestPostsByCategory_Pagination(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author12")
	catID := blogCatID(t, db)

	for i := range 5 {
		slug := "paged-post-" + string(rune('a'+i))
		makePost(t, db, catID, u.ID, slug, "published")
	}

	posts, total, err := model.PostsByCategory(db, catID, 1, 3)
	if err != nil {
		t.Fatalf("PostsByCategory page 1: %v", err)
	}
	if total != 5 {
		t.Errorf("total: got %d, want 5", total)
	}
	if len(posts) != 3 {
		t.Errorf("page 1 count: got %d, want 3", len(posts))
	}

	posts2, total2, err := model.PostsByCategory(db, catID, 2, 3)
	if err != nil {
		t.Fatalf("PostsByCategory page 2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("total page 2: got %d, want 5", total2)
	}
	if len(posts2) != 2 {
		t.Errorf("page 2 count: got %d, want 2", len(posts2))
	}
}

func TestPostsByCategory_ExcludesDraftAndDeleted(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author13")
	catID := blogCatID(t, db)

	makePost(t, db, catID, u.ID, "visible", "published")
	makePost(t, db, catID, u.ID, "hidden-draft", "draft")

	deleted := makePost(t, db, catID, u.ID, "hidden-deleted", "published")
	if err := model.SoftDeletePost(db, deleted.ID, u.ID); err != nil {
		t.Fatalf("SoftDeletePost: %v", err)
	}

	posts, total, err := model.PostsByCategory(db, catID, 1, 10)
	if err != nil {
		t.Fatalf("PostsByCategory: %v", err)
	}
	if total != 1 {
		t.Errorf("total: got %d, want 1", total)
	}
	if len(posts) != 1 || posts[0].Slug != "visible" {
		t.Errorf("expected only 'visible', got %v", posts)
	}
}

// --- PostsByAuthor ---

func TestPostsByAuthor_Pagination(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author14")
	catID := blogCatID(t, db)

	for i := range 4 {
		slug := "auth-post-" + string(rune('a'+i))
		makePost(t, db, catID, u.ID, slug, "published")
	}

	posts, total, err := model.PostsByAuthor(db, u.ID, 1, 3)
	if err != nil {
		t.Fatalf("PostsByAuthor page 1: %v", err)
	}
	if total != 4 {
		t.Errorf("total: got %d, want 4", total)
	}
	if len(posts) != 3 {
		t.Errorf("page 1 count: got %d, want 3", len(posts))
	}

	posts2, _, err := model.PostsByAuthor(db, u.ID, 2, 3)
	if err != nil {
		t.Fatalf("PostsByAuthor page 2: %v", err)
	}
	if len(posts2) != 1 {
		t.Errorf("page 2 count: got %d, want 1", len(posts2))
	}
}

// --- RecentPosts ---

func TestRecentPosts(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author15")
	catID := blogCatID(t, db)

	for i := range 5 {
		slug := "recent-" + string(rune('a'+i))
		makePost(t, db, catID, u.ID, slug, "published")
	}
	makePost(t, db, catID, u.ID, "recent-draft", "draft")

	posts, err := model.RecentPosts(db, 3)
	if err != nil {
		t.Fatalf("RecentPosts: %v", err)
	}
	if len(posts) != 3 {
		t.Errorf("expected 3 recent posts, got %d", len(posts))
	}
	for _, p := range posts {
		if p.Status != "published" {
			t.Errorf("RecentPosts returned non-published post: %q", p.Status)
		}
	}
}

// --- SoftDeletePost / RestorePost ---

func TestSoftDeletePost(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author16")
	catID := blogCatID(t, db)

	p := makePost(t, db, catID, u.ID, "to-soft-delete", "published")

	if err := model.SoftDeletePost(db, p.ID, u.ID); err != nil {
		t.Fatalf("SoftDeletePost: %v", err)
	}

	found, err := model.PostByID(db, p.ID)
	if err != nil {
		t.Fatalf("PostByID: %v", err)
	}
	if !found.IsDeleted {
		t.Error("expected IsDeleted=true after SoftDeletePost")
	}
	if found.DeletedBy == nil || *found.DeletedBy != u.ID {
		t.Errorf("DeletedBy: got %v, want %d", found.DeletedBy, u.ID)
	}
}

func TestRestorePost(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author17")
	catID := blogCatID(t, db)

	p := makePost(t, db, catID, u.ID, "to-restore", "published")
	if err := model.SoftDeletePost(db, p.ID, u.ID); err != nil {
		t.Fatalf("SoftDeletePost: %v", err)
	}

	if err := model.RestorePost(db, p.ID); err != nil {
		t.Fatalf("RestorePost: %v", err)
	}

	found, err := model.PostByID(db, p.ID)
	if err != nil {
		t.Fatalf("PostByID: %v", err)
	}
	if found.IsDeleted {
		t.Error("expected IsDeleted=false after RestorePost")
	}
	if found.DeletedBy != nil {
		t.Errorf("expected DeletedBy=nil after restore, got %v", found.DeletedBy)
	}
}

// --- HardDeletePost ---

func TestHardDeletePost(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "author18")
	catID := blogCatID(t, db)

	p := makePost(t, db, catID, u.ID, "to-hard-delete", "draft")

	if err := model.HardDeletePost(db, p.ID); err != nil {
		t.Fatalf("HardDeletePost: %v", err)
	}

	found, err := model.PostByID(db, p.ID)
	if err != nil {
		t.Fatalf("PostByID after hard delete: %v", err)
	}
	if found != nil {
		t.Error("expected nil after HardDeletePost")
	}
}
