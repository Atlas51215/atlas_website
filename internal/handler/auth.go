package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/claude/blog/internal/model"
	"github.com/claude/blog/internal/render"
	"github.com/claude/blog/templates"
)

var (
	usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

func RegisterPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if model.UserFromContext(r.Context()) != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		cats, err := model.AllCategories(db)
		if err != nil {
			log.Printf("RegisterPage: fetch categories: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		render.Component(w, r, templates.Register(cats, "dark", templates.RegisterData{}))
	}
}

func RegisterSubmit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if model.UserFromContext(r.Context()) != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirm := r.FormValue("confirm_password")

		data := templates.RegisterData{
			Username: username,
			Email:    email,
		}

		if !usernameRe.MatchString(username) {
			data.ErrUsername = "Username must be 3–30 characters: letters, numbers, and underscores only."
		}
		if !emailRe.MatchString(email) {
			data.ErrEmail = "Enter a valid email address."
		}
		if len(password) < 8 {
			data.ErrPassword = "Password must be at least 8 characters."
		}
		if password != confirm {
			data.ErrConfirm = "Passwords do not match."
		}

		if data.ErrUsername != "" || data.ErrEmail != "" || data.ErrPassword != "" || data.ErrConfirm != "" {
			cats, _ := model.AllCategories(db)
			render.Component(w, r, templates.Register(cats, "dark", data))
			return
		}

		user, err := model.CreateUser(db, username, email, password)
		if err != nil {
			if errors.Is(err, model.ErrDuplicateUsername) {
				data.ErrUsername = "That username is already taken."
			} else if errors.Is(err, model.ErrDuplicateEmail) {
				data.ErrEmail = "That email is already registered."
			} else {
				log.Printf("RegisterSubmit: CreateUser: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			cats, _ := model.AllCategories(db)
			render.Component(w, r, templates.Register(cats, "dark", data))
			return
		}

		token, err := model.CreateSession(db, user.ID)
		if err != nil {
			log.Printf("RegisterSubmit: CreateSession: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			MaxAge:   int(30 * 24 * time.Hour / time.Second),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
