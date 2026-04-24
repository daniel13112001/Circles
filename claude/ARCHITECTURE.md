# Circles — Architecture & Design Reference

## Tech Stack
- **Language:** Go (1.22+)
- **Database:** PostgreSQL (via pgx v5)
- **Auth:** Firebase Auth — phone OTP on client, JWT verification on backend
- **Router:** chi (lightweight, idiomatic Go)
- **Infra:** Docker Compose (Postgres + API)

---

## Project Layout

```
Circles/                        # Repo root
├── docker-compose.yml          # Primary entry point — starts Postgres + API
├── .env.example                # Only FIREBASE_PROJECT_ID needed for Docker
├── README.md                   # Quickstart for UI contributors
├── backend/                    # Go API
│   ├── cmd/api/main.go         # Entry point, wires everything together
│   ├── internal/
│   │   ├── auth/               # Firebase JWT middleware + token parsing
│   │   ├── users/              # User domain (handler, service, repo)
│   │   ├── contacts/           # Contacts domain
│   │   ├── friends/            # Friends domain (derived — no DB table)
│   │   ├── groups/             # Groups domain
│   │   ├── posts/              # Posts domain
│   │   └── db/
│   │       ├── db.go           # pgxpool setup
│   │       ├── migrate.go      # Migration runner (embed.FS)
│   │       └── migrations/     # SQL files embedded into binary at compile time
│   ├── secrets/                # Service account keys — gitignored
│   ├── docker-compose.yml      # Backend-only alternative (uses local Postgres)
│   ├── Dockerfile
│   └── .env.example            # Full env vars for local Go development
├── client/                     # Mobile/web frontend (TBD)
├── claude/                     # Design docs & session reference
└── docs/                       # Spec, whitepaper
```

Each domain follows a three-layer pattern:
```
handler.go   — HTTP (decode request, call service, encode response)
service.go   — Business logic + authorization checks
repo.go      — SQL queries only, no business logic
```

---

## Database Schema

### `users`
| Column        | Type      | Notes                              |
|---------------|-----------|------------------------------------|
| id            | UUID PK   | gen_random_uuid()                  |
| firebase_uid  | TEXT UNIQUE | from Firebase JWT sub claim      |
| display_name  | TEXT      |                                    |
| phone_hash    | TEXT UNIQUE | SHA-256 of phone number (client) |
| created_at    | TIMESTAMPTZ |                                  |

### `contacts`
| Column     | Type    | Notes                                          |
|------------|---------|------------------------------------------------|
| id         | UUID PK |                                                |
| owner_id   | UUID FK → users.id                             |
| phone_hash | TEXT    | the hash the owner added                       |
| created_at | TIMESTAMPTZ |                                            |
| UNIQUE(owner_id, phone_hash)                                      |

> Friendships are NOT stored. They are derived at query time.
> Two users A and B are friends iff:
>   - A has a contact entry with phone_hash = B.phone_hash
>   - B has a contact entry with phone_hash = A.phone_hash

### `groups`
| Column     | Type      | Notes             |
|------------|-----------|-------------------|
| id         | UUID PK   |                   |
| name       | TEXT      |                   |
| created_by | UUID FK → users.id |          |
| created_at | TIMESTAMPTZ |               |

### `group_members`
| Column    | Type      | Notes                   |
|-----------|-----------|-------------------------|
| group_id  | UUID FK → groups.id     |
| user_id   | UUID FK → users.id      |
| joined_at | TIMESTAMPTZ |               |
| PRIMARY KEY (group_id, user_id)          |

### `posts`
| Column     | Type      | Notes                       |
|------------|-----------|-----------------------------|
| id         | UUID PK   |                             |
| group_id   | UUID FK → groups.id         |
| author_id  | UUID FK → users.id          |
| image_url  | TEXT      | required                    |
| caption    | TEXT      | nullable                    |
| created_at | TIMESTAMPTZ |                           |

