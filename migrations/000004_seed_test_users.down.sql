-- Removes seed users created in 000004_seed_test_users.up.sql.
-- CASCADE on social.profiles, social.user_settings cleans related rows automatically.
DELETE FROM social.users WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);
