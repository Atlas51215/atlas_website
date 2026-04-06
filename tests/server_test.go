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
	// Change to project root so static files are found.
	os.Chdir("..")
	os.Exit(m.Run())
}

func TestHome_OK(t *testing.T) {
	srv := httptest.NewServer(handler.NewRouter("static"))
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
	if !strings.Contains(body, "Atlas is running") {
		t.Errorf("expected body to contain 'Atlas is running', got: %s", body)
	}
}

func TestHTMX_StaticFile(t *testing.T) {
	srv := httptest.NewServer(handler.NewRouter("static"))
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
	srv := httptest.NewServer(handler.NewRouter("static"))
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
