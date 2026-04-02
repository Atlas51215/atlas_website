# Blog Platform — Technical Plan

## Overview

A multi-category blog platform with admin authoring and user commenting, built with **Go + Templ + HTMX + SQLite**.

---

## Stack Breakdown

| Layer | Technology | Role |
|-------|-----------|------|
| Language | **Go** | Backend logic, HTTP server, all business rules |
| Templating | **Templ** | Type-safe HTML components compiled into Go code |
| Interactivity | **HTMX** | Dynamic UI (comments, moderation, filtering) without writing JS |
| Database | **SQLite** | Single-file storage for all content, users, comments |
| Router | **Chi** | Lightweight HTTP router with middleware support |
| CSS | **Tailwind CSS** (standalone CLI) | Utility-first styling, no Node.js required |
| Auth | **Session cookies + bcrypt** | Admin login, user accounts |

---

## Content Categories

Each category shares the same underlying `post` schema but has its own URL namespace and optional extra fields.

| Category | Route Prefix | Extra Fields |
|----------|-------------|--------------|
| Reviews — Movies | `/reviews/movies` | rating, director, release_year |
| Reviews — Video Games | `/reviews/games` | rating, platform, developer |
| Reviews — TV Shows | `/reviews/tv` | rating, seasons, network |
| Reviews — Products | `/reviews/products` | rating, price_range, purchase_link |
| Blog (general) | `/blog` | — |
| Dev Blog | `/dev` | language_tags, repo_link |
| Tech | `/tech` | — |

Categories are stored in the database, not hardcoded — new ones can be added by admins without code changes.

---

## Database Schema (SQLite)

```sql
-- Categories are data, not code
CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,       -- "movies", "games", "dev"
    name TEXT NOT NULL,              -- "Movie Reviews"
    group_name TEXT,                 -- "reviews", "blog", "tech"
    extra_fields TEXT                -- JSON schema for category-specific fields
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',  -- "admin" | "user"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    author_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    body TEXT NOT NULL,              -- Markdown stored, rendered on read
    extra_data TEXT,                 -- JSON blob for category-specific fields (rating, director, etc.)
    status TEXT NOT NULL DEFAULT 'draft',  -- "draft" | "published"
    published_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(category_id, slug)
);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',  -- "pending" | "approved" | "rejected"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE post_tags (
    post_id INTEGER REFERENCES posts(id),
    tag_id INTEGER REFERENCES tags(id),
    PRIMARY KEY (post_id, tag_id)
);

-- Indexes for common queries
CREATE INDEX idx_posts_category_status ON posts(category_id, status, published_at DESC);
CREATE INDEX idx_comments_post_status ON comments(post_id, status);
CREATE INDEX idx_posts_slug ON posts(slug);
```

---

## How Each Technology Is Used

### Go (Backend)

All server logic lives in Go. No separate API layer — the server renders HTML directly.

- **HTTP server**: Go's `net/http` with Chi router. Production-ready without a reverse proxy.
- **Markdown rendering**: `goldmark` library converts post body from Markdown to HTML on read.
- **Image handling**: Admin uploads go to a local `./uploads/` directory. Go serves them as static files.
- **Session management**: Encrypted cookies via `gorilla/sessions` or a lightweight custom implementation.
- **Database access**: `crawshaw.io/sqlite` or `modernc.org/sqlite` (pure Go, no CGO needed for easy cross-compilation).

```
cmd/
  server/
    main.go              -- entry point, wires everything together

internal/
  handler/
    home.go              -- landing page
    posts.go             -- list, view, create, edit posts
    comments.go          -- submit, moderate comments
    auth.go              -- login, register, logout
    admin.go             -- admin dashboard, category management
  model/
    post.go              -- Post struct, DB queries
    user.go              -- User struct, auth queries
    comment.go           -- Comment struct, moderation queries
    category.go          -- Category struct, DB queries
  middleware/
    auth.go              -- require login, require admin
    logging.go           -- request logging
  render/
    render.go            -- helper to execute templ components with common layout data

templates/
  layout.templ           -- base HTML shell (head, nav, footer)
  home.templ             -- homepage
  post_list.templ        -- category listing page
  post_view.templ        -- single post with comments
  post_form.templ        -- admin editor (create/edit)
  comment.templ          -- single comment component (for HTMX swap)
  admin_dashboard.templ  -- moderation queue, post management

static/
  css/output.css         -- compiled Tailwind
  htmx.min.js            -- single JS file, served locally

uploads/                 -- user-uploaded images
```

### Templ (HTML Templating)

Templ compiles `.templ` files into Go functions. Every page and component is a Go function with type-checked parameters — no `interface{}` or string keys.

**How it works in practice:**

```templ
// templates/post_view.templ

templ PostView(post model.Post, comments []model.Comment, user *model.User) {
    @Layout("Post Title") {
        <article class="prose">
            <h1>{ post.Title }</h1>
            <div>@templ.Raw(post.RenderedBody)</div>

            if post.ExtraData.Rating > 0 {
                <div class="rating">Rating: { fmt.Sprint(post.ExtraData.Rating) }/10</div>
            }
        </article>

        <section id="comments">
            for _, c := range comments {
                @CommentCard(c)
            }
        </section>

        if user != nil {
            @CommentForm(post.ID)
        }
    }
}
```

