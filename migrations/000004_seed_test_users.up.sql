-- Seed users for development/testing only.
-- Password for ALL users: "password"
-- Hashes generated via: go run ./cmd/seedhash/ -password=password -count=3
-- DO NOT USE IN PRODUCTION (this migration is gated by APP_ENV in deployment).

-- 1. Создаем пользователей в social.users
-- Пароль у всех: password
INSERT INTO social.users (id, phone_hash, phone_encrypted, password_hash, verification_level, trust_score)
VALUES
    ('11111111-1111-1111-1111-111111111111', '\x0102', '\x0304', '$argon2id$v=19$m=65536,t=3,p=4$fB6YKf59djDjL/k3P5qM5A$OnQssL8RlZw40RkRACczO10iInG1B+CAnr9xxNJ9akc', 'photo_verified', 95),
    ('22222222-2222-2222-2222-222222222222', '\x0506', '\x0708', '$argon2id$v=19$m=65536,t=3,p=4$6iV4cSjZMbL/Xym5aa7x8w$IXOot9DNrO+9OtQyEzMSWgJwYqyVhUuxDa0mF2i4L9I', 'none', 88),
    ('33333333-3333-3333-3333-333333333333', '\x0910', '\x1112', '$argon2id$v=19$m=65536,t=3,p=4$dQpZ7Wh6W3uo7r3YBLDNmQ$A62P8je13QKOIIzhRpKZkwvmbbSfl6uFHS9I25KPxGM', 'id_verified', 100)
ON CONFLICT (id) DO NOTHING;

-- 2. Создаем профили в social.profiles
-- Добавляем ST_SetSRID(ST_MakePoint(lon, lat), 4326) для корректной работы PostGIS
INSERT INTO social.profiles (user_id, display_name, bio, city, gender, looking_for, avatar_url, location)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'Diana', 'Student of AITU, love coding', 'Astana', 'female', 'male', 'users/diana_avatar.png', ST_SetSRID(ST_MakePoint(71.430, 51.128), 4326)),
    ('22222222-2222-2222-2222-222222222222', 'Alex', 'Backend Developer, Go lover', 'Astana', 'male', 'female', 'users/alex_avatar.png', ST_SetSRID(ST_MakePoint(71.440, 51.130), 4326)),
    ('33333333-3333-3333-3333-333333333333', 'Mira', 'Designer and artist', 'Astana', 'female', 'male', 'users/mira_avatar.png', ST_SetSRID(ST_MakePoint(71.420, 51.125), 4326))
ON CONFLICT (user_id) DO NOTHING;

-- 3. Настройки поиска
INSERT INTO social.user_settings (user_id, age_range_min, age_range_max, max_distance_km)
VALUES
    ('11111111-1111-1111-1111-111111111111', 18, 30, 100),
    ('22222222-2222-2222-2222-222222222222', 18, 30, 100),
    ('33333333-3333-3333-3333-333333333333', 18, 30, 100)
ON CONFLICT (user_id) DO NOTHING;
