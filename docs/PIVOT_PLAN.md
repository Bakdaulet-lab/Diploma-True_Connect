# TrueConnect — Halal Pivot Plan (docs/PIVOT_PLAN.md)

## Context

TrueConnect is being pivoted from a general-purpose dating app to a niche Muslim-focused product for Kazakhstan (20–35 лет, ханафитский мазхаб, городская среда). The backend (Sprints 0–5) is nearly complete with 99 passing tests. The Flutter app has 33/44 empty files and must be built from scratch. The web (Next.js) is frozen as legacy — no deletion, just marked deprecated.

All work is on branch `app-v7`. New migrations start at `000013`. Existing migrations `000001–000012` are never modified.

---

## Milestones

| # | Name | After Sprint | Condition |
|---|------|-------------|-----------|
| M1 | Backend Islamic Identity Complete | 8 | New profile fields (niyyah, madhab, languages), mahram registration, no-photo mode in DB + API. Flutter onboarding + niyyah selection screens done. |
| M2 | Halal Matching + Chat Live | 10 | Niyyah-compatible filter, madhab boost, halal text filter, 90-day timer, mahram group chat deployed. Flutter discovery + chat screens updated. Demo seed data loaded. |
| M3 | Full Feature Set + Demo Ready | 13 | Imam Connect, Whisper Network, Family Intro milestone, flutter advanced screens all done. "Айгерим meets Алихан in 90 seconds" demo runs clean. |

---

## Sprint 6 — Database Foundation (Islamic Identity Fields)
**Duration:** 5 days | **Type:** Migration + Domain

### Goal
Three new migrations add all Islamic identity data to the schema. Domain structs and repository interfaces updated.

### Tasks

**1. `migrations/000013_halal_identity_fields.up/down.sql`** (new)
- Add ENUMs: `social.niyyah` (`nikah_year`, `serious_marriage`, `friendship`), `social.madhab` (`hanafi`, `shafii`, `maliki`, `hanbali`, `none`)
- Add columns to `social.profiles`: `niyyah social.niyyah`, `madhab social.madhab`, `languages TEXT[]`, `no_photo_mode BOOLEAN NOT NULL DEFAULT false`
- Add index: `idx_profiles_niyyah ON social.profiles(niyyah)`
- Add to `social.user_settings`: `modesty_level INT DEFAULT 0`, `niyyah_filter TEXT`, `madhab_filter TEXT`

**2. `migrations/000014_mahram.up/down.sql`** (new)
- New table `social.mahrams`: `id UUID PK`, `woman_user_id UUID FK → users`, `mahram_phone_encrypted BYTEA`, `mahram_phone_hash BYTEA UNIQUE`, `telegram_chat_id BIGINT`, `verification_status VARCHAR(20) DEFAULT 'pending'`, `verified_at TIMESTAMPTZ`, `created_at TIMESTAMPTZ`

**3. `migrations/000015_halal_matches_ext.up/down.sql`** (new)
- Add to `social.matches`: `niyyah_timer_ends_at TIMESTAMPTZ`, `family_intro_done BOOLEAN DEFAULT false`, `imam_confirmed BOOLEAN DEFAULT false`
- Add to `social.profiles`: `marital_status VARCHAR(20) NOT NULL DEFAULT 'single'` (values: `single`, `married_via_app`, `divorced`) — index `idx_profiles_marital_status`
- New table `social.mahram_chat_rooms`: `id UUID PK`, `match_id UUID UNIQUE FK → matches`, `mahram_user_id UUID FK → users`, `created_at TIMESTAMPTZ`
- New table `social.mahram_messages`: `id UUID PK`, `room_id UUID FK → mahram_chat_rooms`, `sender_id UUID FK → users`, `content_encrypted BYTEA`, `created_at TIMESTAMPTZ`
- New table `social.whisper_reports`: `id UUID PK`, `reporter_id UUID FK`, `reported_id UUID FK`, `meeting_match_id UUID FK → matches`, `feedback_encrypted BYTEA`, `strike_weight INT NOT NULL DEFAULT 1`, `strike_counted BOOLEAN DEFAULT false`, `admin_flagged BOOLEAN DEFAULT false`, `created_at TIMESTAMPTZ`

