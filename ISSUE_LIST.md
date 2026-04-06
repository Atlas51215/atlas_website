# Atlas — Issue List

Derived from `PLAN.md`. Each issue is scoped to roughly **1–3 hours** of developer work. Issues are ordered by dependency — earlier issues must be completed before later ones that depend on them. Each issue lists its dependencies explicitly.

---

## Phase 1: Foundation & Project Scaffold

### ISSUE-001: Restructure Project to Target Layout

**Depends on:** Nothing (starting point)

**Goal:** Migrate from the current flat `main.go` + `internal/blog/` structure to the target project layout defined in `PLAN.md`.

**What to do:**

1. Create the full directory structure:
   ```
   cmd/server/
   internal/handler/
   internal/model/
   internal/middleware/
   internal/render/
   internal/jobs/
   templates/
   static/
   uploads/
   migrations/
   tests/
   ```
2. Move the entry point from `main.go` (root) to `cmd/server/main.go`.
3. Delete `internal/blog/blog.go` and the old `main.go` — the code will be replaced entirely.
4. Create a minimal `cmd/server/main.go` that:
   - Imports the Chi router (`github.com/go-chi/chi/v5`).
   - Starts an HTTP server on `:8080`.
   - Serves a placeholder `"Atlas is running"` response at `/`.
   - Serves static files from `static/`.
5. Update `go.mod` to add the `chi` dependency.
6. Place `htmx.min.js` (version 2.x) in `static/`.
7. Verify the server starts and responds at `http://localhost:8080`.
8. Write a basic test in `tests/` that starts the server and checks the `/` route returns 200.

**Acceptance criteria:**

- `go run ./cmd/server` starts without errors.
- `GET /` returns 200 with "Atlas is running".
- `GET /static/htmx.min.js` returns the HTMX file.
- Old `main.go` and `internal/blog/` are removed.
- Tests pass via `go test -v ./...`.

---

### ISSUE-002: SQLite Database Setup & Migration System

**Depends on:** ISSUE-001

**Goal:** Set up SQLite with WAL mode and create the initial migration containing the full database schema from `PLAN.md`.

**What to do:**

1. Add `github.com/mattn/go-sqlite3` (or `modernc.org/sqlite` for pure Go) as a dependency.
2. Create `migrations/001_initial.sql` containing the **entire** schema from `PLAN.md`:
   - All tables: `categories`, `users`, `posts`, `comments`, `tags`, `post_tags`, `follows`, `notifications`, `email_queue`, `verification_tokens`, `about_pages`.
   - All indexes listed in the plan.
3. Create a database initialization function in `internal/model/db.go` that:
   - Opens (or creates) `blog.db`.
   - Sets the three required pragmas: `journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout=5000`.
   - Reads and executes `migrations/001_initial.sql`.
   - Tracks which migrations have been applied (create a `schema_migrations` table with the migration filename and applied timestamp).
   - Skips migrations that have already been applied.
4. Call this initialization function from `cmd/server/main.go` at startup.
5. Write tests that:
   - Open an in-memory SQLite database.
   - Run the migration.
   - Verify all tables exist by querying `sqlite_master`.
   - Verify the pragmas are set correctly.

**Acceptance criteria:**

- Server creates `blog.db` on first run.
- All tables and indexes from the schema exist.
- Running the server a second time does NOT re-apply migrations.
- Tests pass.

---

### ISSUE-003: Seed Default Categories

**Depends on:** ISSUE-002

**Goal:** Populate the `categories` table with the default categories from `PLAN.md` and create the model layer for reading them.

**What to do:**

1. Create `migrations/002_seed_categories.sql` that inserts all 11 default categories:
   - **Reviews group:** Movies (slug: `movies`, extra: `rating`, `director`, `release_year`), Video Games (`games`, extra: `rating`, `platform`, `developer`), TV Shows (`tv`, extra: `rating`, `seasons`, `network`), Products (`products`, extra: `rating`, `price_range`, `purchase_link`).
   - **Blog group:** Movies, Video Games, TV Shows, Products, General, Dev, Hardware/Tech.
   - Each review category's `extra_fields` should be a JSON array like: `[{"name":"rating","type":"float","label":"Rating","min":0,"max":10}, {"name":"director","type":"text","label":"Director"}, {"name":"release_year","type":"number","label":"Release Year"}]`.
   - Blog categories have `extra_fields` set to `NULL`.
   - Set `sort_order` so they appear in the order listed above.
2. Create `internal/model/category.go` with:
   - A `Category` struct matching the DB schema (including parsed `ExtraFields`).
   - An `ExtraField` struct: `Name`, `Type`, `Label`, and optional `Min`, `Max` fields.
   - Function `AllCategories(db) → []Category` — returns all categories ordered by `sort_order`.
   - Function `CategoriesByGroup(db, groupName) → []Category` — returns categories for a specific group.
   - Function `CategoryBySlugAndGroup(db, slug, groupName) → Category` — returns a single category.
3. Write tests for each query function using a test database seeded with the migration.

**Acceptance criteria:**

- After server start, `SELECT COUNT(*) FROM categories` returns 11.
- `CategoriesByGroup(db, "reviews")` returns 4 categories.
- `CategoriesByGroup(db, "blog")` returns 7 categories.
- `CategoryBySlugAndGroup(db, "movies", "reviews")` returns the correct category with parsed extra fields.
- Tests pass.

---

### ISSUE-004: User Model & Password Hashing

**Depends on:** ISSUE-002

**Goal:** Create the user model layer with account creation, lookup, and bcrypt password handling.

**What to do:**

1. Create `internal/model/user.go` with:
   - A `User` struct matching the DB columns: `ID`, `Username`, `Email`, `PasswordHash`, `Role`, `Bio`, `Theme`, `EmailNotifyPosts`, `EmailNotifyReplies`, `IsBanned`, `VerifiedAt`, `LastActiveAt`, `CreatedAt`.
   - Function `CreateUser(db, username, email, plainPassword) → (User, error)`:
     - Hash the password with bcrypt (cost 12).
     - Insert into users table with role `"unverified"`.
     - Return the created user.
   - Function `UserByID(db, id) → (User, error)`.
   - Function `UserByEmail(db, email) → (User, error)`.
   - Function `UserByUsername(db, username) → (User, error)`.
   - Function `CheckPassword(user, plainPassword) → bool` — uses `bcrypt.CompareHashAndPassword`.
   - Function `UpdateUserRole(db, userID, newRole) → error`.
   - Function `UpdateLastActive(db, userID) → error` — sets `last_active_at` to now.
2. Add `golang.org/x/crypto` dependency for bcrypt.
3. Write tests:
   - Create a user and verify fields.
   - Verify password check works (correct and incorrect passwords).
   - Verify duplicate username/email returns an error.
   - Verify role update works.

**Acceptance criteria:**

- Users can be created with hashed passwords.
- Password verification works correctly.
- All lookup functions return the correct user.
- Duplicate usernames and emails are rejected.
- Tests pass.

---

### ISSUE-005: Templ Setup & Base Layout Template

**Depends on:** ISSUE-001

**Goal:** Set up the Templ templating engine and create the base HTML layout that all pages will use.

**What to do:**

1. Add the `github.com/a-h/templ` dependency.
2. Install the `templ` CLI tool (document the install command in a comment in `main.go`).
3. Create `templates/layout.templ`:
   - Full HTML5 document with `<head>` (meta charset, viewport, title, link to CSS, HTMX script).
   - `<body>` with a theme class (default `"dark"`).
   - A two-column layout: sidebar slot (left) and main content slot (right).
   - Use templ's `children` for the main content slot.
   - Include a `<div id="sidebar">` placeholder that `nav.templ` will fill.
4. Create `templates/nav.templ`:
   - Sidebar container with the structure from `PLAN.md`:
     - Logo area (placeholder text "Atlas" for now — logo SVG comes later).
     - Search bar: a `<form>` with `action="/search"` and an input field, using `hx-get="/search"` and `hx-target="#main-content"` for HTMX.
     - Nav links: Home (`/`), About Me (`/about`), Blog (`/blog`) with expandable sub-categories, Reviews (`/reviews`) with expandable sub-categories.
     - Ad slot placeholder `<div>` at the bottom.
   - Accept a list of categories (grouped by `blog`/`reviews`) so the sub-category links are dynamic.
   - Mark the current page's link as active (accept a `currentPath` parameter).
5. Create `internal/render/render.go`:
   - A helper function that renders a templ component to an `http.ResponseWriter`.
6. Update `cmd/server/main.go` to serve the layout at `/` with the nav sidebar.
7. Run `templ generate` to compile templates and verify no errors.
8. Write a test that renders the layout and checks the HTML output contains expected elements.

**Acceptance criteria:**

