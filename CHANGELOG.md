# Changelog

All notable changes to the Castellers API are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

The API version is defined in [`VERSION`](VERSION) and exposed at `GET /api/v1/version`.

## [0.28.0] - 2026-11-13

Dues, second step: collections tied to events, the reminder of those events, the address for Interac transfers and the attendance of the members. Needs UI 1.12.0 to be used; an older UI ignores it.

### Added

- An event can point to the collection of dues that takes place at it: `duesCollectionUuid` in the payload of `POST /events` and `PUT /events/{uuid}`, and in the answers of `GET /events` and `GET /events/{uuid}`. An unknown collection is refused (400). In a series of recurring events, only the first one keeps the collection.
- Reminder of an event. When dues are collected at the event, the reminder email (automatic or sent by an admin) gets one more paragraph for the regular members and admins who have not settled them: the name of the dues, the amount and, when it is set, the address for Interac transfers. Members who paid or are exempt, the canalla and guests do not get it.
- `GET /dues/settings` (any member) and `PUT /dues/settings` (admin): `{"interacEmail": "..."}`, the address to which Interac transfers are sent. An empty value removes it.
- Attendance. `GET /members/{uuid}/attendance` returns `{"memberUuid", "firstVisit", "count"}`: the number of events at which the presence of the member was confirmed, and the start of the first one. A member can only read their own (403 otherwise), an admin can read anyone's. `GET /dues/attendance` (admin) returns it for every member who came at least once. An answer "yes" without confirmed presence does not count.

### Upgrade notes

- The migration adds `events.dues_collection_uuid` (empty for existing events) and the table `dues_settings`.
- An older UI does not send `duesCollectionUuid`: saving an event with it removes the collection of that event.

## [0.27.1] - 2026-11-13

No change in the application. The merge of 0.27.0 reached `master` without starting the deploy workflow; this release starts it, so that 0.27.0 (creation date of the members) gets tested and deployed.

## [0.27.0] - 2026-11-13

### Added

- Members have a creation date. `GET /members` returns it to admins as `createdAt` (Unix seconds, `0` when it is not known). It is set by the server when a member is added and cannot be sent by the client. The dues page of the UI (1.11.0) uses it to leave out accounts that were created long ago and never activated.

### Upgrade notes

- The migration adds `members.created_at` and fills it, for the members who are already there, with the date of their first registration email. Members added without an email keep `0`.

## [0.26.0] - 2026-11-12

Dues registry (cotisations), first step. The application does not handle payments: an admin records which dues a member has settled, and every change is kept in a journal. Needs UI 1.10.0 to be used; an older UI ignores it.

### Added

- Collections (cueillettes). A season follows the membership year of the Cercle culturel catalan: season `2026` runs from October 1st 2026 to September 30th 2027. It holds at most four collections, one of each `kind`: `cercle` (annual membership of the Cercle, collected in the autumn), `winter`, `spring` and `summer`. Each has an `amount` in cents and a `dueDate`.
  - `GET /dues/collections?season=2026` (any member): the collections of a season, or of all seasons without `season`.
  - `PUT /dues/collections/{season}` (admin): creates or updates the collections given in `{"collections": [{"kind", "amount", "dueDate"}]}`. A kind that is not given is left as it is. Collections cannot be removed.
- Recording (admin). `POST /dues/entries` with `{"memberUuid", "collectionUuids": [...], "action", "note"}` sets one or several collections to `paid`, `unpaid` or `exempt` for a member, all or nothing. A collection that is already in that state is skipped, so repeating a request changes nothing. Returns `{"entries": n}`, the number of journal entries written.
- `GET /members/{uuid}/dues`: every collection with the member's `status` (`paid`, `unpaid` or `exempt`) and the date of the last change. A member can only read their own (403 otherwise), an admin can read anyone's. It does not say who recorded the change.
- `GET /dues/states?season=2026` (admin): the current state of every member, for the pairs member/collection recorded at least once. A missing pair is unpaid.
- Journal. `GET /dues/log` (admin) lists who set what, for whom and when, most recent first, with the names of the member and of the author. Filters: `member`, `collection`, `actor`, `season`; paging with `limit` (100 by default, 500 at most) and `offset`. There is no endpoint to modify or remove an entry: a mistake is fixed by a new entry.

### Notes

- Dues only apply to regular members and admins: recording for a member of the canalla or a guest is refused (400).
- The current state is not stored: it is the last journal entry of the member for the collection.
- The note of an entry is encrypted in the database, like the personal data of the members. The journal keeps naming an admin whose account was deleted.

