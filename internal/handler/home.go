package handler

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/claude/blog/internal/model"
	"github.com/claude/blog/internal/render"
	"github.com/claude/blog/templates"
)

func Home(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cats, err := model.AllCategories(db)
		if err != nil {
			log.Printf("Home: fetch categories: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		render.Component(w, r, templates.Home(cats))
	}
}
