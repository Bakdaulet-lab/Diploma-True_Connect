-- Persistent swipe rejections (passes) so rejected profiles never reappear.
CREATE TABLE IF NOT EXISTS social.swipe_rejections (
    user_id   UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, target_id)
);

CREATE INDEX IF NOT EXISTS idx_swipe_rejections_user ON social.swipe_rejections(user_id);

-- Blocked users: bidirectional exclusion from feed, discovery and chat.
CREATE TABLE IF NOT EXISTS social.blocked_users (
    blocker_id UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id)
);

CREATE INDEX IF NOT EXISTS idx_blocked_users_blocker ON social.blocked_users(blocker_id);
CREATE INDEX IF NOT EXISTS idx_blocked_users_blocked ON social.blocked_users(blocked_id);