### Upgrade notes

- The migration adds two empty tables, `dues_collections` and `dues_log`. Nothing else changes.

## [0.25.0] - 2026-11-10

First step of pinya planning: the pilar de 4 gets pinya positions. Needs UI 1.8.0 to be drawn; an older UI ignores them.

### Added

- `GET /castells/types/1d4` now also returns 20 positions with `part: "pinya"`: the nucli (`agulla`, `contrafort`, 2 `crossa`, all with `cordon: 0`) and 8 rengles of 2 cordons (2 `mans`, 2 `vent`, 4 `lateral`, with `cordon: 1` and `2`). `column` tells the rengles of the same name apart; the convention is described at the top of `sql/0.25.0.sql`.
- Members are assigned to these positions through `position_members`, exactly like tronc and pom positions. No endpoint and no payload format changed.

### Upgrade notes

- The other castell types have no pinya positions and are unchanged.
- The migration only adds rows to `castell_positions` and `castell_positions_in_castells`. Existing castell models are not modified.

## [0.24.5] - 2026-11-08

JWT hardening: server-side sessions, so tokens can actually be revoked. No UI change is required.

### Security

- Each login now creates a session in Redis (`session:<id>`, plus an index `user_sessions:<member>`). Access and refresh tokens carry its id (`session_id` claim) and are accepted only while the session exists. The session expires after `jwt.refresh_ttl_days` without refresh.
- `POST /logout` ends the whole session. Every access and refresh token of that session stops working at once, including those issued by earlier refreshes. Other sessions of the member (other devices) are not affected.
- Changing the password (`/change_password`) ends the member's other sessions and keeps the current one. A reset through a registration or "forgot password" link ends all of them.
- Deleting a member ends their sessions. A refresh by a deleted member now gets 401 instead of 500.
- A revoked session gets the same response as an expired access token (403 with `code: "token_expired"`): the UI tries to refresh, the refresh fails with 401 and it goes back to the login page (or, on public pages, continues anonymously).
- If Redis cannot be reached, authenticated requests get 503 instead of a misleading 401/500, and `GET /events` and `GET /events/{uuid}` answer as for an anonymous request.

### Fixed

- `GET /events/{uuid}` with an expired access token returned 401 without `code: "token_expired"`, so the UI never refreshed the token on that page. It now returns 403 with the code. Other invalid tokens still get 401.

### Upgrade notes

- Refresh tokens issued before this version keep working: their first refresh moves them to a new session, and their Redis key then points to it, so logging out of that session also revokes them and refreshing them again reuses the same session. Once a member's sessions are revoked (password change, reset link, deletion), their pre-upgrade refresh tokens are refused.
- Refresh tokens are not rotated yet: a refresh token stays valid until it expires or its session ends. Rotation needs the UI to share tokens between tabs first.
- Every authenticated request now checks its session in Redis (`EXISTS`; twice for handlers that read the token again).

## [0.24.4] - 2026-11-08

First step of the JWT hardening. No API change is required from the UI.

### Security

- Registration and "forgot password" links (`reset_credentials` tokens) can now be used only once. Before, the "single use" check deleted a Redis key that never existed, so a registration link (valid 7 days) could reset the password and username any number of times. Each link token now has a Redis key with the link's TTL, consumed atomically before the password is changed, and restored if the change fails (e.g. the chosen username is taken) so the member can try again. Links sent before this version are accepted once more after the upgrade, even if they were already used (at most until they expire: 7 days for registration links).
- Link tokens (registration, password reset, participation) no longer create a refresh token. Each one used to leave a Redis key with no expiry.
- `createToken` returned `(nil, nil)` when Redis failed, which crashed the request, or the whole process when called from the scheduler. It now returns the error.
- The `Authorization` header must use the `Bearer` scheme (case-insensitive). `Bearer: <token>`, which the UI currently sends, is still accepted, as are extra spaces after the scheme. Other schemes such as `Basic <token>` used to be accepted and are now rejected.

### Fixed

- `POST /logout` crashed (nil error dereference) and never succeeded. It now revokes the refresh token linked to the access token it receives and returns 202. It also works with an expired access token and when the session was already revoked, and returns 401 for a missing or invalid token. Access tokens, and refresh tokens issued by earlier `/refresh` calls of the same session, stay valid until they expire: refresh rotation and session revocation are planned next.

### Operator actions

