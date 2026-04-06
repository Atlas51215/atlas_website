package main

import (
	"log"
	"net/http"

	"github.com/claude/blog/internal/handler"
)

func main() {
	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler.NewRouter("static")))
}
