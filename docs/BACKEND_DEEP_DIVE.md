# TrueConnect Backend — Полное техническое объяснение

> **Для кого этот документ:** Для разработчиков, которые хотят полностью понять архитектуру, логику и код бэкенда TrueConnect — чтобы объяснить его другим, расширить функциональность или отладить проблемы.

---

## Содержание

1. [Общая картина — что такое этот бэкенд](#section-1)
2. [Слоёная архитектура (Clean Architecture)](#section-2)
3. [Разбор каждого модуля](#section-3)
4. [Жизненный цикл запроса (5 примеров)](#section-4)
5. [Схема базы данных](#section-5)
6. [Аутентификация и безопасность](#section-6)
7. [WebSocket Hub — полное объяснение](#section-7)
8. [Конфигурация и инфраструктура](#section-8)
9. [Паттерны и решения по дизайну](#section-9)
10. [Практические советы разработчику](#section-10)

---

<a name="section-1"></a>
## РАЗДЕЛ 1: Общая картина — Что такое этот бэкенд?

### Архитектурный паттерн

Бэкенд — это **Clean Architecture Modular Monolith** (модульный монолит с чистой архитектурой).

Все зависимости текут в одном направлении:

```
Запрос → Handler → Service → Repository (интерфейс) ← Adapter (реализация)
                                                     ↑
                                               Domain (чистые структуры)
```

Никакого фреймворка для DI нет. `cmd/api/main.go` — это **450 строк ручного внедрения зависимостей**. Каждый объект создаётся вручную и передаётся вниз по цепочке. Никакого `wire`, `fx`, `dig`.

### Сколько модулей и что они делают

| Модуль | Что он отвечает |
|---|---|
| **Auth** | Регистрация, вход, refresh токены, брутфорс-защита |
| **User** | Управление аккаунтом (`/users/me`, FCM токен, мягкое удаление) |
| **Profile** | Профиль, фото, геолокация (PostGIS), Niyyah/Madhab |
| **Matching** | Лента свайпов, лайк/пасс, создание мэтча |
| **Settings** | Фильтры пользователя (возраст, расстояние, мазхаб) |
| **Interactions** | Рейтинги (1–5 звёзд), подтверждение, трастовые события |
| **Reputation** | Вычисление траст-скора, рейтинговая таблица |
| **Chat** | WebSocket Hub, зашифрованные сообщения, статус прочтения |
| **MahramChat** | 3-сторонний групповой чат (женщина + мужчина + махрам) |
| **Posts** | Социальная лента, лайки, комментарии |
| **Notifications** | In-app уведомления + FCM push |
| **KYC** | Загрузка документов, Sumsub webhook |
| **Mahram** | Регистрация и верификация хранителя |
| **Whisper** | Анонимные жалобы сообщества |
| **Imam** | Встроенный каталог имамов, подтверждение никаха |
| **Admin** | Обзор Sybil кластеров, очередь KYC, управление пользователями |
| **Reports** | Жалобы на пользователей → ребро `REPORTED` в Neo4j |
| **Venues** | Статический каталог халяль мест встреч |

### От чего зависит бэкенд (внешние сервисы)

```
[Flutter / Next.js клиент]
           |
           | HTTPS (порт 80)
           ↓
      [nginx:alpine]
           |
           | HTTP (порт 8080)
           ↓
┌──────────────────────────────────────────────────────────┐
│         Go API (Gin HTTP + WebSocket Hub)                │
│  Auth  Profile  Matching  Chat  Feed  Reputation ...     │
│                                                          │
│  cmd/api/main.go ← ручное внедрение зависимостей        │
└──┬───────────┬───────────┬────────────┬──────────────────┘
   │           │           │            │
   ▼           ▼           ▼            ▼
[PostgreSQL] [Neo4j]   [Redis]      [MinIO]
PostGIS      GDS       go-redis     minio-go
порт 5433    7687/     6379         9000
             7474
                              [Firebase FCM] (опционально)
```

### Точка входа и старт приложения

Единственная точка входа: `cmd/api/main.go` → функция `run()`.

Порядок запуска:
1. Загрузить конфиг из переменных окружения
2. Подключиться к PostgreSQL, Neo4j (+ создать constraints), Redis, MinIO
3. Декодировать ключ шифрования из hex-строки
4. Создать все репозитории (адаптеры, оборачивающие соединения с БД)
5. Создать все сервисы (инъекция интерфейсов репозиториев)
6. Запустить фоновые горутины: `TrustEngine`, `PushWorker`, ticker очистки токенов
7. Создать `Hub` (WebSocket) и все HTTP обработчики
8. Создать роутер со всеми обработчиками
9. Запустить `http.Server` на `0.0.0.0:8080`
10. Ждать `SIGINT`/`SIGTERM` → плавное завершение (10 секунд)

---

<a name="section-2"></a>
## РАЗДЕЛ 2: Слоёная архитектура

### Слой 1: Handler (Контроллер)

**Задача:** Получить HTTP запрос, распарсить и валидировать входные данные, вызвать сервис, вернуть JSON.

**Может вызывать:** только сервисы.

**Не может вызывать:** базу данных, Redis, Neo4j — ни напрямую, ни через адаптеры.

**Реальный пример** — функция `Register` в [`internal/handler/auth_handler.go`](../internal/handler/auth_handler.go):

```go
func (h *AuthHandler) Register(c *gin.Context) {
    var req registerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        errorResponse(c, http.StatusBadRequest, "INVALID_JSON", ...)
        return
    }
    if err := validator.Validate.Struct(req); err != nil {
        errorResponse(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", ...)
        return
    }
    result, err := h.authService.Register(c.Request.Context(), service.RegisterInput{...})
    ...
    h.setRefreshTokenCookie(c, result.RefreshToken)
    c.JSON(http.StatusCreated, gin.H{"data": authResponse{...}})
}
```

Паттерн всегда одинаковый: **bind → validate → вызов сервиса → маппинг ошибок → ответ**.

**Формат ответа:** Все ответы обёрнуты:
- Успех: `{"data": ...}`
- Ошибка: `{"error": {"code": "...", "message": "..."}}`

Функция `errorResponse()` в `auth_handler.go` применяется во всех обработчиках для единообразия.

**Применение middleware:** Через Gin `.Use()` на группах маршрутов. Защищённая группа получает `authMW.Authenticate()` + `AuditLogMiddleware`. Отдельные эндпоинты дополнительно получают `userRL` (rate limiter per-user) как параметр маршрута.

---

### Слой 2: Service (Бизнес-логика)

**Задача:** Применять бизнес-правила, оркестрировать репозитории, координировать побочные эффекты.

**Может вызывать:** интерфейсы репозиториев, другие сервисы (по прямому указателю на структуру), утилиты из `pkg/`.

**Не может вызывать:** пакеты `adapter/` напрямую — сервисы никогда не импортируют конкретные реализации адаптеров.

**Реальное бизнес-правило** — защита от брутфорса при входе в [`internal/service/auth_service.go`](../internal/service/auth_service.go):

```go
func (s *AuthService) Login(ctx context.Context, phone, password string) (*AuthResult, error) {
    phoneHash := crypto.SHA256Hash([]byte(phone))
    phoneHashHex := hex.EncodeToString(phoneHash)

    // Блокировка после 5 неудачных попыток за 15 минут
    failCount, err := s.sessionStore.GetAuthFailureCount(ctx, phoneHashHex)
    if failCount >= 5 {
        return nil, fmt.Errorf("login: %w", domain.ErrRateLimitExceeded)
    }

    user, err := s.userRepo.GetByPhoneHash(ctx, phoneHash)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            // Инкрементируем даже для неизвестного телефона
            // чтобы предотвратить перечисление пользователей через тайминг
            _, _ = s.sessionStore.IncrementAuthFailure(ctx, phoneHashHex, 15*time.Minute)
            return nil, fmt.Errorf("login: %w", domain.ErrInvalidCredentials)
        }
    }
    ...
```

**Взаимодействие между модулями:** Сервисы вызывают другие сервисы по конкретному указателю. Например, `MatchingService` хранит `*NotificationService` для создания уведомлений при мэтче. Интерфейсы используются только на границе с репозиторием, не между сервисами — это прагматичное решение.

---

### Слой 3: Repository (Доступ к данным)

**Задача:** Выполнять SQL-запросы и преобразовывать строки БД в Go структуры.

**Библиотека:** `pgx/v5` с `pgxpool.Pool` — современный высокопроизводительный PostgreSQL драйвер. Никакого ORM. **Везде чистый SQL.**

**Транзакции:** Через паттерн UoW (Unit of Work). `postgres.UoW.Do(ctx, fn)` инжектирует `pgx.Tx` в контекст. Репозитории вызывают `runner(ctx, pool)`, который возвращает либо транзакцию из контекста, либо пул:

```go
// internal/adapter/postgres/uow.go
func runner(ctx context.Context, pool *pgxpool.Pool) dbRunner {
    if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
        return tx
    }
    return pool
}
```

✅ **Отличный паттерн:** Любая функция репозитория работает внутри транзакции или без неё без каких-либо изменений в своём коде.

**Обработка ошибок БД:**
- `pgx.ErrNoRows` → `domain.ErrNotFound`
- PostgreSQL код `23505` (нарушение уникальности) → `domain.ErrAlreadyExists`

Эти доменные ошибки всплывают до обработчика, который маппит их на HTTP статус коды.

**Реальный запрос** — поиск кандидатов с PostGIS в [`internal/adapter/postgres/profile_repo.go`](../internal/adapter/postgres/profile_repo.go):

```sql
SELECT p.user_id, p.display_name, ...
FROM social.profiles p
JOIN social.users u ON u.id = p.user_id
WHERE p.gender = $lookingFor
  AND ST_DWithin(
        p.location,
        ST_SetSRID(ST_MakePoint($lon, $lat), 4326)::geography,
        $maxDistMeters
      )
  AND p.user_id NOT IN ($excludeIDs)
  AND COALESCE(p.niyyah::text,'') = ANY($allowedNiyyahs)
ORDER BY u.trust_score DESC
LIMIT 20
```

---

### Слой 4: Domain (Доменная модель)

**Задача:** Чистые Go структуры без каких-либо внешних зависимостей. Единый источник истины для типов и sentinel ошибок.

**Нет разделения на DTO и доменную модель.** Одни и те же `domain.User`, `domain.Profile` используются на всех слоях. Обработчики создают лёгкие response-структуры (например, `authResponse`, `CandidateView`) inline — они не находятся в доменном пакете.

**Ключевые доменные структуры:**

```go
// domain/user.go
type User struct {
    ID                uuid.UUID
    PhoneHash         []byte         // SHA-256 телефона — для быстрого поиска
    PhoneEncrypted    []byte         // AES-256-GCM зашифрованный текст
    PasswordHash      string         // Argon2id encoded строка
    PublicKey         *string        // X25519 ключ для E2E шифрования
    VerificationLevel VerificationLevel
    TrustStatus       TrustStatus    // normal/under_review/suspended/banned
    TrustScore        int            // 0-100, кэшировано из Neo4j
    IsAdmin           bool
    FCMToken          *string
}

// domain/profile.go
type Profile struct {
    UserID      uuid.UUID
    Niyyah      Niyyah    // "nikah_year" | "serious_marriage" | "friendship"
    Madhab      Madhab    // "hanafi" | "shafii" | "maliki" | "hanbali" | "none"
    NoPhotoMode bool      // B3: скрыть аватар в ленте свайпов
    Latitude    *float64  // PostGIS координаты
    Longitude   *float64
    Languages   []string
    ...
}
```

**Sentinel ошибки** в `domain/errors.go` — всего 8 доменных ошибок:

```go
var (
    ErrNotFound           = errors.New("resource not found")
    ErrAlreadyExists      = errors.New("resource already exists")
    ErrInvalidInput       = errors.New("invalid input")
    ErrUnauthorized       = errors.New("unauthorized")
    ErrForbidden          = errors.New("forbidden")
    ErrAccountSuspended   = errors.New("account suspended")
    ErrRateLimitExceeded  = errors.New("rate limit exceeded")
    ErrInvalidCredentials = errors.New("invalid credentials")
)
```

---

### Слой 5: Repository Interface (Граница)

**Задача:** Определить, какие операции с данными существуют, не указывая как они реализованы.

Каждый репозиторий — это интерфейс в `internal/repository/`, его конкретная реализация — в `internal/adapter/postgres/`, `redis/`, `neo4j/`.

```go
// internal/repository/user_repo.go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByPhoneHash(ctx context.Context, phoneHash []byte) (*domain.User, error)
    UpdateTrustScore(ctx context.Context, id uuid.UUID, score int) error
    UpdateTrustStatus(ctx context.Context, id uuid.UUID, status domain.TrustStatus) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    // ...
}
```

Каждый адаптер доказывает соответствие интерфейсу на этапе компиляции:

```go
// internal/adapter/postgres/user_repo.go:20
var _ repository.UserRepository = (*UserRepo)(nil)
```

✅ **Отличный паттерн:** Эта compile-time проверка (`var _ Interface = (*Concrete)(nil)`) присутствует в **каждом** файле адаптера. Если метод в интерфейсе добавлен, а в адаптере нет — проект не скомпилируется.

---

<a name="section-3"></a>
## РАЗДЕЛ 3: Разбор каждого модуля

---

### МОДУЛЬ AUTH

**Назначение:** Управление идентичностью и сессиями. Регистрация, аутентификация, выдача и ротация JWT токенов, защита от брутфорса.

**Файлы:**
- Handler: [`internal/handler/auth_handler.go`](../internal/handler/auth_handler.go)
- Service: [`internal/service/auth_service.go`](../internal/service/auth_service.go)
- Repo interfaces: `internal/repository/user_repo.go`, `refresh_token_repo.go`, `session_store.go`
- Adapters: `internal/adapter/postgres/user_repo.go`, `refresh_token_repo.go`, `internal/adapter/redis/session_store.go`

**Ключевые структуры:**
- `AuthService` — хранит `userRepo`, `tokenRepo`, `sessionStore`, `graphRepo`, `uow`, `jwt`, `encryptionKey`, `refreshExpiry`
- `AuthResult` — `AccessToken`, `RefreshToken`, `ExpiresIn` (900 секунд), `UserID`
- `RegisterInput` — `Phone`, `Password`, `DisplayName`, `PublicKey`

**Функция `Register(ctx, input)`** — пошагово:
1. Валидирует `DisplayName` (не пустой)
2. Хэширует пароль с Argon2id (64MB памяти, 3 итерации, 4 параллельных потока)
3. SHA-256 хэширует телефон для поиска (`phone_hash`)
4. AES-256-GCM шифрует сырой номер телефона (`phone_encrypted`)
5. Выполняет UoW транзакцию: создаёт строку в `social.users` + строку в `social.profiles` атомарно
6. Создаёт узел пользователя в Neo4j (не фатально если ошибка — engine повторит)
7. Вызывает `issueTokens` → возвращает `AuthResult`

**Функция `Login(ctx, phone, password)`** — пошагово:
1. SHA-256 хэширует телефон, проверяет Redis счётчик брутфорса (ключ: `auth:fail:<hex>`)
2. Блокирует при 5 неудачах — возвращает `ErrRateLimitExceeded`
3. Получает пользователя по `phone_hash` — инкрементирует счётчик даже для неизвестного телефона (защита от timing атак на перечисление пользователей)
4. Верифицирует Argon2id хэш с константным сравнением по времени
5. Очищает счётчик при успехе
6. Вызывает `issueTokens`

**Функция `Refresh(ctx, rawRefreshToken)`** — пошагово:
1. SHA-256 хэширует токен
2. Получает сохранённый токен по хэшу из PostgreSQL
3. **Обнаружение компрометации**: если `revoked=true`, немедленно отзывает ВСЕ токены этого пользователя (family revocation) и возвращает 401
4. Помечает использованный токен как отозванный (one-time-use)
5. Выдаёт новую пару токенов

**Функция `issueTokens(ctx, user)`** (приватная):
1. Генерирует JWT через `jwtManager.Generate`
2. Генерирует 32-байтовую криптографически случайную hex-строку как refresh token
3. SHA-256 хэширует её для хранения в PostgreSQL (`token_hash`)
4. Сохраняет хэш + дату истечения в `social.refresh_tokens`
5. Также сохраняет в Redis session store (для быстрого поиска при отзыве)
6. Возвращает сырой refresh token (единственный момент, когда он виден незашифрованным)

**API Эндпоинты:**
```
POST /v1/auth/register     → создаёт аккаунт, устанавливает куки refresh_token, возвращает access token
POST /v1/auth/login        → аутентификация, устанавливает куки, возвращает access token
POST /v1/auth/refresh      → читает куки refresh_token, ротирует токены
POST /v1/auth/logout       → отзывает refresh token, очищает куки
POST /v1/auth/verify-phone → только для dev, принимает любой 6-значный код
```

Auth эндпоинты имеют вторичный rate limit: 20 req/min по IP (против 120 для других публичных маршрутов).

**Бизнес-правила:**
- Телефон никогда не хранится в открытом виде — SHA-256 для поиска, AES-256-GCM для восстановления
- Пользователь + профиль создаются всегда вместе (атомарная UoW транзакция)
- Заблокированные/приостановленные пользователи не могут войти
- Refresh токены одноразовые и хэшируются перед хранением
- Повторное использование токена (replay атака) вызывает полный отзыв сессии

---

### МОДУЛЬ MATCHING

**Назначение:** Лента свайпов. Получение пакета кандидатов с множеством фильтров, запись лайков/пассов, обнаружение взаимных мэтчей.

**Файлы:**
- Handler: [`internal/handler/matching_handler.go`](../internal/handler/matching_handler.go)
- Service: [`internal/service/matching_service.go`](../internal/service/matching_service.go)
- Adapter: [`internal/adapter/postgres/match_repo.go`](../internal/adapter/postgres/match_repo.go)

**Ключевые структуры:**
- `CandidateView` — карточка в ленте свайпов: `UserID`, `DisplayName`, `AvatarURL`, `AvatarBlurred`, `TrustScore`, `Niyyah`, `Madhab`, `IsKYCVerified`
- `LikeResult` — `Matched bool`, `MatchID uuid.UUID`

**Функция `GetCandidates(ctx, userID)`** — наиболее сложная в кодовой базе:

1. Загружает профиль и настройки запрашивающего
2. Получает `seenIDs` из Redis (TTL 24ч) — уже показанные кандидаты
3. Получает `matchedIDs` — уже замэтченные пользователи
4. Получает `rejectedIDs` из таблицы `social.swipe_rejections` — постоянные пассы
5. Получает `blockedIDs` — двунаправленные блокировки
6. **B1: Фильтр Ниях** — если у запрашивающего `nikah_year`, показывать только `nikah_year` + `serious_marriage` кандидатов
7. **Автоопределение пола** — если `LookingFor` не задан, выводим из гендера (исламский стандарт по умолчанию)
8. Выполняет `profileRepo.FindCandidates(opts)` — PostGIS `ST_DWithin` запрос со всеми фильтрами
9. **Fallback 1:** Если пусто и есть seenIDs — повторить без seenIDs
10. **Fallback 2:** Если всё равно пусто — убрать фильтр расстояния (глобальный пул)
11. **B2: Буст мазхаба** — добавить +10 к скору (макс 100) если кандидат того же мазхаба
12. **B3: Режим без фото** — сбросить `AvatarURL` и поставить `AvatarBlurred=true` для кандидатов с privacy режимом
13. Добавляет показанных кандидатов в Redis seen set

**Функция `RecordLike` в match_repo.go** — обработка гонки при взаимном лайке:

```go
// Сортируем пару чтобы user_a_id < user_b_id (всегда один порядок)
userA, userB := orderPair(userID, targetID)

// Advisory lock по паре UUID — предотвращает дублирование мэтча
SELECT pg_advisory_xact_lock($lockKey)

INSERT INTO social.matches (user_a_id, user_b_id, user_a_liked)
VALUES ($1, $2, true)
ON CONFLICT (user_a_id, user_b_id) DO UPDATE
    SET user_a_liked = true,
        matched_at = CASE
            WHEN social.matches.user_b_liked = true THEN COALESCE(matched_at, NOW())
            ELSE matched_at
        END,
        niyyah_timer_ends_at = CASE
            WHEN social.matches.user_b_liked = true
            THEN COALESCE(niyyah_timer_ends_at, NOW() + INTERVAL '90 days')
            ELSE niyyah_timer_ends_at
        END
RETURNING id, user_b_liked, matched_at
```

✅ **Отличный паттерн:** Ограничение `user_a_id < user_b_id` (CHECK constraint в БД) означает всегда ровно одну строку на пару пользователей. Advisory lock предотвращает два параллельных лайка, оба видящих `matched_at IS NULL`. UPSERT атомарно устанавливает флаг и обнаруживает взаимность в одном запросе.

**API Эндпоинты:**
```
GET  /v1/matching/candidates        → лента свайпов (с фильтрами, пагинация)
GET  /v1/matching/graph-candidates  → Neo4j рекомендации friends-of-friends
GET  /v1/matching/likes             → кто лайкнул меня (ожидающие взаимный лайк)
POST /v1/matching/like              → свайп вправо (rate limit: 30/мин на пользователя)
POST /v1/matching/pass              → свайп влево
GET  /v1/matches                    → список взаимных мэтчей
POST /v1/matches/:id/family-intro   → отметить этап знакомства с семьёй выполненным
POST /v1/matches/:id/unmatch        → удалить мэтч
POST /v1/users/:id/block            → заблокировать пользователя
```

**Бизнес-правила:**
- Нельзя лайкнуть самого себя (`userID == targetID → ErrInvalidInput`)
- Мэтч создаётся только когда оба лайкнули друг друга (`matched_at` устанавливается при втором лайке)
- При взаимном мэтче: уведомление получателю + FCM push событие (неблокирующая отправка в канал)
- Пасс: записать в `social.swipe_rejections` постоянно + добавить в Redis seen set
- Блок удаляет существующий мэтч + записывает в `social.blocked_users`; оба направления исключены из ленты

---

### МОДУЛЬ CHAT (WebSocket Hub)

**Назначение:** Real-time зашифрованный чат. `Hub` управляет in-memory WebSocket соединениями и маршрутизирует сообщения.

**Файлы:**
- Handler/Hub: [`internal/handler/chat_handler.go`](../internal/handler/chat_handler.go)
- Service: [`internal/service/chat_service.go`](../internal/service/chat_service.go)
- Adapter: `internal/adapter/postgres/message_repo.go`

**Структура Hub:**
```go
type Hub struct {
    mu             sync.RWMutex
    connections    map[uuid.UUID]*websocket.Conn  // userID → живое соединение
    chatSvc        *service.ChatService
    matchSvc       *service.MatchingService
    reputationSvc  *service.ReputationService
    mahramChatSvc  *service.MahramChatService
    rdb            *redis.Client                  // для Pub/Sub
    jwtManager     *tcjwt.Manager
    allowedOrigins []string
    isDev          bool
}
```

**Типы входящих WebSocket сообщений:**

| `type` | Действие |
|---|---|
| `chat_msg` | Отправить зашифрованное сообщение партнёру по мэтчу |
| `mahram_chat_msg` | Отправить в 3-сторонний махрам чат |
| `typing` | Передать индикатор набора партнёру |
| `read` | Отметить сообщения прочитанными, уведомить партнёра |
| `webrtc_offer` / `webrtc_answer` / `webrtc_ice_candidate` | Переслать WebRTC signaling |
| `pong` | Обновить TTL Redis presence |

**Поток `handleChatMsg`:**
1. Распарсить `{match_id, content}` из payload
2. Запустить `halalfilter.CheckMessage(content)` — жёсткая блокировка при откровенном контенте
3. Вызвать `chatSvc.SendMessage(ctx, senderID, matchID, content, isToxic)` — шифрует AES-256-GCM, сохраняет в БД
4. Если `isToxic`: вычесть 10 из траст-скора отправителя
5. Подтвердить отправителю: `sendTo(senderID, outMsg)` — прямая запись если подключён
6. Найти получателя из мэтча: `matchSvc.GetMatchByID` → `RecipientID(match, senderID)`
7. Опубликовать в Redis: `PUBLISH ws:user:<recipientID> <json>` — достигнет получателя даже на другом инстансе

**Мультиинстансная доставка через Redis:**
```
Инстанс A (отправитель)          Инстанс B (получатель)
    │                                  │
    ├─ sendTo(sender) → ACK            │
    │                                  │
    └─ rdb.Publish("ws:user:B", msg)   │
                                       │
                         ┌─── redisSubLoop подписан на "ws:user:B"
                         │
                         └─ hub.sendTo(recipientID, msg) → WebSocket
```

**Хранение сообщений:** Сообщения сохраняются в PostgreSQL **до** доставки в `chatSvc.SendMessage`. Контент шифруется AES-256-GCM с случайным nonce, хранится в `content_encrypted BYTEA`. Если получатель офлайн — сообщение в БД, он получит его через `GET /v1/matches/:id/messages`.

---

### МОДУЛЬ REPUTATION (Траст-скор)

**Назначение:** Вычисление, хранение и выдача траст-скора (0–100), который отображается на каждом профиле.

**Файлы:**
- Service: [`internal/service/reputation_service.go`](../internal/service/reputation_service.go)
- Neo4j Adapter: [`internal/adapter/neo4j/trust_graph_repo.go`](../internal/adapter/neo4j/trust_graph_repo.go)

**Формула траст-скора** (Cypher запрос в `trust_graph_repo.go`):

```cypher
MATCH (u:User {uid: $uid})

// 1. KYC бонус платформы
WITH u,
     CASE u.verification_level
       WHEN 'id_verified' THEN 10.0
       WHEN 'photo_verified' THEN 5.0
       ELSE 0.0
     END AS kyc_bonus

// 2. Агрегирование верифицированных рейтингов
OPTIONAL MATCH (u)<-[r:RATED]-(rater:User)
WHERE r.verified = true
WITH u, kyc_bonus, rater, r,
     CASE
       WHEN rater.verification_level IN ['id_verified', 'photo_verified'] THEN 1.5
       ELSE 1.0
     END AS id_weight,
     (COALESCE(rater.trust_score, 50) / 100.0) AS trust_weight

WITH u, kyc_bonus,
     SUM(r.score * id_weight * trust_weight) AS sum_effective,
     COUNT(r) AS rating_count

// Байесовское сглаживание (нейтральная база 2.5/5 из 5 виртуальных рейтингов)
WITH u, kyc_bonus,
     (sum_effective + (2.5 * 5.0)) / (rating_count + 5.0) AS smoothed

// 3. Штраф за жалобы
OPTIONAL MATCH (u)<-[rep:REPORTED]-(reporter:User)
WITH u, kyc_bonus, smoothed, COUNT(DISTINCT reporter) AS report_count

// 4. Финальный расчёт (0-100)
RETURN toInteger(
    CASE
      WHEN ((smoothed * 20.0) + kyc_bonus - (report_count * 15.0)) > 100.0 THEN 100.0
      WHEN ((smoothed * 20.0) + kyc_bonus - (report_count * 15.0)) < 0.0 THEN 0.0
      ELSE ((smoothed * 20.0) + kyc_bonus - (report_count * 15.0))
    END
) AS trust_score
```

**Где хранится скор:** После вычисления скор записывается в три места:
1. `social.users.trust_score` (PostgreSQL) — для SQL join при листинге профилей
2. Neo4j `User.trust_score` свойство — используется как `trust_weight` при расчёте других
3. Redis через `matchingCache.CacheTrustScore` — TTL 10 минут для быстрых чтений

**Event-driven пересчёт:** При подтверждении взаимодействия, `interactionSvc` отправляет UUID пользователя в `eventCh chan uuid.UUID`. `TrustEngine` воркер читает из этого канала и вызывает `reputeSvc.RecalculateScore`. Также есть 24-часовой ticker для пакетного пересчёта всех пользователей.

---

### МОДУЛЬ SYBIL DETECTION

**Назначение:** Обнаружение сетей фейковых аккаунтов с помощью обнаружения сообществ в графе.

Воркер `SybilDetector` запускается по таймеру. При каждом тике:

1. Удаляет устаревшую проекцию графа `trust-net` из Neo4j GDS
2. Проецирует все рёбра `RATED` и `MET_WITH` как ненаправленные
3. Запускает **обнаружение сообществ Louvain**: `CALL gds.louvain.stream('trust-net')`
4. Фильтрует кластеры где: `size >= 3` И `external_connections < size * 0.3`
   (менее 30% связей выходят за пределы кластера — паттерн изолированного клика)
5. Для каждого подозрительного кластера: устанавливает `TrustStatus = "under_review"` для всех участников в PostgreSQL и инвалидирует их Redis кэш траст-статуса

---

### МОДУЛЬ MAHRAM CHAT

**Назначение:** 3-сторонний зашифрованный групповой чат между женщиной, её потенциальным мужем и махрамом (мужчина-опекун).

Сообщения шифруются AES-256-GCM и хранятся в `social.mahram_chat_rooms` / `social.mahram_messages`. Доставка использует Redis Pub/Sub канал `mahram_room:<roomID>` — `MahramChatService.SendMessage` сам обрабатывает публикацию.

**Бизнес-правила:**
- Только участники мэтча или их зарегистрированные махрамы могут вступить в комнату
- Создание комнаты: `POST /v1/mahram-rooms` (с rate limit)
- Все три участника видят сообщения друг друга в разных цветах пузырей (во Flutter UI)

---

### МОДУЛЬ IMAM CONNECT

**Назначение:** Просмотр статического каталога имамов и запись подтверждения никаха.

Каталог имамов (`internal/pkg/imam/catalog.json`) встроен в бинарник на этапе компиляции через `//go:embed`. Никакой базы данных для имамов нет — данные живут в бинарнике.

```go
//go:embed catalog.json
var catalogJSON []byte

func init() {
    var entries []catalogEntry
    if err := json.Unmarshal(catalogJSON, &entries); err != nil {
        panic("imam catalog: invalid JSON: " + err.Error())
    }
    // Парсим один раз при запуске, паникуем при ошибке (fail-fast)
}
```

✅ **Отличный паттерн:** Встроить неизменяемые данные в бинарник. Ноль обращений к БД для статического списка. Добавление имама требует деплоя кода — приемлемо для курируемого списка.

**Эндпоинты:**
```
GET  /v1/imams                      → список имамов по параметру city
POST /v1/matches/:id/nikah-confirm  → imam_confirmed=true на мэтче, married_via_app на обоих профилях
```

---

<a name="section-4"></a>
## РАЗДЕЛ 4: Жизненный цикл запроса (5 полных примеров)

---

### Жизненный цикл 1: Вход пользователя

```
POST /v1/auth/login
  {"phone":"+77071234567","password":"Secure123!"}

→ Цепочка глобальных middleware:
    RequestID()     → инжектирует "X-Request-ID" заголовок
    Logger()        → логирует начало запроса
    Recovery()      → восстановление после паники
    CORS()          → устанавливает Access-Control заголовки
    RateLimit(120/мин, по IP)  → Redis INCR rl:ip:<ip>:<window>
    RateLimit(20/мин, по IP)   → вторичный для группы /auth

→ AuthHandler.Login
    c.ShouldBindJSON(&req)
    validator.Validate.Struct(req) → проверяет kz_phone формат, required

→ AuthService.Login
    phoneHash = SHA256("+77071234567")   → 32 байта
    phoneHashHex = hex.EncodeToString(phoneHash)

    sessionStore.GetAuthFailureCount(ctx, phoneHashHex)
    → Redis GET auth:fail:<hex>
    → если count ≥ 5 → ErrRateLimitExceeded → HTTP 429

    userRepo.GetByPhoneHash(ctx, phoneHash)
    → PostgreSQL: SELECT * FROM social.users WHERE phone_hash = $1
    → если не найден → инкрементировать счётчик → ErrInvalidCredentials → HTTP 401

    crypto.VerifyPassword(password, user.PasswordHash)
    → парсим "$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>"
    → пересчитываем argon2.IDKey → сравнение за константное время
    → если не совпадает → инкрементировать счётчик → HTTP 401

    sessionStore.ClearAuthFailures(ctx, phoneHashHex) → Redis DEL
    userRepo.UpdateLastLogin(ctx, user.ID)
    → UPDATE social.users SET last_login_at = NOW() WHERE id = $1

    issueTokens(ctx, user)
    → jwtManager.Generate(userID, verLevel, trustStatus, isAdmin)
       → jwt.NewWithClaims(HS256, Claims{sub, iat, exp, ver, tst, adm})
       → jwt.SignedString(secret)
    → generateSecureToken(32) → crypto/rand → hex строка
    → tokenHash = SHA256(rawRefreshToken)
    → tokenRepo.Create(ctx, userID, tokenHash, expiresAt)
       → INSERT INTO social.refresh_tokens (user_id, token_hash, expires_at)
    → sessionStore.StoreRefreshToken → Redis SETEX
    → возвращает AuthResult

→ AuthHandler.Login (продолжение)
    h.setRefreshTokenCookie(c, result.RefreshToken)
    → c.SetCookie("refresh_token", rawToken, 604800, "/v1/auth", "", true, true)
       HttpOnly=true, Secure=true (prod), SameSite=Strict, Path=/v1/auth

    c.JSON(200, {"data": {"access_token":"eyJ...", "expires_in":900, "user_id":"uuid"}})
```

---

### Жизненный цикл 2: Получение ленты свайпов

```
GET /v1/matching/candidates

→ Middleware: RequestID → Logger → Recovery → CORS → RateLimit(120/мин, IP)
→ AuthMiddleware.Authenticate()
    → заголовок: "Authorization: Bearer eyJ..."
    → jwtManager.Verify(token) → парсинг HS256, проверка истечения → Claims{sub:userID}
    → fetchTrustStatus(ctx, userID):
        TrustStatusCache.Get(ctx, userID) → Redis GET trust:<userID>
        при промахе → userRepo.GetByID(ctx, userID) → SELECT FROM social.users
        если banned/suspended/under_review → HTTP 403
    → c.Set("user_id", userID); c.Next()
→ AuditLogMiddleware() → логирует действие в таблицу audit_log

→ MatchingHandler.GetCandidates
    userID = middleware.GetUserID(c)

→ MatchingService.GetCandidates(ctx, userID)
    profileRepo.GetByUserID → профиль запрашивающего
    settingsRepo.Get → настройки (возраст, расстояние, мазхаб)
    matchingCache.GetSeenIDs → Redis SMEMBERS seen:<userID>
    matchRepo.ListMatches → уже замэтченные IDs
    matchRepo.GetRejectedIDs → SELECT target_id FROM social.swipe_rejections
    matchRepo.GetBlockedIDs → UNION запрос в обоих направлениях

    B1: niyyahCompatible(requesterProfile.Niyyah) → разрешённые значения Ниях
    Автовывод: lookingFor = противоположный гендер если не задан

    profileRepo.FindCandidates(FindCandidatesOpts{
        RequesterID, LookingFor, AgeRangeMin/Max,
        MaxDistanceMeters, RequesterLat/Lon,
        ExcludeIDs, Limit:20, AllowedNiyyahs, MadhabFilter
    })
    → PostGIS запрос ST_DWithin со всеми фильтрами
    → ORDER BY trust_score DESC LIMIT 20

    Fallback 1: если пусто и есть seenIDs → повторить без seenIDs
    Fallback 2: если всё равно пусто → убрать фильтр расстояния

    Для каждой строки:
        B2: score += 10 если мазхаб совпадает (макс 100)
        B3: если NoPhotoMode → пустой avatarURL, blurred=true
        → построить CandidateView

    matchingCache.AddSeen(ctx, userID, newSeenIDs, 24ч)
    → Redis SADD seen:<userID> <ids...>; EXPIRE 24h

→ c.JSON(200, {"data": [CandidateView, ...]})
```

---

### Жизненный цикл 3: Свайп вправо (лайк)

```
POST /v1/matching/like
  {"target_id":"uuid-of-target"}

→ Auth middleware → per-user rate limit (30/мин через Redis)
→ MatchingHandler.Like
    userID = middleware.GetUserID(c)
    targetID из тела запроса

→ MatchingService.Like(ctx, userID, targetID)
    если userID == targetID → ErrInvalidInput → HTTP 400

    matchRepo.RecordLike(ctx, userID, targetID)
    → orderPair(userID, targetID) → сортировка лексикографически → (userA, userB)
    → setCol = "user_a_liked" или "user_b_liked"
    → lockKey = int64 из первых байтов обоих UUID

    → PostgreSQL транзакция:
        SELECT pg_advisory_xact_lock($lockKey)  ← сериализуем эту пару
        INSERT INTO social.matches (user_a_id, user_b_id, user_a_liked)
        VALUES ($userA, $userB, true)
        ON CONFLICT DO UPDATE SET
            user_a_liked = true,
            matched_at = CASE WHEN user_b_liked=true THEN NOW() ELSE NULL END,
            niyyah_timer_ends_at = CASE WHEN user_b_liked=true
                                   THEN NOW()+90 days ELSE NULL END
        RETURNING id, user_b_liked, matched_at

        → если returned matched_at IS NOT NULL И otherLiked=true:
            matched=true, matchID=возвращённый UUID

    matchingCache.AddSeen(ctx, userID, [targetID], 24ч)

    если matched:
        notifSvc.Create → уведомление получателю типа "match"
    иначе:
        notifSvc.Create → уведомление получателю типа "like"

    pushCh <- PushEvent{UserID:targetID, Title:"Жаңа мэтч!", ...}
    → PushWorker получает → FCM.Send(ctx, fcmToken, title, body)

    → возвращает LikeResult{Matched:true, MatchID:uuid}

→ c.JSON(200, {"data": {"matched":true,"match_id":"uuid"}})
```

---

### Жизненный цикл 4: Отправка сообщения (WebSocket)

```
Установка WebSocket соединения GET /v1/ws:
    Hub.HandleWS()
    → websocket.Accept(w, r, opts)  -- библиотека coder/websocket
    → читать первое сообщение (таймаут 10с):
        {"type":"auth","token":"eyJ..."}
    → jwtManager.Verify(token) → Claims{sub:userID}
    → hub.register(userID, conn)
        mu.Lock()
        если старое соединение есть: old.Close(StatusPolicyViolation, "new connection")
        connections[userID] = conn
    → rdb.Set("presence:"+userID, "1", 90s)
    → wsjson.Write(ctx, conn, {type:"auth_ok"})
    → go pingLoop(ctx, conn, userID, pingDone)    -- пинг каждые 30с
    → go redisSubLoop(ctx, userID, pingDone)      -- подписка ws:user:<userID>
    → readLoop(ctx, conn, userID)

Пользователь отправляет:
    {"type":"chat_msg","payload":{"match_id":"uuid","content":"Сәлем!"}}

→ readLoop читает msg
→ switch msg.Type → case "chat_msg":
    → hub.handleChatMsg(ctx, senderID, msg.Payload)
        распарсить → wsChatPayload{MatchID, Content}

        halalfilter.CheckMessage("Сәлем!")
        → проверить списки blocked/warned слов (lowercase)
        → "Сәлем" чисто → blocked=false, warned=false

        chatSvc.SendMessage(ctx, senderID, matchID, content, false)
        → matchRepo.IsMatched(ctx, ...) → проверить что мэтч взаимный
        → crypto.Encrypt([]byte(content), encryptionKey)
           → AES-256-GCM, случайный 12-байтовый nonce
           → nonce + ciphertext = итоговый BYTEA
        → messageRepo.Save → INSERT INTO social.messages (match_id, sender_id, content_encrypted)
        → возвращает *domain.Message

        outMsg = wsOutgoing{Type:"chat_msg", Payload: dm}

        hub.sendTo(senderID, outMsg)   -- ACK отправителю
        → mu.RLock(); conn = connections[senderID]; mu.RUnlock()
        → wsjson.Write(5с таймаут, conn, outMsg)

        matchSvc.GetMatchByID(ctx, matchID, senderID)
        recipientID = RecipientID(match, senderID)

        hub.publish(ctx, recipientID, outMsg)
        → json.Marshal(outMsg)
        → rdb.Publish(ctx, "ws:user:"+recipientID, bytes)

        На инстансе получателя:
        → redisSubLoop получает из pubsub канала
        → json.Unmarshal → wsOutgoing
        → hub.sendTo(recipientID, outMsg)
        → wsjson.Write в WebSocket получателя
```

---

### Жизненный цикл 5: Обновление токена (Refresh)

```
POST /v1/auth/refresh
  (без тела — refresh token в HttpOnly куке)

→ Глобальный middleware: RequestID, Logger, Recovery, CORS, RateLimit(20/мин, IP)
  ВАЖНО: Auth middleware НЕТ — эндпоинт в публичной группе /auth

→ AuthHandler.Refresh
    refreshToken, err = c.Cookie("refresh_token")
    если ошибка или пусто → HTTP 401 "MISSING_REFRESH_TOKEN"

→ AuthService.Refresh(ctx, rawRefreshToken)
    tokenHash = SHA256(rawRefreshToken)

    tokenRepo.GetByTokenHash(ctx, tokenHash)
    → SELECT * FROM social.refresh_tokens WHERE token_hash = $1
    → если не найден → HTTP 401

    если storedToken.Revoked == true:
        tokenRepo.RevokeAllForUser → UPDATE SET revoked=true WHERE user_id=$1
        sessionStore.RemoveAllRefreshTokens → Redis DEL session:<userID>
        → HTTP 401 "token reuse detected"
        (Обнаружен replay attack — полный отзыв сессии)

    tokenRepo.Revoke(ctx, tokenHash)
    → UPDATE social.refresh_tokens SET revoked=true WHERE token_hash=$1

    userRepo.GetByID(ctx, storedToken.UserID)
    → SELECT * FROM social.users WHERE id=$1 AND is_active=true

    если suspended/banned → HTTP 403

    issueTokens(ctx, user)
    → генерируем НОВЫЙ access JWT + НОВЫЙ refresh token
    → сохраняем новый хэш refresh token в БД

→ AuthHandler.Refresh (продолжение)
    setRefreshTokenCookie(c, newRefreshToken)  -- заменяет старую куку
    c.JSON(200, {"data": {"access_token":"eyJ...", "expires_in":900, "user_id":"uuid"}})
```

---

<a name="section-5"></a>
## РАЗДЕЛ 5: Схема базы данных

### `social.users` — Таблица пользователей

**Назначение:** Основная запись идентичности. Хранит только хэшированные/зашифрованные данные — никакого plaintext PII.

| Колонка | Тип | Назначение |
|---|---|---|
| `id` | UUID PK | Генерируется `uuid_generate_v4()` |
| `phone_hash` | BYTEA UNIQUE | SHA-256 телефона. O(1) поиск без хранения plaintext |
| `phone_encrypted` | BYTEA | AES-256-GCM шифротекст. Расшифровывается только сервером |
| `email_encrypted` | BYTEA | То же, nullable |
| `password_hash` | TEXT | Argon2id encoded строка в PHC формате |
| `verification_level` | ENUM | `none` → `phone_verified` → `id_verified` → `photo_verified` |
| `trust_status` | ENUM | `normal` / `under_review` / `suspended` / `banned` |
| `trust_score` | SMALLINT (0–100) | Кэш из Neo4j; также используется в SQL JOIN |
| `is_admin` | BOOLEAN | Флаг администратора прямо в строке пользователя (нет таблицы ролей) |
| `is_active` | BOOLEAN | Флаг мягкого удаления; все запросы фильтруют `is_active = true` |
| `fcm_token` | TEXT | Firebase токен для push уведомлений |
| `public_key` | TEXT | X25519 публичный ключ для клиентского E2E шифрования |

### `social.profiles` — Профили

**Назначение:** Публичные данные профиля для отображения в ленте свайпов.

| Колонка | Тип | Назначение |
|---|---|---|
| `user_id` | UUID PK/FK | 1:1 с users |
| `location` | `GEOGRAPHY(POINT,4326)` | PostGIS geography; метровая точность дистанции |
| `niyyah` | TEXT | `nikah_year` / `serious_marriage` / `friendship` |
| `madhab` | TEXT | Исламская правовая школа |
| `languages` | `TEXT[]` | Массив языковых кодов |
| `no_photo_mode` | BOOLEAN | B3 функция приватности |
| `prompts` | JSONB | Массив `{question, answer}` объектов |

**Индекс:** `idx_profiles_location USING GIST (location)` — обязателен для производительности `ST_DWithin`.

### `social.matches` — Мэтчи

**Назначение:** Отслеживает отношение свайпа между любыми двумя пользователями. Одна строка на упорядоченную пару.

| Колонка | Тип | Назначение |
|---|---|---|
| `user_a_id`, `user_b_id` | UUID | Всегда отсортированы: `user_a_id < user_b_id` (CHECK ограничение) |
| `user_a_liked`, `user_b_liked` | BOOLEAN | Отдельные флаги для каждого направления |
| `matched_at` | TIMESTAMPTZ | NULL до взаимного лайка |
| `niyyah_timer_ends_at` | TIMESTAMPTZ | 90 дней после мэтча — дедлайн обязательства Ниях |
| `family_intro_done` | BOOLEAN | Этап: знакомство с семьёй завершено |
| `imam_confirmed` | BOOLEAN | Этап: никах подтверждён имамом |

**Почему не две строки?** Одна каноническая строка на пару избегает неоднозначности "в каком направлении лайк?". Функция `orderPair` в адаптере обеспечивает последовательную сортировку.

### `social.messages` — Сообщения

| Колонка | Тип | Назначение |
|---|---|---|
| `content_encrypted` | BYTEA | AES-256-GCM шифротекст. Сервер не может прочесть без ключа шифрования |
| `read_at` | TIMESTAMPTZ | NULL = непрочитано; устанавливается когда получатель шлёт `{type:"read"}` |

### `social.refresh_tokens` — Refresh токены

| Колонка | Тип | Назначение |
|---|---|---|
| `token_hash` | BYTEA UNIQUE | SHA-256 сырого токена. Сервер никогда не хранит сырой токен |
| `revoked` | BOOLEAN | Позволяет семейный отзыв при обнаружении replay |
| `expires_at` | TIMESTAMPTZ | Задание очистки удаляет истёкшие строки ежечасно |

### `social.interactions` — Взаимодействия/Рейтинги

Ограничение: `CONSTRAINT no_self_rating CHECK (rater_id != rated_id)` — принудительно на уровне БД.

### `social.swipe_rejections` — Постоянные пассы

Постоянная запись свайпов влево. Удерживает пройденных кандидатов из ленты навсегда, даже после истечения Redis TTL.

### `social.blocked_users` — Блокировки

Отношения блокировки. Оба направления (`blocker→blocked` и `blocked→blocker`) исключены из ленты открытий.

### `identity_vault.verifications` — Хранилище KYC

**Назначение:** Чувствительные данные KYC в отдельной ограниченной схеме. Хранит ИИН (казахстанский национальный ID), полное имя и хэш документа — всё AES-256-GCM зашифровано.

### Карта отношений

```
social.users ─1:1──► social.profiles
             ─1:many► social.media
             ─1:many► social.posts
             ─1:many► social.interactions (как rater)
             ─1:many► social.interactions (как rated)
             ─1:1──► social.user_settings
             ─1:many► social.refresh_tokens
             ─1:1──► identity_vault.verifications

social.matches (user_a_id, user_b_id оба FK к users, UNIQUE pair)
  └─1:many──► social.messages
  └─1:many──► social.mahram_chat_rooms ─1:many► social.mahram_messages

social.posts ─1:many► social.post_likes
             ─1:many► social.post_comments

social.users ─1:many► social.swipe_rejections
             ─1:many► social.blocked_users
             ─1:many► social.reports (как reporter и reported)
             ─1:many► social.whisper_reports
```

---

<a name="section-6"></a>
## РАЗДЕЛ 6: Аутентификация и безопасность

### Структура JWT токена

**Алгоритм:** HS256 (HMAC-SHA256). Секрет — строка ≥32 символов из `JWT_SECRET` env var.

**Клеймы (payload):**
```go
type Claims struct {
    jwt.RegisteredClaims          // sub (UUID пользователя), iat, exp
    VerificationLevel string      // "none"|"phone_verified"|"id_verified"|"photo_verified"
    TrustStatus       string      // "normal"|"under_review"|"suspended"|"banned"
    IsAdmin           bool        // true если администратор
}
```

| Клейм | Значение |
|---|---|
| `sub` | UUID строка пользователя |
| `iat` | Время выдачи (Unix timestamp) |
| `exp` | iat + 15 минут |
| `ver` | Уровень верификации на момент выдачи |
| `tst` | Траст-статус на момент выдачи |
| `adm` | Флаг администратора |

- **Access token:** истекает через 15 минут (настраивается через `JWT_ACCESS_EXPIRY`)
- **Refresh token:** истекает через 7 дней (настраивается через `JWT_REFRESH_EXPIRY`)
- **Секрет:** переменная окружения `JWT_SECRET`, проверяется ≥32 символов при старте

### Auth Middleware — Полный код

```go
// internal/handler/middleware/auth_middleware.go
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Извлечь Bearer токен
        header := c.GetHeader("Authorization")
        if header == "" || !strings.HasPrefix(header, "Bearer ") {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing_token"})
            return
        }
        token := strings.TrimPrefix(header, "Bearer ")

        // 2. Верифицировать подпись и срок действия JWT
        claims, err := m.jwt.Verify(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid_token"})
            return
        }

        // 3. Парсить userID из claims.Subject
        userID, err := uuid.Parse(claims.Subject)

        // 4. Получить trust_status: кэш Redis → PostgreSQL как fallback
        //    Если Redis недоступен — идём в БД
        //    Если БД недоступна — fail-closed (500, не пропускаем)
        status, err := m.fetchTrustStatus(c.Request.Context(), userID)

        // 5. Заблокировать ограниченные аккаунты
        switch status {
        case "banned", "suspended", "under_review":
            c.AbortWithStatusJSON(403, gin.H{"error": "account_restricted"})
            return
        }

        // 6. Инжектировать в gin контекст
        c.Set("user_id", userID)
        c.Set("verification_level", claims.VerificationLevel)
        c.Set("trust_status", status)
        c.Set("is_admin", claims.IsAdmin)
        c.Next()
    }
}
```

Обработчики извлекают userID через:
```go
userID, ok := middleware.GetUserID(c)  // возвращает uuid.UUID из gin контекста
```

### Механизм Refresh токенов

- Хранится в **HttpOnly куке** с именем `refresh_token`
- Область куки: `Path=/v1/auth` — отправляется только к auth эндпоинтам, не к каждому запросу
- `Secure=true` в production, `false` в development
- `SameSite=Strict` — CSRF защита
- **Ротация:** Каждый refresh отзывает старый токен и выдаёт новый
- **Семейный отзыв:** Если предъявлен отозванный токен (replay атака), ВСЕ токены этого пользователя немедленно отзываются

### Контроль доступа по ролям

Две роли: **user** (все) и **admin** (`is_admin=true`).

Флаг admin хранится в `social.users.is_admin` (BOOLEAN). Попадает в JWT при выдаче токена как `claims.IsAdmin`.

Проверка роли:
```go
// internal/handler/middleware/auth_middleware.go
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        isAdmin, _ := c.Get(ContextKeyIsAdmin)
        if admin, ok := isAdmin.(bool); !ok || !admin {
            c.AbortWithStatusJSON(403, gin.H{"error": "admin_required"})
            return
        }
        c.Next()
    }
}
```

Применяется к группе `/v1/admin/*` **после** `Authenticate()`, читает `is_admin` из уже проверенного контекста.

### Безопасность паролей

**Алгоритм: Argon2id** с параметрами OWASP:
- Память: 64 MB
- Итерации: 3
- Параллелизм: 4
- Соль: 16 случайных байт
- Длина ключа: 32 байта
- Формат: `$argon2id$v=19$m=65536,t=3,p=4$<base64salt>$<base64hash>`

Верификация использует `constantTimeEqual` — XOR цикл с одинаковым временем выполнения независимо от места несовпадения, предотвращая timing атаки.

### Безопасность PII данных

**AES-256-GCM** для телефона, email, ИИН, полного имени в KYC:
- Ключ: 32 байта из `ENCRYPTION_KEY` env var (64-символьная hex строка)
- Nonce: 12 случайных байт на каждое поле
- Хранение: `[12-байтовый nonce][ciphertext+16-байтовый auth tag]`

Формат:
```go
func Encrypt(plaintext, key []byte) ([]byte, error) {
    aesGCM, _ := cipher.NewGCM(aes.NewCipher(key))
    nonce := make([]byte, aesGCM.NonceSize())  // 12 байт
    io.ReadFull(rand.Reader, nonce)
    // Seal(dst, nonce, plaintext, additionalData)
    // Prepend nonce: [nonce][ciphertext+tag]
    return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}
```

---

<a name="section-7"></a>
## РАЗДЕЛ 7: WebSocket Hub — Полное объяснение

### Структура Hub

```go
type Hub struct {
    mu             sync.RWMutex
    connections    map[uuid.UUID]*websocket.Conn  // userID → живое соединение
    chatSvc        *service.ChatService
    matchSvc       *service.MatchingService
    reputationSvc  *service.ReputationService
    mahramChatSvc  *service.MahramChatService
    rdb            *redis.Client
    jwtManager     *tcjwt.Manager
    log            *slog.Logger
    allowedOrigins []string
    isDev          bool
}
```

### Регистрация клиента

Каждое соединение — одна запись в `connections map[uuid.UUID]*websocket.Conn`.

**Нет отдельной горутины для записи** — Hub пишет напрямую в `websocket.Conn` через `wsjson.Write` с 5-секундным таймаутом контекста.

**Есть горутина для чтения** (`readLoop`) и **горутина для ping** (`pingLoop`) — по одной на соединение. Redis подписка также работает в горутине (`redisSubLoop`).

**Single-device enforcement:** При новом соединении от того же userID, старое соединение принудительно закрывается:
```go
func (h *Hub) register(userID uuid.UUID, conn *websocket.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if old, ok := h.connections[userID]; ok {
        old.Close(websocket.StatusPolicyViolation, "new connection")
    }
    h.connections[userID] = conn
}
```

### Auth Handshake

WebSocket маршрут `GET /v1/ws` является **публичным** (вне защищённой группы). Auth происходит через первое WebSocket сообщение:

```
Почему первое сообщение должно быть auth?
→ WebSocket upgrade происходит с GET запросом
→ Браузеры не позволяют устанавливать произвольные заголовки при WS handshake
→ Токен не может быть в URL параметре (попадает в логи)
→ Решение: принять соединение, затем требовать auth сообщение в течение 10 секунд
→ При неудаче — закрыть соединение с StatusPolicyViolation
```

### Маршрутизация сообщений

При `handleChatMsg` на Инстансе A для пользователя на Инстансе B:

```
1. hub.sendTo(senderID, outMsg)   ← ACK отправителю (прямо на этом инстансе)
2. hub.publish(ctx, recipientID, outMsg)
   → rdb.Publish("ws:user:<recipientID>", json)

На Инстансе B:
3. redisSubLoop получает сообщение из pubsub канала
4. hub.sendTo(recipientID, outMsg)   ← доставляет получателю
```

Если получатель офлайн (не в `connections`), `sendTo` тихо возвращается — сообщение уже сохранено в PostgreSQL.

### Ключи Redis для WebSocket

| Ключ | Тип | Назначение |
|---|---|---|
| `ws:user:<userID>` | Pub/Sub канал | Кросс-инстансная доставка сообщений |
| `user:banned` | Pub/Sub канал | Real-time принудительный бан |
| `presence:<userID>` | STRING (TTL 90с) | Индикатор онлайн присутствия |

### Обнаружение и отключение забаненных

Hub слушает Redis `user:banned` канал:
```go
func (h *Hub) listenForBans() {
    pubsub := h.rdb.Subscribe(ctx, "user:banned")
    for msg := range pubsub.Channel() {
        uid, _ := uuid.Parse(msg.Payload)
        h.mu.Lock()
        if conn, ok := h.connections[uid]; ok {
            conn.Close(websocket.StatusPolicyViolation, "account restricted")
            delete(h.connections, uid)
        }
        h.mu.Unlock()
    }
}
```

---

<a name="section-8"></a>
## РАЗДЕЛ 8: Конфигурация и инфраструктура

### Загрузка конфигурации

`internal/config/config.go:Load()` читает **только переменные окружения** — нет YAML, TOML, конфигурационных файлов. Валидация при старте падает сразу если обязательные секреты отсутствуют.

**Все конфигурационные переменные:**

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `SERVER_HOST` | `0.0.0.0` | Адрес привязки |
| `SERVER_PORT` | `8080` | HTTP порт |
| `APP_ENV` | `development` | Влияет на CORS, куки, WS origin |
| `CORS_ORIGINS` | `http://localhost:3000` | Разрешённые origins через запятую |
| `PG_HOST/PORT/USER/PASSWORD/DBNAME/SSLMODE` | различные | Подключение к PostgreSQL |
| `NEO4J_URI/USER/PASSWORD` | `bolt://localhost:7687` | Подключение к Neo4j |
| `REDIS_HOST/PORT/PASSWORD/DB` | `localhost:6379` | Подключение к Redis |
| `MINIO_ENDPOINT/ACCESS_KEY/SECRET_KEY/USE_SSL/BUCKET` | различные | Объектное хранилище |
| `JWT_SECRET` | обязательно, ≥32 символов | HMAC ключ для подписи JWT |
| `JWT_ACCESS_EXPIRY` | `15m` | Длительность access токена |
| `JWT_REFRESH_EXPIRY` | `168h` (7 дней) | Длительность refresh токена |
| `ENCRYPTION_KEY` | обязательно, 64 hex символа | AES-256-GCM ключ для PII полей |
| `FIREBASE_CREDENTIALS` | пусто | Путь к Firebase service account JSON |

### Внедрение зависимостей

Ручное внедрение через конструкторы в `cmd/api/main.go`. Никакого фреймворка. Порядок:
1. Подключить всю инфраструктуру (PG, Neo4j, Redis, MinIO)
2. Создать адаптеры (репозитории), оборачивающие соединения
3. Создать сервисы, принимающие интерфейсы репозиториев
4. Создать сервисы, принимающие другие сервисы (например, `MatchingService` принимает `NotificationService`)
5. Создать обработчики, принимающие сервисы
6. Создать роутер со всеми обработчиками

### Docker Compose — 7 сервисов

| Сервис | Образ | Порты | Роль |
|---|---|---|---|
| `migrate` | Dockerfile проекта | — | Запускает `golang-migrate up`, завершается |
| `api` | Dockerfile проекта | `8080:8080` | Основной API сервер |
| `worker` | Dockerfile проекта | — | Фоновые воркеры |
| `postgres` | `postgis/postgis:16-3.4-alpine` | `5433:5432` | Основная БД с PostGIS |
| `neo4j` | `neo4j:5-community` + GDS плагин | `7474, 7687` | Граф доверия |
| `redis` | `redis:7-alpine` | `6379` | Кэш, Pub/Sub, сессии |
| `minio` | `minio/minio:latest` | `9000, 9001` | Объектное хранилище + консоль |
| `nginx` | `nginx:alpine` | `80:80` | Обратный прокси |

**Порядок запуска:** `migrate` ждёт `postgres` → `api` ждёт `migrate` завершения + все 4 БД здоровы → `nginx` ждёт `api`.

Redis настроен: `maxmemory 256mb --maxmemory-policy allkeys-lru`.

### Миграции БД

Инструмент: `golang-migrate/migrate/v4`. 19 миграций охватывают полную эволюцию схемы:

| Миграция | Что добавляет |
|---|---|
| 000001 | Основная схема: users, profiles, matches, messages, posts, interactions |
| 000002 | FCM токен |
| 000005 | Public key (X25519 для E2E) |
| 000006 | Флаг `is_toxic` на сообщениях |
| 000009 | Таблица notifications |
| 000010 | Admin/KYC функции |
| 000011 | Sybil кластеры |
| 000012 | `is_admin` флаг |
| 000013 | Halal поля: niyyah, madhab, languages, no_photo_mode |
| 000014 | Mahram таблицы |
| 000015 | Расширение мэтчей: family_intro_done, imam_confirmed, niyyah_timer |
| 000017 | swipe_rejections и blocked_users |
| 000018 | Индексы производительности |
| 000019 | Каскадные удаления |

---

<a name="section-9"></a>
## РАЗДЕЛ 9: Паттерны и решения по дизайну

### Паттерн Repository

**Зачем:** Отделяет бизнес-логику от технологии хранения. `AuthService` вызывает `repository.UserRepository` — ему всё равно, это PostgreSQL, in-memory мок или что-то ещё.

**Тестируемость:** Каждый файл тестов сервиса содержит ручные мок-реализации интерфейсов репозиториев. Нет фреймворка для генерации моков (no mockery, no gomock) — чистые рукописные моки.

### Паттерн Unit of Work (транзакции)

`UoW.Do(ctx, fn)` инжектирует `pgx.Tx` в контекст через `context.WithValue(ctx, txKey, tx)`. Каждая функция репозитория вызывает `runner(ctx, pool)` который проверяет этот ключ сначала.

```go
// Использование в регистрации: атомарное создание user + profile
s.uow.Do(ctx, func(txCtx context.Context) error {
    s.userRepo.Create(txCtx, user)        // выполняется внутри транзакции
    s.profileRepo.Upsert(txCtx, profile)  // та же транзакция
    return nil
})
// Если любая функция возвращает ошибку → автоматический rollback
// Если обе успешны → commit
```

### Advisory Locks для мэтчей

Ключ блокировки вычисляется из UUID пары:
```go
lockKey := int64(userA[0])<<56 | int64(userA[1])<<48 | int64(userB[0])<<8 | int64(userB[1])
```

Это позволяет параллельным лайкам от разных пар работать одновременно, сериализуя только одинаковые пары. Намного лучше глобальной блокировки таблицы.

### Обработка ошибок

**Паттерн:** `fmt.Errorf("имя операции: %w", err)` — оборачивание с контекстом на каждой границе слоя. Обработчики используют `errors.Is(err, domain.ErrXxx)` для разворачивания.

Цепочка ошибок строит стек-трейс-подобное описание: `"like: recording like: upsert: unique violation"`.

Никаких кастомных типов ошибок кроме 8 sentinel переменных в `domain/errors.go`.

### Propagation контекста

`context.Context` является первым параметром каждой функции репозитория и сервиса. Несёт:
- Отмену (через `signal.NotifyContext` в main)
- Транзакцию (`pgx.Tx` через паттерн UoW)
- Deadline запроса (из таймаутов `http.Server`)

Ничего бизнес-доменного не хранится в контексте. `userID` передаётся как явный параметр функции, не через контекст. Контекст Gin хранит идентичность пользователя (`user_id`, `is_admin`) — они извлекаются на уровне обработчика до вызова сервисов.

### Встроенные статические данные

Имамский каталог и каталог мест встреч используют `//go:embed`:
```go
//go:embed catalog.json
var catalogJSON []byte

func init() {
    // Парсим JSON при старте, паникуем при ошибке
    // Panic = fail-fast: лучше упасть при запуске чем в runtime
}
```

---

<a name="section-10"></a>
## РАЗДЕЛ 10: Практические советы разработчику

### Как добавить новый API эндпоинт

**5-шаговый процесс с именами файлов:**

1. **Доменная модель** в `internal/domain/` если нужны новые структуры данных
2. **Метод интерфейса репозитория** в `internal/repository/<domain>_repo.go`
3. **Реализация в адаптере** в `internal/adapter/postgres/<domain>_repo.go` (написать чистый SQL)
4. **Метод сервиса** в `internal/service/<domain>_service.go` (бизнес-логика)
5. **Метод обработчика + регистрация маршрута** в `internal/handler/<domain>_handler.go` + `router.go`

Для **нового модуля**: также создать доменный файл, файл интерфейса репо, файл адаптера, файл сервиса, файл обработчика, и подключить всё в `cmd/api/main.go`.

### Как добавить новую колонку в БД

1. Создать новую пару миграций: `migrations/0000XX_описание.up.sql` / `.down.sql`
2. Написать `ALTER TABLE social.xxx ADD COLUMN ...` в up файле
3. Обновить доменную структуру в `internal/domain/`
4. Обновить интерфейс репозитория если нужен новый метод запроса
5. Обновить SQL запросы в адаптере (обновить списки `SELECT`, `INSERT`/`UPDATE`)
6. Запустить `make migrate-up` локально

### Где самая сложная бизнес-логика

**Читать эти файлы особенно внимательно:**

1. [`internal/adapter/postgres/match_repo.go:31`](../internal/adapter/postgres/match_repo.go) — `RecordLike` с advisory lock + UPSERT — логика обнаружения взаимного мэтча
2. [`internal/adapter/neo4j/trust_graph_repo.go:127`](../internal/adapter/neo4j/trust_graph_repo.go) — `ComputeTrustScore` Cypher — взвешенная байесовская формула доверия
3. [`internal/service/auth_service.go:188`](../internal/service/auth_service.go) — `Refresh` — семейный отзыв токенов при replay атаке
4. [`internal/handler/chat_handler.go:59`](../internal/handler/chat_handler.go) — `Hub` — маршрутизация WebSocket и Redis pub/sub
5. [`internal/service/matching_service.go:114`](../internal/service/matching_service.go) — `GetCandidates` — вся логика фильтрации

### Текущие проблемы и технический долг

🐛 **`fixAvatarURL` — захардкоженный URL** в [`matching_service.go:22`](../internal/service/matching_service.go):
```go
func fixAvatarURL(url string) string {
    return strings.ReplaceAll(url, "minio:9000", "localhost:9000")
}
```
Заменяет Docker внутренний hostname на `localhost` — не будет работать в production где MinIO URL должен быть CDN или внешним доменом.

🐛 **`RecalculateAllScores` использует `fmt.Printf`** вместо структурированного логгера:
```go
fmt.Printf("RecalculateAllScores error at offset %d: %v\n", offset, err)
```
Обходит slog observability pipeline.

⚠️ **SybilDetector не подключён в API процессе** — `cmd/api/main.go` запускает `TrustEngine` и `PushWorker` но не создаёт `SybilDetector`. Нужно подключить в `cmd/api/main.go` или `cmd/worker/main.go`.

⚠️ **Нет настройки пула соединений** — используются дефолты `pgxpool`. В production нужно настроить `MaxConns`, `MinConns`, `MaxConnLifetime`.

⚠️ **`identity_vault` схема не ограничена на уровне PostgreSQL ролей** — дизайн предполагает отдельную роль БД, но приложение использует единого пользователя `tc_app` с доступом к обеим схемам.

⚠️ **WebSocket доставляет сообщения at-most-once** — если `wsjson.Write` падает, сообщение логируется но не повторяется. Получатель должен получить пропущенные через REST историю.

⚠️ **`VerifyPhone` — только dev stub** — в production возвращает 501. Реальная верификация через SMS (OTP) не реализована.

⚠️ **`is_admin` в JWT может устареть** — JWT выдаётся на 15 минут. Если admin статус отозван, пользователь сохраняет доступ до истечения токена. Auth middleware не перепроверяет `is_admin` из БД на каждом запросе (только `trust_status` перепроверяется через кэш).

### Узкие места при масштабировании (100,000 пользователей)

**Что сломается первым, по порядку:**

1. **Коллизия ключей advisory lock** — ключ блокировки использует только 4 байта UUID пары. Несвязанные пары могут хэшироваться в один ключ, вызывая ложную сериализацию. При высокой нагрузке свайпов — узкое место.

2. **In-memory `connections` map в Hub** — при нескольких API инстансах карта каждого инстанса неполна. Redis pub/sub обрабатывает кросс-инстансную доставку, но прямая запись должна всегда идти через Redis.

3. **`RecalculateAllScores` — полный проход по таблице** — 24-часовой пересчёт траст-скора итерирует каждого пользователя со статусом `normal` пакетами по 100, делая Neo4j Cypher запрос на каждого. При 100k пользователях = 1,000 Neo4j запросов на запуск.

4. **Fallback запросы в `GetCandidates`** — три потенциально дорогих PostGIS запроса на загрузку страницы свайпов, один из которых убирает фильтр расстояния (полный скан таблицы).

5. **Redis maxmemory 256MB** — seen-sets, счётчики rate limit, presence ключи и кэш траст-статуса конкурируют за это место. При 100k активных сессий — быстро заполняется.

6. **Neo4j GDS Louvain на полном графе** — `DetectSybilClusters` проецирует весь граф `RATED+MET_WITH` в память. При масштабе может исчерпать Neo4j heap.

---

## Приложение: Полный список API эндпоинтов

```
PUBLIC (без JWT):
  GET  /v1/health
  POST /v1/kyc/webhook
  POST /v1/auth/register
  POST /v1/auth/login
  POST /v1/auth/refresh
  POST /v1/auth/logout
  POST /v1/auth/verify-phone
  GET  /v1/ws              ← WebSocket (auth через первое сообщение)

PROTECTED (JWT обязателен):
  Profiles:
    GET    /v1/profiles/me
    PUT    /v1/profiles/me
    GET    /v1/profiles/me/photos
    POST   /v1/profiles/me/photos
    DELETE /v1/profiles/me/photos/:photoID
    GET    /v1/profiles/:id

  Matching:
    GET  /v1/matching/candidates
    GET  /v1/matching/graph-candidates
    GET  /v1/matching/likes
    POST /v1/matching/like       ← rate limited
    POST /v1/matching/pass       ← rate limited
    GET  /v1/matches
    POST /v1/matches/:id/family-intro  ← rate limited
    POST /v1/matches/:id/unmatch       ← rate limited
    POST /v1/users/:id/block           ← rate limited

  Settings:
    GET   /v1/settings
    PATCH /v1/settings

  User Account:
    GET    /v1/users/me
    POST   /v1/users/me/fcm-token
    DELETE /v1/users/me

  Interactions & Reputation:
    POST /v1/interactions              ← rate limited
    POST /v1/interactions/:id/confirm
    GET  /v1/users/:id/reputation
    GET  /v1/reputation/leaderboard

  Social Feed:
    GET    /v1/posts
    POST   /v1/posts                   ← rate limited
    GET    /v1/posts/:id
    DELETE /v1/posts/:id
    POST   /v1/posts/:id/like          ← rate limited
    DELETE /v1/posts/:id/like
    POST   /v1/posts/:id/comments      ← rate limited
    GET    /v1/posts/:id/comments

  Chat:
    GET /v1/matches/:id/messages

  Notifications:
    GET   /v1/notifications
    PATCH /v1/notifications/:id/read
    POST  /v1/notifications/read-all
    GET   /v1/notifications/unread-count

  KYC:
    POST /v1/kyc/submit   ← rate limited
    GET  /v1/kyc/status

  Reports:
    POST /v1/reports      ← rate limited

  Mahram Chat:
    POST /v1/mahram-rooms                       ← rate limited
    GET  /v1/mahram-rooms/:room_id/messages

  Mahram Registration:
    POST /v1/mahram                  ← rate limited
    GET  /v1/mahram
    POST /v1/mahram/:id/verify       ← rate limited

  Whisper:
    POST /v1/whisper                 ← rate limited

  Imam Connect:
    GET  /v1/imams
    POST /v1/matches/:id/nikah-confirm  ← rate limited

  Venues:
    GET /v1/venues

ADMIN (JWT + is_admin=true):
  GET  /v1/admin/users/under-review
  POST /v1/admin/users/:id/review
  GET  /v1/admin/sybil-clusters
  GET  /v1/admin/analytics
  GET  /v1/admin/users
  GET  /v1/admin/kyc/pending
  GET  /v1/admin/kyc/:id/document
  POST /v1/admin/kyc/:id/review
  GET  /v1/admin/whisper-flags
```

---

## Глоссарий терминов

| Термин | Значение в контексте проекта |
|---|---|
| **Ниях (Niyyah)** | Намерение пользователя: `nikah_year` (никах в течение года), `serious_marriage` (серьёзный брак), `friendship` (дружба) |
| **Мазхаб (Madhab)** | Исламская правовая школа: ханафи, шафии, малики, ханбали |
| **Махрам** | Законный мужской опекун женщины-мусульманки |
| **KYC** | Know Your Customer — верификация личности по документам |
| **ИИН** | Индивидуальный Идентификационный Номер — казахстанский аналог паспортного номера |
| **Trust Score** | Число от 0 до 100, показывающее репутацию пользователя в сети |
| **Sybil атака** | Создание множества фейковых аккаунтов для манипуляции системой репутации |
| **UoW** | Unit of Work — паттерн для атомарного выполнения нескольких операций с БД |
| **Advisory Lock** | PostgreSQL механизм блокировки на уровне приложения без блокировки строк таблицы |
| **Bayesian Smoothing** | Статистический метод: добавление виртуальных "нейтральных" рейтингов чтобы новые пользователи не получали экстремальные скоры |
| **GDS Louvain** | Neo4j Graph Data Science — алгоритм обнаружения сообществ в графе |
| **FCM** | Firebase Cloud Messaging — сервис Google для push уведомлений |
| **PostGIS** | Расширение PostgreSQL для геопространственных запросов (расстояние, радиус) |
| **AES-256-GCM** | Алгоритм симметричного шифрования с аутентификацией (Authenticated Encryption) |
| **Argon2id** | Современный алгоритм хэширования паролей, устойчивый к GPU/ASIC атакам |
| **At-most-once** | Доставка сообщения максимум один раз (возможна потеря, но не дублирование) |

---

*Документ создан на основе полного анализа исходного кода TrueConnect backend.*  
*Дата: Май 2026. Версия: Sprint 11 (feature-complete v1).*
