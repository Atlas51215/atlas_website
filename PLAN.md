# Atlas — Technical Plan

## Overview

A multi-category blog and review platform with curator authoring, moderation, user commenting, and follow-based notifications. Built with **Go + Templ + HTMX + SQLite**.

---

## Stack

| Layer | Technology | Role |
|-------|-----------|------|
| Language | **Go 1.25** | Backend logic, HTTP server, all business rules |
| Templating | **Templ** | Type-safe compiled HTML components |
| Interactivity | **HTMX** | Dynamic UI via HTML fragment swaps, no custom JS |
| Database | **SQLite** (WAL mode) | Single-file storage for all data |
| Router | **Chi** | HTTP router with middleware |
| CSS | **Tailwind CSS** (standalone CLI) | Utility-first styling, no Node.js |
| Markdown | **goldmark** | Render post/comment bodies from Markdown to HTML |
| Email | **Resend** (free tier, 3k/month) | Verification emails, follow notifications |
| Auth | **Session cookies + bcrypt** | Login, registration, role enforcement |

---

## Pages

| Page | Route | Description |
|------|-------|-------------|
| Home | `/` | Recent posts from curators the user follows (or all, if not logged in) |
| About | `/about/:curator-slug` | Per-curator about page, editable by that curator |
| Blog | `/blog` | Blog landing — links to sub-categories |
| Blog Sub-category | `/blog/:category` | Post listing for a blog sub-category |
| Reviews | `/reviews` | Review landing — links to sub-categories |
| Review Sub-category | `/reviews/:category` | Post listing for a review sub-category |
| Single Post | `/:group/:category/:slug` | Full post view with threaded comments |
| Search Results | `/search?q=` | Search results page |
| Login | `/login` | Login form |
| Register | `/register` | Registration form |
| User Settings | `/settings` | Username, password, theme, avatar, bio, notification prefs |
| Admin Dashboard | `/admin` | Post management, moderation queue, category management |
| Post Editor | `/admin/posts/new` | Plain Markdown textarea, preview available after saving as draft |
| Category Manager | `/admin/categories` | Create/edit categories with custom fields |
| User Manager | `/admin/users` | Promote/demote, ban/unban users |

---

## Content Categories

Each category belongs to a **group** (`reviews` or `blog`). Categories are stored in the database and managed by curators — no hardcoding.

### Default Categories

| Group | Category | Slug | Extra Fields |
|-------|----------|------|--------------|
| Reviews | Movies | `movies` | rating, director, release_year |
| Reviews | Video Games | `games` | rating, platform, developer |
| Reviews | TV Shows | `tv` | rating, seasons, network |
| Reviews | Products | `products` | rating, price_range, purchase_link |
| Blog | Movies | `movies` | — |
| Blog | Video Games | `games` | — |
| Blog | TV Shows | `tv` | — |
| Blog | Products | `products` | — |
| Blog | General | `general` | — |
| Blog | Dev | `dev` | — |
| Blog | Hardware/Tech | `tech` | — |

### Curator-Defined Custom Fields

When creating a new category, curators define custom fields by specifying a **name** and **type**:

| Field Type | Example |
|------------|---------|
| `text` | director, developer, publisher |
| `number` | release_year, seasons, player_count |
| `float` | rating (0–10, max 2 decimal places, no trailing zeros: 4.0 → "4", 2.20 → "2.2") |
| `url` | purchase_link, repo_link |

Custom fields are stored as a JSON schema on the category and rendered dynamically in the post editor and post view.

### Rating Display

Ratings are `float` type, 0–10 range, max 2 decimal places. Display rules:
- `4.0` → "4"
- `2.20` → "2.2"
- `3.14` → "3.14"
- `10.0` → "10"

---

## Roles & Permissions

Four user roles, from highest to lowest:

