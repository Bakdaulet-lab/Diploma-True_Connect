DROP INDEX IF EXISTS social.idx_messages_toxic;
DROP INDEX IF EXISTS social.idx_messages_sender;

ALTER TABLE social.messages
    DROP CONSTRAINT IF EXISTS messages_sender_id_fkey;
ALTER TABLE social.messages
    ADD CONSTRAINT messages_sender_id_fkey
        FOREIGN KEY (sender_id) REFERENCES social.users(id);

ALTER TABLE social.reports
    DROP CONSTRAINT IF EXISTS reports_reporter_id_fkey,
    DROP CONSTRAINT IF EXISTS reports_reported_id_fkey;
ALTER TABLE social.reports
    ADD CONSTRAINT reports_reporter_id_fkey
        FOREIGN KEY (reporter_id) REFERENCES social.users(id),
    ADD CONSTRAINT reports_reported_id_fkey
        FOREIGN KEY (reported_id) REFERENCES social.users(id);

ALTER TABLE social.interactions
    DROP CONSTRAINT IF EXISTS interactions_rater_id_fkey,
    DROP CONSTRAINT IF EXISTS interactions_rated_id_fkey;
ALTER TABLE social.interactions
    ALTER COLUMN rater_id SET NOT NULL,
    ALTER COLUMN rated_id SET NOT NULL;
ALTER TABLE social.interactions
    ADD CONSTRAINT interactions_rater_id_fkey
        FOREIGN KEY (rater_id) REFERENCES social.users(id),
    ADD CONSTRAINT interactions_rated_id_fkey
        FOREIGN KEY (rated_id) REFERENCES social.users(id);
