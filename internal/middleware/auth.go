package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/claude/blog/internal/model"
)

const sessionCookieName = "session"

// AuthMiddleware reads the session cookie and loads the matching user into the
// request context. Visitors (no cookie or invalid token) get a nil user.
func AuthMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := model.SessionUser(db, cookie.Value)
			if err != nil && !errors.Is(err, model.ErrSessionNotFound) {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if user != nil {
				model.UpdateLastActive(db, user.ID) //nolint:errcheck
				r = r.WithContext(model.WithUser(r.Context(), user))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth rejects unauthenticated requests with 401.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if model.UserFromContext(r.Context()) == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole rejects requests whose user does not have one of the allowed roles.
// Returns 401 if not logged in at all, 403 if logged in with the wrong role.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := model.UserFromContext(r.Context())
			if user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !allowed[user.Role] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireVerified requires role verified, moderator, or curator.
func RequireVerified(next http.Handler) http.Handler {
	return RequireRole("verified", "moderator", "curator")(next)
}

// RequireModerator requires role moderator or curator.
func RequireModerator(next http.Handler) http.Handler {
	return RequireRole("moderator", "curator")(next)
}

// RequireCurator requires role curator.
func RequireCurator(next http.Handler) http.Handler {
	return RequireRole("curator")(next)
}