| Action | Curator | Moderator | Verified User | Unverified User |
|--------|---------|-----------|---------------|-----------------|
| Read all published content | yes | yes | yes | yes |
| Comment on posts | yes | yes | yes | no |
| Follow curators | yes | yes | yes | no |
| Write/edit/delete own posts | yes | no | no | no |
| Create/edit categories & custom fields | yes | no | no | no |
| Soft-delete any post or comment | yes | yes | no | no |
| Hard-delete any post or comment | yes | no | no | no |
| Restore soft-deleted content | yes | no | no | no |
| Ban/unban users from commenting | yes | yes | no | no |
| Promote/demote user roles | yes | no | no | no |
| Manage users | yes | no | no | no |
| Edit own About page | yes | no | no | no |

**Notes:**
- **Curator** is the superuser role (site owner and trusted authors).
- **Moderator** can moderate content and ban users but cannot create posts.
- **Verified User** has confirmed their email and can comment.
- **Unverified User** can read everything but cannot comment or generate content. Accounts inactive for 1 month are cleaned up.
- Middleware enforces roles before handlers execute — no permission checks in templates.

---

## Authentication & Email

### Registration Flow

1. User submits registration form (username, email, password).
2. Account created with role `unverified`.
3. Verification email sent via **Resend** API.
4. If Resend's free tier limit (3k/month) is reached, the email is queued in the database and sent the next day via a scheduled job.
5. User clicks verification link → role promoted to `verified`.
6. Unverified accounts with no activity for 1 month are automatically purged.

### Sessions

- Session cookies with `HttpOnly`, `Secure`, `SameSite=Strict`.
- Passwords hashed with bcrypt.

---

## Notifications

### In-Site Notifications

- Bell icon in the UI with unread count.
- Notification types:
  - New post from a followed curator
  - Reply to your comment
  - Comment approved/rejected (for the commenter)
  - Account banned/unbanned

### Email Notifications

- Same events as in-site, controlled by user's email notification preferences in settings.
- Sent via Resend. If daily limit reached, queued for next day.

### Follow System

- Verified users and above can follow curators.
- Following a curator triggers notifications (in-site + email based on prefs) when that curator publishes a new post.

---

## Navigation — Left Sidebar

### Layout

The navbar is a left-side sidebar with the following structure (top to bottom):

1. **Logo & Title** — The Atlas "A" logo followed by "tlas" text. The "A" is the custom logo (black, red, white). Favicon is the same logo.
2. **Hide/Show Button** — Always visible in the top-right corner of the viewport. Toggles sidebar visibility. On mobile, sidebar starts hidden.
3. **Search Bar** — Text input, searches post titles first; if no matches, searches post bodies.
4. **Nav Links** (with sub-category expansion):
   - Home
   - About Me
   - Blog → expands to show sub-categories when active
   - Reviews → expands to show sub-categories when active
5. **Ad Slot** — Placeholder at the bottom of the sidebar. This is the **only** ad placement on the entire site. When the sidebar is hidden (including on mobile), the ad is hidden too.

### Behavior

- **Desktop**: Sidebar visible by default. Toggle button hides/shows it. When hidden, content stays centered at the same max-width (does not expand to fill).
- **Mobile**: Sidebar hidden by default. Toggle button shows/hides it as an overlay.
- **Sub-category expansion**: When navigating into Blog or Reviews, that section's sub-categories expand in the nav. Other sections stay collapsed.

---

## Theme

### Dark Mode (Default)

- Background: very dark gray
- Text: white
- Accents: red

### Light Mode

- Background: very light gray
- Text: black
- Accents: red

### Behavior

- Always starts in dark mode (no OS preference detection).
- Users can toggle in the navbar or user settings.
- Preference stored per-user in the database (for logged-in users) and in a cookie (for visitors).

---

## User Settings Page

Located at `/settings`. Available to all logged-in users.

| Setting | Description |
|---------|-------------|
| Username | Change display name |
| Password | Change password (requires current password) |
| Email | Display only (cannot change) |
| Theme | Toggle dark/light mode |
| Avatar | Auto-generated initials/identicon (not uploadable, reduces moderation burden) |
| Bio | Short text bio |
| Email Notifications | Toggle notifications for: new posts from followed curators, comment replies |

---

## Database Schema

