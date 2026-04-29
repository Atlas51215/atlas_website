package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/claude/blog/internal/handler"
	"github.com/claude/blog/internal/model"
)

func TestPostView_Published_OK(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "pv-author1")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, u.ID, "hello-post", "published")

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/general/" + p.Slug)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, p.Title) {
		t.Errorf("expected post title %q in response", p.Title)
	}
}

func TestPostView_RendersMarkdown(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "pv-author2")
	catID := blogCatID(t, db)
	p, err := model.CreatePost(db, catID, u.ID, "Markdown Post", "markdown-post",
		"## Hello\n\nThis is **bold**.", "", "published")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/general/" + p.Slug)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "<h2") {
		t.Error("expected rendered Markdown heading")
	}
	if !strings.Contains(body, "<strong>bold</strong>") {
		t.Error("expected rendered Markdown bold")
	}
}

func TestPostView_ShowsBreadcrumb(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "pv-author3")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, u.ID, "breadcrumb-post", "published")

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/general/" + p.Slug)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if !strings.Contains(body, `href="/blog"`) {
		t.Error("expected breadcrumb link to /blog")
	}
	if !strings.Contains(body, `href="/blog/general"`) {
		t.Error("expected breadcrumb link to /blog/general")
	}
}

func TestPostView_Draft_Returns404ForVisitor(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "pv-author4")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, u.ID, "secret-draft", "draft")

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/general/" + p.Slug)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for draft as visitor, got %d", resp.StatusCode)
	}
}

func TestPostView_Deleted_Returns404ForVisitor(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "pv-author5")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, u.ID, "deleted-post", "published")
	if err := model.SoftDeletePost(db, p.ID, u.ID); err != nil {
		t.Fatalf("SoftDeletePost: %v", err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/general/" + p.Slug)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for deleted post as visitor, got %d", resp.StatusCode)
	}
}

func TestPostView_Draft_VisibleToCurator(t *testing.T) {
	db := openTestDB(t)
	author := makeUser(t, db, "pv-author6")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, author.ID, "curator-draft", "draft")

	curator := makeUser(t, db, "pv-curator1")
	if err := model.UpdateUserRole(db, curator.ID, "curator"); err != nil {
		t.Fatal(err)
	}
	token, err := model.CreateSession(db, curator.ID)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/blog/general/"+p.Slug, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for draft as curator, got %d", resp.StatusCode)
	}
}

func TestPostView_Draft_VisibleToAuthor(t *testing.T) {
	db := openTestDB(t)
	author := makeUser(t, db, "pv-author7")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, author.ID, "author-draft", "draft")

	token, err := model.CreateSession(db, author.ID)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/blog/general/"+p.Slug, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for draft as author, got %d", resp.StatusCode)
	}
}

func TestPostView_Draft_Returns404ForOtherUser(t *testing.T) {
	db := openTestDB(t)
	author := makeUser(t, db, "pv-author8")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, author.ID, "other-draft", "draft")

	otherUser := makeUser(t, db, "pv-other1")
	token, err := model.CreateSession(db, otherUser.ID)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/blog/general/"+p.Slug, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for draft as unrelated user, got %d", resp.StatusCode)
	}
}

func TestPostView_UnknownCategory_Returns404(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/no-such-cat/some-slug")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown category, got %d", resp.StatusCode)
	}
}

func TestPostView_UnknownSlug_Returns404(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blog/general/no-such-post")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown slug, got %d", resp.StatusCode)
	}
}

func TestPostView_ReviewWithRating(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "pv-author9")

	cat, err := model.CategoryBySlugAndGroup(db, "movies", "reviews")
	if err != nil || cat == nil {
		t.Fatal("reviews/movies category not found")
	}

	p, err := model.CreatePost(db, cat.ID, u.ID, "Great Film", "great-film",
		"Great movie.", `{"rating":8.5}`, "published")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/reviews/movies/" + p.Slug)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "8.5") {
		t.Error("expected rating 8.5 in response")
	}
	if !strings.Contains(body, "★") {
		t.Error("expected star icon in response")
	}
}
