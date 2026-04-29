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

func postLoginForm(t *testing.T, srv *httptest.Server, email, password string) *http.Response {
	t.Helper()
	form := url.Values{
		"email":    {email},
		"password": {password},
	}
	resp, err := noRedirectClient.PostForm(srv.URL+"/login", form)
	if err != nil {
		t.Fatalf("POST /login: %v", err)
	}
	return resp
}

func TestLoginPage_GET(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/login")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	for _, want := range []string{
		`name="email"`,
		`name="password"`,
		`action="/login"`,
		`href="/register"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain %q", want)
		}
	}
}

func TestLoginPage_GET_RedirectsIfLoggedIn(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	u, err := model.CreateUser(db, "login_loggedin", "login_loggedin@example.com", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	token, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/login", nil)
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

func TestLoginSubmit_Valid(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	if _, err := model.CreateUser(db, "loginuser", "login@example.com", "securepass"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	resp := postLoginForm(t, srv, "login@example.com", "securepass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Errorf("expected redirect to /, got %q", loc)
	}

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
}

func TestLoginSubmit_WrongPassword(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	if _, err := model.CreateUser(db, "wrongpass_user", "wrongpass@example.com", "correctpass"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	resp := postLoginForm(t, srv, "wrongpass@example.com", "wrongpass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "Invalid email or password") {
		t.Error("expected 'Invalid email or password' error in body")
	}
}

func TestLoginSubmit_UnknownEmail(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postLoginForm(t, srv, "nobody@example.com", "anypassword")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "Invalid email or password") {
		t.Error("expected generic error; must not reveal which field is wrong")
	}
}

func TestLoginSubmit_BannedUser(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	u, err := model.CreateUser(db, "banned_user", "banned@example.com", "securepass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if _, err := db.Exec(`UPDATE users SET is_banned = 1 WHERE id = ?`, u.ID); err != nil {
		t.Fatalf("ban user: %v", err)
	}

	resp := postLoginForm(t, srv, "banned@example.com", "securepass")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with error, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "suspended") {
		t.Error("expected suspended message for banned user")
	}
}

func TestLoginSubmit_EmptyFields(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postLoginForm(t, srv, "", "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with errors, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "form-error") {
		t.Error("expected form-error class in body for empty fields")
	}
}

func TestLoginSubmit_RepopulatesEmail(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp := postLoginForm(t, srv, "check@example.com", "wrongpass")
	defer resp.Body.Close()

	if !strings.Contains(readBody(t, resp), "check@example.com") {
		t.Error("expected email to be repopulated on login error")
	}
}

func TestLogoutSubmit(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	u, err := model.CreateUser(db, "logout_user", "logout@example.com", "securepass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	token, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	resp, err := noRedirectClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 after logout, got %d", resp.StatusCode)
	}

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be cleared (MaxAge=-1)")
	}
	if sessionCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge=-1 to clear cookie, got %d", sessionCookie.MaxAge)
	}

	// Session must be deleted from DB.
	_, err = model.SessionUser(db, token)
	if err == nil {
		t.Error("expected session to be deleted from DB after logout")
	}
}

func TestLogoutSubmit_WithoutSession(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := noRedirectClient.Post(srv.URL+"/logout", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302 even without session, got %d", resp.StatusCode)
	}
}