```sql
-- Categories are data, not code
CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL,                -- "movies", "games", "dev"
    name TEXT NOT NULL,                -- "Movie Reviews", "General Blog"
    group_name TEXT NOT NULL,          -- "reviews", "blog"
    extra_fields TEXT,                 -- JSON array defining custom fields: [{"name":"rating","type":"float","label":"Rating","min":0,"max":10}, ...]
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_name, slug)
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'unverified',  -- "curator" | "moderator" | "verified" | "unverified"
    bio TEXT DEFAULT '',
    theme TEXT NOT NULL DEFAULT 'dark',       -- "dark" | "light"
    email_notify_posts INTEGER DEFAULT 1,     -- notify on new posts from followed curators
    email_notify_replies INTEGER DEFAULT 1,   -- notify on comment replies
    is_banned INTEGER DEFAULT 0,
    verified_at DATETIME,
    last_active_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    author_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    body TEXT NOT NULL,                       -- Markdown source
    extra_data TEXT,                          -- JSON blob matching category's extra_fields
    status TEXT NOT NULL DEFAULT 'draft',     -- "draft" | "published"
    is_deleted INTEGER DEFAULT 0,            -- soft delete flag
    deleted_by INTEGER REFERENCES users(id),
    published_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(category_id, slug)
);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    parent_id INTEGER REFERENCES comments(id),  -- NULL = top-level, otherwise threaded reply
    body TEXT NOT NULL,                          -- basic Markdown
    status TEXT NOT NULL DEFAULT 'pending',      -- "pending" | "approved" | "rejected"
    is_deleted INTEGER DEFAULT 0,
    deleted_by INTEGER REFERENCES users(id),
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

CREATE TABLE follows (
    follower_id INTEGER NOT NULL REFERENCES users(id),
    curator_id INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, curator_id)
);

CREATE TABLE notifications (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    type TEXT NOT NULL,           -- "new_post" | "comment_reply" | "comment_approved" | "comment_rejected" | "banned" | "unbanned"
    payload TEXT,                 -- JSON with context (post_id, comment_id, curator name, etc.)
    is_read INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE email_queue (
    id INTEGER PRIMARY KEY,
    to_email TEXT NOT NULL,
    subject TEXT NOT NULL,
    body_html TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',   -- "pending" | "sent" | "failed"
    scheduled_for DATETIME DEFAULT CURRENT_TIMESTAMP,
    attempts INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE verification_tokens (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    token TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE about_pages (
    id INTEGER PRIMARY KEY,
    curator_id INTEGER UNIQUE NOT NULL REFERENCES users(id),
    body TEXT NOT NULL DEFAULT '',   -- Markdown
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_posts_category_status ON posts(category_id, status, published_at DESC);
CREATE INDEX idx_posts_author ON posts(author_id, status);
CREATE INDEX idx_posts_deleted ON posts(is_deleted);
CREATE INDEX idx_comments_post ON comments(post_id, status, is_deleted);
CREATE INDEX idx_comments_parent ON comments(parent_id);
CREATE INDEX idx_follows_curator ON follows(curator_id);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);
CREATE INDEX idx_email_queue_status ON email_queue(status, scheduled_for);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_last_active ON users(last_active_at);
CREATE INDEX idx_verification_tokens_token ON verification_tokens(token);
```

---

## Project Structure