- Link tokens issued before this version left Redis keys with no expiry. Once 7 days have passed after the upgrade (the longest link TTL), they can be removed with a one-off scan for keys whose `TTL` is `-1`; sessions always have a TTL.

### Added

- Responses for an expired access token (403) now include `"code": "token_expired"`, so the UI can tell them apart from other 403 responses before the status codes are fixed.

## [0.24.3] - 2026-11-07

### Security

- `memberRegistration` notifications no longer store the member and the author (names, email, phone, height, weight, roles…) in clear text in `notifications.payload`. They now store only `{"memberUuid", "authorUuid"}`, and the scheduler loads both members when it sends the email. A member who was deleted, or who no longer has an email, is not emailed. If the author was deleted, the email is sent without their name.
- Migration `sql/0.24.3.sql` rewrites existing `memberRegistration` notifications to keep only these UUIDs. Migrations now run with SQLite `secure_delete`, so the old content is zeroed in the database file, and the database is `VACUUM`ed afterwards (best effort).
- Logs no longer contain personal data or tokens:
  - member and notification log lines show UUIDs instead of the whole record. `Member` and `Notification` now implement `String()` so formatting them with `%v` never prints personal data or payloads;
  - emails (including the one typed on "forgot password"), height and weight values are no longer logged;
  - email addresses are redacted from SMTP errors;
  - the `Authorization` header (access token) is no longer logged at debug level.
- Production log level is now `info` instead of `debug` (`deployment/fly.io/fly.toml`).

### Operator actions

- Litestream backups (90-day retention) still hold the old clear-text notifications. To remove them now, delete the old generations from the replica bucket or recreate the replica.
- Logs already collected by Fly (and any log drain or export) may contain personal data and access tokens. Access tokens expire after 15 minutes. Rotate `APP_ACCESS_SECRET` and `APP_REFRESH_SECRET` if a log export may exist.
- Do not roll back to 0.24.2 once this version has run: 0.24.2 cannot read the new notification payload and would fail to send pending registration emails.

## [0.24.2] - 2026-11-06

### Security

- Bump Go to 1.26.10 (`go.mod`), fixing standard library vulnerabilities in `net/http`, `net/url`, `crypto/x509`, `crypto/tls` and others reported by `govulncheck`.
- Replace `github.com/golang-jwt/jwt` v3 (unmaintained, GO-2025-3553: excessive memory allocation while parsing headers) with `github.com/golang-jwt/jwt/v5` 5.3.1. Token verification now only accepts `HS256` and rejects tokens without an `exp` claim; HS384/HS512 tokens and tokens without `exp` used to be accepted (this API never issued such tokens).
- Bump `github.com/redis/go-redis/v9` and `redisotel` 9.3.1 → 9.23.0 (GO-2025-3540).
- Bump OpenTelemetry (`otel`, `otel/sdk`, `otlptracegrpc`) to 1.47.0 (GO-2026-6505) and `otelmux` 0.46.1 → 0.72.0. Span names are kept as the route template; HTTP span attributes now follow the stable semantic conventions (e.g. `http.request.method` instead of `http.method`).
- Bump `golang.org/x/net` 0.59.0 → 0.62.0, `golang.org/x/crypto` 0.57.0 → 0.58.0, `golang.org/x/text` 0.42.0 → 0.43.0.

### Fixed

- Detect expired access tokens with `errors.Is(err, jwt.ErrTokenExpired)` (new `controller.IsTokenExpired`) instead of comparing the error text, which changed in jwt v5. Expired tokens keep returning 403, which the UI uses to trigger `/refresh`.

### Added

- Tests for `/api/v1/refresh` and access token validation, including expired, unsigned (`alg: none`), wrongly signed and `exp`-less tokens, and a regression test that an expired access token returns 403.

## [0.24.1] - 2026-09-30

### Security

- Bump Go to 1.26 (`go.mod`, `Dockerfile`, `Dockerfile.dev`).
- Bump `golang.org/x/net` 0.19.0 → 0.59.0, `golang.org/x/crypto` 0.17.0 → 0.57.0, `golang.org/x/text` 0.14.0 → 0.42.0, `golang.org/x/sys` 0.15.0 → 0.48.0.
- Bump `google.golang.org/grpc` 1.60.1 → 1.84.0, `google.golang.org/protobuf` 1.31.0 → 1.36.12 and `genproto`; this pulls OpenTelemetry core (`otel`, `otel/sdk`, `otel/trace`, `otel/metric`) from 1.21.0 to 1.44.0.