---

## Friendship Query (SQL)

```sql
-- Friends of user :uid
SELECT u.*
FROM users u
JOIN contacts c_me  ON c_me.phone_hash  = u.phone_hash  AND c_me.owner_id  = :uid
JOIN contacts c_them ON c_them.phone_hash = (SELECT phone_hash FROM users WHERE id = :uid)
                     AND c_them.owner_id  = u.id
WHERE u.id != :uid;
```

---

## Circle Check (Group Join)

When user X is being added to group G by member M:
1. M must be a member of G.
2. X must be friends with every current member of G.
3. Every current member of G must be friends with X.

Since friendship is symmetric by derivation (mutual contact), steps 2 and 3 are equivalent. One SQL check suffices:

```sql
-- Count members of G who are NOT friends with X
SELECT COUNT(*)
FROM group_members gm
JOIN users member ON member.id = gm.user_id
WHERE gm.group_id = :group_id
  AND NOT EXISTS (
    SELECT 1 FROM contacts c1
    JOIN contacts c2 ON TRUE
    WHERE c1.owner_id = :x_id   AND c1.phone_hash = member.phone_hash
      AND c2.owner_id = member.id AND c2.phone_hash = (SELECT phone_hash FROM users WHERE id = :x_id)
  );
-- Result must be 0 to allow join
```

---

## API Endpoints

All endpoints require `Authorization: Bearer <firebase_jwt>`.

### Users
| Method | Path           | Description                         |
|--------|----------------|-------------------------------------|
| POST   | /users         | Register (first login)              |
| GET    | /users/me      | Get own profile                     |
| GET    | /users/:id     | Get friend's profile (friends only) |

### Contacts
| Method | Path              | Description            |
|--------|-------------------|------------------------|
| POST   | /contacts         | Add a phone hash       |
| GET    | /contacts         | List own contacts      |
| DELETE | /contacts/:id     | Remove a contact       |

### Friends
| Method | Path      | Description       |
|--------|-----------|-------------------|
| GET    | /friends  | List friends      |

### Groups
| Method | Path                        | Description                     |
|--------|-----------------------------|---------------------------------|
| POST   | /groups                     | Create a group                  |
| GET    | /groups                     | List groups I belong to         |
| GET    | /groups/:id/members         | List members (member only)      |
| POST   | /groups/:id/members         | Add a friend to group           |
| DELETE | /groups/:id/members/me      | Leave group                     |

### Posts
| Method | Path                  | Description                        |
|--------|-----------------------|------------------------------------|
| POST   | /groups/:id/posts     | Create post in group               |
| GET    | /groups/:id/posts     | Group feed (member only)           |
| GET    | /feed                 | Global feed (all my groups)        |

---

## Auth Flow

1. Client authenticates via Firebase phone OTP → gets Firebase ID token (JWT).
2. Every API request includes `Authorization: Bearer <token>`.
3. Go middleware calls Firebase Admin SDK `VerifyIDToken()`.
4. Extracts `uid` (sub claim) → looks up internal user record.
5. Injects `userID` into request context for downstream handlers.

---

## Key Design Decisions

- **No friendship table:** friendships are derived via a JOIN on the contacts table. Simpler schema, no sync issues.
- **Circle enforced at join time only:** post-join friendship dissolution doesn't auto-remove from group; users can voluntarily leave.
- **No raw phone numbers on backend:** backend only ever sees SHA-256 hashes.
- **Image URLs only (MVP):** no file upload infra needed.
- **UUIDs for all PKs:** avoids enumeration attacks.
- **pgx/v5 with pgxpool:** connection pooling, no ORM.
- **chi router:** lightweight, middleware-friendly, idiomatic.
- **Migrations via hand-rolled embed.FS runner:** SQL files are embedded into the binary at compile time; the runner tracks applied files in a `schema_migrations` table and runs on startup. No external tool needed.
