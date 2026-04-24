# Circles — Frontend Reference

Everything the client needs to integrate with the backend API.

---

## Base URL

```
http://localhost:8080       (local dev)
https://api.circles.app     (prod — TBD)
```

---

## Authentication

Every request (except `GET /health`) requires a Firebase ID token:

```
Authorization: Bearer <firebase_id_token>
```

- Obtain via Firebase Auth SDK — phone number + OTP flow.
- Tokens expire after 1 hour; Firebase SDK handles refresh automatically.
- The backend verifies the token on every request and rejects expired/invalid tokens with `401`.

---

## Phone Number Hashing

**The client must hash phone numbers before sending them to the backend.**
The backend never receives raw phone numbers.

```js
// JavaScript / React Native example
import { createHash } from 'crypto'; // or use expo-crypto

const phoneHash = createHash('sha256')
  .update('+15551234567')   // E.164 format, normalized
  .digest('hex');
```

- Normalize to E.164 before hashing: `+[country code][number]`, no spaces or dashes.
- The hash is what gets sent in all contact-related API calls.
- On registration, the user's own phone hash is also sent so others can find them.

---

## Registration Flow

On first sign-in (after Firebase OTP succeeds):

```
POST /users
{
  "display_name": "Alice",
  "phone_hash": "abc123..."    // SHA-256 of user's own phone number
}
```

This is idempotent — safe to call on every login. Returns the user object.

---

## API Endpoints

### Users

| Method | Path          | Auth | Body / Notes |
|--------|---------------|------|--------------|
| POST   | /users        | ✓    | `{ display_name, phone_hash }` — register or re-register |
| GET    | /users/me     | ✓    | Returns own profile |
| GET    | /users/:id    | ✓    | Returns friend's profile. 403 if not friends. |

**User object:**
```json
{
  "id": "uuid",
  "display_name": "Alice",
  "created_at": "2026-04-12T00:00:00Z"
}
```

---

### Contacts

| Method | Path             | Auth | Body / Notes |
|--------|------------------|------|--------------|
| POST   | /contacts        | ✓    | `{ phone_hash }` — add a contact |
| GET    | /contacts        | ✓    | List own contacts |
| DELETE | /contacts/:id    | ✓    | Remove a contact (breaks friendship if mutual) |

**Contact list item:**
```json
{
  "id": "uuid",
  "phone_hash": "abc123...",
  "display_name": "Bob",     // present if matched; null if pending
  "matched": true,
  "created_at": "..."
}
```

---

### Friends

| Method | Path      | Auth | Notes |
|--------|-----------|------|-------|
| GET    | /friends  | ✓    | Lists all mutual-match friends |

**Friend object:** same shape as User object.

A friendship exists automatically when both users have added each other's phone hash. No separate action needed.

---

### Circles (Groups)

| Method | Path                       | Auth | Body / Notes |
|--------|----------------------------|------|--------------|
| POST   | /groups                    | ✓    | `{ name }` — create a circle |
| GET    | /groups                    | ✓    | List circles you belong to |
| GET    | /groups/:id/members        | ✓    | List members (must be a member) |
| POST   | /groups/:id/members        | ✓    | `{ user_id }` — add a friend to circle |
| DELETE | /groups/:id/members/me     | ✓    | Leave the circle |

**Circle object:**
```json
{
  "id": "uuid",
  "name": "SF Squad",
  "created_by": "uuid",
  "created_at": "..."
}
```

**Add member rules (enforced by backend):**
- You must be a member of the circle.
- The person you're adding must be your friend.
- The person you're adding must be friends with **every** current member.
- Backend returns `403` if the circle check fails.

---

### Posts

| Method | Path                  | Auth | Body / Notes |
|--------|-----------------------|------|--------------|
| POST   | /groups/:id/posts     | ✓    | `{ image_url, caption? }` — create post |
| GET    | /groups/:id/posts     | ✓    | Group feed — most recent first |
| GET    | /feed                 | ✓    | Global feed — all circles, most recent first |

**Post object:**
```json
{
  "id": "uuid",
  "group_id": "uuid",
  "author": { "id": "uuid", "display_name": "Alice" },
  "image_url": "https://...",
  "caption": "optional text",
  "created_at": "..."
}
```

MVP: `image_url` is a link to an already-hosted image. No upload endpoint.

---

## Error Responses

All errors return JSON:

```json
{ "error": "human-readable message" }
```

| Status | Meaning |
|--------|---------|
| 400    | Bad request — missing or invalid fields |
| 401    | Missing or invalid Firebase token |
| 403    | Forbidden — access control violation |
| 404    | Resource not found |
| 409    | Conflict — e.g. duplicate contact |
| 500    | Server error |

---

## Postman Quick-Start

1. Set base URL variable: `{{base_url}} = http://localhost:8080`
2. Get a Firebase ID token (easiest: use Firebase Auth REST API or the Firebase Emulator)
3. Set collection-level header: `Authorization: Bearer {{token}}`
4. Hit `GET /health` first — should return `{"status":"ok"}` with no auth needed

### Getting a test token (dev only)
Firebase REST API — exchange a test phone number:
```
POST https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=<WEB_API_KEY>
```
Or use the Firebase Emulator Suite locally for full offline testing.

---

## Not in MVP (post-MVP features)

- Image/video upload (posts use URLs only)
- Push notifications
- Pagination (feeds return all results)
- Profile editing
- Group renaming / deletion
- Edit / delete posts
