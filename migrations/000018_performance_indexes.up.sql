-- D1: Composite index for interaction cooldown check
-- Query: WHERE rater_id=$1 AND rated_id=$2 AND created_at > $3
CREATE INDEX IF NOT EXISTS idx_interactions_cooldown
    ON social.interactions (rater_id, rated_id, created_at DESC);

-- D2: Index for trust score leaderboard queries (ORDER BY trust_score DESC)
CREATE INDEX IF NOT EXISTS idx_users_trust_score
    ON social.users (trust_score DESC, last_login_at DESC NULLS LAST);

-- Extra: speed up the pending-likes query (matches where one side has liked)
CREATE INDEX IF NOT EXISTS idx_matches_pending_likes_a
    ON social.matches (user_b_id, user_a_liked, user_b_liked)
    WHERE user_a_liked = true AND user_b_liked = false;

CREATE INDEX IF NOT EXISTS idx_matches_pending_likes_b
    ON social.matches (user_a_id, user_b_liked, user_a_liked)
    WHERE user_b_liked = true AND user_a_liked = false;
