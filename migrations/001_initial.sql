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