- `templ generate` runs without errors.
- `GET /` returns a full HTML page with sidebar nav, search bar, and nav links.
- Nav links for Blog and Reviews include sub-category links.
- Tests pass.

---

### ISSUE-006: Tailwind CSS Setup (Standalone CLI)

**Depends on:** ISSUE-005

**Goal:** Set up Tailwind CSS using the standalone CLI (no Node.js) and create the base styles for dark/light themes.

**What to do:**

1. Download the Tailwind CSS standalone CLI binary for the project's target platform.
2. Create `static/css/input.css` with:
   - `@tailwind base;`, `@tailwind components;`, `@tailwind utilities;`.
   - Custom theme colors as CSS variables:
     - Dark mode: background `#1a1a1a` (very dark gray), text `#ffffff`, accent `#dc2626` (red).
     - Light mode: background `#f5f5f5` (very light gray), text `#000000`, accent `#dc2626` (red).
   - Apply theme via a `.dark` / `.light` class on `<body>`.
3. Create `tailwind.config.js` with:
   - Content paths pointing to `templates/**/*.templ` and `templates/**/*.go` (templ generates Go files).
   - Dark mode set to `class` strategy.
4. Run `tailwindcss -i static/css/input.css -o static/css/output.css` and verify it generates the output file.
5. Update `templates/layout.templ` to link to `static/css/output.css`.
6. Style the sidebar nav from ISSUE-005 with Tailwind classes:
   - Dark background for sidebar, proper text colors, hover states on nav links, active link highlight in red.
   - Responsive: sidebar visible on desktop, hidden on mobile.
7. Add a build instruction comment or Makefile target for regenerating CSS.

**Acceptance criteria:**

- `static/css/output.css` is generated and included in the HTML.
- The sidebar is styled: dark background, white text, red accents.
- The page looks correct in both dark and light modes (toggle by changing the body class manually).
- On narrow viewports, the sidebar is hidden.
- No Node.js is required.

---

## Phase 2: Authentication & Sessions

### ISSUE-007: Session Management Infrastructure

**Depends on:** ISSUE-002, ISSUE-004

**Goal:** Implement cookie-based session management for login persistence.

**What to do:**

1. Create a `sessions` table in a new migration (`003_sessions.sql`):
   ```sql
   CREATE TABLE sessions (
       token TEXT PRIMARY KEY,
       user_id INTEGER NOT NULL REFERENCES users(id),
       created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
       expires_at DATETIME NOT NULL
   );
   CREATE INDEX idx_sessions_user ON sessions(user_id);
   ```
2. Create `internal/model/session.go` with:
   - Function `CreateSession(db, userID) → (token string, error)`:
     - Generate a 32-byte random token, base64-encode it.
     - Insert into sessions with expiry 30 days from now.
     - Return the token.
   - Function `SessionUser(db, token) → (User, error)` — look up session token, check expiry, return the associated user. Delete expired sessions.
   - Function `DeleteSession(db, token) → error` — for logout.
   - Function `DeleteUserSessions(db, userID) → error` — for "log out everywhere".
3. Write tests:
   - Create a session, retrieve the user from it.
   - Verify expired sessions return an error.
   - Verify deleted sessions return an error.

**Acceptance criteria:**

- Sessions are created with a secure random token.
- Valid tokens return the correct user.
- Expired tokens are rejected.
- Logout deletes the session.
- Tests pass.

---

### ISSUE-008: Auth Middleware (Role-Based Access)

**Depends on:** ISSUE-007

**Goal:** Create middleware that reads the session cookie, loads the user, and enforces role-based access.

**What to do:**

1. Create `internal/middleware/auth.go` with:
   - A context key type for storing the current user.
   - `AuthMiddleware(db)` — reads the `session` cookie, calls `SessionUser`, stores the user in the request context. If no cookie or invalid session, context has no user (visitor).
   - `RequireAuth(next)` — middleware that returns 401 if no user in context.
   - `RequireRole(roles ...string)` — middleware that returns 403 if the user's role is not in the allowed list.
   - `RequireVerified(next)` — shorthand for `RequireRole("verified", "moderator", "curator")`.
   - `RequireModerator(next)` — shorthand for `RequireRole("moderator", "curator")`.
   - `RequireCurator(next)` — shorthand for `RequireRole("curator")`.
   - Helper `UserFromContext(ctx) → *User` — returns nil if no user.
2. Also update `last_active_at` for the user on each authenticated request (call `UpdateLastActive`).
3. Write tests:
   - Request with no cookie → context has nil user.
   - Request with valid session cookie → context has the correct user.
   - `RequireAuth` blocks unauthenticated requests with 401.
   - `RequireRole("curator")` blocks verified users with 403.
   - `RequireRole("curator")` allows curators.

**Acceptance criteria:**

- Middleware correctly reads sessions and populates context.
- Role checks return correct HTTP status codes.
- `last_active_at` is updated on authenticated requests.
- Tests pass.

---

### ISSUE-009: Registration Handler & Page

**Depends on:** ISSUE-004, ISSUE-005, ISSUE-008

**Goal:** Build the user registration form page and the POST handler that creates an account.

**What to do:**