**4. `internal/domain/profile.go`** — add `Niyyah`, `Madhab string`, `Languages []string`, `NoPhotoMode bool`, `MaritalStatus string`; declare string constants for each enum value (marital: `MaritalSingle`, `MaritalMarriedViaApp`, `MaritalDivorced`)

**5. `internal/domain/` new files**:
- `mahram.go`: `Mahram` struct (ID, WomanUserID, MahramPhoneHash, TelegramChatID, VerificationStatus, VerifiedAt)
- `imam.go`: `Imam` struct (ID, Name, City, Mosque, Languages []string, AvailabilityNotes)

**6. `internal/domain/settings.go`** — add `ModestyLevel int`, `NiyyahFilter *string`, `MadhabFilter *string`

**7. `internal/repository/profile_repo.go`** — extend `FindCandidatesOpts` with `AllowedNiyyahs []string`, `MadhabFilter *string`, `LanguageFilter []string`; extend `CandidateRow` with `Niyyah`, `Madhab`, `Languages`, `NoPhotoMode`

**8. New repository interfaces**:
- `internal/repository/mahram_repo.go`: `MahramRepository` — `Create`, `GetByWomanID`, `GetByID`, `UpdateStatus`, `SetTelegramChatID`
- `internal/repository/whisper_repo.go`: `WhisperRepository` — `Create`, `GetByReportedUser`, `CountWeightedStrikes(reportedID) int`, `FlagForAdmin`

### Dependencies
- Sprints 0–5 complete (baseline `000012_add_is_admin`)

---

## Sprint 7 — Postgres Adapters + New Services
**Duration:** 5 days | **Type:** Backend adapter + service

### Goal
All new domain models have working Postgres adapters. New services for mahram registration and whisper network implemented.

### Tasks

**1. `internal/adapter/postgres/profile_repo.go`** (extend)
- Update `Upsert` to include `niyyah`, `madhab`, `languages`, `no_photo_mode`, `marital_status`
- Update `FindCandidates` SQL:
  - Marital status: `WHERE p.marital_status = 'single'` (always — married_via_app/divorced excluded from discovery)
  - Niyyah WHERE: `p.niyyah::text = ANY($allowedNiyyahs)` (service pre-computes allowed array)
  - Madhab filter: `($madhabFilter IS NULL OR p.madhab::text = $madhabFilter)`
  - Languages overlap: `($languageFilter IS NULL OR p.languages && $languageFilter)`
  - No-photo mode: if `row.NoPhotoMode AND NOT mutual_like` → return `avatar_url = ''` in CandidateRow

**2. `internal/adapter/postgres/mahram_repo.go`** (new)
- `Create`: INSERT into `social.mahrams`. Encrypt phone with existing `crypto.Encrypt`; hash with `crypto.SHA256Hash`
- `GetByWomanID`, `UpdateStatus`, `SetTelegramChatID`

**3. `internal/adapter/postgres/whisper_repo.go`** (new)
- `Create`, `CountStrikes(reportedID)`, `FlagForAdmin(reportedID)`

**4. `internal/adapter/postgres/settings_repo.go`** (extend)
- Include `modesty_level`, `niyyah_filter`, `madhab_filter` in `Get` / `Upsert`

**5. `internal/service/profile_service.go`** (extend)
- Add `Niyyah`, `Madhab`, `Languages`, `NoPhotoMode` to `UpsertProfileInput` and `ProfileView`
- Validate enum values in service layer (return `domain.ErrValidation` if invalid)

**6. `internal/service/settings_service.go`** (extend)
- Add new halal fields to `UpdateSettingsInput`

**7. `internal/service/mahram_service.go`** (new)
- `RegisterMahram(ctx, womanUserID, phone string)`: encrypt phone → hash → INSERT → Telegram bot stub (dev: log OTP "123456")
- `VerifyMahram(ctx, mahramID, code string)`: in dev mode accept any 6-digit code; set status=verified
- `GetMahrams(ctx, womanUserID)`

