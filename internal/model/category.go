package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type ExtraField struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`  // "float" | "int" | "text"
	Label string  `json:"label"`
	Min   float64 `json:"min,omitempty"`
	Max   float64 `json:"max,omitempty"`
}

type Category struct {
	ID          int64
	Slug        string
	Name        string
	GroupName   string
	ExtraFields []ExtraField
	SortOrder   int
}

func AllCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query(`
		SELECT id, slug, name, group_name, extra_fields, sort_order
		FROM categories
		ORDER BY group_name, sort_order`)
	if err != nil {
		return nil, fmt.Errorf("AllCategories: %w", err)
	}
	defer rows.Close()
	return scanCategories(rows)
}

func CategoriesByGroup(db *sql.DB, group string) ([]Category, error) {
	rows, err := db.Query(`
		SELECT id, slug, name, group_name, extra_fields, sort_order
		FROM categories
		WHERE group_name = ?
		ORDER BY sort_order`, group)
	if err != nil {
		return nil, fmt.Errorf("CategoriesByGroup: %w", err)
	}
	defer rows.Close()
	return scanCategories(rows)
}

func CategoryBySlugAndGroup(db *sql.DB, slug, group string) (*Category, error) {
	row := db.QueryRow(`
		SELECT id, slug, name, group_name, extra_fields, sort_order
		FROM categories
		WHERE slug = ? AND group_name = ?`, slug, group)

	c, err := scanCategory(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("CategoryBySlugAndGroup: %w", err)
	}
	return c, nil
}

func scanCategories(rows *sql.Rows) ([]Category, error) {
	var cats []Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		cats = append(cats, *c)
	}
	return cats, rows.Err()
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanCategory(s scanner) (*Category, error) {
	var c Category
	var extraRaw sql.NullString
	if err := s.Scan(&c.ID, &c.Slug, &c.Name, &c.GroupName, &extraRaw, &c.SortOrder); err != nil {
		return nil, err
	}
	if extraRaw.Valid && extraRaw.String != "" {
		if err := json.Unmarshal([]byte(extraRaw.String), &c.ExtraFields); err != nil {
			return nil, fmt.Errorf("parse extra_fields for %s/%s: %w", c.GroupName, c.Slug, err)
		}
	}
	return &c, nil
}