## [0.24.0] - 2026-09-30

### Added

- New badge `castanyada2026` (Castanyada 2026) in the `events` series, seeded via migration `sql/0.24.0.sql`, awarded for taking part in the 2026 Castanyada.

## [0.23.0] - 2026-07-19

### Added

- Admin option `notifyByEmail` on `POST /api/v1/badges/{badge_uuid}/members` to queue a `badgeAwarded` notification for newly awarded members who are subscribed to emails. The email congratulates them and links to `/myBadges`.
- Event field `uniformRequired` (migration `sql/0.23.0.sql`). When set, automatic and manual reminder emails include uniform guidance (official shirt and white trousers, similar colours if needed, second-hand shirts available to borrow).

## [0.22.1] - 2026-07-06

### Changed

- CI: migrated the GitHub Actions workflows off the deprecated Node 20 actions. `job_check_version` now uses `actions/checkout@v5`, detects a `VERSION` change with native git (replacing the archived `technote-space/get-diff-action`), and cancels the run via the GitHub CLI (`gh run cancel`) plus fails the job when `VERSION` was not bumped (replacing `andymckay/cancel-action`). Bumped `actions/checkout` to `v5` in the test, docker and deploy workflows.

## [0.22.0] - 2026-07-06

### Added

- New badge series `events` (Évènements) seeded via migration `sql/0.22.0.sql`, with its first badge `mcc2026` (Montréal Complètement Cirque 2026), awarded for taking part in the castells workshops during the 2026 festival.

## [0.21.0] - 2026-07-04

### Added

- Badges feature: new tables `badge_series`, `badges` and `member_badges` (migration `sql/0.21.0.sql`), seeded with the `welcome` series and its 7 badges (`casal`, `camisa`, `uniformeCasteller`, `motxilla`, `amunt`, `primeraDiada`, `primerCastell`). Badge names and descriptions are not stored; the UI derives i18n keys from the badge `code`.
- Endpoint `GET /api/v1/badges` returning every badge series with its badges (any authenticated member).
- Endpoint `GET /api/v1/members/{member_uuid}/badges` returning the badges a member has unlocked (viewable by any authenticated member).
- Admin endpoint `GET /api/v1/badges/{badge_uuid}/members` returning the UUIDs of the members holding a badge.
- Admin endpoints `POST` and `DELETE /api/v1/badges/{badge_uuid}/members` to grant or revoke a badge for a batch of members (payload `{ "memberUuids": [...] }`). Granting is idempotent.
- Automatic award of the `amunt` badge when a member confirms a participation to an event through the app (any answer). It is granted to the member for whom the participation is recorded when they act for themselves or a parent acts for a dependent, but never when an admin answers on behalf of another member. Awarding is best-effort and idempotent, so it never interferes with recording the participation.

### Changed

- `GET /api/v1/members/{member_uuid}` now lets any authenticated member view any profile. Admins get the full profile, a member sees their own as before, and other members only receive first name, last name and UUID (all other fields are stripped).
- `GET /api/v1/members` now returns a sanitized list (UUID, first name, last name only) to non-admin members so they can browse and search profiles without exposing roles or contact details.

## [0.20.1] - 2026-06-11

### Changed

- Refactored SQL queries in the members, participation, and events models to bind all compared values as query parameters instead of interpolating them into the statement string, hardening against SQL injection regressions.

## [0.20.0] - 2026-05-30

### Added

- Admin endpoint `PUT /api/v1/members/{member_uuid}/status` to manually set a member status to `active` or `paused`.
- Column `members.last_activity_date` tracking the last manual reactivation, so the automatic pause job treats a manual reactivation like a recent participation.

### Changed

- `pauseAbsentMembers` now compares the inactivity delay against the most recent of the last participated event and `last_activity_date`, preventing a manually reactivated member from being paused again right away.

## [0.19.0] - 2026-05-18

### Added

- Admin endpoint `POST /api/v1/events/{event_uuid}/reminders` to queue manual event reminder emails.
- Notification type `manualEventReminder` with audience presets:
  - `default` — same recipients as the automatic two-day reminder (active and paused members with `subscribed = 1`).
  - `no_answer_active` — active members with no participation answer.
  - `no_answer_active_paused` — active and paused members with no participation answer.
  - `members` — explicit list of member UUIDs.
- Shared helper `sendReminderEmailsToMembers` for automatic and manual reminder delivery.

### Changed

- Refactored upcoming-event reminder sending in the scheduler to use the shared delivery helper.
