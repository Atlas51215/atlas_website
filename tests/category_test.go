package tests

import (
	"testing"

	"github.com/claude/blog/internal/model"
)

func TestCategories_TotalCount(t *testing.T) {
	db := openTestDB(t)

	cats, err := model.AllCategories(db)
	if err != nil {
		t.Fatalf("AllCategories: %v", err)
	}
	if len(cats) != 11 {
		t.Errorf("expected 11 categories, got %d", len(cats))
	}
}

func TestCategories_ByGroup(t *testing.T) {
	db := openTestDB(t)

	reviews, err := model.CategoriesByGroup(db, "reviews")
	if err != nil {
		t.Fatalf("CategoriesByGroup reviews: %v", err)
	}
	if len(reviews) != 4 {
		t.Errorf("expected 4 review categories, got %d", len(reviews))
	}

	blog, err := model.CategoriesByGroup(db, "blog")
	if err != nil {
		t.Fatalf("CategoriesByGroup blog: %v", err)
	}
	if len(blog) != 7 {
		t.Errorf("expected 7 blog categories, got %d", len(blog))
	}
}

func TestCategories_BySlugAndGroup(t *testing.T) {
	db := openTestDB(t)

	cat, err := model.CategoryBySlugAndGroup(db, "movies", "reviews")
	if err != nil {
		t.Fatalf("CategoryBySlugAndGroup: %v", err)
	}
	if cat == nil {
		t.Fatal("expected reviews/movies, got nil")
	}
	if cat.Name != "Movie Reviews" {
		t.Errorf("expected 'Movie Reviews', got %q", cat.Name)
	}
	if len(cat.ExtraFields) != 3 {
		t.Errorf("expected 3 extra fields, got %d", len(cat.ExtraFields))
	}
}

func TestCategories_BlogHasNoExtraFields(t *testing.T) {
	db := openTestDB(t)

	cat, err := model.CategoryBySlugAndGroup(db, "general", "blog")
	if err != nil {
		t.Fatalf("CategoryBySlugAndGroup: %v", err)
	}
	if cat == nil {
		t.Fatal("expected blog/general, got nil")
	}
	if len(cat.ExtraFields) != 0 {
		t.Errorf("expected no extra fields for blog category, got %d", len(cat.ExtraFields))
	}
}

func TestCategories_NotFound(t *testing.T) {
	db := openTestDB(t)

	cat, err := model.CategoryBySlugAndGroup(db, "nonexistent", "blog")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat != nil {
		t.Errorf("expected nil for missing category, got %+v", cat)
	}
}