**8. `internal/service/whisper_service.go`** (new)
- `SubmitWhisper(ctx, reporterID, reportedID, matchID, feedback string)`:
  - Guard: reporter must have a **mutual match** with reported user (not just any historical contact) — query `social.matches WHERE (user1_id=$r OR user2_id=$r) AND (user1_id=$t OR user2_id=$t) AND status='matched'`
  - `strike_weight = 1`; if the match has `proof_of_meeting_confirmed = true` (physical meeting verified) → `strike_weight = 2`
  - Encrypt feedback AES-256-GCM → INSERT into `whisper_reports` with computed `strike_weight`
  - `CountWeightedStrikes(reportedID)` → SUM(strike_weight) WHERE strike_counted=false; if total ≥ 3: `FlagForAdmin` + admin notification via `NotificationService`

**9. Tests**:
- `mahram_service_test.go` (new): register + verify flow
- Extend `profile_service_test.go`: niyyah/madhab validation

### Dependencies
- Sprint 6 migrations and domain structs

---

## Sprint 8 — API Handlers, Routing & KYC Ban-Evasion (E1)
**Duration:** 5 days | **Type:** Backend handler + routing | **→ Milestone M1**

### Goal
All new features have HTTP endpoints. KYC IIN uniqueness constraint implemented (E1). Web marked legacy.

### Tasks

**1. `internal/handler/profile_handler.go`** (extend)
- `UpsertProfile` accepts + passes `niyyah`, `madhab`, `languages`, `no_photo_mode`
- `GetProfile` response includes new fields

**2. `internal/handler/settings_handler.go`** (extend)
- `UpdateSettings` accepts `modesty_level`, `niyyah_filter`, `madhab_filter`

**3. `internal/handler/mahram_handler.go`** (new)
- `POST /v1/mahram` — register mahram
- `GET /v1/mahram` — list own mahrams
- `POST /v1/mahram/:id/verify` — submit OTP code

**4. `internal/handler/whisper_handler.go`** (new)
- `POST /v1/whisper` — submit anonymous feedback (caller must have a historical match with reported user)

**5. `internal/handler/imam_handler.go`** (new stub — real logic Sprint 11)
- `GET /v1/imams?city=...`

**6. `internal/handler/router.go`** (extend)
- Add `Mahram *MahramHandler`, `Whisper *WhisperHandler`, `Imam *ImamHandler` to `RouterDeps`
- Register routes, add `RequireAdmin()` to new admin whisper endpoint

**7. `cmd/api/main.go`** (extend)
- Wire new services and handlers

**8. E1 — IIN uniqueness**:
- `migrations/000016_iin_hash_unique.up/down.sql`: add `iin_hash BYTEA NOT NULL DEFAULT ''::bytea` column + `UNIQUE INDEX idx_verifications_iin_hash` on `identity_vault.verifications`
- `internal/adapter/kyc/`: compute `iin_hash = SHA256(normalizedIIN)` before INSERT; catch `UNIQUE_VIOLATION` → return `domain.ErrConflict`
- KYC handler: return HTTP 409 on `ErrConflict`

**9. `README.md`** — add note: "Web (Next.js) is frozen as legacy from v6, no longer actively maintained. Mobile-first via Flutter."

**10. `docs/ARCHITECTURE.md`** — update frontend section (web = deprecated), add new tables to DDL summary

### Dependencies
- Sprint 7 services complete

---

## Sprint 9 — Niyyah Matching (B1, B2), No-Photo Mode (B3), Halal Chat Filter (C1), 90-day Timer (C3)
**Duration:** 5 days | **Type:** Backend service + pkg

### Goal
Core halal algorithms implemented and tested. Matching engine respects niyyah and madhab. Chat blocks prohibited content. 90-day timer column set on match creation.

### Tasks

**1. B1 — Niyyah-compatible filter in `internal/service/matching_service.go`**
- Compatibility matrix:
  - `nikah_year` → allowed: `[nikah_year, serious_marriage]`
  - `serious_marriage` → allowed: `[nikah_year, serious_marriage, friendship]`
  - `friendship` → allowed: `[nikah_year, serious_marriage, friendship]`
- Compute `AllowedNiyyahs []string` from requester's profile before calling `profileRepo.FindCandidates`

