-- Mock users seed (safe to run multiple times; inserts new random users)
-- Generates 40 users (20 female, 20 male) around Astana coordinates.

WITH female_data AS (
    SELECT
        gen_random_uuid() as u_id,
        decode(md5(random()::text), 'hex') as p_hash,
        decode(md5(random()::text), 'hex') as p_enc,
        (ARRAY[
            'Amina','Adiya','Madina','Tomiris','Aruzhan','Ayazhan','Kamila','Zarina','Aida','Dana',
            'Sabina','Alina','Dariga','Aisha','Inkar','Malika','Dilnaz','Aigerim','Saule','Kuralai'
        ])[i] as d_name,
        71.38 + (random() * 0.12) as lon,
        51.08 + (random() * 0.12) as lat
    FROM generate_series(1, 20) as i
),
male_data AS (
    SELECT
        gen_random_uuid() as u_id,
        decode(md5(random()::text), 'hex') as p_hash,
        decode(md5(random()::text), 'hex') as p_enc,
        (ARRAY[
            'Alikhan','Nursultan','Dias','Arman','Timur','Islam','Aldiyar','Aslan','Rinat','Serik',
            'Daniyar','Murat','Eldar','Bekzat','Rauan','Ayan','Ilyas','Zhandos','Kairat','Yerkebulan'
        ])[i] as d_name,
        71.38 + (random() * 0.12) as lon,
        51.08 + (random() * 0.12) as lat
    FROM generate_series(1, 20) as i
),
all_data AS (
    SELECT u_id, p_hash, p_enc, d_name, lon, lat,
           'female'::social.gender as gender,
           'male'::social.gender as looking_for
    FROM female_data
    UNION ALL
    SELECT u_id, p_hash, p_enc, d_name, lon, lat,
           'male'::social.gender as gender,
           'female'::social.gender as looking_for
    FROM male_data
),
inserted_users AS (
    INSERT INTO social.users (id, phone_hash, phone_encrypted, password_hash, verification_level, trust_score)
    SELECT
        u_id,
        p_hash,
        p_enc,
        '$2a$10$dummyhashformockusers1234567890',
        'photo_verified'::social.verification_level,
        (70 + (random() * 30))::smallint
    FROM all_data
    RETURNING id
),
inserted_profiles AS (
    INSERT INTO social.profiles (user_id, display_name, bio, gender, city, location, looking_for, avatar_url)
    SELECT
        u_id,
        d_name,
        'Hello! Looking for new connections in Astana. Love coffee and evening walks.',
        gender,
        'Astana',
        ST_SetSRID(ST_MakePoint(lon, lat), 4326)::geography,
        looking_for,
        'users/mock_avatar.png'
    FROM all_data
    RETURNING user_id
),
inserted_settings AS (
    INSERT INTO social.user_settings (user_id, age_range_min, age_range_max, max_distance_km)
    SELECT u_id, 18, 40, 200
    FROM all_data
    RETURNING user_id
)
SELECT count(*) AS inserted_users FROM inserted_users;