- Every template is a **compiled Go function** — typos in field names are caught at build time.
- Components are composable: `CommentCard`, `PostCard`, `Layout` are reused across pages.
- No template inheritance confusion — just function calls.

### HTMX (Interactivity)

HTMX makes the UI feel dynamic by swapping HTML fragments from the server. No JavaScript written by hand.

**Where HTMX is used:**

| Feature | HTMX Pattern | How It Works |
|---------|-------------|--------------|
| Submit comment | `hx-post="/comments"` | Form posts to server, server returns the new comment HTML, HTMX appends it to `#comments` |
| Moderate comment | `hx-put="/admin/comments/5/approve"` | Admin clicks approve, server returns updated comment with "approved" badge, swaps in place |
| Delete comment | `hx-delete="/admin/comments/5"` | Removes the comment element from the DOM |
| Load more posts | `hx-get="/blog?page=2" hx-swap="beforeend"` | Appends next page of posts below current list |
| Filter by tag | `hx-get="/blog?tag=go" hx-target="#post-list"` | Replaces post list with filtered results |
| Live search | `hx-get="/search?q=..." hx-trigger="keyup changed delay:300ms"` | Debounced search, replaces results div |
| Preview post (admin) | `hx-post="/admin/preview" hx-target="#preview"` | Admin writes markdown, sees rendered preview without page reload |

**Example — Comment submission:**

```html
<!-- The form -->
<form hx-post="/comments" hx-target="#comments" hx-swap="beforeend" hx-on::after-request="this.reset()">
    <input type="hidden" name="post_id" value="42"/>
    <textarea name="body" required></textarea>
    <button type="submit">Post Comment</button>
</form>
```

The server endpoint returns **only** the new comment's HTML fragment (not the whole page). HTMX appends it to the comment list. The page never fully reloads.

### SQLite (Database)

SQLite is the entire data layer. No separate database server to install or manage.

**Why it works here:**
- A blog is read-heavy, write-light — SQLite handles this easily.
- WAL mode enables concurrent readers with a single writer (more than enough for a blog).
- Backup is `cp blog.db blog.db.bak`.
- The database ships alongside the binary — one folder contains the entire application.

**Configuration applied at startup:**
```go
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA foreign_keys=ON")
db.Exec("PRAGMA busy_timeout=5000")
```

---

## Roles and Permissions

| Action | Admin | User | Visitor |
|--------|-------|------|---------|
| Read published posts | yes | yes | yes |
| Comment on posts | yes | yes | no |
| Create/edit/delete posts (any category) | yes | no | no |
| Manage categories | yes | no | no |
| Moderate comments (approve/reject) | yes | no | no |
| Manage users | yes | no | no |

Middleware checks role before handler executes — no permission logic scattered in templates.

---

## Page Routes

```
GET   /                           -- homepage (featured + recent across categories)
GET   /:group/:category           -- post listing (e.g., /reviews/movies)
GET   /:group/:category/:slug     -- single post (e.g., /reviews/movies/dune-2)
GET   /search?q=                  -- search results

GET   /login                      -- login form
POST  /login                      -- authenticate
GET   /register                   -- registration form
POST  /register                   -- create account
POST  /logout                     -- end session

POST  /comments                   -- submit comment (HTMX)

GET   /admin                      -- dashboard: recent posts, pending comments
GET   /admin/posts/new            -- post editor
POST  /admin/posts                -- create post
GET   /admin/posts/:id/edit       -- edit post
PUT   /admin/posts/:id            -- update post
DELETE /admin/posts/:id           -- delete post
GET   /admin/comments             -- moderation queue
PUT   /admin/comments/:id/:action -- approve/reject comment (HTMX)
```

---

## Build and Deploy

```bash
# Development
templ generate --watch &     # recompile templates on save
go run ./cmd/server           # start server with hot reload (via air)

# Production build
templ generate
go build -o blog ./cmd/server

# Deploy (the entire deployment)
scp blog server:/opt/blog/
scp blog.db server:/opt/blog/
ssh server "systemctl restart blog"
```

The production deployment is:
1. One binary (`blog`)
2. One database file (`blog.db`)
3. One uploads directory (`uploads/`)
4. A systemd unit file

No Docker required (though it works trivially if you want it — `FROM scratch` + binary).

---

## Summary

- **Go** handles all logic, serves HTML directly, compiles to a single binary.
- **Templ** provides type-safe, composable HTML components that catch errors at compile time.
- **HTMX** makes comments, moderation, search, and filtering feel instant — zero custom JS.
- **SQLite** stores everything in one file — no database server, backup is a file copy.
- **Tailwind** (standalone CLI) handles styling with no Node.js dependency.

The entire application is one binary + one database file. Deploy with `scp`.
