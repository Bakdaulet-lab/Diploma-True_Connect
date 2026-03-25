-- 1. Создаем пользователей в social.users
-- Пароль у всех: password
INSERT INTO social.users (id, phone_hash, phone_encrypted, password_hash, verification_level, trust_score)
VALUES 
    ('11111111-1111-1111-1111-111111111111', '\x0102', '\x0304', '$2a$10$8K1p/a06jl7z7YvK5WpXyeXmK2jP.h5fQy5Xw5Xw5Xw5Xw5Xw5Xw', 'photo_verified', 95),
    ('22222222-2222-2222-2222-222222222222', '\x0506', '\x0708', '$2a$10$8K1p/a06jl7z7YvK5WpXyeXmK2jP.h5fQy5Xw5Xw5Xw5Xw5Xw5Xw', 'none', 88),
    ('33333333-3333-3333-3333-333333333333', '\x0910', '\x1112', '$2a$10$8K1p/a06jl7z7YvK5WpXyeXmK2jP.h5fQy5Xw5Xw5Xw5Xw5Xw5Xw', 'id_verified', 100)
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