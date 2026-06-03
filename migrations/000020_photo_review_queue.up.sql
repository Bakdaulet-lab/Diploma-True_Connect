-- Admin re-review queue for profile photos rejected by the CV verifier
-- (no face / NSFW). Rejected uploads are held here (not shown publicly) until an
-- admin approves (promotes to a real photo/avatar) or rejects (deletes it).
CREATE TABLE social.photo_review_queue (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    object_key  TEXT NOT NULL,
    reasons     TEXT[] NOT NULL DEFAULT '{}',
    nsfw_score  REAL NOT NULL DEFAULT 0,
    face_count  INT  NOT NULL DEFAULT 0,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | approved | rejected
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES social.users(id)
);

CREATE INDEX idx_photo_review_pending
    ON social.photo_review_queue (created_at)
    WHERE status = 'pending';