**2. B2 — Madhab score boost in `matching_service.go` `GetCandidates`**
- After fetching candidates, if `row.Madhab == requesterProfile.Madhab` AND madhab ≠ "none": `displayScore = min(row.TrustScore + 10, 100)`
- Boost is display-only, not persisted

**3. B3 — No-Photo mode enforcement** in `profile_repo.FindCandidates` (adapter already updated Sprint 7): ensure `avatar_url = ""` returned when conditions met; matching_service maps to `CandidateView.AvatarBlurred = true`

**4. C1 — Halal text filter**:
- `internal/pkg/halalfilter/filter.go` (new): `HalalFilter` with two-tier embed:
  - `//go:embed wordlist.example.txt` always compiled in (safe for repo — fictional/test terms only)
  - At `New()`: attempt `os.ReadFile("wordlist.txt")`; if found → use real list; else → fall back to embedded example list and log warn
- `CheckMessage(text string) (blocked bool, warned bool)`: severe terms → `blocked`; mild → `warned`
- `internal/pkg/halalfilter/wordlist.example.txt` (new): ~50 placeholder KZ/RU/AR terms for tests (non-offensive, clearly fictional)
- `internal/pkg/halalfilter/wordlist.txt` added to `.gitignore` — real production list managed outside repo
- `README.md`: section "Halal Filter Wordlist" — "Real wordlist managed outside repo, see deployment docs. Place `wordlist.txt` alongside the binary."
- `internal/handler/chat_handler.go` `handleChatMsg`: call `HalalFilter.CheckMessage`; if blocked → send `{"type":"content_blocked"}` WS error, don't persist; if warned → persist with `is_toxic=true`, prepend `{"type":"content_warning"}`

**5. C3 — 90-day Niyyah Timer**:
- `internal/adapter/postgres/match_repo.go` `RecordLike`: when match created, set `niyyah_timer_ends_at = NOW() + INTERVAL '90 days'`
- `internal/service/matching_service.go` `ListMatches`: add `NiyyahTimerEndsAt *time.Time` to `MatchView`
- `internal/worker/niyyah_timer.go` (new): daily cron at 09:00, queries matches where timer ends within 3 days, sends push via existing `NotificationService`

**6. Tests**:
- `matching_service_test.go`: table-driven niyyah compatibility cases
- `halalfilter/filter_test.go`: blocked/warned/pass scenarios

### Affected files
- `internal/service/matching_service.go`
- `internal/adapter/postgres/profile_repo.go` (SQL WHERE for niyyah array)
- `internal/adapter/postgres/match_repo.go`
- `internal/handler/chat_handler.go`
- `internal/pkg/halalfilter/filter.go` + `wordlist.txt` (new)
- `internal/worker/niyyah_timer.go` (new)

### Dependencies
- Sprint 8 (new profile fields in DB + domain)

---

## Sprint 10 — Mahram Group Chat (C2) + Family Intro Milestone (D2) + Demo Seed
**Duration:** 5 days | **Type:** Backend WS extension + data | **→ Milestone M2**

### Goal
WebSocket Hub extended for 3-way mahram rooms. Family Introduction milestone triggers trust score bonus. Demo seed data loaded.

### Tasks

**1. C2 — Mahram group chat WS extension in `internal/handler/chat_handler.go`**
- New WS message types in `readLoop`:
  - `mahram_chat_msg`: `{type, room_id, content}` → `handleMahramChatMsg`
  - Server emits `mahram_room_invite` to all 3 participants when room is created
- `handleMahramChatMsg`: validate sender is one of {woman, man, mahram} via `mahramChatSvc.GetRoom`; encrypt content; store in `mahram_messages`; publish to all 3 participants via existing Redis pub/sub (each user already has own `ws:user:{id}` channel)

**2. `internal/service/mahram_chat_service.go`** (new)
- `CreateRoom(ctx, matchID, mahramUserID)`: INSERT into `mahram_chat_rooms`, publish `mahram_room_invite` to all 3 users
- `SendMessage(ctx, roomID, senderID, plaintext)`: encrypt + INSERT into `mahram_messages`
- `GetMessages(ctx, roomID, userID, cursor, limit)`: pagination with decryption

**3. REST endpoints for mahram chat**:
- `POST /v1/mahram-rooms` (woman only, creates room for a match)
- `GET /v1/mahram-rooms/:room_id/messages`

