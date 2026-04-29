package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/claude/blog/internal/middleware"
	"github.com/claude/blog/internal/model"
)

// okHandler is a trivial handler that writes 200.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestAuthMiddleware_NoCookie(t *testing.T) {
	db := openTestDB(t)

	handler := middleware.AuthMiddleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if model.UserFromContext(r.Context()) != nil {
			t.Error("expected nil user for request without cookie")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_ValidSession(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "mw_alice", "mw_alice@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	token, err := model.CreateSession(db, u.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	handler := middleware.AuthMiddleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := model.UserFromContext(r.Context())
		if got == nil {
			t.Error("expected user in context, got nil")
		} else if got.ID != u.ID {
			t.Errorf("user ID: got %d, want %d", got.ID, u.ID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	db := openTestDB(t)

	handler := middleware.AuthMiddleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if model.UserFromContext(r.Context()) != nil {
			t.Error("expected nil user for invalid token")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bogus-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireAuth_NoUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	middleware.RequireAuth(okHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_WithUser(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "mw_bob", "mw_bob@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(model.WithUser(req.Context(), u))
	rr := httptest.NewRecorder()
	middleware.RequireAuth(okHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireRole_NoUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	middleware.RequireRole("curator")(okHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRequireRole_WrongRole(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "mw_carol", "mw_carol@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	// Default role is "unverified", not "curator".

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(model.WithUser(req.Context(), u))
	rr := httptest.NewRecorder()
	middleware.RequireRole("curator")(okHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestRequireRole_CorrectRole(t *testing.T) {
	db := openTestDB(t)

	u, err := model.CreateUser(db, "mw_dave", "mw_dave@example.com", "pass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := model.UpdateUserRole(db, u.ID, "curator"); err != nil {
		t.Fatalf("UpdateUserRole: %v", err)
	}
	u.Role = "curator"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(model.WithUser(req.Context(), u))
	rr := httptest.NewRecorder()
	middleware.RequireRole("curator")(okHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireVerified(t *testing.T) {
	db := openTestDB(t)

	cases := []struct {
		role string
		want int
	}{
		{"unverified", http.StatusForbidden},
		{"verified", http.StatusOK},
		{"moderator", http.StatusOK},
		{"curator", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			u, err := model.CreateUser(db, "mw_v_"+tc.role, "mw_v_"+tc.role+"@example.com", "pass")
			if err != nil {
				t.Fatalf("CreateUser: %v", err)
			}
			u.Role = tc.role

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(model.WithUser(req.Context(), u))
			rr := httptest.NewRecorder()
			middleware.RequireVerified(okHandler).ServeHTTP(rr, req)

			if rr.Code != tc.want {
				t.Errorf("role %q: got %d, want %d", tc.role, rr.Code, tc.want)
			}
		})
	}
}

func TestRequireModerator(t *testing.T) {
	db := openTestDB(t)

	cases := []struct {
		role string
		want int
	}{
		{"unverified", http.StatusForbidden},
		{"verified", http.StatusForbidden},
		{"moderator", http.StatusOK},
		{"curator", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			u, err := model.CreateUser(db, "mw_m_"+tc.role, "mw_m_"+tc.role+"@example.com", "pass")
			if err != nil {
				t.Fatalf("CreateUser: %v", err)
			}
			u.Role = tc.role

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(model.WithUser(req.Context(), u))
			rr := httptest.NewRecorder()
			middleware.RequireModerator(okHandler).ServeHTTP(rr, req)

			if rr.Code != tc.want {
				t.Errorf("role %q: got %d, want %d", tc.role, rr.Code, tc.want)
			}
		})
	}
}

func TestRequireCurator(t *testing.T) {
	db := openTestDB(t)

	cases := []struct {
		role string
		want int
	}{
		{"unverified", http.StatusForbidden},
		{"verified", http.StatusForbidden},
		{"moderator", http.StatusForbidden},
		{"curator", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			u, err := model.CreateUser(db, "mw_c_"+tc.role, "mw_c_"+tc.role+"@example.com", "pass")
			if err != nil {
				t.Fatalf("CreateUser: %v", err)
			}
			u.Role = tc.role

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(model.WithUser(req.Context(), u))
			rr := httptest.NewRecorder()
			middleware.RequireCurator(okHandler).ServeHTTP(rr, req)

			if rr.Code != tc.want {
				t.Errorf("role %q: got %d, want %d", tc.role, rr.Code, tc.want)
			}
		})
	}
}