```
cmd/server/main.go              -- entry point, wires everything together

internal/
  handler/
    home.go                     -- homepage (recent from followed curators)
    posts.go                    -- list, view posts (generic for all categories)
    comments.go                 -- submit, thread, moderate comments
    auth.go                     -- login, register, logout, email verification
    admin.go                    -- dashboard, category/user management
    about.go                    -- curator about pages
    search.go                   -- search handler (titles first, then body)
    settings.go                 -- user settings
    notifications.go            -- notification list, mark-as-read
    upload.go                   -- image upload endpoint
  model/
    post.go                     -- Post struct, DB queries
    user.go                     -- User struct, auth queries
    comment.go                  -- Comment struct, threading, moderation
    category.go                 -- Category struct, custom field schema
    follow.go                   -- Follow/unfollow queries
    notification.go             -- Notification CRUD
    email.go                    -- Email queue queries
    about.go                    -- About page queries
    tag.go                      -- Tag queries
  middleware/
    auth.go                     -- role-based access (require verified, moderator, curator)
    logging.go                  -- request logging
    theme.go                    -- inject theme preference into context
  render/
    render.go                   -- templ rendering helpers, common layout data
  jobs/
    cleanup.go                  -- purge inactive unverified users (1 month)
    email_sender.go             -- process email queue, respect daily limits

templates/
  layout.templ                  -- base HTML shell (sidebar nav, theme, notification bell)
  nav.templ                     -- sidebar component (logo, search, links, ad slot)
  home.templ                    -- homepage
  about.templ                   -- curator about page
  post_list.templ               -- generic post listing (used by all categories)
  post_view.templ               -- generic single post with threaded comments
  post_form.templ               -- curator post editor (Markdown textarea)
  post_preview.templ            -- draft preview
  comment.templ                 -- single comment component (threaded)
  comment_form.templ            -- comment submission form
  search.templ                  -- search results page
  settings.templ                -- user settings form
  auth_login.templ              -- login page
  auth_register.templ           -- registration page
  admin/
    dashboard.templ             -- admin dashboard
    categories.templ            -- category manager with custom field editor
    users.templ                 -- user management (roles, bans)
    moderation.templ            -- pending comments/posts queue
  components/
    post_card.templ             -- reusable post card (used in listings, home, search)
    comment_thread.templ        -- recursive threaded comment display
    pagination.templ            -- numbered page navigation
    notification_bell.templ     -- bell icon with unread count
    rating.templ                -- rating display (formatted float)
    custom_fields.templ         -- dynamic field renderer for category extra fields
    ad_slot.templ               -- ad placeholder

static/
  css/output.css                -- compiled Tailwind
  htmx.min.js                   -- single JS file
  logo.svg                      -- Atlas "A" logo (to be added later)
  favicon.ico                   -- favicon (same logo, to be added later)

uploads/                        -- user-uploaded images

migrations/
  001_initial.sql               -- full schema
```

---

## Generic Templates

Most pages share the same base components to minimize duplication:

- **`post_list.templ`** — one template drives all category listing pages. Receives category info, posts, and pagination data. Blog and review listings use the same template; review listings additionally render extra fields (rating, etc.) via `custom_fields.templ`.
- **`post_view.templ`** — one template for all single-post pages. Renders Markdown body, extra fields (if any), and threaded comments.
- **`post_card.templ`** — reusable card used on home, listing, and search pages.
- **`pagination.templ`** — shared numbered pagination component.
- **`comment_thread.templ`** — recursive component for nested comments.

If a specific category or page needs something extra, it composes the generic component and adds to it — the generic components themselves stay lean and shared. Security fixes (e.g., sanitization) only need to happen in one place.

---

## HTMX Interactions

| Feature | Trigger | Endpoint | Swap |
|---------|---------|----------|------|
| Submit comment | form submit | `POST /comments` | append to thread |
| Reply to comment | form submit | `POST /comments` (with parent_id) | append under parent |
| Approve/reject comment | button click | `PUT /admin/comments/:id/:action` | swap in place |
| Delete comment (mod) | button click | `DELETE /comments/:id` | remove element |
| Load page of posts | page link click | `GET /:group/:category?page=N` | replace list |
| Search | form submit | `GET /search?q=...` | replace results |
| Toggle follow | button click | `POST /follow/:curator_id` | swap button state |
| Mark notification read | click | `PUT /notifications/:id/read` | swap item |
| Upload image | file input change | `POST /upload` | insert Markdown tag into textarea |
| Preview draft | button click | `GET /admin/posts/:id/preview` | show in preview pane |
| Toggle sidebar | button click | — | client-side CSS toggle (no server round-trip) |
| Toggle theme | button click | `PUT /settings/theme` | swap body class |

---