**4. D2 — Family Introduction milestone in `internal/service/matching_service.go`**
- `FamilyIntroductionDone(ctx, matchID, callerID)`:
  - Set `social.matches.family_intro_done = true`
  - `reputeSvc.RecalculateDepth1Targets(ctx, callerID)` (existing) + INSERT special interaction `context="family_introduction"` with rating 5 for both participants → triggers +15 trust score naturally via existing trust engine
- `POST /v1/matches/:id/family-intro` endpoint in `matching_handler.go`

**5. Demo seed data `migrations/seed_halal_demo.sql`** (new, not a numbered migration — run separately):
- Айгерим: female, nikah_year, hanafi, [kazakh, russian], Almaty, trust_score=82, phone_verified
- Алихан: male, nikah_year, hanafi, [kazakh], Almaty, trust_score=78, phone_verified
- Both in `social.user_settings` with appropriate age/distance ranges
- UUIDs chosen to be memorable for demo (e.g., `a1g3r1m0-...`, `a1ikh4n0-...`)
- Алихан pre-liked Айгерим (existing `swipes` row) so match fires immediately on Айгерим's swipe

**6. `Makefile`** (extend):
- `make seed-halal`: runs `docker-compose exec -T postgres psql -U postgres -d trueconnect -f /migrations/seed_halal_demo.sql`
- `make reset-demo`: `make docker-reset && make migrate-up && make seed-halal`

**7. `README.md`** — add section **"Running the Demo"**:
```
## Running the Demo
1. make docker-up && make migrate-up
2. make seed-halal          # loads Айгерим + Алихан demo users
3. flutter run -d android   # or -d chrome
4. Login: Айгерим +7 111 111 1111 / password
5. Follow the 90-second demo scenario in docs/PIVOT_PLAN.md
make reset-demo             # full reset + reseed when needed
```

### Dependencies
- Sprint 9 (Halal filter, niyyah matching active), Sprint 7 (mahram service)

---

## Sprint 11 — Imam Connect (D3) + Whisper Admin (E2)
**Duration:** 4 days | **Type:** Backend service + admin

### Goal
Imam Connect catalog served via API with nikah confirmation flow. Whisper Network admin flags complete.

### Tasks

**1. D3 — Imam Connect**:
- `internal/pkg/imam/catalog.json` (new): 10 fictional imam entries across KZ cities (Almaty, Astana, Shymkent, Karagandy, Aktobe)
- `internal/pkg/imam/catalog.go`: `//go:embed catalog.json`, `ListByCity(city string)` function
- `internal/service/imam_service.go`:
  - `ListImams(ctx, city)`: returns filtered catalog
  - `ConfirmNikah(ctx, matchID, imamID, callerID)`:
    - Sets `social.matches.imam_confirmed = true`
    - Sets `social.profiles.marital_status = 'married_via_app'` for both users (**NOT** `is_active=false` — accounts stay active)
    - Users remain logged in: can view chat history with spouse, download their data; hidden from discovery via `marital_status='single'` filter
    - Records `interactions` with `context="nikah_confirmation"` + rating 5 for both → triggers +20 trust via existing engine
    - Future-proof: `marital_status='divorced'` re-enters discovery (post-MVP "Mark as Divorced" feature)
- `internal/handler/imam_handler.go` (complete implementation):
  - `GET /v1/imams?city=...`
  - `POST /v1/matches/:id/nikah-confirm {imam_id: string}`

**2. E2 — Whisper admin flow**:
- Whisper service 3-strike logic already in Sprint 7. Complete the admin side:
- `internal/service/admin_service.go` (extend): `GetWhisperFlags(ctx)` → query `social.whisper_reports WHERE admin_flagged = true`
- `internal/handler/admin_handler.go` (extend): `GET /v1/admin/whisper-flags` (RequireAdmin)
- Admin then uses existing `POST /v1/admin/users/:id/review` to ban/clear

**3. Wire in `cmd/api/main.go`**: `imamSvc`, `imamHandler`

### Dependencies
- Sprint 10 complete

---

## Sprint 12 — Flutter Infrastructure + Auth + Onboarding
**Duration:** 6 days | **Type:** Flutter core

