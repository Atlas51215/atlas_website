package main

import (
	"log"
	"net/http"

	"github.com/claude/blog/internal/handler"
	"github.com/claude/blog/internal/model"
)

func main() {
	db, err := model.Open("blog.db", "migrations")
	if err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer db.Close()

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler.NewRouter("static")))
}
