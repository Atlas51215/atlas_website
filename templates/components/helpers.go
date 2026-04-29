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
