package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/claude/blog/internal/handler"
	"github.com/claude/blog/internal/model"
)

// --- PostsByGroup model tests ---

func TestPostsByGroup_Empty(t *testing.T) {
	db := openTestDB(t)
	posts, total, err := model.PostsByGroup(db, "blog", 1, 10)
	if err != nil {
		t.Fatalf("PostsByGroup: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total=0, got %d", total)
	}
	if len(posts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(posts))
	}
}

func TestPostsByGroup_ReturnsOnlyPublished(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "groupauthor")
	catID := blogCatID(t, db)

	makePost(t, db, catID, u.ID, "draft-post", "draft")
	makePost(t, db, catID, u.ID, "pub-post", "published")

	posts, total, err := model.PostsByGroup(db, "blog", 1, 10)
	if err != nil {
		t.Fatalf("PostsByGroup: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].Slug != "pub-post" {
		t.Errorf("expected pub-post, got %q", posts[0].Slug)
	}
}

func TestPostsByGroup_Pagination(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "paginator")
	catID := blogCatID(t, db)

	for i := range 5 {
		makePost(t, db, catID, u.ID, "pg-post-"+string(rune('a'+i)), "published")
	}

	posts, total, err := model.PostsByGroup(db, "blog", 1, 3)
	if err != nil {
		t.Fatalf("PostsByGroup page 1: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total=5, got %d", total)
	}
	if len(posts) != 3 {
		t.Errorf("expected 3 posts on page 1, got %d", len(posts))
	}

	posts2, _, err := model.PostsByGroup(db, "blog", 2, 3)
	if err != nil {
		t.Fatalf("PostsByGroup page 2: %v", err)
	}
	if len(posts2) != 2 {
		t.Errorf("expected 2 posts on page 2, got %d", len(posts2))
	}
}

func TestPostsByGroup_OnlyGroupPosts(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "multigroup")

	blogCat, err := model.CategoryBySlugAndGroup(db, "general", "blog")
	if err != nil || blogCat == nil {
		t.Fatal("blog/general not found")
	}
	reviewCat, err := model.CategoryBySlugAndGroup(db, "movies", "reviews")
	if err != nil || reviewCat == nil {
		t.Fatal("reviews/movies not found")
	}

	makePost(t, db, blogCat.ID, u.ID, "blog-only", "published")
	makePost(t, db, reviewCat.ID, u.ID, "review-only", "published")

	blogPosts, blogTotal, err := model.PostsByGroup(db, "blog", 1, 10)
	if err != nil {
		t.Fatalf("blog group: %v", err)
	}
	if blogTotal != 1 || len(blogPosts) != 1 {
		t.Errorf("blog: expected 1 post, got total=%d len=%d", blogTotal, len(blogPosts))
	}

	revPosts, revTotal, err := model.PostsByGroup(db, "reviews", 1, 10)
	if err != nil {
		t.Fatalf("reviews group: %v", err)
	}
	if revTotal != 1 || len(revPosts) != 1 {
		t.Errorf("reviews: expected 1 post, got total=%d len=%d", revTotal, len(revPosts))
	}
}

// --- Handler tests ---

func TestBlogLanding_OK(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "<html") {
		t.Error("expected full HTML page")
	}
	if !strings.Contains(body, "Blog") {
		t.Error("expected heading to contain 'Blog'")
	}
}

func TestBlogLanding_WithPosts(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "blogger")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, u.ID, "hello-world", "published")

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if !strings.Contains(body, p.Title) {
		t.Errorf("expected post title %q in response body", p.Title)
	}
}

func TestBlogCategory_OK(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/general")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /blog/general, got %d", resp.StatusCode)
	}
}

func TestBlogCategory_NotFound(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/nonexistent-slug")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown category, got %d", resp.StatusCode)
	}
}

func TestReviewsLanding_OK(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/reviews")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Reviews") {
		t.Error("expected heading to contain 'Reviews'")
	}
}

func TestReviewsCategory_OK(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/reviews/movies")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /reviews/movies, got %d", resp.StatusCode)
	}
}

func TestReviewsCategory_NotFound(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/reviews/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestBlogLanding_HTMXRequest_ReturnsFragment(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/blog", nil)
	req.Header.Set("HX-Request", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if strings.Contains(body, "<html") {
		t.Error("HTMX response should not contain <html> wrapper")
	}
}

func TestBlogCategory_HTMXRequest_ReturnsFragment(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/blog/general", nil)
	req.Header.Set("HX-Request", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if strings.Contains(body, "<html") {
		t.Error("HTMX response should not contain <html> wrapper")
	}
}

func TestBlogLanding_Pagination(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "pgblogger")
	catID := blogCatID(t, db)
	for i := range 12 {
		makePost(t, db, catID, u.ID, "pg-post-"+string(rune('a'+i)), "published")
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog?page=2")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "pagination") {
		t.Error("expected pagination controls in response")
	}
	// Page 2 should not contain the page=1 link as the active page.
	if strings.Contains(body, `pagination-current" aria-current="page">1<`) {
		t.Error("page 2 response should not show page 1 as current")
	}
}
