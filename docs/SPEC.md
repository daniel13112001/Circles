# Circles — MVP Specification

A phone-based social app where friendships are implicit (mutual contact matching), every group is a trust circle, and all content lives within groups.

**Stack:** Go, PostgreSQL, Firebase Auth, Docker Compose
**Target:** Deployable product + backend engineer portfolio piece

---

## Core Concepts

- **Identity** is phone-number-based. Users authenticate via Firebase Auth (phone OTP). The backend never sees raw phone numbers or passwords.
- **Privacy-first contact matching.** Phone numbers are hashed (SHA-256) on the client before reaching the backend. The backend only stores and matches on hashes. At registration, the user's own phone number hash is stored so others can find them.
- **Friendship** is implicit. Two users become friends when both have added each other's phone number hash. There is no request/accept flow.
- **Groups** are trust circles. A user can only be added to a group if they are friends with every existing member. This is enforced at join time only.
- **Leaving** is voluntary. If a friendship breaks after group formation, the affected user can choose to leave. The system does not force removals.

---

## How Phone Matching Works

1. Alice signs up via Firebase Auth with her phone number. The client hashes her number (`SHA256("+15551234567")` → `abc123...`) and sends it to the backend along with her Firebase UID and display name.
2. The backend stores: `alice_uid` owns hash `abc123...`.
3. Bob wants to add Alice. He types `+15551234567` into the app. The client hashes it to `abc123...` and sends: "Bob wants to connect with `abc123...`".
4. The backend stores this as a contact entry and checks: does the owner of `abc123...` (Alice) have a contact entry for Bob's hash? If yes → mutual match → friends.

**The backend never sees a raw phone number.** All hashing happens on the client. The backend only compares opaque hash strings.

---

## MVP Functional Requirements

### 1. Authentication & Users

1.1. Authentication is handled entirely by Firebase Auth (phone number + OTP).
1.2. The backend verifies Firebase JWTs on every request. It does not manage passwords, sessions, or OTP.
1.3. On first login, the client sends the Firebase UID, display name, and the SHA-256 hash of the user's phone number. The backend creates a user record.
1.4. A user can view their own profile (display name).
1.5. A user can view the profile of a friend (display name only).

### 2. Contacts

2.1. A user can add a phone number hash to their contacts list. The client hashes the number before sending it.
2.2. A user cannot add their own phone number hash.
2.3. A user can view the list of contacts they have added (displayed as display names for matched users, or "pending" for unmatched hashes).
2.4. A user can remove a contact entry.

### 3. Friendships

3.1. A friendship exists between two users when both have added each other's phone number hash. This is derived, not stored as a separate record.
3.2. A user can view their list of friends (display names).
3.3. A user cannot manually create or delete a friendship directly — it is a consequence of mutual contact entries.
3.4. Removing a contact breaks the mutual match and dissolves the friendship.

### 4. Groups

4.1. A user can create a group with a name. The creator is automatically a member.
4.2. Any member of a group can add a friend to the group.
4.3. A user can only be added to a group if they are friends with **every** current member of that group (circle check, enforced at join time only).
4.4. A user can view the list of groups they belong to.
4.5. A user can view the members of a group they belong to.
4.6. A user can leave a group.

### 5. Posts

5.1. A user can create a post in a group they belong to.
5.2. A post contains: an image URL (required), an optional text caption, author, timestamp, and group reference.
5.3. A user can view posts in a group they belong to. Posts are displayed as an image feed (Instagram-style).
5.4. A user cannot post in a group they are not a member of.
5.5. **MVP shortcut:** Images are referenced by URL. No file upload endpoint — users provide a link to an already-hosted image.

### 6. Feeds

**Group Feed**
6.1. A user can view a feed of posts within a specific group, ordered by most recent first.

**Global Feed**
6.2. A user can view a feed of posts from all groups they belong to, ordered by most recent first.
6.3. No duplicate posts appear in the global feed.

### 7. Access Control

7.1. A user can only view groups they are a member of.
7.2. A user can only view posts in groups they belong to.
7.3. A user can only view profiles of their friends.
7.4. All endpoints require a valid Firebase JWT.

---

## Out of Scope (Post-MVP)

These are acknowledged but **not included** in the MVP build.

| Feature | Notes |
|---|---|
| Image / video upload | Posts reference images by URL. File upload (presigned S3 URLs, image resizing, CDN) is post-MVP. |
| Video posts | MVP is image-only. Video adds transcoding, streaming, and storage complexity. |
| Custom feeds (subset of groups) | Stretch goal from original spec. |
| Notifications | No alerts for new friendships, posts, or group additions. |
| Pagination | Feeds return all results. Add cursor-based pagination later. |
| Edit / delete posts | Posts are immutable in MVP. |
| User profile editing | Display name is set at registration and fixed. |
| Group renaming / deletion | Groups are permanent once created. |
| Rate limiting | Not enforced in MVP. |
| Frontend | API-only. A minimal UI may follow separately. |
| Salted hashing / pepper | MVP uses plain SHA-
56 on client. A production hardening pass would add a server-side pepper or use a KDF. |