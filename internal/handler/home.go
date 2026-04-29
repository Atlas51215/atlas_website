package handler

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/claude/blog/internal/model"
	"github.com/claude/blog/internal/render"
	"github.com/claude/blog/templates"
)

const homePageSize = 10

func Home(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cats, err := model.AllCategories(db)
		if err != nil {
			log.Printf("Home: fetch categories: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		theme := "dark"
		u := model.UserFromContext(r.Context())
		if u != nil && u.Theme != "" {
			theme = u.Theme
		}

		d := templates.HomeData{
			Page:       1,
			TotalPages: 1,
			IsLoggedIn: u != nil,
		}

		if u != nil {
			curators, err := model.FollowedCurators(db, u.ID)
			if err != nil {
				log.Printf("Home: followed curators: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			d.HasFollows = len(curators) > 0
		}

		if d.HasFollows {
			page := parsePage(r)
			posts, total, err := model.PostsByFollowedCurators(db, u.ID, page, homePageSize)
			if err != nil {
				log.Printf("Home: PostsByFollowedCurators: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			totalPages := (total + homePageSize - 1) / homePageSize
			if totalPages == 0 {
				totalPages = 1
			}
			d.Posts = posts
			d.Page = page
			d.TotalPages = totalPages
		} else {
			posts, err := model.RecentPosts(db, 20)
			if err != nil {
				log.Printf("Home: RecentPosts: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			d.Posts = posts
		}

		if r.Header.Get("HX-Request") == "true" {
			render.Component(w, r, templates.HomeInner(d))
			return
		}
		render.Component(w, r, templates.Home(cats, theme, d))
	}
}
