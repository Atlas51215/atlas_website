package handler

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/claude/blog/internal/model"
	"github.com/claude/blog/internal/render"
	"github.com/claude/blog/templates"
)

const defaultPageSize = 10

func BlogLanding(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postListHandler(db, "blog", nil, w, r)
	}
}

func BlogCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "category")
		cat, err := model.CategoryBySlugAndGroup(db, slug, "blog")
		if err != nil {
			log.Printf("BlogCategory: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if cat == nil {
			http.NotFound(w, r)
			return
		}
		postListHandler(db, "blog", cat, w, r)
	}
}

func ReviewsLanding(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postListHandler(db, "reviews", nil, w, r)
	}
}

func ReviewsCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "category")
		cat, err := model.CategoryBySlugAndGroup(db, slug, "reviews")
		if err != nil {
			log.Printf("ReviewsCategory: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if cat == nil {
			http.NotFound(w, r)
			return
		}
		postListHandler(db, "reviews", cat, w, r)
	}
}

func postListHandler(db *sql.DB, group string, cat *model.Category, w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)

	var (
		posts   []model.Post
		total   int
		err     error
		baseURL string
	)

	if cat != nil {
		posts, total, err = model.PostsByCategory(db, cat.ID, page, defaultPageSize)
		baseURL = "/" + group + "/" + cat.Slug
	} else {
		posts, total, err = model.PostsByGroup(db, group, page, defaultPageSize)
		baseURL = "/" + group
	}

	if err != nil {
		log.Printf("postListHandler: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	totalPages := (total + defaultPageSize - 1) / defaultPageSize
	if totalPages == 0 {
		// Ensure at least one page so callers always have a valid range.
		totalPages = 1
	}

	d := templates.PostListData{
		Group:      group,
		Category:   cat,
		Posts:      posts,
		Page:       page,
		TotalPages: totalPages,
		BaseURL:    baseURL,
	}

	if r.Header.Get("HX-Request") == "true" {
		render.Component(w, r, templates.PostListInner(d))
		return
	}

	cats, err := model.AllCategories(db)
	if err != nil {
		log.Printf("postListHandler categories: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	theme := "dark"
	if u := model.UserFromContext(r.Context()); u != nil && u.Theme != "" {
		theme = u.Theme
	}

	render.Component(w, r, templates.PostList(cats, theme, d))
}

func parsePage(r *http.Request) int {
	// Atoi returns 0 on non-numeric input; the clamp below handles it.
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if p < 1 {
		p = 1
	}
	return p
}
