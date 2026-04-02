package main

import (
	"log"
	"net/http"

	"github.com/claude/blog/internal/blog"
)

func main() {
	blog.Init("templates/*.html")
	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", blog.NewMux()))
}