### Goal
Flutter app has working network layer, routing, Islamic theme, and auth/onboarding through Niyyah selection.

### Tasks

**1. `frontend/lib/core/network/dio_client.dart`** (implement):
- `DioClient` with `BaseOptions` (baseUrl from `api_constants.dart`, 30s timeout)
- Interceptors: attach JWT from `flutter_secure_storage` on every request; on 401 → call `POST /auth/refresh` → retry once; log in debug mode

**2. `frontend/lib/core/router/app_router.dart`** (implement):
- GoRouter routes: `/splash` → `/onboarding` → `/auth/phone` → `/auth/otp` → `/niyyah` → `/home` (shell) → `/chat/:matchId` → `/mahram-chat/:roomId` → `/kyc` → `/imams` → `/profile/:userId` → `/whisper/:matchId`
- Auth guard: redirect to `/auth/phone` if no stored token

**3. `frontend/lib/core/theme/app_theme.dart`** (implement):
- Primary: deep green `#1B5E20`, secondary: gold `#F9A825`, background: warm white `#FAFAF5`
- Islamic geometric pattern SVG in `frontend/assets/patterns/`

**4. `frontend/lib/models/`** (implement all empty files):
- `profile.dart`: all fields including `niyyah`, `madhab`, `languages`, `noPhotoMode`
- `match.dart`: `niyyahTimerEndsAt`, `familyIntroDone`
- `settings.dart`: halal fields
- `user.dart`, `interaction.dart`, `post.dart` (minimal for MVP)

**5. `frontend/lib/providers/auth_provider.dart`** (implement):
- `StateNotifier<AsyncValue<User?>>`, methods: `register`, `login`, `verifyOTP`, `logout`, `refreshToken`
- JWT stored in `flutter_secure_storage`, FCM token sent on login

**6. Screens**:
- `frontend/lib/screens/splash/splash_screen.dart`: Islamic geometric animation (2s) → route-guard redirect
- `frontend/lib/screens/onboarding/onboarding_screen.dart` (new): 3-slide PageView
  - Slide 1: "Знакомства с намерением" + Quran 30:21 (Arabic + KZ)
  - Slide 2: "Безопасность" (KYC + Mahram explanation)
  - Slide 3: "Для Казахстана" + city map illustration
- `frontend/lib/screens/auth/login_screen.dart`: phone input (+7 KZ prefix), OTP on next screen
- `frontend/lib/screens/auth/register_screen.dart`: name, birth date, city
- `frontend/lib/screens/onboarding/niyyah_selection_screen.dart` (new): 3 large cards, Arabic + KZ labels, mandatory post-registration step
- Mahram registration screen UI text: "Mahram получит уведомление в Telegram через нашего бота **@MahabbatVerifyBot**"
  - Backend dev mode: logs OTP to console stdout + returns `mahram_id` in JSON response (Flutter dev reads OTP from server logs)
  - Defense talking point: "В production — реальный Telegram Bot API integration, в MVP — dev-friendly OTP visible in server logs"

**7. `frontend/lib/main.dart`** (rewrite):
- `ProviderScope` (Riverpod) + `MaterialApp.router` + `AppRouter` + `AppTheme`

### Dependencies
- Backend Sprints 6–9 deployed locally. `pubspec.yaml` already has Riverpod, GoRouter, Dio, flutter_secure_storage.

---

## Sprint 13 — Flutter Feature Screens + Demo Polish
**Duration:** 7 days | **Type:** Flutter screens + demo | **→ Milestone M3**

### Goal
All MVP screens implemented. Demo scenario runs end-to-end in 90 seconds.

### Tasks

**1. `frontend/lib/providers/matching_provider.dart`** (implement):
- `candidatesProvider`: FutureProvider → `GET /v1/matching/candidates`
- `likeProvider`, `passProvider`: StateNotifierProvider

**2. `frontend/lib/screens/discovery/discovery_screen.dart`** (extend existing partial impl):
- Wire Riverpod provider replacing hardcoded list
- Card: niyyah badge (moon icon), madhab label, language chips, Trust Score badge (color-coded)
- No-photo mode: if `avatarBlurred=true`, show blurred placeholder "Фото после взаимного лайка"
- Filter bottom sheet: niyyah, madhab, language multiselect

