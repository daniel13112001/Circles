# Circles

Phone-based social app — backend in Go, UI in progress.

**Backend API:** Go · PostgreSQL · Firebase Auth  
**Base URL (local):** `http://localhost:8080`

---

## Quickstart (Docker — for UI contributors)

You need [Docker Desktop](https://www.docker.com/products/docker-desktop/).

We all use the **same Firebase project**. Ask the team for the `FIREBASE_PROJECT_ID`. You generate your own service account key for it — each person has their own copy of the key but they all point to the same project.

### 1. Copy the env file

```bash
cp .env.example .env
```

Open `.env` and set the project ID (ask the team for this value):

```env
FIREBASE_PROJECT_ID=circles-ac313
```

### 2. Generate your Firebase service account key

Go to **[Firebase Console](https://console.firebase.google.com) → (select the Circles project) → Project Settings → Service Accounts → Generate new private key**

Save the downloaded file as:

```
backend/secrets/firebase-service-account.json
```

The `backend/secrets/` directory is gitignored — never commit credentials.

### 3. Start the backend

```bash
docker compose up
```

This starts a local Postgres database and the API server. On first run, all database migrations apply automatically. You should see:

```
api-1  | {"level":"INFO","msg":"migrations applied"}
api-1  | {"level":"INFO","msg":"server starting","port":"8080"}
```

### 4. Verify

```
curl http://localhost:8080/health
# → {"status":"ok"}
```

---

## API

All endpoints except `GET /health` require:
```
Authorization: Bearer <firebase_id_token>
```

See [`backend/README.md`](backend/README.md) for the full endpoint list and [`claude/FRONTEND.md`](claude/FRONTEND.md) for request/response shapes.

---

## Project Structure

```
Circles/
├── backend/          # Go REST API
├── client/           # UI (in progress)
├── claude/           # Architecture docs and frontend contract
└── docs/             # Product spec and whitepaper
```

---

## Backend development (Go)

See [`backend/README.md`](backend/README.md).
