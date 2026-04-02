# Atlas

A multi-category blog platform built with Go, Templ, HTMX, and SQLite.

## About

Atlas is a content platform supporting multiple post categories — reviews (movies, games, TV, products), a general blog, dev blog, and tech articles. Admins author and manage posts, users can comment (with moderation), and visitors browse published content.

The entire application compiles to a single binary with a single database file. No Docker, no Node.js, no JavaScript frameworks.

## Stack

| Layer | Technology |
|-------|------------|
| Backend | Go |
| Templating | Templ (type-safe, compiled HTML components) |
| Interactivity | HTMX (dynamic UI via HTML fragment swaps) |
| Database | SQLite (WAL mode) |
| Routing | Chi |
| Styling | Tailwind CSS (standalone CLI) |
| Auth | Session cookies + bcrypt |

## Categories

| Category | Route |
|----------|-------|
| Movie Reviews | `/reviews/movies` |
| Game Reviews | `/reviews/games` |
| TV Show Reviews | `/reviews/tv` |
| Product Reviews | `/reviews/products` |
| Blog | `/blog` |
| Dev Blog | `/dev` |
| Tech | `/tech` |

Categories are stored in the database and can be added by admins without code changes.

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Templ CLI](https://templ.guide/quick-start/installation)
- [Tailwind CSS standalone CLI](https://tailwindcss.com/blog/standalone-cli)

### Development

```bash
templ generate --watch &
go run ./cmd/server
```

### Production Build

```bash
templ generate
go build -o blog ./cmd/server
./blog
```

## Project Structure

```
cmd/server/main.go        Entry point
internal/
  handler/                 HTTP handlers (posts, comments, auth, admin)
  model/                   Data structs and DB queries
  middleware/              Auth, logging
  render/                  Templ rendering helpers
templates/                 .templ component files
static/                    CSS, htmx.min.js
uploads/                   User-uploaded images
migrations/                SQL migration files
docs/                      Project documentation
tests/                     Tests
```

## Deployment

The production deployment is:

1. One binary (`blog`)
2. One database file (`blog.db`)
3. One uploads directory (`uploads/`)

```bash
scp blog server:/opt/blog/
scp blog.db server:/opt/blog/
```

## License

See [LICENSE](LICENSE) for details.
