CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- =========================================================
-- USERS
-- =========================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(30) NOT NULL UNIQUE,

    display_name VARCHAR(100) NOT NULL,
    bio TEXT,
    avatar_url TEXT,

    human_verified BOOLEAN NOT NULL DEFAULT FALSE,

    status VARCHAR(20) NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_status_check
        CHECK (status IN ('active', 'restricted', 'suspended', 'deleted'))
);


-- =========================================================
-- INNER CIRCLE
-- =========================================================

CREATE TABLE inner_circle_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, member_id),

    CHECK (user_id <> member_id)
);

CREATE INDEX idx_inner_circle_user
    ON inner_circle_members(user_id);


-- =========================================================
-- INTERESTS
-- =========================================================

CREATE TABLE interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE user_interests (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES interests(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, interest_id)
);

CREATE INDEX idx_user_interests_interest
    ON user_interests(interest_id);


-- =========================================================
-- POSTS
-- =========================================================

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    content TEXT NOT NULL,

    visibility VARCHAR(20) NOT NULL DEFAULT 'public',

    status VARCHAR(20) NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT posts_visibility_check
        CHECK (visibility IN ('public', 'inner_circle')),

    CONSTRAINT posts_status_check
        CHECK (status IN ('active', 'hidden', 'deleted'))
);

CREATE INDEX idx_posts_author_created
    ON posts(author_id, created_at DESC);

CREATE INDEX idx_posts_created
    ON posts(created_at DESC);


-- =========================================================
-- POST INTERESTS
-- =========================================================

CREATE TABLE post_interests (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES interests(id) ON DELETE CASCADE,

    PRIMARY KEY (post_id, interest_id)
);

CREATE INDEX idx_post_interests_interest
    ON post_interests(interest_id);


-- =========================================================
-- COMMENTS
-- =========================================================

CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    parent_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,

    content TEXT NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT comments_status_check
        CHECK (status IN ('active', 'hidden', 'deleted'))
);

CREATE INDEX idx_comments_post_created
    ON comments(post_id, created_at ASC);


-- =========================================================
-- REACTIONS
-- =========================================================

CREATE TABLE post_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    reaction_type VARCHAR(30) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (post_id, user_id, reaction_type)
);

CREATE INDEX idx_post_reactions_post
    ON post_reactions(post_id);


-- =========================================================
-- REPUTATION EVENTS
-- =========================================================

CREATE TABLE reputation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    event_type VARCHAR(50) NOT NULL,

    source_type VARCHAR(50),
    source_id UUID,

    weight NUMERIC(10, 2) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reputation_user_created
    ON reputation_events(user_id, created_at DESC);


-- =========================================================
-- REPUTATION SNAPSHOTS
-- =========================================================

CREATE TABLE reputation_snapshots (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    score NUMERIC(10, 2) NOT NULL DEFAULT 0,

    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- =========================================================
-- REPORTS
-- =========================================================

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    target_type VARCHAR(30) NOT NULL,
    target_id UUID NOT NULL,

    reason VARCHAR(50) NOT NULL,
    description TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT reports_status_check
        CHECK (status IN ('pending', 'reviewing', 'resolved', 'dismissed'))
);

CREATE INDEX idx_reports_status
    ON reports(status);


-- =========================================================
-- MODERATION ACTIONS
-- =========================================================

CREATE TABLE moderation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    target_type VARCHAR(30) NOT NULL,
    target_id UUID NOT NULL,

    moderator_id UUID REFERENCES users(id) ON DELETE SET NULL,

    action VARCHAR(50) NOT NULL,
    reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- =========================================================
-- NOTIFICATIONS
-- =========================================================

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    type VARCHAR(50) NOT NULL,

    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,

    target_type VARCHAR(30),
    target_id UUID,

    read_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_created
    ON notifications(user_id, created_at DESC);