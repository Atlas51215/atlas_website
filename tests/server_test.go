package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/claude/blog/internal/handler"
)

func TestMain(m *testing.M) {
	// Change to project root so static files and migrations are found.
	os.Chdir("..")
	os.Exit(m.Run())
}

func TestHome_OK(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	for _, want := range []string{
		"<!doctype html>",
		`id="main-content"`,
		`class="sidebar"`,
		`hx-get="/search"`,
		`href="/"`,
		`href="/about"`,
		`/blog/`,
		`/reviews/`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain %q", want)
		}
	}
}

func TestHome_NavSubCategories(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)

	blogCats := []string{"movies", "games", "tv", "products", "general", "dev", "tech"}
	for _, slug := range blogCats {
		want := "/blog/" + slug
		if !strings.Contains(body, want) {
			t.Errorf("expected blog sub-category link %q in nav", want)
		}
	}

	reviewCats := []string{"movies", "games", "tv", "products"}
	for _, slug := range reviewCats {
		want := "/reviews/" + slug
		if !strings.Contains(body, want) {
			t.Errorf("expected reviews sub-category link %q in nav", want)
		}
	}
}

func TestHTMX_StaticFile(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/htmx.min.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for htmx.min.js, got %d", resp.StatusCode)
	}
}

func TestStatic_NotFound(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/nonexistent.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
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
