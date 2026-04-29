package handler

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/claude/blog/internal/middleware"
)

func NewRouter(db *sql.DB, staticDir string) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(middleware.AuthMiddleware(db))
	r.Get("/", Home(db))
	r.Get("/register", RegisterPage(db))
	r.Post("/register", RegisterSubmit(db))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	r.Handle("/favicon.ico", http.FileServer(http.Dir(staticDir)))
	return r
}