**3. `frontend/lib/screens/matches/matches_screen.dart`** (implement):
- Match list with avatar, name, trust badge
- 90-day timer: "Осталось X дней" in orange/red

**4. `frontend/lib/providers/chat_provider.dart`** (implement):
- Uses existing `WebSocketService`
- Handles new WS types: `content_warning`, `content_blocked`, `mahram_chat_msg`, `mahram_room_invite`

**5. `frontend/lib/screens/chat/chat_screen.dart`** (extend existing partial impl):
- Content warning banner for `is_toxic` messages
- "Пригласить Махрама" button in AppBar

**6. `frontend/lib/screens/chat/mahram_chat_screen.dart`** (new):
- 3-way chat UI: distinct avatar bubbles for woman (green), man (blue), mahram (gold)
- Invite banner if mahram not yet joined

**7. `frontend/lib/screens/profile/profile_screen.dart`** (own profile, implement):
- Niyyah / madhab / languages displayed
- Trust badge widget (color-coded by score tier)

**8. `frontend/lib/screens/profile/profile_detail_screen.dart`** (extend existing partial impl):
- Niyyah / madhab / language chips below bio
- Whisper Report FAB (ghost icon, bottom-right)

**9. `frontend/lib/screens/kyc/kyc_screen.dart`** (implement):
- File picker → `POST /v1/kyc/submit`; status polling → display pending/verified/rejected

**10. `frontend/lib/screens/settings/settings_screen.dart`** (implement):
- Toggle: No-Photo Mode
- Dropdown: Niyyah change (confirmation dialog)
- Slider: Modesty Level
- Section: Mahram Management (list + add + OTP verify)

**11. `frontend/lib/screens/imams/imam_connect_screen.dart`** (new):
- Imam list from `GET /v1/imams?city=...`
- "Подтвердить Никах" button → confirmation dialog → `POST /v1/matches/:id/nikah-confirm`

**12. `frontend/lib/widgets/whisper_report_modal.dart`** (new):
- Bottom sheet, anonymous text field, submit → `POST /v1/whisper`
- Confirmation: "Отзыв отправлен анонимно"

**13. `frontend/lib/screens/home/home_shell.dart`** (implement):
- Bottom nav: Discovery (crescent icon), Matches (heart), Profile (person), Settings (gear)

**14. Demo validation**:
- Run `migrations/seed_halal_demo.sql` against local Docker stack
- Walk through 90-second demo scenario (table below)
- Smoke test checklist: auth → match → chat → mahram → imam → nikah

---

## Demo Scenario: "Айгерим meets Алихан in 90 seconds"

| Time | Action | Expected System Response |
|------|--------|--------------------------|
| 0s | App launch as Айгерим | Splash → Discovery feed |
| 5s | Алихан's card appears | Card: niyyah "Никах 🌙", madhab "Ханафи", Trust Score 88 (78+10 madhab boost displayed) |
| 10s | Swipe right | Like recorded, `niyyah_timer_ends_at = now + 90d` written |
| 12s | (Алихан pre-liked) | Match! "Взаимный интерес. 90 дней до решения." |
| 20s | Open chat | Niyyah Timer badge: "90 дней" |
| 25s | Айгерим: "Ассалам алейкум" | Message sent, no filter trigger |
| 30s | Алихан replies | Chat working via WebSocket |
| 40s | Tap "Пригласить Махрама" | Phone input screen |
| 45s | Enter mahram phone + OTP "123456" | Mahram verified (dev mode); 3-way room created |
| 55s | Mahram chat opens | 3 participants visible with distinct bubbles |
| 65s | Open Imam Connect | 10 imams in Almaty listed |
| 75s | Tap "Подтвердить Никах" | Confirmation dialog: imam name + warning |
| 85s | Confirm | marital_status → 'married_via_app', hidden from discovery, trust +20, nikah interaction recorded |
| 90s | Final screen | "Никах подтверждён. Барака Аллаху фикум 🌙" (accounts remain active, chat history preserved) |

