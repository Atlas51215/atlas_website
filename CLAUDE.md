# CLAUDE.md

## Project

Multi-category blog platform with admin authoring and user commenting.  
See `PLAN.md` for full architecture and design decisions.

## Stack

- **Go 1.25** — backend, HTTP server, all business logic
- **Templ** — type-safe compiled HTML templates (migrating from `html/template`)
- **HTMX** — dynamic UI via HTML fragment swaps, no custom JS
- **SQLite** — single-file database (WAL mode)
- **Chi** — HTTP router with middleware
- **Tailwind CSS** (standalone CLI) — styling, no Node.js dependency

## Project Structure

```
cmd/server/main.go          -- entry point
internal/
  handler/                   -- HTTP handlers (posts, comments, auth, admin)
  model/                     -- data structs and DB queries
  middleware/                -- auth, logging
  render/                    -- templ rendering helpers
templates/                   -- .templ files (layout, pages, components)
static/                      -- CSS, htmx.min.js
uploads/                     -- user-uploaded images
migrations/                  -- SQL migration files
```

Note: The project is currently a minimal scaffold (`main.go` at root with `html/template`). The structure above is the target layout per `PLAN.md`.

## Build & Run

```bash
# Development
templ generate --watch &
go run ./cmd/server

# Production
templ generate
go build -o blog ./cmd/server
./blog
```

## Database

SQLite with these pragmas at startup:
```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
```

## Workflow

- After writing or modifying code, always update the unit tests in `tests/` to cover the changes.
- Run `go test -v ./...` to verify all tests pass before considering the work complete.

## Key Conventions

- **Router**: Chi. Group routes by concern (`/admin/*`, `/reviews/*`, etc.).
- **Templates**: Templ components are Go functions — keep them small and composable.
- **HTMX endpoints**: Return HTML fragments, not full pages. Use `hx-target` and `hx-swap` to control placement.
- **Categories**: Stored in the database, not hardcoded. Each has a slug, display name, group, and optional extra fields (JSON).
- **Posts**: Body stored as Markdown, rendered to HTML on read via `goldmark`.
- **Auth**: Session cookies + bcrypt. Middleware enforces roles before handlers run.
- **Comments**: Default status is `pending`. Admins approve/reject via HTMX actions.
- **Errors**: Return appropriate HTTP status codes. Log server errors, show user-friendly messages.
- **No JS frameworks**: All interactivity is HTMX. The only JS file is `htmx.min.js`.

## Roles

| Role | Can do |
|------|--------|
| Admin | Write/edit/delete posts in any category, moderate comments, manage categories and users |
| User | Comment on published posts (comments require approval) |
| Visitor | Read published posts, browse categories |

## Deploy

Single binary + `blog.db` + `uploads/` directory. No Docker required.
