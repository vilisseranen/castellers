# Changelog

All notable changes to the Castellers API are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

The API version is defined in [`VERSION`](VERSION) and exposed at `GET /api/v1/version`.

## [0.24.3] - 2026-11-07

### Security

- `memberRegistration` notifications no longer store the member and the author (names, email, phone, height, weight, roles…) in clear text in `notifications.payload`. They now store only `{"memberUuid", "authorUuid"}`, and the scheduler loads both members when it sends the email.
- Migration `sql/0.24.3.sql` rewrites existing `memberRegistration` notifications to keep only these UUIDs. The database is then `VACUUM`ed so the old content is removed from the database file. **Older copies remain in Litestream backups until their retention expires (90 days)**, unless the replica is purged or recreated.
- Logs no longer contain personal data or tokens:
  - member and notification log lines show UUIDs instead of the whole record. `Member` and `Notification` now implement `String()` so formatting them with `%v` never prints personal data or payloads;
  - emails (including the one typed on "forgot password"), height and weight values are no longer logged;
  - the `Authorization` header (access token) is no longer logged at debug level.
- Production log level is now `info` instead of `debug` (`deployment/fly.io/fly.toml`).

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
