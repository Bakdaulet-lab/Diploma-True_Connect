-- Сброс любых предыдущих зависших транзакций
ROLLBACK; 

BEGIN;

-- =============================================================================
-- ЧАСТЬ 1: ДЕМО-ПОЛЬЗОВАТЕЛИ (Айгерим и Алихан)
-- Исправлено: теперь переменные шифров реально используются в INSERT и UPDATE
-- =============================================================================
DO $$
DECLARE
    aigerim_id  UUID := 'aaaaaaaa-0000-4000-8000-000000000001';
    alikhan_id  UUID := 'bbbbbbbb-0000-4000-8000-000000000002';
    pass_hash   TEXT := '$argon2id$v=19$m=65536,t=3,p=4$GCXKRBz86eApTEwv2tQoSA$vYKu2wy8/rJlDk0JQqoeHWT0wHpU/kOztwjYljuq5U4'; 
    
    -- Твои сгенерированные шифры:
    phone_enc_1 BYTEA := '\x9d5fb9f08ee4732fb38fe168a0a25bc908cfd8873022e742acbc837fa04e07e50a44456b6414a618';
    phone_enc_2 BYTEA := '\x8e5d751123958e72f08adf523f33a37d69d87970819c645e39975ea56a3c4a62c9e63fca9bd9bc0b';
BEGIN
    -- Вставляем данные, используя переменные (phone_enc_1 и phone_enc_2)
    INSERT INTO social.users (id, phone_hash, phone_encrypted, password_hash, is_active, verification_level, trust_score)
    VALUES
        (aigerim_id, sha256('+77000000001'::bytea), phone_enc_1, pass_hash, true, 'photo_verified'::social.verification_level, 85),
        (alikhan_id, sha256('+77000000002'::bytea), phone_enc_2, pass_hash, true, 'photo_verified'::social.verification_level, 80)
    ON CONFLICT (id) DO UPDATE SET 
        phone_hash = EXCLUDED.phone_hash,
        phone_encrypted = EXCLUDED.phone_encrypted, -- ОБЯЗАТЕЛЬНО обновляем шифр
        password_hash = EXCLUDED.password_hash,
        is_active = true, 
        verification_level = 'photo_verified'::social.verification_level;

    -- Profiles (Оставляем как было)
    INSERT INTO social.profiles (user_id, display_name, gender, city, looking_for, niyyah, madhab, languages, location, bio)
    VALUES
        (aigerim_id, 'Айгерим', 'female'::social.gender, 'Almaty', 'male'::social.gender, 'nikah_year'::social.niyyah, 'hanafi'::social.madhab, ARRAY['kazakh', 'russian'], ST_SetSRID(ST_MakePoint(76.899, 43.238), 4326), 'Ищу спутника жизни. Ханафи.'),
        (alikhan_id, 'Алихан', 'male'::social.gender, 'Almaty', 'female'::social.gender, 'nikah_year'::social.niyyah, 'hanafi'::social.madhab, ARRAY['kazakh'], ST_SetSRID(ST_MakePoint(76.901, 43.239), 4326), 'Практикующий мусульманин. Алматы.')
    ON CONFLICT (user_id) DO NOTHING;

    -- Settings (Оставляем как было)
    INSERT INTO social.user_settings (user_id, age_range_min, age_range_max, max_distance_km)
    VALUES (aigerim_id, 18, 40, 100), (alikhan_id, 18, 40, 100)
    ON CONFLICT (user_id) DO NOTHING;

    -- Matches (Оставляем как было)
    INSERT INTO social.matches (user_a_id, user_b_id, user_a_liked, user_b_liked)
    VALUES (aigerim_id, alikhan_id, false, true)
    ON CONFLICT DO NOTHING;
END $$;


