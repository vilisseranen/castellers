# Changelog

All notable changes to the Castellers API are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

The API version is defined in [`VERSION`](VERSION) and exposed at `GET /api/v1/version`.

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
