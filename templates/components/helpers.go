package components

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/claude/blog/internal/model"
)

type fieldItem struct {
	Label string
	Value string
	Type  string
}

func parseExtraData(s string) map[string]any {
	if s == "" {
		return nil
	}
	var m map[string]any
	if json.Unmarshal([]byte(s), &m) != nil {
		return nil
	}
	return m
}

func formatFieldValue(field model.ExtraField, data map[string]any) (string, bool) {
	if data == nil {
		return "", false
	}
	v, ok := data[field.Name]
	if !ok {
		return "", false
	}
	switch field.Type {
	case "float":
		f, ok := v.(float64)
		if !ok {
			return "", false
		}
		return strconv.FormatFloat(f, 'f', -1, 64), true
	case "int":
		f, ok := v.(float64)
		if !ok {
			return "", false
		}
		return strconv.Itoa(int(f)), true
	case "text", "url":
		s, ok := v.(string)
		if !ok {
			return "", false
		}
		return s, true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

func fieldItems(fields []model.ExtraField, extraData string) []fieldItem {
	data := parseExtraData(extraData)
	var items []fieldItem
	for _, f := range fields {
		if val, ok := formatFieldValue(f, data); ok {
			items = append(items, fieldItem{Label: f.Label, Value: val, Type: f.Type})
		}
	}
	return items
}

var (
	reFence  = regexp.MustCompile("(?s)```.*?```")
	reLink   = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	reHeader = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	reInline = regexp.MustCompile(`[*_` + "`" + `~]`)
)

func bodyPreview(body string) string {
	s := reFence.ReplaceAllString(body, " ")
	s = reHeader.ReplaceAllString(s, "")
	s = reLink.ReplaceAllString(s, "$1")
	s = reInline.ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(s), " ")
	runes := []rune(s)
	if len(runes) > 200 {
		return string(runes[:200]) + "…"
	}
	return s
}

func extractRating(extraData string) (string, bool) {
	if extraData == "" {
		return "", false
	}
	var data map[string]any
	if json.Unmarshal([]byte(extraData), &data) != nil {
		return "", false
	}
	r, ok := data["rating"].(float64)
	if !ok {
		return "", false
	}
	return strconv.FormatFloat(r, 'f', -1, 64), true
}

func postURL(p model.Post) string {
	return "/" + path.Join(p.GroupName, p.CategorySlug, p.Slug)
}

func pageURL(base string, page int) string {
	return fmt.Sprintf("%s?page=%d", base, page)
}

func publishedDate(p model.Post) string {
	if p.PublishedAt != nil {
		return p.PublishedAt.Format("Jan 2, 2006")
	}
	return p.CreatedAt.Format(time.DateOnly)
}