## Search

1. User types query in the sidebar search bar, submits.
2. Server searches `posts.title` with `LIKE %query%` (published, non-deleted only).
3. If **zero** title matches, fall back to searching `posts.body`.
4. Results rendered using the generic `post_card.templ`, with pagination.
5. Future enhancement: SQLite FTS5 for full-text search performance.

---

## Soft Delete & Hard Delete

| Action | Who | Effect |
|--------|-----|--------|
| Soft-delete post | Curator, Moderator | Sets `is_deleted = 1`. Post hidden from public. |
| Soft-delete comment | Curator, Moderator | Sets `is_deleted = 1`. Comment hidden from public. |
| Restore (undo soft-delete) | Curator only | Sets `is_deleted = 0`. Content visible again. |
| Hard-delete | Curator only | Row removed from database permanently. |

---

## Scheduled Jobs

| Job | Frequency | Description |
|-----|-----------|-------------|
| Email sender | Every few minutes | Process `email_queue` table, send pending emails via Resend. If daily limit hit, skip until next day. |
| Inactive cleanup | Daily | Delete unverified users with `last_active_at` older than 1 month. |

Implementation: Go goroutines with `time.Ticker` launched at server startup. No external cron needed.

---

## Build & Deploy

```bash
# Development
templ generate --watch &
go run ./cmd/server

# Production
templ generate
go build -o atlas ./cmd/server
./atlas
```

### Deployment Artifacts

1. Single binary (`atlas`)
2. Database file (`blog.db`)
3. Uploads directory (`uploads/`)
4. Environment config (Resend API key, etc.)

Hosting TBD (leaning AWS). No Docker required.

---

## Route Summary

```
# Public
GET   /                                -- home (recent from followed curators)
GET   /about/:curator-slug             -- curator about page
GET   /blog                            -- blog landing
GET   /blog/:category                  -- blog sub-category listing
GET   /reviews                         -- reviews landing
GET   /reviews/:category               -- review sub-category listing
GET   /:group/:category/:slug          -- single post view
GET   /search                          -- search results

# Auth
GET   /login                           -- login form
POST  /login                           -- authenticate
GET   /register                        -- registration form
POST  /register                        -- create account
POST  /logout                          -- end session
GET   /verify/:token                   -- email verification

# User
GET   /settings                        -- user settings page
PUT   /settings/profile                -- update username, bio
PUT   /settings/password               -- change password
PUT   /settings/theme                  -- toggle dark/light
PUT   /settings/notifications          -- update email prefs

# Interactions
POST  /comments                        -- submit comment (HTMX)
POST  /follow/:curator_id              -- toggle follow (HTMX)
PUT   /notifications/:id/read          -- mark notification read (HTMX)

# Curator
GET   /admin                           -- dashboard
GET   /admin/posts/new                 -- post editor
POST  /admin/posts                     -- create post
GET   /admin/posts/:id/edit            -- edit post
PUT   /admin/posts/:id                 -- update post
GET   /admin/posts/:id/preview         -- preview draft
DELETE /admin/posts/:id                -- hard delete post
GET   /admin/about/edit                -- edit own about page
PUT   /admin/about                     -- save about page
GET   /admin/categories                -- category manager
POST  /admin/categories                -- create category
PUT   /admin/categories/:id            -- update category
POST  /upload                          -- image upload

# Moderation (Moderator + Curator)
GET   /admin/moderation                -- pending content queue
PUT   /admin/comments/:id/:action      -- approve/reject comment (HTMX)
DELETE /comments/:id                   -- soft-delete comment (HTMX)
DELETE /posts/:id                      -- soft-delete post
PUT   /admin/posts/:id/restore         -- restore soft-deleted post (curator only)
PUT   /admin/comments/:id/restore      -- restore soft-deleted comment (curator only)

# User Management (Curator only)
GET   /admin/users                     -- user list
PUT   /admin/users/:id/role            -- promote/demote
PUT   /admin/users/:id/ban             -- ban user
PUT   /admin/users/:id/unban           -- unban user
```
