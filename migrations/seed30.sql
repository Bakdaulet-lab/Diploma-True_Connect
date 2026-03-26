WITH generated_data AS (
    SELECT
        gen_random_uuid() as u_id,
        decode(md5(random()::text), 'hex') as p_hash,
        decode(md5(random()::text), 'hex') as p_enc,
        (ARRAY['Амина', 'Адия', 'Мадина', 'Томирис', 'Аружан', 'Аяжан', 'Камила', 'Зарина', 'Аида', 'Дана', 'Сабина', 'Алина', 'Дарига', 'Аиша', 'Инкар', 'Малика', 'Дильназ', 'Айгерим', 'Сауле', 'Куралай', 'Гульжан', 'Асель', 'Ботагоз', 'Айнур', 'Алия', 'Мария', 'София', 'Хадиша', 'Эльнара', 'Лола'])[i] as d_name,
        -- Астана: генерируем разброс координат (lon ~71.4, lat ~51.12)
        71.38 + (random() * 0.1) as lon,
        51.08 + (random() * 0.1) as lat
    FROM generate_series(1, 30) as i
),
inserted_users AS (
    INSERT INTO social.users (id, phone_hash, phone_encrypted, password_hash, verification_level, trust_score)
    SELECT u_id, p_hash, p_enc, '$2a$10$dummyhashformockusers1234567890', 'photo_verified'::social.verification_level, (80 + (random() * 20))::smallint
    FROM generated_data
    RETURNING id
),
inserted_profiles AS (
    INSERT INTO social.profiles (user_id, display_name, bio, gender, city, location, looking_for, avatar_url)
    SELECT u_id, d_name, 'Привет! Ищу интересные знакомства в Астане. Люблю кофе и вечерние прогулки 🌅', 'female'::social.gender, 'Astana', ST_SetSRID(ST_MakePoint(lon, lat), 4326)::geography, 'male'::social.gender, 'users/mira_avatar.png'
    FROM generated_data
    RETURNING user_id
)
INSERT INTO social.user_settings (user_id, age_range_min, age_range_max, max_distance_km)
SELECT u_id, 18, 35, 100
FROM generated_data;