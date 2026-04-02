package blog

import (
	"html/template"
	"net/http"
	"time"
)

type Post struct {
	Slug      string
	Title     string
	Body      template.HTML
	CreatedAt time.Time
}

var posts = []Post{
	{
		Slug:      "hello-world",
		Title:     "Hello, World!",
		Body:      "<p>Welcome to my blog. This is the first post.</p>",
		CreatedAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	},
}

var templates *template.Template

// Init parses templates from the given glob pattern.
func Init(templateGlob string) {
	templates = template.Must(template.ParseGlob(templateGlob))
}

// NewMux returns a configured ServeMux with all blog routes registered.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/post/", handlePost)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

// SetPosts replaces the global posts slice. Returns a restore function.
func SetPosts(p []Post) func() {
	original := posts
	posts = p
	return func() { posts = original }
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[len("/post/"):]
	for _, p := range posts {
		if p.Slug == slug {
			if err := templates.ExecuteTemplate(w, "post.html", p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	http.NotFound(w, r)
}
