# Circles — Backend API

Go REST API for the Circles social app. Phone-based identity, implicit friendships, trust-circle groups, image feed.

**Stack:** Go 1.22 · PostgreSQL (Supabase) · Firebase Auth · chi router · Docker

---

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (optional — for local Postgres)
- A [Firebase project](https://console.firebase.google.com) with Phone Auth enabled
- A [Supabase](https://supabase.com) project (or any Postgres instance)

---

## Setup

### 1. Clone & enter the backend directory

```bash
cd backend
```

All `go` commands must be run from the `backend/` directory.

### 2. Copy and fill the env file

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Supabase session pooler URL (port 6543)
DATABASE_URL=postgresql://postgres.[PROJECT-REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres

FIREBASE_PROJECT_ID=your-project-id
FIREBASE_SERVICE_ACCOUNT_FILE=./secrets/your-service-account-key.json

PORT=8080
```

### 3. Add your Firebase service account key

Download from: Firebase Console → Project Settings → Service Accounts → Generate new private key

```
backend/secrets/<your-key-file>.json
```

The `secrets/` directory is gitignored — never commit credentials.

### 4. Run

```bash
go run ./cmd/api
```

On first start, all database migrations run automatically. You should see:

```
{"level":"INFO","msg":"migrations applied"}
{"level":"INFO","msg":"server starting","port":"8080"}
```

---

## Running with Docker (local Postgres)

```bash
docker compose up
```

This starts a local Postgres container and the API. Useful for fully offline dev.

---

## API Overview

Base URL: `http://localhost:8080`

All endpoints except `GET /health` require:
```
Authorization: Bearer <firebase_id_token>
```

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check (no auth) |
| POST | /users | Register / re-register |
| GET | /users/me | Own profile |
| GET | /users/:id | Friend's profile (403 if not friends) |
| POST | /contacts | Add a phone hash |
| GET | /contacts | List contacts (with match status) |
| DELETE | /contacts/:id | Remove a contact |
| GET | /friends | List mutual friends |
| POST | /groups | Create a circle |
| GET | /groups | List circles you belong to |
| GET | /groups/:id/members | List circle members |
| POST | /groups/:id/members | Add a friend to circle |
| DELETE | /groups/:id/members/me | Leave a circle |
| POST | /groups/:id/posts | Create a post |
| GET | /groups/:id/posts | Circle feed |
| GET | /feed | Global feed (all circles) |

See `claude/FRONTEND.md` for full request/response shapes and error codes.

---

## Key Design Notes

- **Phone numbers are never stored.** The client hashes them (SHA-256, E.164 format) before sending.
- **Friendships are implicit** — mutual contact match, no request/accept flow.
- **Circle check** — to join a circle, you must be friends with every current member. Enforced at join time only.
- **Migrations are embedded** in the binary via `embed.FS` — no external migration tool needed.

---

## Project Structure

```
backend/
├── cmd/api/main.go          # Entry point
├── internal/
│   ├── auth/                # Firebase JWT middleware
│   ├── users/               # Users domain + RequireUser middleware
│   ├── contacts/            # Contacts domain
│   ├── friends/             # Friends domain
│   ├── groups/              # Groups domain (circle check lives here)
│   ├── posts/               # Posts domain
│   └── db/
│       ├── db.go            # pgxpool setup
│       ├── migrate.go       # Migration runner
│       └── migrations/      # SQL files (embedded at compile time)
├── secrets/                 # Gitignored — put service account key here
├── Dockerfile
├── docker-compose.yml
└── .env.example
```
