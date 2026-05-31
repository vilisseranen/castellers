# Changelog

All notable changes to the Castellers API are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

The API version is defined in [`VERSION`](VERSION) and exposed at `GET /api/v1/version`.

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