1. Create `templates/auth_register.templ`:
   - Registration form with fields: Username, Email, Password, Confirm Password.
   - Form submits via standard POST to `/register` (not HTMX — full page response).
   - Display validation errors inline (username taken, email taken, passwords don't match, password too short).
   - Link to login page: "Already have an account? Log in."
2. Create `internal/handler/auth.go` with:
   - `GET /register` — renders the registration form. If user is already logged in, redirect to `/`.
   - `POST /register`:
     - Validate: username 3–30 chars (alphanumeric + underscores), email format, password min 8 chars, password confirmation matches.
     - Call `CreateUser(db, username, email, password)`.
     - On duplicate username/email, re-render form with error.
     - On success, create a session (call `CreateSession`), set the cookie (`HttpOnly`, `Secure`, `SameSite=Strict`, path `/`, 30-day max-age), redirect to `/`.
   - Note: email verification is a separate issue — for now, accounts are created as `"unverified"` but functional.
3. Register routes in `cmd/server/main.go` using Chi.
4. Write tests:
   - `GET /register` returns 200 with a form.
   - `POST /register` with valid data creates a user and sets a session cookie.
   - `POST /register` with duplicate username returns the form with an error message.
   - `POST /register` with mismatched passwords returns the form with an error message.

**Acceptance criteria:**

- Registration page renders correctly.
- Valid registration creates a user, starts a session, and redirects.
- Validation errors are shown inline on the form.
- Tests pass.

---

### ISSUE-010: Login & Logout Handlers

**Depends on:** ISSUE-009

**Goal:** Build the login form page, the POST handler for authentication, and the logout endpoint.

**What to do:**

1. Create `templates/auth_login.templ`:
   - Login form with fields: Email, Password.
   - Standard POST to `/login`.
   - Display error: "Invalid email or password" (don't reveal which is wrong).
   - Link to registration: "Don't have an account? Register."
2. Add to `internal/handler/auth.go`:
   - `GET /login` — render the login form. If already logged in, redirect to `/`.
   - `POST /login`:
     - Look up user by email.
     - Check password with `CheckPassword`.
     - If banned, show error: "Your account has been suspended."
     - On success, create session, set cookie, redirect to `/`.
     - On failure, re-render with "Invalid email or password."
   - `POST /logout`:
     - Delete the session from the database.
     - Clear the session cookie (set max-age to -1).
     - Redirect to `/`.
3. Register routes.
4. Write tests:
   - `POST /login` with correct credentials sets a session cookie and redirects.
   - `POST /login` with wrong password returns the form with an error.
   - `POST /logout` clears the session and redirects.
   - Banned user cannot log in.

**Acceptance criteria:**

- Login page renders.
- Correct credentials create a session.
- Wrong credentials show a generic error message.
- Logout clears the session.
- Tests pass.

---

## Phase 3: Core Content — Posts & Categories

### ISSUE-011: Post Model Layer

**Depends on:** ISSUE-002, ISSUE-003

**Goal:** Create the model layer for creating, reading, updating, and listing posts.

**What to do:**

1. Create `internal/model/post.go` with:
   - A `Post` struct: `ID`, `CategoryID`, `AuthorID`, `Title`, `Slug`, `Body` (Markdown source), `ExtraData` (JSON string), `Status` (`"draft"` or `"published"`), `IsDeleted`, `DeletedBy`, `PublishedAt`, `CreatedAt`, `UpdatedAt`.
   - Also include joined fields for display: `AuthorUsername`, `CategorySlug`, `CategoryName`, `GroupName`.
   - Function `CreatePost(db, categoryID, authorID, title, slug, body, extraData, status) → (Post, error)`.
   - Function `UpdatePost(db, postID, title, slug, body, extraData, status) → error`. Set `updated_at` to now. If status changes from `"draft"` to `"published"`, set `published_at`.
   - Function `PostBySlug(db, categorySlug, groupName, postSlug) → (Post, error)` — joins with categories for routing.
   - Function `PostByID(db, postID) → (Post, error)`.
   - Function `PostsByCategory(db, categoryID, page, perPage) → ([]Post, totalCount, error)` — published, non-deleted posts, ordered by `published_at DESC`. Include pagination.
   - Function `PostsByAuthor(db, authorID, page, perPage) → ([]Post, totalCount, error)` — all posts (including drafts) for the author.
   - Function `RecentPosts(db, page, perPage) → ([]Post, totalCount, error)` — all published non-deleted posts across all categories.
   - Function `SoftDeletePost(db, postID, deletedByUserID) → error`.
   - Function `RestorePost(db, postID) → error`.
   - Function `HardDeletePost(db, postID) → error`.
   - Slug generation helper: `GenerateSlug(title) → string` — lowercase, replace spaces with hyphens, strip non-alphanumeric characters.
2. Write tests for each function, especially:
   - Pagination returns correct pages and total count.
   - Soft delete hides posts from listings.
   - Status filtering (only published posts in public queries).

**Acceptance criteria:**

- All CRUD operations work correctly.
- Pagination returns correct results.
- Soft-deleted posts are excluded from public queries.
- Draft posts are excluded from public queries.
- Slug generation produces clean URL-friendly strings.
- Tests pass.

---

### ISSUE-012: Markdown Rendering with Goldmark

**Depends on:** ISSUE-001

**Goal:** Set up Markdown-to-HTML rendering for post and comment bodies.

**What to do:**

1. Add `github.com/yuin/goldmark` as a dependency.
2. Create `internal/render/markdown.go` with:
   - Function `RenderMarkdown(markdownSource string) → (template.HTML, error)`:
     - Use goldmark with these extensions: GFM (GitHub Flavored Markdown) for tables, strikethrough, autolinks, and task lists.
     - Enable safe mode (strip raw HTML to prevent XSS).
     - Return the rendered HTML.
   - Function `RenderMarkdownUnsafe(markdownSource string) → (template.HTML, error)`:
     - Same but without safe mode — for curator posts where raw HTML is intentional.
     - Note: only curators can create posts, so this is acceptable.
3. Write tests:
   - Basic Markdown: headings, bold, italic, links, images, code blocks.
   - GFM features: tables, strikethrough, task lists.
   - Safe mode strips `<script>` tags and other dangerous HTML.
   - Unsafe mode preserves raw HTML.

**Acceptance criteria:**

- Markdown renders to correct HTML.
- Safe mode blocks XSS vectors.
- GFM extensions work.
- Tests pass.

---

### ISSUE-013: Post Listing Pages (Blog & Reviews Landing)

**Depends on:** ISSUE-005, ISSUE-006, ISSUE-011, ISSUE-003

**Goal:** Build the blog and reviews landing pages, and the sub-category listing pages.

**What to do:**

1. Create `templates/components/post_card.templ`:
   - Displays: title (linked to post), author name, published date, category name, and a short body preview (first 200 chars of rendered Markdown, stripped of HTML tags).
   - For review categories: also display rating (if present) using the formatting rules from the plan (4.0 → "4", 2.20 → "2.2").
2. Create `templates/components/pagination.templ`:
   - Numbered page links: Previous, 1, 2, ..., N, Next.
   - Use HTMX (`hx-get`, `hx-target="#post-list"`, `hx-swap="innerHTML"`) so page changes don't reload the full page.
   - Highlight the current page.
3. Create `templates/post_list.templ`:
   - Takes: category info (or nil for landing pages), list of posts, pagination data, current group.
   - Heading: category name (or "Blog" / "Reviews" for landing pages).
   - List of `post_card` components.
   - Pagination at the bottom.
4. Create `internal/handler/posts.go` with:
   - `GET /blog` — landing page listing all published blog posts (most recent first), paginated.
   - `GET /blog/:category` — posts in a specific blog sub-category.
   - `GET /reviews` — landing page listing all published review posts.
   - `GET /reviews/:category` — posts in a specific review sub-category.
   - Each handler checks `HX-Request` header: if present, return just the post list fragment; otherwise, return the full page with layout.
5. Register routes using Chi route groups.
6. Write tests:
   - Landing pages return 200.
   - Sub-category pages return 200.
   - Invalid category slug returns 404.
   - HTMX requests return fragments (no `<html>` wrapper).

**Acceptance criteria:**

- `/blog` and `/reviews` show paginated post listings.
- `/blog/movies` shows only blog posts in the movies category.
- Post cards show the correct info.
- Pagination works with HTMX.
- Tests pass.

---

### ISSUE-014: Single Post View Page

**Depends on:** ISSUE-012, ISSUE-013

**Goal:** Build the single post view page that displays a full post with rendered Markdown and category-specific extra fields.

**What to do:**

1. Create `templates/components/custom_fields.templ`:
   - Takes the category's `ExtraFields` schema and the post's `ExtraData` JSON.
   - Renders each field as a labeled value:
     - `text` → plain text.
     - `number` → formatted number.
     - `float` → formatted with the rating display rules (strip trailing zeros).
     - `url` → clickable link.
   - Used for review posts to show rating, director, platform, etc.
2. Create `templates/components/rating.templ`:
   - Takes a float64 value and renders it with the display rules.
   - Example: `4.0` → "4", `2.20` → "2.2", `3.14` → "3.14".
3. Create `templates/post_view.templ`:
   - Full post: title, author, published date, category breadcrumb.
   - Rendered Markdown body.
   - Custom fields section (if the category has extra fields).
   - Comment section placeholder (comments are a separate issue).
4. Add to `internal/handler/posts.go`:
   - `GET /:group/:category/:slug` — look up the post by group, category slug, and post slug. Render the post view page.
   - If the post is a draft or soft-deleted, return 404 (unless the viewer is the author or a curator).
   - Render the Markdown body to HTML using goldmark.
5. Write tests:
   - Published post returns 200 with rendered HTML.
   - Draft post returns 404 for visitors.
   - Post with extra fields displays them correctly.
   - Rating formatting is correct.

**Acceptance criteria:**

- Single post page shows the full rendered post.
- Extra fields (rating, director, etc.) are displayed for review posts.
- Ratings follow the formatting rules.
- Drafts are hidden from the public.
- Tests pass.

---

### ISSUE-015: Homepage — Recent Posts Feed

**Depends on:** ISSUE-013, ISSUE-008

**Goal:** Build the homepage that shows recent posts from followed curators (or all posts for visitors).

**What to do:**

1. Create `internal/model/follow.go` with:
   - Function `Follow(db, followerID, curatorID) → error`.
   - Function `Unfollow(db, followerID, curatorID) → error`.
   - Function `IsFollowing(db, followerID, curatorID) → bool`.
   - Function `FollowedCurators(db, userID) → []int` — returns IDs of curators the user follows.
2. Create `templates/home.templ`:
   - If user is logged in and follows curators: show recent posts from those curators.
   - If user is logged in but follows no one: show all recent posts with a message "Follow curators to personalize your feed."
   - If visitor (not logged in): show all recent posts.
   - Use `post_card.templ` and `pagination.templ`.
3. Add to `internal/model/post.go`:
   - Function `PostsByFollowedCurators(db, followerID, page, perPage) → ([]Post, totalCount, error)` — posts from curators the user follows.
4. Create `internal/handler/home.go`:
   - `GET /` — determine if user is logged in and has follows, then query the appropriate posts.
5. Write tests:
   - Visitor sees all recent posts.
   - User following curators sees only those curators' posts.
   - User following nobody sees all posts.

**Acceptance criteria:**

- Homepage shows the correct posts based on login and follow state.
- Post cards render correctly.
- Pagination works.
- Tests pass.

---

## Phase 4: Admin & Content Management

### ISSUE-016: Admin Dashboard Page

**Depends on:** ISSUE-008, ISSUE-011

**Goal:** Build the admin dashboard with an overview of posts, comments, and quick actions.

**What to do:**

1. Create `templates/admin/dashboard.templ`:
   - Heading: "Admin Dashboard".
   - Quick stats: total posts (published/draft), total comments (pending/approved), total users by role.
   - Recent posts list (last 10) with edit/delete links.
   - Link to: Post Editor, Category Manager, User Manager, Moderation Queue.
2. Create `internal/handler/admin.go`:
   - `GET /admin` — curator-only. Query stats from the database, render the dashboard.
3. Set up route group `/admin/*` in `cmd/server/main.go` with `RequireCurator` middleware.
4. Write tests:
   - Curator can access `/admin` and sees stats.
   - Non-curator gets 403.
   - Stats match the data in the database.

**Acceptance criteria:**

- `/admin` shows dashboard with correct stats.
- Only curators can access it.
- Links to other admin pages are present.
- Tests pass.

---

### ISSUE-017: Post Editor — Create & Edit Posts

**Depends on:** ISSUE-011, ISSUE-016, ISSUE-003

**Goal:** Build the post creation and editing form for curators.

**What to do:**

1. Create `templates/post_form.templ`:
   - Category dropdown (grouped by blog/reviews).
   - Title input.
   - Slug input (auto-generated from title via JS-free approach: display generated slug, allow manual override).
   - Body: plain Markdown `<textarea>` with generous height.
   - Dynamic extra fields section: when a category is selected, show the category's custom fields. Use HTMX — when the category dropdown changes, `hx-get="/admin/categories/:id/fields"` returns the field inputs.
   - Status radio: Draft or Published.
   - Save button.
   - Display validation errors inline.
2. Add to `internal/handler/admin.go`:
   - `GET /admin/posts/new` — render blank post form with category list.
   - `POST /admin/posts` — validate and create the post. On success, redirect to the post view page.
   - `GET /admin/posts/:id/edit` — render the form pre-filled with the post's data.
   - `PUT /admin/posts/:id` — validate and update the post.
   - `GET /admin/categories/:id/fields` — HTMX endpoint that returns the HTML form fields for a category's custom fields.
3. Validation:
   - Title required, max 200 chars.
   - Slug required, must be unique within the category.
   - Body required.
   - Extra fields validated against the category schema (e.g., rating must be 0–10).
4. Write tests:
   - Create a post and verify it exists in the database.
   - Edit a post and verify changes.
   - Validation errors are returned correctly.
   - Changing category loads the correct extra fields.

**Acceptance criteria:**

- Curators can create and edit posts.
- Extra fields render dynamically based on category selection.
- Validation works correctly.
- Publishing a post sets `published_at`.
- Tests pass.

---

### ISSUE-018: Post Preview for Drafts

**Depends on:** ISSUE-017, ISSUE-012

**Goal:** Allow curators to preview a saved draft before publishing.

**What to do:**

1. Create `templates/post_preview.templ`:
   - Same layout as `post_view.templ` but with a banner: "This is a preview. This post is not published yet."
   - Buttons: "Edit" (link to edit form) and "Publish" (HTMX PUT to update status).
2. Add to `internal/handler/admin.go`:
   - `GET /admin/posts/:id/preview` — curator only. Render the draft post using the preview template.
   - The "Publish" button: `hx-put="/admin/posts/:id"` with `status=published`, then redirect to the live post URL.
3. Write tests:
   - Preview renders the Markdown body.
   - Preview shows the "not published" banner.
   - Non-curators cannot access preview.

**Acceptance criteria:**

- Curators can preview drafts with fully rendered Markdown.
- Preview clearly indicates the post is not published.
- "Publish" button publishes the post.
- Tests pass.

---

### ISSUE-019: Category Manager (Admin)

**Depends on:** ISSUE-003, ISSUE-016

**Goal:** Build the admin page for creating and editing categories with custom fields.

**What to do:**

1. Add to `internal/model/category.go`:
   - Function `CreateCategory(db, slug, name, groupName, extraFields, sortOrder) → (Category, error)`.
   - Function `UpdateCategory(db, categoryID, name, slug, extraFields, sortOrder) → error`.
2. Create `templates/admin/categories.templ`:
   - List all categories grouped by Blog/Reviews.
   - Each category shows: name, slug, group, number of custom fields, sort order, edit button.
   - "Add Category" button opens an inline form (or a separate section).
   - Category form:
     - Group dropdown: "blog" or "reviews".
     - Name, slug inputs.
     - Custom fields editor: a repeatable section where each row has: field name, field type dropdown (`text`, `number`, `float`, `url`), field label, optional min/max (for number/float). "Add field" button appends a new row. "Remove" button on each row.
     - Sort order input.
     - Save button.
3. Add to `internal/handler/admin.go`:
   - `GET /admin/categories` — render the category manager page.
   - `POST /admin/categories` — validate and create a new category.
   - `PUT /admin/categories/:id` — validate and update a category.
4. Validation:
   - Slug must be unique within the group.
   - Name required.
   - Extra fields JSON must be well-formed.
5. Write tests:
   - Create a category and verify it appears in the list.
   - Update a category's name and custom fields.
   - Duplicate slug in the same group is rejected.

**Acceptance criteria:**

- Curators can create and edit categories.
- Custom fields can be added/removed dynamically.
- Validation works.
- Tests pass.

---

### ISSUE-020: User Manager (Admin)

**Depends on:** ISSUE-004, ISSUE-016

**Goal:** Build the admin page for managing users — promoting, demoting, banning, and unbanning.

**What to do:**

1. Add to `internal/model/user.go`:
   - Function `AllUsers(db, page, perPage) → ([]User, totalCount, error)`.
   - Function `BanUser(db, userID) → error` — sets `is_banned = 1`.
   - Function `UnbanUser(db, userID) → error` — sets `is_banned = 0`.
2. Create `templates/admin/users.templ`:
   - Paginated list of users showing: username, email, role, banned status, last active date.
   - Each user row has action buttons:
     - Role dropdown (or buttons): promote to curator, moderator, verified, or demote to unverified.
     - Ban/Unban toggle button.
   - Actions use HTMX to swap the updated row in-place.
3. Add to `internal/handler/admin.go`:
   - `GET /admin/users` — curator only. Render the user list.
   - `PUT /admin/users/:id/role` — curator only. Update the user's role. Return the updated row as HTMX fragment.
   - `PUT /admin/users/:id/ban` — moderator+curator. Ban the user. Return the updated row.
   - `PUT /admin/users/:id/unban` — moderator+curator. Unban the user. Return the updated row.
4. Prevent curators from demoting themselves.
5. Write tests:
   - User list renders with correct data.
   - Role promotion/demotion works.
   - Ban/unban works.
   - Non-curators cannot access user management.
   - Curators cannot demote themselves.

**Acceptance criteria:**

- User list shows all users with their roles and statuses.
- Role changes and bans work via HTMX.
- Self-demotion is prevented.
- Tests pass.

---

## Phase 5: Comments & Moderation

### ISSUE-021: Comment Model Layer

**Depends on:** ISSUE-002

**Goal:** Create the model layer for threaded comments with moderation status.

**What to do:**

1. Create `internal/model/comment.go` with:
   - A `Comment` struct: `ID`, `PostID`, `UserID`, `ParentID` (nullable), `Body`, `Status` (`"pending"`, `"approved"`, `"rejected"`), `IsDeleted`, `DeletedBy`, `CreatedAt`. Also include `Username` (joined).
   - Function `CreateComment(db, postID, userID, parentID, body) → (Comment, error)`:
     - Status defaults to `"pending"`.
     - If `parentID` is provided, verify the parent comment exists and belongs to the same post.
   - Function `CommentsByPost(db, postID) → ([]Comment, error)` — returns all approved, non-deleted comments ordered by `created_at`. Include parent-child relationships for threading.
   - Function `PendingComments(db, page, perPage) → ([]Comment, totalCount, error)` — for moderation queue.
   - Function `ApproveComment(db, commentID) → error`.
   - Function `RejectComment(db, commentID) → error`.
   - Function `SoftDeleteComment(db, commentID, deletedByUserID) → error`.
   - Function `RestoreComment(db, commentID) → error`.
   - Function `HardDeleteComment(db, commentID) → error`.
2. Write tests:
   - Create a top-level comment and a reply.
   - Verify threading: reply's `ParentID` matches parent's `ID`.
   - Pending comments don't appear in `CommentsByPost`.
   - Approve → comment appears. Reject → still hidden.
   - Soft delete and restore work.

**Acceptance criteria:**

- Comments can be created with optional parent.
- Threading works correctly.
- Only approved comments appear in public queries.
- Moderation actions change status correctly.
- Tests pass.

---

### ISSUE-022: Comment Submission & Display (HTMX)

**Depends on:** ISSUE-021, ISSUE-014, ISSUE-008

**Goal:** Add comment submission and threaded display to the single post page.

**What to do:**

1. Create `templates/comment.templ`:
   - Renders a single comment: username, date, rendered Markdown body, and a "Reply" button.
   - If the current user is a moderator or curator: show "Delete" button.
   - Reply button toggles a reply form inline (HTMX or CSS toggle).
2. Create `templates/components/comment_thread.templ`:
   - Recursive component: render a comment, then indent and render its children.
   - Accept a flat list of comments and build the tree in the template (or accept a pre-built tree).
3. Create `templates/comment_form.templ`:
   - Textarea for comment body (Markdown).
   - Hidden input for `post_id` and optional `parent_id`.
   - Submit via HTMX: `hx-post="/comments"`, `hx-target="#comments-section"`, `hx-swap="innerHTML"`.
   - Show login prompt if user is not logged in.
   - Show "your email must be verified" if user is unverified.
4. Create `internal/handler/comments.go`:
   - `POST /comments` — verified users only. Create the comment. Return the updated comment thread as an HTMX fragment.
5. Update `templates/post_view.templ` to include the comment thread and comment form.
6. Write tests:
   - Submitting a comment creates it with `"pending"` status.
   - Comment appears in the thread after approval.
   - Replies are nested under the parent.
   - Unauthenticated users see a login prompt.
   - Unverified users see a verification prompt.

**Acceptance criteria:**

- Comments submit via HTMX without full page reload.
- Threaded comments display correctly.
- Pending comments are not shown to other users.
- Role checks are enforced.
- Tests pass.

---

### ISSUE-023: Moderation Queue

**Depends on:** ISSUE-021, ISSUE-022, ISSUE-016

**Goal:** Build the moderation page where moderators and curators can approve/reject pending comments and manage soft-deleted content.

**What to do:**

1. Create `templates/admin/moderation.templ`:
   - **Pending comments section:** list of comments awaiting approval. Each shows: post title (linked), commenter username, comment body preview, submitted date. Action buttons: Approve, Reject (HTMX, swap the item in-place).
   - **Soft-deleted content section:** list of soft-deleted posts and comments with "Restore" (curator only) and "Hard Delete" (curator only) buttons.
2. Add to `internal/handler/admin.go` or create `internal/handler/moderation.go`:
   - `GET /admin/moderation` — moderator+curator. Render the queue.
   - `PUT /admin/comments/:id/approve` — approve comment, return updated fragment.
   - `PUT /admin/comments/:id/reject` — reject comment, return updated fragment.
   - `DELETE /comments/:id` — soft-delete comment (moderator+curator).
   - `PUT /admin/comments/:id/restore` — restore soft-deleted comment (curator only).
   - `DELETE /posts/:id` — soft-delete post (moderator+curator).
   - `PUT /admin/posts/:id/restore` — restore soft-deleted post (curator only).
3. Write tests:
   - Pending comments appear in the queue.
   - Approving a comment changes its status and removes it from the queue.
   - Soft-deleted content appears in the deleted section.
   - Restoring content makes it visible again.
   - Moderators cannot restore (403).

**Acceptance criteria:**

- Moderation queue shows all pending comments.
- Approve/reject works via HTMX.
- Soft-deleted content can be restored by curators.
- Role checks are correct.
- Tests pass.

---

## Phase 6: User Features

### ISSUE-024: User Settings Page

**Depends on:** ISSUE-008, ISSUE-005

**Goal:** Build the user settings page where users can update their profile and preferences.

**What to do:**

1. Add to `internal/model/user.go`:
   - Function `UpdateUsername(db, userID, newUsername) → error`.
   - Function `UpdatePassword(db, userID, newPasswordHash) → error`.
   - Function `UpdateBio(db, userID, bio) → error`.
   - Function `UpdateTheme(db, userID, theme) → error`.
   - Function `UpdateEmailPrefs(db, userID, notifyPosts, notifyReplies) → error`.
2. Create `templates/settings.templ`:
   - Sections (each can be submitted independently via HTMX):
     - **Username:** input field, save button. `hx-put="/settings/profile"`.
     - **Bio:** textarea, save button. Same endpoint.
     - **Password:** current password, new password, confirm new password. `hx-put="/settings/password"`.
     - **Email:** display only (not editable).
     - **Theme:** toggle switch (dark/light). `hx-put="/settings/theme"`.
     - **Avatar:** display auto-generated initials/identicon. Not editable (as per plan — reduces moderation burden).
     - **Email Notifications:** checkboxes for "New posts from followed curators" and "Comment replies". `hx-put="/settings/notifications"`.
   - Show success/error feedback inline after each save.
3. Create `internal/handler/settings.go`:
   - `GET /settings` — render the settings page for the logged-in user.
   - `PUT /settings/profile` — update username and/or bio. Validate username uniqueness.
   - `PUT /settings/password` — verify current password, hash new password, update. Return success/error.
   - `PUT /settings/theme` — toggle between "dark" and "light". Also set a cookie for visitors.
   - `PUT /settings/notifications` — update email notification preferences.
4. Write tests:
   - Username change works and rejects duplicates.
   - Password change requires correct current password.
   - Theme toggle updates the database and sets a cookie.
   - Notification preferences update correctly.

**Acceptance criteria:**

- All settings sections render and save correctly.
- Validation errors are shown inline.
- Theme changes immediately affect the page.
- Tests pass.

---

### ISSUE-025: Theme Toggle (Dark/Light Mode)

**Depends on:** ISSUE-006, ISSUE-024

**Goal:** Implement theme switching in the UI with persistence for both logged-in users and visitors.

**What to do:**

1. Create `internal/middleware/theme.go`:
   - Middleware that determines the current theme:
     - If user is logged in → use `user.Theme` from the database.
     - If visitor → read a `theme` cookie. Default to `"dark"` if no cookie.
   - Store the theme in the request context so templates can access it.
2. Update `templates/layout.templ`:
   - Set the `<body>` class to `"dark"` or `"light"` based on the theme from context.
3. Add a theme toggle button to `templates/nav.templ`:
   - A sun/moon icon button in the sidebar.
   - For logged-in users: `hx-put="/settings/theme"`, then swap the body class.
   - For visitors: use a minimal inline `<script>` (acceptable since it's purely cosmetic) to toggle the class and set a cookie. Alternatively, make the PUT endpoint work for visitors by just setting the cookie.
4. Update `PUT /settings/theme` in `internal/handler/settings.go`:
   - If user is logged in: update the database.
   - Always set/update the `theme` cookie (for consistency and for visitors).
   - Return an HTMX response that swaps the body class (use `HX-Trigger` or return a small fragment).
5. Write tests:
   - Default theme is dark.
   - Logged-in user's theme preference is read from the database.
   - Visitor's theme preference is read from a cookie.
   - Toggle changes the theme and persists it.

**Acceptance criteria:**

- Site starts in dark mode.
- Theme toggle works for both logged-in users and visitors.
- Preference persists across page loads.
- Tests pass.

---

### ISSUE-026: Follow/Unfollow Curators (HTMX)

**Depends on:** ISSUE-015

**Goal:** Add follow/unfollow buttons on curator content so verified users can follow curators.

**What to do:**

1. Add a "Follow" / "Unfollow" button to:
   - `templates/post_view.templ` — near the author name.
   - `templates/about.templ` — on the curator's about page (built later, but wire the component now).
2. Create a `templates/components/follow_button.templ`:
   - Takes: `curatorID`, `isFollowing` (bool), and `isLoggedIn` (bool).
   - If not logged in or unverified: show disabled button or "Log in to follow".
   - If following: show "Unfollow" button with `hx-post="/follow/:curator_id"`.
   - If not following: show "Follow" button with `hx-post="/follow/:curator_id"`.
   - After toggle, swap the button in-place with the new state.
3. Create `internal/handler/follow.go` (or add to an existing handler file):
   - `POST /follow/:curator_id` — verified users only. Toggle follow state (if following, unfollow; if not, follow). Return the updated button as an HTMX fragment.
4. Write tests:
   - Following a curator creates a follow record.
   - Unfollowing removes it.
   - The button state reflects the current follow status.
   - Unverified users cannot follow.

**Acceptance criteria:**

- Follow/unfollow toggles via HTMX.
- Button state updates without page reload.
- Only verified+ users can follow.
- Tests pass.

---

### ISSUE-027: Search Handler & Page

**Depends on:** ISSUE-011, ISSUE-005, ISSUE-013

**Goal:** Implement the search feature — search post titles first, fall back to body search.

**What to do:**

1. Add to `internal/model/post.go`:
   - Function `SearchPosts(db, query, page, perPage) → ([]Post, totalCount, error)`:
     - First search `posts.title LIKE %query%` (published, non-deleted only).
     - If zero title matches, search `posts.body LIKE %query%`.
     - Return results with pagination.
2. Create `templates/search.templ`:
   - Show the search query at the top: "Results for 'query'".
   - If no results: "No posts found for 'query'."
   - List results using `post_card.templ` with pagination.
3. Create `internal/handler/search.go`:
   - `GET /search?q=...` — run the search, render results.
   - If `HX-Request` header is present, return just the results fragment.
   - Empty query redirects to `/`.
4. Update the search bar in `templates/nav.templ`:
   - Use `hx-get="/search"`, `hx-target="#main-content"`, `hx-trigger="submit"`, with `hx-include` to include the search input.
5. Write tests:
   - Search by title returns matching posts.
   - When no title matches, body matches are returned.
   - No results shows the empty state.
   - Pagination works.

**Acceptance criteria:**

- Search finds posts by title first, then body.
- Results render with post cards and pagination.
- HTMX search works from the sidebar.
- Tests pass.

---

### ISSUE-028: Curator About Pages

**Depends on:** ISSUE-008, ISSUE-012

**Goal:** Build the per-curator about page that curators can edit.

**What to do:**

1. Create `internal/model/about.go`:
   - Function `GetAboutPage(db, curatorID) → (AboutPage, error)`.
   - Function `UpsertAboutPage(db, curatorID, body) → error` — insert or update.
   - An `AboutPage` struct: `ID`, `CuratorID`, `Body` (Markdown), `UpdatedAt`.
2. Add to `internal/model/user.go`:
   - Function `UserBySlug(db, slug) → (User, error)` — the "slug" for about pages is the lowercase username. (Or add a `slug` column to users — simpler to just use the username.)
3. Create `templates/about.templ`:
   - Curator's username/display name at the top.
   - Rendered Markdown body.
   - If the logged-in user is this curator: show "Edit" link.
   - Follow button (from ISSUE-026).
4. Create `internal/handler/about.go`:
   - `GET /about/:curator-slug` — look up curator, get their about page, render.
   - `GET /admin/about/edit` — curator only. Render a Markdown textarea with current body.
   - `PUT /admin/about` — curator only. Save the about page body.
5. Write tests:
   - About page renders for a valid curator.
   - Invalid curator slug returns 404.
   - Editing saves the content.
   - Rendered Markdown appears correctly.

**Acceptance criteria:**

- Each curator has an about page at `/about/:slug`.
- Curators can edit their own about page.
- Markdown is rendered.
- Tests pass.

---

## Phase 7: Notifications

### ISSUE-029: Notification Model & In-Site Bell Icon

**Depends on:** ISSUE-002, ISSUE-008

**Goal:** Build the notification system model and the bell icon UI.

**What to do:**

1. Create `internal/model/notification.go`:
   - A `Notification` struct: `ID`, `UserID`, `Type`, `Payload` (JSON), `IsRead`, `CreatedAt`.
   - Notification types (constants): `"new_post"`, `"comment_reply"`, `"comment_approved"`, `"comment_rejected"`, `"banned"`, `"unbanned"`.
   - Function `CreateNotification(db, userID, notifType, payload) → error`.
   - Function `UserNotifications(db, userID, page, perPage) → ([]Notification, totalCount, error)` — ordered by `created_at DESC`.
   - Function `UnreadCount(db, userID) → (int, error)`.
   - Function `MarkRead(db, notificationID) → error`.
   - Function `MarkAllRead(db, userID) → error`.
2. Create `templates/components/notification_bell.templ`:
   - Bell icon with unread count badge (red circle with number).
   - Clicking opens a dropdown or navigates to `/notifications`.
   - Uses HTMX polling or loads notifications on click: `hx-get="/notifications"`, `hx-trigger="click"`.
3. Create `templates/notifications.templ` (or inline dropdown):
   - List of notifications with icon/text based on type:
     - `new_post`: "CuratorName published: PostTitle"
     - `comment_reply`: "Username replied to your comment on PostTitle"
     - `comment_approved`/`rejected`: "Your comment on PostTitle was approved/rejected"
     - `banned`/`unbanned`: "Your account has been banned/unbanned"
   - Each notification links to the relevant content.
   - "Mark all read" button.
4. Create `internal/handler/notifications.go`:
   - `GET /notifications` — render notification list (HTMX fragment or full page).
   - `PUT /notifications/:id/read` — mark single notification as read.
   - `PUT /notifications/read-all` — mark all as read.
5. Add the bell icon to `templates/layout.templ` (visible only to logged-in users).
6. Write tests:
   - Creating notifications and reading them.
   - Unread count is accurate.
   - Mark read changes the count.

**Acceptance criteria:**

- Bell icon shows unread count.
- Notification list shows correct items.
- Mark-as-read works.
- Tests pass.

---

### ISSUE-030: Trigger Notifications on Events

**Depends on:** ISSUE-029, ISSUE-015, ISSUE-022

**Goal:** Wire up notification creation to actual events — publishing posts, approving comments, banning users, etc.

**What to do:**

1. When a post is **published** (status changes from draft to published):
   - Query all followers of the post's author.
   - Create a `"new_post"` notification for each follower.
   - Payload: `{"post_id": N, "post_title": "...", "curator_name": "..."}`.
2. When a comment is **approved**:
   - If the comment is a reply, create a `"comment_reply"` notification for the parent comment's author.
   - Create a `"comment_approved"` notification for the commenter.
   - Payload includes `post_id`, `post_title`, `comment_id`.
3. When a comment is **rejected**:
   - Create a `"comment_rejected"` notification for the commenter.
4. When a user is **banned/unbanned**:
   - Create a `"banned"` or `"unbanned"` notification for the user.
5. Add these notification triggers in the appropriate handler functions (or model functions).
6. Write tests:
   - Publishing a post creates notifications for all followers.
   - Approving a reply creates a notification for the parent commenter.
   - Banning creates a notification.

**Acceptance criteria:**

- All notification types are triggered by the correct events.
- Notifications contain accurate payload data.
- Tests pass.

---

## Phase 8: Email System

### ISSUE-031: Email Queue & Resend Integration

**Depends on:** ISSUE-002

**Goal:** Build the email queue system and integrate with the Resend API for sending.

**What to do:**

1. Create `internal/model/email.go`:
   - Function `QueueEmail(db, toEmail, subject, bodyHTML) → error` — insert into `email_queue` with status `"pending"`.
   - Function `PendingEmails(db, limit) → ([]QueuedEmail, error)` — returns pending emails ordered by `scheduled_for`.
   - Function `MarkEmailSent(db, emailID) → error`.
   - Function `MarkEmailFailed(db, emailID) → error` — increment attempts.
2. Create `internal/jobs/email_sender.go`:
   - A function that runs on a `time.Ticker` (every 2 minutes).
   - Fetches pending emails from the queue.
   - Sends each via the Resend HTTP API (`POST https://api.resend.com/emails` with `Authorization: Bearer <API_KEY>`).
   - On success, mark as sent. On failure (non-2xx), mark as failed.
   - If the response indicates rate limiting (429), stop processing until next tick.
   - Read the Resend API key from an environment variable `RESEND_API_KEY`.
   - If the env var is missing, log a warning and skip sending (don't crash).
3. Start the email sender goroutine from `cmd/server/main.go`.
4. Write tests:
   - Queuing an email inserts a row.
   - `PendingEmails` returns only pending items.
   - Marking sent/failed updates status correctly.
   - (Use a mock HTTP server for Resend API tests.)

**Acceptance criteria:**

- Emails can be queued and are processed by the background job.
- Resend API integration works (tested with mock).
- Rate limiting is handled gracefully.
- Missing API key doesn't crash the server.
- Tests pass.

---

### ISSUE-032: Email Verification Flow

**Depends on:** ISSUE-031, ISSUE-009

**Goal:** Implement the email verification flow — send a verification email on registration, handle the verification link.

**What to do:**

1. Add to `internal/model/user.go` or create a new file:
   - Function `CreateVerificationToken(db, userID) → (token string, error)`:
     - Generate a 32-byte random token, hex-encode it.
     - Insert into `verification_tokens` with expiry 24 hours from now.
   - Function `VerifyToken(db, token) → (userID int, error)`:
     - Look up the token, check expiry.
     - If valid, update the user's role to `"verified"`, set `verified_at`, delete the token.
     - If expired, delete the token and return an error.
2. Update `POST /register` in `internal/handler/auth.go`:
   - After creating the user, create a verification token.
   - Queue a verification email via `QueueEmail` with subject "Verify your Atlas account" and a link: `https://yourdomain.com/verify/{token}`.
   - Show a message: "Check your email to verify your account."
3. Add to `internal/handler/auth.go`:
   - `GET /verify/:token` — call `VerifyToken`. On success, redirect to `/` with a flash message "Email verified!" On failure, show "Invalid or expired verification link."
4. Write tests:
   - Registration queues a verification email.
   - Valid token verifies the user and promotes role.
   - Expired token returns an error.
   - Used token cannot be reused.

**Acceptance criteria:**

- Registration sends a verification email.
- Clicking the link verifies the account.
- Expired links are handled gracefully.
- Tests pass.

---

### ISSUE-033: Email Notifications for Events

**Depends on:** ISSUE-031, ISSUE-030

**Goal:** Send email notifications for events, respecting user preferences.

**What to do:**

1. When a `"new_post"` notification is created (ISSUE-030):
   - Check if the follower has `email_notify_posts = 1`.
   - If yes, queue an email: "CuratorName published a new post: PostTitle" with a link.
2. When a `"comment_reply"` notification is created:
   - Check if the parent commenter has `email_notify_replies = 1`.
   - If yes, queue an email: "Username replied to your comment on PostTitle" with a link.
3. Create email HTML templates (simple inline-styled HTML strings in Go — no need for a template engine):
   - Verification email template.
   - New post notification template.
   - Comment reply notification template.
4. Write tests:
   - User with notifications enabled gets an email queued.
   - User with notifications disabled does NOT get an email queued.

**Acceptance criteria:**

- Email notifications respect user preferences.
- Emails are queued (not sent directly) for async processing.
- Tests pass.

---

## Phase 9: Remaining Features

### ISSUE-034: Image Upload Endpoint

**Depends on:** ISSUE-008

**Goal:** Build the image upload endpoint for curators to embed images in posts.

**What to do:**

1. Create `internal/handler/upload.go`:
   - `POST /upload` — curator only.
   - Accept a multipart file upload.
   - Validate: file is an image (check MIME type: `image/jpeg`, `image/png`, `image/gif`, `image/webp`), max size 5MB.
   - Generate a unique filename: `{timestamp}_{random}.{ext}`.
   - Save to `uploads/` directory.
   - Return a Markdown image tag: `![alt](uploads/{filename})` as plain text (the editor will insert it into the textarea).
2. Serve the `uploads/` directory as static files in `cmd/server/main.go`.
3. Write tests:
   - Valid image upload returns a Markdown tag.
   - Non-image file is rejected.
   - Oversized file is rejected.
   - Non-curators get 403.

**Acceptance criteria:**

- Curators can upload images.
- Uploaded images are served as static files.
- Validation rejects invalid files.
- Tests pass.

---

### ISSUE-035: Tags Model & Post Tagging

**Depends on:** ISSUE-011

**Goal:** Implement the tag system for posts.

**What to do:**

1. Create `internal/model/tag.go`:
   - A `Tag` struct: `ID`, `Name`.
   - Function `CreateTag(db, name) → (Tag, error)` — normalize name to lowercase.
   - Function `TagByName(db, name) → (Tag, error)`.
   - Function `AllTags(db) → ([]Tag, error)`.
   - Function `TagsForPost(db, postID) → ([]Tag, error)`.
   - Function `SetPostTags(db, postID, tagNames []string) → error`:
     - Create tags that don't exist.
     - Delete existing post_tags for this post.
     - Insert new post_tags associations.
   - Function `PostsByTag(db, tagName, page, perPage) → ([]Post, totalCount, error)`.
2. Update the post editor (`templates/post_form.templ`) from ISSUE-017:
   - Add a "Tags" input field (comma-separated).
   - When saving a post, parse the tags and call `SetPostTags`.
3. Update `templates/post_view.templ` and `templates/components/post_card.templ`:
   - Display tags as clickable links (link to search or a tag-specific listing).
4. Write tests:
   - Creating and associating tags with posts.
   - Querying posts by tag.
   - Duplicate tag names are handled.

**Acceptance criteria:**

- Posts can be tagged with multiple tags.
- Tags display on post cards and post views.
- Tag links lead to filtered results.
- Tests pass.

---

### ISSUE-036: Inactive User Cleanup Job

**Depends on:** ISSUE-004, ISSUE-031

**Goal:** Implement the scheduled job that purges inactive unverified accounts.

**What to do:**

1. Create `internal/jobs/cleanup.go`:
   - Function that runs on a `time.Ticker` (every 24 hours).
   - Queries for users where: `role = 'unverified'` AND `last_active_at < NOW() - 30 days`.
   - Deletes these users and their associated data (sessions, comments, follows).
   - Logs the number of purged accounts.
2. Start the cleanup goroutine from `cmd/server/main.go`.
3. Write tests:
   - User inactive for 30+ days is purged.
   - User active within 30 days is NOT purged.
   - Verified users are never purged regardless of activity.
   - Associated data is cleaned up.

**Acceptance criteria:**

- Inactive unverified users are automatically purged.
- Active and verified users are never affected.
- Tests pass.

---

### ISSUE-037: Sidebar Toggle & Mobile Responsive Layout

**Depends on:** ISSUE-006, ISSUE-005

**Goal:** Implement the sidebar hide/show behavior for desktop and mobile.

**What to do:**

1. Add a toggle button in the top-right corner of the viewport (always visible, even when sidebar is hidden).
2. **Desktop behavior:**
   - Sidebar visible by default.
   - Clicking the toggle hides the sidebar.
   - When sidebar is hidden, the main content stays centered at the same max-width (does NOT expand).
   - This is a CSS-only toggle — no server round-trip.
3. **Mobile behavior:**
   - Sidebar hidden by default.
   - Clicking the toggle shows the sidebar as an overlay on top of content.
   - Clicking outside the sidebar or clicking the toggle again hides it.
4. Implementation: Use a minimal inline `<script>` (2–3 lines) to toggle a CSS class on a parent element. The actual show/hide logic is in Tailwind CSS classes.
5. When the sidebar is hidden, the ad slot is also hidden (it's inside the sidebar).
6. Persist the sidebar state in a cookie so it's remembered across page loads.
7. Write tests (or manual verification):
   - Desktop: sidebar toggles on click.
   - Mobile: sidebar overlays on click.
   - Main content width is consistent.

**Acceptance criteria:**

- Sidebar toggle works on desktop and mobile.
- Content does not reflow when sidebar is toggled on desktop.
- Mobile sidebar is an overlay.
- Ad slot hides with the sidebar.
- State persists via cookie.

---

### ISSUE-038: Post Soft-Delete & Hard-Delete from UI

**Depends on:** ISSUE-011, ISSUE-016, ISSUE-023

**Goal:** Wire up soft-delete and hard-delete actions for posts from the admin UI.

**What to do:**

1. Add delete buttons to:
   - `templates/admin/dashboard.templ` — on each post in the recent posts list.
   - `templates/post_view.templ` — visible only to curators/moderators.
2. Soft-delete:
   - Moderator + Curator can soft-delete.
   - Button uses HTMX: `hx-delete="/posts/:id"`.
   - Handler sets `is_deleted = 1` and `deleted_by` to the current user.
   - Post is hidden from public views.
   - Return an HTMX fragment indicating the post was deleted (e.g., replace the post card with a "deleted" badge).
3. Hard-delete:
   - Curator only.
   - Button on the moderation page (soft-deleted content section) or dashboard.
   - `hx-delete="/admin/posts/:id"` — permanently removes the row.
   - Requires confirmation (use `hx-confirm="Are you sure? This cannot be undone."`).
4. Restore:
   - Curator only.
   - Button on the moderation page's soft-deleted section.
   - `hx-put="/admin/posts/:id/restore"`.
5. Write tests:
   - Soft-delete hides post from public.
   - Hard-delete removes the row entirely.
   - Restore makes the post visible again.
   - Moderators can soft-delete but NOT hard-delete or restore.

**Acceptance criteria:**

- All delete/restore actions work via HTMX.
- Role restrictions are enforced.
- Confirmation dialog on hard delete.
- Tests pass.

---

### ISSUE-039: Sub-Category Navigation Expansion

**Depends on:** ISSUE-005, ISSUE-003

**Goal:** Implement the sidebar nav behavior where Blog/Reviews expand to show sub-categories when active.

**What to do:**

1. Update `templates/nav.templ`:
   - When the current URL is under `/blog` or `/blog/*`, the Blog section expands to show its sub-category links.
   - When the current URL is under `/reviews` or `/reviews/*`, the Reviews section expands.
   - Only one section expands at a time (the other stays collapsed).
   - The active sub-category link is highlighted.
2. Implementation: use the `currentPath` parameter already passed to the nav to determine which section to expand. This is a server-rendered decision — no JS needed.
3. Style the expanded/collapsed states with Tailwind:
   - Collapsed: sub-categories are hidden (`hidden` class).
   - Expanded: sub-categories are visible with indentation.
   - Active link: red accent or underline.
4. Write tests:
   - Request to `/blog/movies` → Blog section is expanded, "Movies" is highlighted.
   - Request to `/reviews` → Reviews section is expanded.
   - Request to `/` → both sections are collapsed.

**Acceptance criteria:**

- Nav expands the correct section based on current URL.
- Only one section is expanded at a time.
- Active link is visually highlighted.
- Tests pass.

---

### ISSUE-040: Request Logging Middleware

**Depends on:** ISSUE-001

**Goal:** Add structured request logging for all HTTP requests.

**What to do:**

1. Create `internal/middleware/logging.go`:
   - Middleware that logs each request: method, path, status code, response time, client IP.
   - Use Go's `slog` package (standard library structured logging).
   - Format: `INFO  GET /blog/movies 200 12ms 192.168.1.1`.
   - Log 5xx errors at ERROR level.
   - Log 4xx at WARN level.
   - Log 2xx/3xx at INFO level.
2. Wrap the `http.ResponseWriter` to capture the status code.
3. Apply this middleware globally in `cmd/server/main.go` (first in the middleware chain).
4. Write tests:
   - Verify the middleware doesn't alter the response.
   - Verify log output contains the expected fields.

**Acceptance criteria:**

- All requests are logged with method, path, status, and duration.
- Log levels match the status code range.
- Tests pass.

---

## Phase 10: Polish & Launch Prep

### ISSUE-041: Error Pages (404, 403, 500)

**Depends on:** ISSUE-005

**Goal:** Create custom error pages that match the site's design.

**What to do:**

1. Create `templates/error.templ`:
   - Takes: status code, title, and message.
   - Uses the site layout with sidebar.
   - 404: "Page not found. The page you're looking for doesn't exist."
   - 403: "Access denied. You don't have permission to view this page."
   - 500: "Something went wrong. We're working on it."
2. Create error handler functions in `internal/handler/` or `internal/render/`:
   - `RenderError(w, r, statusCode, message)` — renders the error template.
3. Update all handlers to use `RenderError` instead of `http.Error`.
4. Add a catch-all 404 handler in the router for unmatched routes.
5. Add a recovery middleware that catches panics and renders 500.
6. Write tests:
   - Unknown routes return 404 with the custom page.
   - Forbidden actions return 403 with the custom page.

**Acceptance criteria:**

- Error pages are styled consistently with the rest of the site.
- Panics are caught and show 500 instead of crashing.
- Tests pass.

---

### ISSUE-042: Ad Slot Placeholder

**Depends on:** ISSUE-005

**Goal:** Add the ad placeholder at the bottom of the sidebar.

**What to do:**

1. Create `templates/components/ad_slot.templ`:
   - A `<div>` styled as a placeholder with text: "Ad Space" or a subtle border/background.
   - This is the **only** ad placement on the entire site.
   - Hidden when the sidebar is hidden (naturally, since it's inside the sidebar).
2. Add the ad slot component to `templates/nav.templ` at the bottom.
3. Style with Tailwind: subtle background, centered text, fixed height.

**Acceptance criteria:**

- Ad placeholder appears at the bottom of the sidebar.
- It is hidden when the sidebar is hidden.
- Minimal and non-intrusive styling.

---

### ISSUE-043: Seed Initial Curator Account

**Depends on:** ISSUE-004, ISSUE-002

**Goal:** Create a seeding mechanism so the first curator account can be created on a fresh database.

**What to do:**

1. Create `migrations/003_seed_curator.sql` (or a Go-based seed):
   - Check if any curator exists. If not, create one:
     - Username: read from env var `ATLAS_ADMIN_USERNAME` (default: `admin`).
     - Email: read from env var `ATLAS_ADMIN_EMAIL` (default: `admin@atlas.local`).
     - Password: read from env var `ATLAS_ADMIN_PASSWORD` (default: `changeme`).
     - Role: `curator`, `verified_at` set to now.
   - Since bcrypt hashing can't be done in SQL, this should be a Go function called at startup.
2. Create a `SeedCurator` function in `internal/model/user.go` (or a `seeds.go` file):
   - Only runs if `SELECT COUNT(*) FROM users WHERE role = 'curator'` returns 0.
   - Creates the curator with the hashed password.
   - Logs: "Created initial curator account: admin".
3. Call this function after running migrations in `cmd/server/main.go`.
4. Write tests:
   - Fresh database gets a curator.
   - Database with an existing curator does not create a duplicate.

**Acceptance criteria:**

- First run creates a curator account.
- Subsequent runs skip seeding.
- Credentials are configurable via environment variables.
- Tests pass.

---

### ISSUE-044: Wire All Routes in Main Entry Point

**Depends on:** All handler issues

**Goal:** Final wiring of all routes, middleware, and startup tasks in `cmd/server/main.go`.

**What to do:**

1. Organize `cmd/server/main.go` with clear sections:
   - Database initialization (open, pragmas, migrations).
   - Seed curator account.
   - Start background jobs (email sender, inactive cleanup).
   - Configure Chi router with middleware chain:
     - Logging (first).
     - Auth (reads session, populates context).
     - Theme (reads theme preference).
   - Route groups:
     - **Public routes**: home, about, blog, reviews, post view, search, static files, uploads.
     - **Auth routes**: login, register, logout, verify.
     - **User routes** (RequireAuth): settings, comments, follow, notifications.
     - **Moderator routes** (RequireModerator): moderation queue, soft-delete, approve/reject.
     - **Curator routes** (RequireCurator): admin dashboard, post editor, category manager, user manager, hard-delete, restore, upload.
2. Graceful shutdown: listen for `SIGINT`/`SIGTERM`, stop background jobs, close the database.
3. Read port from env var `PORT` (default 8080).
4. Write an integration test that starts the full server and hits a few key routes.

**Acceptance criteria:**

- All routes from the Route Summary in `PLAN.md` are registered.
- Middleware is applied in the correct order.
- Server starts, handles requests, and shuts down gracefully.
- Tests pass.

---

### ISSUE-045: Logo & Favicon Integration

**Depends on:** ISSUE-005, ISSUE-006

**Goal:** Replace the nav placeholder text with the real Atlas SVG logo and wire all favicon/manifest assets into the layout `<head>`.

**What to do:**

1. Update `templates/nav.templ`:
   - Replace the placeholder text `"Atlas"` with an `<img src="/static/logo.svg" alt="Atlas" ...>` tag.
   - Size the logo appropriately for the sidebar (e.g., `h-12 w-auto`).
   - Wrap it in an `<a href="/">` so clicking the logo goes home.
2. Update `templates/layout.templ` `<head>` to include:
   - `<link rel="icon" type="image/x-icon" href="/static/favicon.ico">`
   - `<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">`
   - `<link rel="icon" type="image/png" sizes="96x96" href="/static/favicon-96x96.png">`
   - `<link rel="apple-touch-icon" href="/static/apple-touch-icon.png">`
   - `<link rel="manifest" href="/static/site.webmanifest">`
3. Verify `static/logo.svg`, `static/favicon.ico`, `static/favicon.svg`, `static/favicon-96x96.png`, `static/apple-touch-icon.png`, `static/web-app-manifest-192x192.png`, `static/web-app-manifest-512x512.png`, and `static/site.webmanifest` are all served correctly at their `/static/...` paths.
4. Run `templ generate` and verify no template errors.

**Acceptance criteria:**

- The Atlas SVG logo appears in the sidebar instead of placeholder text.
- Clicking the logo navigates to `/`.
- Browser shows the favicon on the tab.
- iOS and PWA icon assets are referenced correctly.
- `templ generate` passes.

---

### ISSUE-046: Tag Browse Page

**Depends on:** ISSUE-035, ISSUE-013

**Goal:** Create a `GET /tag/:name` route so tag links on post cards and post views lead to a filtered listing of posts with that tag.

**What to do:**

1. Add to `internal/handler/posts.go` (or a new `internal/handler/tags.go`):
   - `GET /tag/:name` — look up the tag by name (case-insensitive), query `PostsByTag`, render the listing.
   - If the tag doesn't exist or has no published posts, return a friendly empty-state page (not 404).
   - If `HX-Request` header is present, return just the post list fragment (same pattern as other listing pages).
2. Register the route in `cmd/server/main.go` under public routes.
3. Update `templates/components/post_card.templ` and `templates/post_view.templ` (from ISSUE-035): ensure each tag link points to `/tag/:name`.
4. Reuse `templates/post_list.templ` for the tag listing page:
   - Pass `nil` for category (since tags span all categories).
   - Set the heading to "Posts tagged: tagname".
5. Write tests:
   - `GET /tag/go` returns 200 with posts tagged "go".
   - `GET /tag/nonexistent` returns 200 with an empty-state message.
   - HTMX requests return just the fragment.
   - Tag name lookup is case-insensitive.

**Acceptance criteria:**

- `/tag/:name` shows all published posts with that tag.
- Non-existent tags show an empty state, not an error.
- Tag links on post cards and post views navigate to the correct URL.
- Tests pass.

---

## Summary

| Phase | Issues | Focus |
|-------|--------|-------|
| 1: Foundation | 001–006 | Project structure, database, templates, CSS |
| 2: Auth | 007–010 | Sessions, middleware, register, login |
| 3: Content | 011–015 | Posts, Markdown, listings, homepage |
| 4: Admin | 016–020 | Dashboard, editor, categories, users |
| 5: Comments | 021–023 | Comment model, HTMX submission, moderation |
| 6: User Features | 024–028 | Settings, theme, follow, search, about |
| 7: Notifications | 029–030 | Bell icon, event triggers |
| 8: Email | 031–033 | Queue, verification, email notifications |
| 9: Remaining | 034–040 | Uploads, tags, cleanup, sidebar, logging |
| 10: Polish | 041–046 | Error pages, ads, seeding, final wiring, logo, tag browse |

**Total: 46 issues**
