package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/claude/blog/internal/handler"
	"github.com/claude/blog/internal/model"
)

// noRedirectClient stops the HTTP client from following redirects so tests can
// inspect the 302 response directly.
var noRedirectClient = &http.Client{
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func postRegisterForm(t *testing.T, srv *httptest.Server, username, email, password, confirm string) *http.Response {
	t.Helper()
	form := url.Values{
		"username":         {username},
		"email":            {email},
		"password":         {password},
		"confirm_password": {confirm},
	}
	resp, err := noRedirectClient.PostForm(srv.URL+"/register", form)
	if err != nil {
		t.Fatalf("POST /register: %v", err)
	}
	return resp
}

func TestRegisterPage_GET(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/register")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	for _, want := range []string{
		`name="username"`,
		`name="email"`,
		`name="password"`,
		`name="confirm_password"`,
		`action="/register"`,
		`href="/login"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain %q", want)
		}
	}
}

func TestRegisterPage_GET_RedirectsIfLoggedIn(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	u, err := model.CreateUser(db, "reg_loggedin", "reg_loggedin@example.com", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	token, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/register", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	resp, err := noRedirectClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302 for logged-in user, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Errorf("expected redirect to /, got %q", loc)
	}
}

func TestRegisterSubmit_Valid(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postRegisterForm(t, srv, "newuser", "newuser@example.com", "securepass", "securepass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Errorf("expected redirect to /, got %q", loc)
	}

	// Session cookie must be set.
	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie in response")
	}
	if sessionCookie.Value == "" {
		t.Error("session cookie value must not be empty")
	}
	if !sessionCookie.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Error("session cookie must be SameSite=Strict")
	}
	if sessionCookie.MaxAge <= 0 {
		t.Error("session cookie must have a positive MaxAge")
	}

	// User must exist in DB.
	u, err := model.UserByUsername(db, "newuser")
	if err != nil {
		t.Fatalf("UserByUsername: %v", err)
	}
	if u == nil {
		t.Fatal("expected user to be created in DB")
	}
}

func TestRegisterSubmit_UsernameTooShort(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postRegisterForm(t, srv, "ab", "ok@example.com", "securepass", "securepass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "form-error") {
		t.Error("expected inline error in response body")
	}
}

func TestRegisterSubmit_InvalidEmail(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postRegisterForm(t, srv, "validuser", "not-an-email", "securepass", "securepass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "form-error") {
		t.Error("expected inline error in response body")
	}
}

func TestRegisterSubmit_PasswordTooShort(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postRegisterForm(t, srv, "validuser", "ok@example.com", "short", "short")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "form-error") {
		t.Error("expected inline error in response body")
	}
}

func TestRegisterSubmit_PasswordMismatch(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postRegisterForm(t, srv, "validuser", "ok@example.com", "securepass1", "securepass2")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "do not match") {
		t.Error("expected password mismatch error in body")
	}
}

func TestRegisterSubmit_DuplicateUsername(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	if _, err := model.CreateUser(db, "taken_user", "taken@example.com", "securepass"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	resp := postRegisterForm(t, srv, "taken_user", "other@example.com", "securepass", "securepass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "already taken") {
		t.Error("expected duplicate username error in body")
	}
}

func TestRegisterSubmit_DuplicateEmail(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	if _, err := model.CreateUser(db, "first_user", "shared@example.com", "securepass"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	resp := postRegisterForm(t, srv, "second_user", "shared@example.com", "securepass", "securepass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "already registered") {
		t.Error("expected duplicate email error in body")
	}
}

func TestRegisterSubmit_RepopulatesFields(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	// Submit with an invalid username so the form re-renders.
	resp := postRegisterForm(t, srv, "ab", "keep@example.com", "securepass", "securepass")
	defer resp.Body.Close()

	body := readBody(t, resp)
	if !strings.Contains(body, "keep@example.com") {
		t.Error("expected email field to be repopulated on validation error")
	}
}
