# Circles — Key Decisions Log

Running log of non-obvious technical decisions. Add entries as they're made.

---

## 2026-04-12 — Naming

**App/group name: "Circles" over "Cliques" or "Friends"**

Three names were considered for both the app and the group concept:

- **Cliques** — technically precise (graph-theory clique = every node connected to every other), but carries negative connotations: exclusivity, high school social dynamics, gatekeeping. Bad brand signal even if the mechanics are clique-like.
- **Friends** — too generic; already owned by Facebook culturally, and doesn't convey the group/trust-circle concept at all.
- **Circles** — chosen. "Circle of friends" is a familiar, positive phrase. Copy reads naturally: "create a Circle", "your Circles", "add to Circle." Warm rather than exclusionary, even though the underlying constraint (must be friends with every member) is strict.

The graph-theory precision of "clique" is an implementation detail. The brand experience matters more.

---

## 2026-04-12 — Initial Design

**No friendship table**
Friendships are derived from the contacts table via a mutual-match JOIN. Avoids a separate table that must be kept in sync, and naturally handles dissolution when a contact is removed.

**Circle check enforced at join time only**
The spec explicitly says: "enforced at join time only." Post-join friendship changes don't auto-remove users from groups. Users can voluntarily leave.

**pgx/v5 over database/sql + ORM**
Raw pgx gives full control over PostgreSQL-specific types (UUID, TIMESTAMPTZ), named parameters, and batch queries without ORM magic. Code is more verbose but easier to audit.

**chi over gin/echo**
chi is closer to the standard library, composes well with net/http middleware, and has zero magic. For a backend portfolio piece, idiomatic standard-library-adjacent Go is preferable.

**UUID PKs over serial int**
Prevents enumeration of users, groups, posts via sequential IDs. Small performance cost, large security benefit.

**No ORM / no code generation (yet)**
MVP scope is small enough that hand-written SQL is fast and clear. Can revisit with sqlc later if schema stabilizes and query count grows.

**golang-migrate for migrations**
Simple, well-known tool. Keeps migrations as plain SQL files — easy to review and version.
