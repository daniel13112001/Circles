# Circles — Build Plan

Ordered list of implementation milestones. Work top-to-bottom.

---

## Phase 1 — Project Skeleton ✅
- [x] `go mod init` — `github.com/danielyakubu/circles`
- [x] Directory structure: `backend/cmd/api`, `backend/internal/`, `backend/internal/db/migrations/`
- [x] `backend/docker-compose.yml` — Postgres + API services (backend-only)
- [x] `docker-compose.yml` (root) — primary entry point for all contributors
- [x] `Dockerfile` — multi-stage Go build
- [x] `.env.example` (root) — only FIREBASE_PROJECT_ID needed for Docker
- [x] `backend/.env.example` — full env vars for local Go development
- [x] `README.md` (root) — quickstart for UI contributors
- [x] `cmd/api/main.go` — HTTP server bootstrap (chi router, graceful shutdown)
- [x] `internal/db/db.go` — pgxpool setup, simple query protocol (Supabase pooler compat)
- [x] `client/` folder scaffolded

## Phase 2 — Auth Middleware ✅
- [x] Firebase Admin SDK added
- [x] `internal/auth/middleware.go` — verify ID token, inject firebaseUID into ctx
- [x] Context helpers: `FirebaseUIDFromCtx`, `WithUserID`, `UserIDFromCtx`

## Phase 3 — Database Migrations ✅
- [x] `001_create_users.sql`
- [x] `002_create_contacts.sql`
- [x] `003_create_groups.sql`
- [x] `004_create_group_members.sql`
- [x] `005_create_posts.sql`
- [x] Hand-rolled migration runner using `embed.FS` (SQL baked into binary)

## Phase 4 — Users Domain ✅
- [x] `internal/users/repo.go` — Upsert, GetByFirebaseUID, GetByID, AreFriends
- [x] `internal/users/service.go` — Register (idempotent), GetMe, GetFriend (403 if not friends)
- [x] `internal/users/handler.go` — POST /users, GET /users/me, GET /users/:id
- [x] `internal/users/middleware.go` — RequireUser (resolves firebaseUID → userID for all other domains)

## Phase 5 — Contacts Domain ✅
- [x] `internal/contacts/repo.go` — Add, List (with display name resolution), Delete
- [x] `internal/contacts/service.go` — validate not self-hash
- [x] `internal/contacts/handler.go` — POST /contacts, GET /contacts, DELETE /contacts/:id

## Phase 6 — Friends Domain ✅
- [x] `internal/friends/repo.go` — ListFriends (SQL JOIN on contacts), AreFriends
- [x] `internal/friends/service.go` — ListFriends, AreFriends
- [x] `internal/friends/handler.go` — GET /friends

## Phase 7 — Groups Domain ✅
- [x] `internal/groups/repo.go` — Create, AddMember, RemoveMember, ListForUser, ListMembers, IsMember, CircleCheckFails
- [x] `internal/groups/service.go` — Create (auto-add creator), AddMember (circle check), Leave, ListForUser, ListMembers
- [x] `internal/groups/handler.go` — POST /groups, GET /groups, GET+POST /groups/:id/members, DELETE /groups/:id/members/me

## Phase 8 — Posts Domain ✅
- [x] `internal/posts/repo.go` — Create, GroupFeed, GlobalFeed
- [x] `internal/posts/service.go` — membership check before create/read
- [x] `internal/posts/handler.go` — POST + GET /groups/:id/posts, GET /feed

## Phase 9 — Integration & Hardening ✅
- [x] All 16 endpoints smoke-tested (auth boundary verified)
- [x] README with setup instructions (`backend/README.md`)
- [ ] Full end-to-end flow tested with real Firebase tokens in Postman

---

## Dependencies (go.mod)
```
github.com/go-chi/chi/v5        ✅
github.com/jackc/pgx/v5         ✅
firebase.google.com/go/v4       ✅
github.com/joho/godotenv        ✅
```

---

## Notes / Watch-outs
- The circle check is the most complex query — write it carefully and test with edge cases (1-member group, 2-member group, etc.).
- `POST /users` is idempotent: ON CONFLICT (firebase_uid) DO UPDATE — returns existing user, never 409.
- Contacts list JOINs against users table to resolve display names; returns `matched: false, display_name: null` for unmatched hashes.
- Global feed: subquery across all groups the user belongs to, ORDER BY created_at DESC. No dedup needed (posts belong to exactly one group).
- Running from `backend/` directory: `go run ./cmd/api`