-- =============================================================================
-- ЧАСТЬ 2: 100 МОК-ПОЛЬЗОВАТЕЛЕЙ (Через надежный цикл)
-- =============================================================================
DO $$
DECLARE
    f_names text[] := ARRAY['Amina','Adiya','Madina','Tomiris','Aruzhan','Ayazhan','Kamila','Zarina','Aida','Dana','Asel','Aliya','Ainur','Dinara','Saltanat','Zhibek','Aknur','Moldir','Bayan','Aizada','Gaukhar','Marzhan','Sholpan','Gulmira','Nazym','Elnara','Galiya','Samal','Laura','Meruert','Indira','Anar','Gulzhan','Balzhan','Asem','Aygul','Damira','Elmira','Fariza','Guldana','Hadisha','Iraida','Janar','Karlygash','Lazzat','Mira','Nargiz','Oksana','Raushan','Saniya'];
    m_names text[] := ARRAY['Alikhan','Nursultan','Dias','Arman','Timur','Islam','Aldiyar','Aslan','Rinat','Serik','Daniyar','Murat','Eldar','Bekzat','Rauan','Ayan','Ilyas','Zhandos','Kairat','Yerkebulan','Azamat','Baurzhan','Maksat','Nurlan','Kuat','Ruslan','Erzhan','Darhan','Marat','Askhat','Samat','Oljas','Abylai','Talgat','Dulat','Erbol','Galym','Sanzhar','Dastan','Berdibek','Kuanysh','Madi','Nurali','Ramazan','Sabyr','Temirlan','Ulan','Yerasyl','Zhanibek','Amir'];
    i int;
    new_uid uuid;
    lon float8; lat float8;
    p_hash bytea; p_enc bytea;
BEGIN
    -- Создаем 50 девушек
    FOR i IN 1..50 LOOP
        new_uid := gen_random_uuid();
        lon := 71.38 + (random() * 0.15);
        lat := 51.08 + (random() * 0.15);
        p_hash := decode(md5(random()::text), 'hex');
        p_enc := decode(md5(random()::text), 'hex');

        INSERT INTO social.users (id, phone_hash, phone_encrypted, password_hash, trust_score, verification_level)
        VALUES (new_uid, p_hash, p_enc, '$2a$10$dummyhashformockusers1234567890', (70 + random()*30)::int, 'photo_verified'::social.verification_level);

        INSERT INTO social.profiles (user_id, display_name, gender, city, location, looking_for, niyyah, madhab, bio)
        VALUES (new_uid, f_names[i], 'female'::social.gender, 'Astana', ST_SetSRID(ST_MakePoint(lon, lat), 4326), 'male'::social.gender, 'serious_marriage'::social.niyyah, 'hanafi'::social.madhab, 'Assalamu alaikum! I am ' || f_names[i] || ' from Astana.');

        INSERT INTO social.user_settings (user_id, age_range_min, age_range_max, max_distance_km)
        VALUES (new_uid, 18, 40, 200);
    END LOOP;

    -- Создаем 50 парней
    FOR i IN 1..50 LOOP
        new_uid := gen_random_uuid();
        lon := 71.38 + (random() * 0.15);
        lat := 51.08 + (random() * 0.15);
        p_hash := decode(md5(random()::text), 'hex');
        p_enc := decode(md5(random()::text), 'hex');

        INSERT INTO social.users (id, phone_hash, phone_encrypted, password_hash, trust_score, verification_level)
        VALUES (new_uid, p_hash, p_enc, '$2a$10$dummyhashformockusers1234567890', (70 + random()*30)::int, 'photo_verified'::social.verification_level);

        INSERT INTO social.profiles (user_id, display_name, gender, city, location, looking_for, niyyah, madhab, bio)
        VALUES (new_uid, m_names[i], 'male'::social.gender, 'Astana', ST_SetSRID(ST_MakePoint(lon, lat), 4326), 'female'::social.gender, 'serious_marriage'::social.niyyah, 'hanafi'::social.madhab, 'Assalamu alaikum! I am ' || m_names[i] || ' from Astana.');

        INSERT INTO social.user_settings (user_id, age_range_min, age_range_max, max_distance_km)
        VALUES (new_uid, 18, 40, 200);
    END LOOP;
END $$;

COMMIT;