# Go Style Guide — Atlas Blog Platform

This guide defines the coding standards for the Atlas blog platform. It follows idiomatic Go conventions, the [Effective Go](https://go.dev/doc/effective_go) guide, and the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) wiki.

---

## Table of Contents

1. [Project Layout](#project-layout)
2. [Naming](#naming)
3. [Formatting](#formatting)
4. [Error Handling](#error-handling)
5. [Functions and Methods](#functions-and-methods)
6. [Packages](#packages)
7. [Database Access](#database-access)
8. [HTTP Handlers](#http-handlers)
9. [Templates (Templ)](#templates-templ)
10. [Middleware](#middleware)
11. [Testing](#testing)
12. [Comments and Documentation](#comments-and-documentation)
13. [Dependencies](#dependencies)
14. [Security](#security)

---

## Project Layout

Follow the standard Go project layout established in `PLAN.md`:

```
cmd/server/main.go       -- entry point only, wiring and startup
internal/                -- all non-exported application code
  handler/               -- HTTP handlers grouped by domain
  model/                 -- data structs, DB queries, business logic
  middleware/            -- HTTP middleware (auth, logging)
  render/                -- templ rendering helpers
templates/               -- .templ component files
static/                  -- compiled CSS, htmx.min.js
migrations/              -- SQL migration files (ordered numerically)
tests/                   -- test files
```

- `cmd/server/main.go` should only wire dependencies and start the server. No business logic.
- All application code belongs under `internal/` to prevent external imports.
- Group files by domain, not by technical role (e.g., `handler/posts.go` not `handlers.go`).

---

## Naming

### General Rules

- Use `MixedCaps` or `mixedCaps`, never underscores in Go names.
- Acronyms should be all caps: `userID`, `httpClient`, `htmlParser`, `URL`, `ID`.
- Keep names short but descriptive. Prefer `post` over `p`, but `i` is fine in a short loop.
- Exported names start with an uppercase letter. Only export what other packages need.

### Packages

- Package names are lowercase, single-word, no underscores or mixedCaps.
- Avoid stutter: `model.Post`, not `model.ModelPost`.
- Name the package after what it provides, not what it contains.

### Variables

```go
// Good
var (
    postCount int
    userID    int64
    db        *sql.DB
)

// Bad
var (
    post_count int
    UserId     int64
    database   *sql.DB  // too verbose for a package-level db handle
)
```

### Interfaces

- Single-method interfaces use the method name plus `-er`: `Reader`, `Writer`, `Stringer`.
- Declare interfaces where they are consumed, not where they are implemented.

### Receiver Names

- Use one or two letter abbreviations consistent within the type: `p` for `Post`, `c` for `Comment`.
- Be consistent: if you start with `p` for `Post`, use `p` everywhere for that type.
- Never use `this` or `self`.

```go
func (p *Post) Publish() error { ... }
func (c *Comment) Approve() error { ... }
```

---

## Formatting

- Run `gofmt` or `goimports` on all code. No exceptions.
- Maximum line length is a soft guideline at ~100 characters. Break long function signatures and struct literals for readability.
- Group imports in three blocks separated by blank lines: standard library, external dependencies, internal packages.

```go
import (
    "context"
    "fmt"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/yuin/goldmark"

    "atlas/internal/model"
    "atlas/internal/render"
)
```

---

## Error Handling

### Always Handle Errors

Never discard errors with `_`. If you truly don't need it, add a comment explaining why.

```go
// Good
if err := db.Ping(); err != nil {
    return fmt.Errorf("database ping: %w", err)
}

// Bad
db.Ping()
```

### Wrap Errors with Context

Use `fmt.Errorf` with `%w` to add context while preserving the error chain.

```go
post, err := model.GetPostBySlug(ctx, db, slug)
if err != nil {
    return fmt.Errorf("fetching post %q: %w", slug, err)
}
```

### Error Messages

- Start with lowercase, no trailing punctuation.
- Describe what failed, not what was attempted: `"fetching post: %w"` not `"failed to fetch post: %w"`.
- Keep messages concise. The error chain provides full context.

### Sentinel Errors

Define package-level sentinel errors for expected conditions that callers need to check:

```go
var ErrPostNotFound = errors.New("post not found")
var ErrUnauthorized = errors.New("unauthorized")
```

Use `errors.Is` to check them:

```go
if errors.Is(err, model.ErrPostNotFound) {
    http.Error(w, "Not Found", http.StatusNotFound)
    return
}
```

---

## Functions and Methods

### Keep Functions Short and Focused

Each function should do one thing. If a function grows beyond ~40 lines, consider splitting it.

### Parameter Order

Follow the convention: `context.Context` first, then inputs, then output dependencies.

```go
func GetPostsByCategory(ctx context.Context, db *sql.DB, categoryID int64, page int) ([]Post, error)
```

### Return Early

Prefer early returns to reduce nesting:

```go
// Good
func (h *PostHandler) View(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, "slug")
    if slug == "" {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }

    post, err := h.posts.GetBySlug(r.Context(), slug)
    if err != nil {
        h.handleError(w, err)
        return
    }

    h.render(w, templates.PostView(post))
}
```

### Avoid Naked Returns

Always use explicit return values. Naked returns obscure what's being returned.

### Accept Interfaces, Return Structs

Functions should accept interfaces for flexibility and return concrete types for clarity.

---

## Packages

### `internal/model`

- Each model file contains the struct, its constructor (if needed), and all DB queries for that type.
- Query functions accept `context.Context` and a database handle as the first parameters.
- Return concrete types, not interfaces.

```go
// model/post.go
type Post struct {
    ID          int64
    CategoryID  int64
    AuthorID    int64
    Title       string
    Slug        string
    Body        string
    ExtraData   json.RawMessage
    Status      string
    PublishedAt sql.NullTime
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func GetPostBySlug(ctx context.Context, db *sql.DB, categorySlug, postSlug string) (*Post, error) {
    // ...
}
```

### `internal/handler`

- Group handlers into structs by domain: `PostHandler`, `CommentHandler`, `AuthHandler`, `AdminHandler`.
- Handler structs hold dependencies (db, logger, template renderer).
- Each handler method matches the `http.HandlerFunc` signature.

```go
type PostHandler struct {
    db     *sql.DB
    logger *slog.Logger
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) { ... }
func (h *PostHandler) View(w http.ResponseWriter, r *http.Request) { ... }
```

### `internal/middleware`

- Middleware follows the `func(next http.Handler) http.Handler` pattern.
- Each middleware does exactly one thing (auth check, logging, etc.).
- Keep middleware side-effect free where possible — read from the request, set context values, call next.

---

## Database Access

### Queries

- Use parameterized queries exclusively. Never interpolate values into SQL strings.
- Use `context.Context` for all database operations.
- Close rows with `defer rows.Close()` immediately after the query call.

```go
func ListPublishedPosts(ctx context.Context, db *sql.DB, categoryID int64) ([]Post, error) {
    rows, err := db.QueryContext(ctx,
        `SELECT id, title, slug, body, published_at
         FROM posts
         WHERE category_id = ? AND status = 'published'
         ORDER BY published_at DESC`, categoryID)
    if err != nil {
        return nil, fmt.Errorf("querying published posts: %w", err)
    }
    defer rows.Close()

    var posts []Post
    for rows.Next() {
        var p Post
        if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.PublishedAt); err != nil {
            return nil, fmt.Errorf("scanning post row: %w", err)
        }
        posts = append(posts, p)
    }
    return posts, rows.Err()
}
```

### Transactions

Wrap multi-step mutations in transactions. Use a helper to ensure rollback on error:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("beginning transaction: %w", err)
}
defer tx.Rollback()

// ... perform operations on tx ...

if err := tx.Commit(); err != nil {
    return fmt.Errorf("committing transaction: %w", err)
}
```

### Migrations

- Store in `migrations/` with numeric prefixes: `001_create_tables.sql`, `002_add_tags.sql`.
- Each file should be idempotent where possible (`CREATE TABLE IF NOT EXISTS`).
- Run migrations at startup before the server begins accepting requests.

---

## HTTP Handlers

### Request Lifecycle

1. Middleware runs (logging, auth).
2. Handler parses input (URL params, form data, query strings).
3. Handler calls model layer for business logic / data access.
4. Handler renders a response (full page or HTMX fragment).

### HTMX Endpoints

- HTMX endpoints return **HTML fragments**, not full pages.
- Check the `HX-Request` header to distinguish HTMX requests from full-page loads when a route serves both.
- Use appropriate HTTP methods: `GET` for reads, `POST` for creation, `PUT` for updates, `DELETE` for removal.

```go
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
    // Parse form, validate, insert into DB...

    // Return only the comment fragment for HTMX to swap in
    render.Component(w, templates.CommentCard(comment))
}
```

### Status Codes

- `200` — successful response
- `201` — resource created
- `204` — successful deletion (no content)
- `400` — bad request / validation failure
- `401` — not authenticated
- `403` — authenticated but not authorized
- `404` — resource not found
- `500` — unexpected server error (log the full error, show a generic message)

---

## Templates (Templ)

### Component Design

- Keep components small and focused. One component per visual unit.
- Pass only the data a component needs — not the entire model if it only uses two fields.
- Use the `@Layout()` wrapper for full pages; return bare components for HTMX fragments.

### Naming

- Component file names match the domain: `post_view.templ`, `comment_card.templ`.
- Component function names are PascalCase: `PostView`, `CommentCard`, `AdminDashboard`.

### Composition

```templ
// Good — small, reusable components
templ PostCard(post model.Post) {
    <article class="border rounded p-4">
        <h2>{ post.Title }</h2>
        <time>{ post.PublishedAt.Format("Jan 2, 2006") }</time>
    </article>
}

templ PostList(posts []model.Post) {
    <div id="post-list">
        for _, p := range posts {
            @PostCard(p)
        }
    </div>
}
```

---

## Middleware

### Auth Middleware

- Use `context.WithValue` to pass the authenticated user through the request context.
- Define typed context keys to avoid collisions:

```go
type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) *model.User {
    user, _ := ctx.Value(userContextKey).(*model.User)
    return user
}
```

### Middleware Ordering

Apply middleware in this order on the router:

1. Request logging
2. Recovery (panic handler)
3. Session loading
4. Auth checks (on protected route groups)

---

## Testing

### Organization

- Place tests in the `tests/` directory as specified in the project workflow.
- Name test files to match what they test: `post_test.go`, `auth_test.go`.
- Run all tests with `go test -v ./...` before considering work complete.

### Table-Driven Tests

Prefer table-driven tests for functions with multiple input/output combinations:

```go
func TestSlugify(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {"simple", "Hello World", "hello-world"},
        {"special chars", "Go & Templ!", "go-templ"},
        {"multiple spaces", "too   many  spaces", "too-many-spaces"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Slugify(tt.input)
            if got != tt.want {
                t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
            }
        })
    }
}
```

### HTTP Handler Tests

Use `net/http/httptest` to test handlers in isolation:

```go
func TestPostHandler_View(t *testing.T) {
    req := httptest.NewRequest("GET", "/blog/my-post", nil)
    w := httptest.NewRecorder()

    handler := &PostHandler{db: testDB}
    handler.View(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
    }
}
```

### Database Tests

- Use an in-memory SQLite database (`:memory:`) for unit tests.
- Run migrations on the test database before each test suite.
- Use `t.Cleanup` to tear down state after tests.

---

## Comments and Documentation

### When to Comment

- Exported functions and types **must** have doc comments following Go conventions.
- Comment the _why_, not the _what_. The code shows what; comments explain non-obvious reasoning.
- Don't comment obvious code.

```go
// GetPostBySlug returns the published post matching the given category and post slugs.
// Returns ErrPostNotFound if no matching published post exists.
func GetPostBySlug(ctx context.Context, db *sql.DB, catSlug, postSlug string) (*Post, error) {
    // ...
}
```

### Doc Comment Style

- Start with the function/type name.
- Use complete sentences.
- No blank line between the comment and the declaration.

---

## Dependencies

### Approved Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/go-chi/chi/v5` | HTTP routing |
| `github.com/a-h/templ` | HTML templating |
| `github.com/yuin/goldmark` | Markdown to HTML |
| `modernc.org/sqlite` (or `crawshaw.io/sqlite`) | SQLite driver (pure Go) |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `log/slog` (stdlib) | Structured logging |

### Dependency Rules

- Minimize external dependencies. Prefer the standard library.
- No JavaScript frameworks. The only JS file is `htmx.min.js`.
- No ORM. Write SQL directly — it's explicit and debuggable.
- Vet new dependencies for maintenance status, license, and transitive dependency count before adding.

---

## Security

### Input Validation

- Validate all user input at the handler layer before passing to the model layer.
- Sanitize Markdown output to prevent XSS (use goldmark with an HTML sanitizer).
- Use parameterized queries for all SQL — never string concatenation.

### Authentication

- Hash passwords with `bcrypt` at a minimum cost of 12.
- Store sessions in encrypted, HTTP-only, secure cookies.
- Set `SameSite=Lax` on session cookies.
- Regenerate session ID after login to prevent session fixation.

### Authorization

- Enforce permissions in middleware, not in templates or handlers.
- Never rely on UI hiding to enforce access control — always check server-side.

### File Uploads

- Restrict allowed file types (images only).
- Limit upload size (e.g., 10 MB max).
- Never serve uploaded files with their original filename in a way that allows path traversal.
- Store uploads outside the application's executable directory.

### HTTP Headers

Set security headers via middleware:

```go
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
```
