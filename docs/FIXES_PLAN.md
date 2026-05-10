# TrueConnect — план фиксов перед защитой диплома

Ветка: app-v7. Каждый фикс — отдельный коммит.

## Сделано ✅

- [x] **P0-1** baseline + UTF-16 → UTF-8 миграции (commits 99ff5f6, 2212685)
- [x] **P0-2** Refresh cookie Secure flag matches APP_ENV (commit b5a3146)
- [x] **P0-3** Seed users: bcrypt → Argon2id, fixed down schema (commits e978778 + cef3d06)
- [x] **P0-4** RATED edge: CREATE → MERGE, no double-count (commit d798b18). Соответствует диплому Section 6.4.
- [x] **P0-5 part 1/3** Trust status cache adapter в Redis (commit f702d64)
- [x] **P0-5 part 2/3** Auth middleware читает trust_status из БД с кешем (commit cffc4f7).
  `AuthMiddleware` struct: cache hit → use; miss/Redis error → DB fallback + warn; DB error → 500 fail-closed.
  Блокирует: banned, suspended, under_review. Заменил старый `Auth(jwt)` (JWT-only, fail-open).
- [x] **P0-5 part 3/3** Invalidate trust status cache on ban/suspend/review (commit bc7d7bf).
  `AdminHandler.ReviewVerdict` → `tsCache.Delete(targetID)`.
  `AdminService.ResolveSybilCluster` → `tsCache.Delete(uid)` для каждого suspect.
  `SybilDetector.detect` → `tsCache.Delete(uid)` после UpdateTrustStatus.

## Дальше 📋

- [ ] **P0-6** Admin role + RequireAdmin middleware. Сейчас любой залогиненный юзер может вызывать `/v1/admin/*`.
  Нужно: миграция + колонка `is_admin BOOLEAN`, JWT claim `adm`, middleware `RequireAdmin()`, привязка к admin route group в router.

## Контекст

- Backend: Go 1.25.1, Gin, Clean Architecture, package alias `redisadapter`.
- Дипломный документ: `Second_Iteration__3_.docx` — Section 6.4 (RATED MERGE), Section 6.6 (Immediate Ban Flow).
- Тестовые юзеры: Diana (11..1), Alex (22..2), Mira (33..3), пароль "password".
- Утилита для хэшей: `go run ./cmd/seedhash/ -password=X -count=N`.
- Backup проекта: `C:\projects\Diploma-True_Connect-app-v6-BACKUP\`.

## Принципы работы

- Каждый фикс — отдельный коммит с conventional commit message.
- Перед каждым изменением кода — разведка (cat файлов, grep структур).
- После каждого изменения — go build, go vet, go test, git diff.
- Никаких git push без явного "PUSH сейчас".
