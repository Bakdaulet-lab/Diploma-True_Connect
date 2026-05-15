-- D2: Add ON DELETE CASCADE so hard-deleting a user propagates to child tables.
-- Uses SET NULL for interactions to preserve rating history with anonymised rater/rated.

-- interactions: keep rows but null out the deleted user's ID
ALTER TABLE social.interactions
    DROP CONSTRAINT IF EXISTS interactions_rater_id_fkey,
    DROP CONSTRAINT IF EXISTS interactions_rated_id_fkey;

ALTER TABLE social.interactions
    ADD CONSTRAINT interactions_rater_id_fkey
        FOREIGN KEY (rater_id) REFERENCES social.users(id) ON DELETE SET NULL,
    ADD CONSTRAINT interactions_rated_id_fkey
        FOREIGN KEY (rated_id) REFERENCES social.users(id) ON DELETE SET NULL;

-- interactions.rater_id / rated_id must now allow NULL
ALTER TABLE social.interactions
    ALTER COLUMN rater_id DROP NOT NULL,
    ALTER COLUMN rated_id DROP NOT NULL;

-- reports: cascade-delete reports when either party is deleted
ALTER TABLE social.reports
    DROP CONSTRAINT IF EXISTS reports_reporter_id_fkey,
    DROP CONSTRAINT IF EXISTS reports_reported_id_fkey;

ALTER TABLE social.reports
    ADD CONSTRAINT reports_reporter_id_fkey
        FOREIGN KEY (reporter_id) REFERENCES social.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT reports_reported_id_fkey
        FOREIGN KEY (reported_id) REFERENCES social.users(id) ON DELETE CASCADE;

-- messages.sender_id: cascade when sender is deleted (match cascade already covers most cases)
ALTER TABLE social.messages
    DROP CONSTRAINT IF EXISTS messages_sender_id_fkey;

ALTER TABLE social.messages
    ADD CONSTRAINT messages_sender_id_fkey
        FOREIGN KEY (sender_id) REFERENCES social.users(id) ON DELETE CASCADE;

-- Additional indexes for content moderation and sender-based queries
CREATE INDEX IF NOT EXISTS idx_messages_sender
    ON social.messages (sender_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_messages_toxic
    ON social.messages (is_toxic, created_at DESC) WHERE is_toxic = true;
