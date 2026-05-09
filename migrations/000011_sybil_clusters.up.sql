CREATE TABLE IF NOT EXISTS social.sybil_clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id INT NOT NULL,
    size INT NOT NULL,
    external_connections INT NOT NULL,
    suspect_uids UUID[] NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);
