package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/claude/blog/internal/blog"
)

func TestMain(m *testing.M) {
	// Change to project root so templates and static files are found.
	os.Chdir("..")
	blog.Init("templates/*.html")
	os.Exit(m.Run())
}

func TestHandleIndex_OK(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Hello, World!") {
		t.Error("index page should contain the post title")
	}
	if !strings.Contains(body, `/post/hello-world`) {
		t.Error("index page should link to the post")
	}
}

func TestHandleIndex_NotFoundForOtherPaths(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandlePost_ExistingSlug(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/post/hello-world")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Hello, World!") {
		t.Error("post page should contain the post title")
	}
	if !strings.Contains(body, "Welcome to my blog") {
		t.Error("post page should contain the post body")
	}
}

func TestHandlePost_NonexistentSlug(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/post/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandlePost_EmptySlug(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/post/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404 for empty slug, got %d", resp.StatusCode)
	}
}

func TestHandleIndex_DisplaysMultiplePosts(t *testing.T) {
	restore := blog.SetPosts([]blog.Post{
		{
			Slug:      "hello-world",
			Title:     "Hello, World!",
			Body:      "<p>First post.</p>",
			CreatedAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Slug:      "second-post",
			Title:     "Second Post",
			Body:      "<p>Another post.</p>",
			CreatedAt: time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		},
	})
	defer restore()

	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Hello, World!") {
		t.Error("index should contain first post title")
	}
	if !strings.Contains(body, "Second Post") {
		t.Error("index should contain second post title")
	}
}

func TestHandleIndex_EmptyPosts(t *testing.T) {
	restore := blog.SetPosts([]blog.Post{})
	defer restore()

	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "No posts yet.") {
		t.Error("empty index should show 'No posts yet.' message")
	}
}

func TestPostStruct(t *testing.T) {
	p := blog.Post{
		Slug:      "test-slug",
		Title:     "Test Title",
		Body:      "<p>Test body</p>",
		CreatedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	if p.Slug != "test-slug" {
		t.Errorf("expected slug 'test-slug', got %q", p.Slug)
	}
	if p.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", p.Title)
	}
	if p.Body != "<p>Test body</p>" {
		t.Errorf("expected body '<p>Test body</p>', got %q", p.Body)
	}
	if p.CreatedAt.Year() != 2026 {
		t.Errorf("expected year 2026, got %d", p.CreatedAt.Year())
	}
}

func TestHandlePost_DateRendered(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/post/hello-world")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if !strings.Contains(body, "April 1, 2026") {
		t.Error("post page should render the formatted date")
	}
}

func TestStaticFileServer(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/style.css")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for style.css, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/css") {
		t.Error("style.css should be served with text/css content type")
	}
}

func TestStaticFileServer_NotFound(t *testing.T) {
	srv := httptest.NewServer(blog.NewMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/nonexistent.css")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
