# Your Task List — Things Claude Cannot Do

This file tracks everything that requires external accounts, credentials, or decisions from you.
Claude will append to this file whenever something new is needed.

---

## Firebase (BLOCKING — needed before auth can be tested)

- [ ] **Create a Firebase project** at https://console.firebase.google.com
  - Project name: "Circles" (or whatever you like)
- [ ] **Enable Phone Authentication**
  - Firebase console → Authentication → Sign-in method → Phone → Enable
- [ ] **Download a service account key (JSON)**
  - Firebase console → Project Settings → Service Accounts → Generate new private key
  - Save the file as `firebase-service-account.json` — do NOT commit this to git
  - Drop it in the repo root (it's already in `.gitignore`)
- [ ] **Note your Firebase Project ID**
  - Found at the top of Project Settings (looks like `circles-abc12`)
  - You'll put this in your `.env` file as `FIREBASE_PROJECT_ID`

---

## Deployment Target (decide before we build infra)

Pick one hosting path. Recommendation in order of simplicity:

- [ ] **Option A — Railway (easiest):** https://railway.app
  - Handles Postgres + Go API in one place, free tier available, deploy from GitHub
- [ ] **Option B — Render:** https://render.com
  - Similar to Railway, good free tier
- [ ] **Option C — Fly.io + managed Postgres:** https://fly.io
  - More control, great for Docker-based Go apps
- [ ] **Option D — Self-hosted VPS (DigitalOcean / Hetzner):**
  - Most control, cheapest long-term, requires more setup

> Once you decide, let Claude know and the Dockerfile/CI config will be tailored to that platform.

---

## Domain (optional but nice for a portfolio piece)

- [ ] Buy a domain if you want one (Namecheap, Cloudflare Registrar, etc.)
- [ ] Point it at wherever you deploy

---

## Local Development Setup

- [ ] **Install Docker Desktop** (if not already): https://www.docker.com/products/docker-desktop
  - This runs Postgres locally so you don't need a cloud DB during development
- [ ] **Install Go 1.22+** (if not already): https://go.dev/dl/
- [ ] **Copy `.env.example` → `.env`** once Claude creates it, and fill in:
  - `FIREBASE_PROJECT_ID`
  - Path to your service account JSON

---

## Git / GitHub

- [ ] Create a GitHub repo (if you want CI/CD or just remote backup)
  - Can be private — this is a portfolio piece so public is fine too
- [ ] Push the initial commit once the skeleton is built

---

*Last updated: 2026-04-12 — initial list*
