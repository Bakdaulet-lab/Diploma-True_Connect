-- =============================================================================
-- TrueConnect — Halal Demo Seed Data
-- Scenario: "Айгерим meets Алихан in 90 seconds"
--
-- Run separately (not a numbered migration):
--   make seed-halal
-- Reset everything and reseed:
--   make reset-demo
-- =============================================================================

BEGIN;

-- ── Demo user UUIDs (memorable for demos) ────────────────────────────────────
-- Айгерим Сейткали  — female, nikah_year, hanafi, Almaty
-- Алихан  Ержан     — male,   nikah_year, hanafi, Almaty

DO $$
DECLARE
    aigerin_id  UUID := 'a1g3r1m0-0000-4000-8000-000000000001';
    alikhan_id  UUID := 'a11kh4n0-0000-4000-8000-000000000002';
BEGIN

-- ── Users ─────────────────────────────────────────────────────────────────────
INSERT INTO social.users (id, phone_hash, phone_encrypted, display_name, is_active, phone_verified, trust_score)
VALUES
    (aigerin_id,  '\xdeadbeef01', '\xcafe01', 'Айгерим', true, true, 82),
    (alikhan_id,  '\xdeadbeef02', '\xcafe02', 'Алихан',  true, true, 78)
ON CONFLICT (id) DO UPDATE
    SET trust_score   = EXCLUDED.trust_score,
        phone_verified = true,
        is_active      = true;

-- ── Profiles ──────────────────────────────────────────────────────────────────
-- PostGIS point: ST_SetSRID(ST_MakePoint(lon, lat), 4326)
-- Almaty centre ≈ 43.238°N, 76.899°E
INSERT INTO social.profiles
    (user_id, display_name, gender, city, looking_for,
     niyyah, madhab, languages, no_photo_mode,
     location, avatar_url, bio)
VALUES
    (aigerin_id, 'Айгерим', 'female', 'Almaty', 'male',
     'nikah_year', 'hanafi', ARRAY['kazakh', 'russian'], false,
     ST_SetSRID(ST_MakePoint(76.899, 43.238), 4326),
     '', 'Ищу серьёзные отношения с намерением жениться в этом году. Ханафи мазхаб. Казашка из Алматы.'),
    (alikhan_id, 'Алихан', 'male', 'Almaty', 'female',
     'nikah_year', 'hanafi', ARRAY['kazakh'], false,
     ST_SetSRID(ST_MakePoint(76.901, 43.239), 4326),
     '', 'Практикующий мусульманин. Ищу жену с намерением никах в этом году.')
ON CONFLICT (user_id) DO UPDATE
    SET display_name  = EXCLUDED.display_name,
        niyyah        = EXCLUDED.niyyah,
        madhab        = EXCLUDED.madhab,
        languages     = EXCLUDED.languages,
        city          = EXCLUDED.city,
        location      = EXCLUDED.location,
        looking_for   = EXCLUDED.looking_for;

-- ── Settings ─────────────────────────────────────────────────────────────────
INSERT INTO social.user_settings
    (user_id, age_range_min, age_range_max, max_distance_km, looking_for)
VALUES
    (aigerin_id, 24, 35, 100, 'male'),
    (alikhan_id, 22, 32, 100, 'female')
ON CONFLICT (user_id) DO UPDATE
    SET age_range_min   = EXCLUDED.age_range_min,
        age_range_max   = EXCLUDED.age_range_max,
        max_distance_km = EXCLUDED.max_distance_km;

-- ── Pre-seeded like: Алихан already liked Айгерим ────────────────────────────
-- Pair is ordered so user_a_id < user_b_id (UUID string comparison).
-- a11kh4n0... < a1g3r1m0... → Алихан is user_a, Айгерим is user_b.
INSERT INTO social.matches (user_a_id, user_b_id, user_a_liked, user_b_liked)
VALUES (alikhan_id, aigerin_id, true, false)
ON CONFLICT (user_a_id, user_b_id) DO UPDATE
    SET user_a_liked = true;

END $$;

COMMIT;

-- Confirm seed loaded.
SELECT
    u.display_name,
    p.niyyah,
    p.madhab,
    p.city,
    u.trust_score
FROM social.users u
JOIN social.profiles p ON p.user_id = u.id
WHERE u.id IN (
    'a1g3r1m0-0000-4000-8000-000000000001',
    'a11kh4n0-0000-4000-8000-000000000002'
)
ORDER BY u.display_name;
