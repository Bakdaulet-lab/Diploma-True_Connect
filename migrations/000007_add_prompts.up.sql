ALTER TABLE social.profiles ADD COLUMN prompts jsonb NOT NULL DEFAULT '[]'::jsonb;