**Seed data users** (`migrations/seed_halal_demo.sql`):
- Айгерим: female, nikah_year, hanafi, [kazakh, russian], Almaty (43.238°N 76.899°E), trust=82, phone_verified
- Алихан: male, nikah_year, hanafi, [kazakh], Almaty (43.239°N 76.901°E), trust=78, phone_verified

---

## Risks & Mitigations

| # | Risk | P | Impact | Mitigation |
|---|------|---|--------|------------|
| R1 | KYC SumSub not real (mock only) | High | Low | Mock acceptable for demo. IIN uniqueness (E1) works on iin_hash. Mark as mock in README. |
| R2 | SMS for mahram expensive | High | Med | **Plan B (default)**: Telegram Bot API — free, widely used in KZ. Dev mode: any 6-digit code accepted. |
| R3 | Halal text filter KZ/RU false positives | Med | Med | Conservative wordlist.example.txt in repo (test/fictional terms). Real wordlist.txt outside repo in .gitignore — ops manages it per-deployment. All blocks logged for admin review. |
| R4 | Mahram group chat WS complexity | Med | Med | Additive changes only — new `case` in existing `readLoop`; each user already has own Redis pub/sub channel. No Hub refactor. |
| R5 | Imam Connect no real API | High | Low | Static JSON embedded via `//go:embed`. 10 fictional entries. Admin can update catalog.json and redeploy. |
| R6 | ML image filter | High | Low | Explicitly deferred. Text-only regex filter for MVP. Image report button as fallback. |
| R7 | 33 empty Flutter files — scope | High | High | Strict MVP scope: only demo-required screens built. Empty post-MVP files (feed, posts) remain empty. Demo does not need social feed. |
| R8 | FCM push for niyyah timer | Med | Med | Timer display is purely client-side (compute from `niyyah_timer_ends_at`). Push requires real FCM token — demo uses in-app display only. |
| R9 | iin_hash NOT NULL breaks existing rows | Low | Med | Migration uses `DEFAULT ''::bytea`. Existing rows unaffected. Uniqueness only enforced on new inserts. |

---

## Sprint Summary

| Sprint | Theme | Duration | Type | Milestone |
|--------|-------|----------|------|-----------|
| 6 | DB Foundation — Islamic identity fields | 5d | Migration + Domain | — |
| 7 | Postgres Adapters + New Services | 5d | Backend adapter/service | — |
| 8 | API Handlers + Routing + E1 (KYC uniqueness) | 5d | Backend handler | M1 |
| 9 | Niyyah filter, Madhab boost, Halal chat, Timer | 5d | Backend algorithms | — |
| 10 | Mahram Group Chat + Family Intro + Demo Seed | 5d | Backend WS + data | M2 |
| 11 | Imam Connect + Whisper Admin | 4d | Backend service | — |
| 12 | Flutter Core + Auth + Onboarding | 6d | Flutter | — |
| 13 | Flutter Feature Screens + Demo Polish | 7d | Flutter + demo | M3 |
| **Total** | | **~42 working days** | | |

---

## Critical Files (most-touched across sprints)

- `internal/adapter/postgres/profile_repo.go` — All matching filter logic (B1, B2, B3, language overlap) concentrated here
- `internal/handler/chat_handler.go` — Integration point for C1 (Halal filter), C2 (mahram_chat_msg), C3 timer push
- `internal/service/matching_service.go` — Madhab boost, AllowedNiyyahs, NiyyahTimerEndsAt in MatchView, FamilyIntroductionDone
- `frontend/lib/screens/discovery/discovery_screen.dart` — Primary demo UI surface; must wire Riverpod, new badges, No-Photo mode, filter sheet
- `migrations/000013_halal_identity_fields.up.sql` — Foundation migration; all subsequent data model work depends on it

---

## Non-Development Tasks

| Task | Sprint | File |
|------|--------|------|
| Mark web as legacy in README | 8 | `README.md` |
| Update ARCHITECTURE.md DDL section | 8 | `docs/ARCHITECTURE.md` |
| Create PIVOT_PLAN.md from this plan | — | `docs/PIVOT_PLAN.md` |
| Prepare diploma section change list (Sections 1, 3, 6.1, 6.2, 6.3) | 13 | External doc notes |
