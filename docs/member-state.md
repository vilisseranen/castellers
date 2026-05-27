# Member types and statuses (backend)

This document describes the **type** and **status** dimensions of a member in
the `castellers` backend, where they are defined, and how the application
reacts to each value.

All members are stored as rows in the `members` SQL table and represented in
Go by the single struct [`model.Member`](../model/members.go). There is no
separate `Guest`, `Canalla` or `Admin` struct — the kind of member is purely
encoded in the `Type` column. Whether the account is enabled, paused or
deleted is encoded in the orthogonal `Status` column.

## Where the constants live

Both sets of constants are declared in [`model/members.go`](../model/members.go):

```go
const (
    MEMBERSTYPEADMIN   = "admin"
    MEMBERSTYPEREGULAR = "member"
    MEMBERSTYPECANALLA = "canalla"
    MEMBERSTYPEGUEST   = "guest"

    MEMBERSSTATUSCREATED   = "created"
    MEMBERSSTATUSACTIVATED = "active"
    MEMBERSSTATUSPAUSED    = "paused"
    MEMBERSSTATUSDELETED   = "deleted"
    MEMBERSSTATUSPURGED    = "purged"
)
```

The SQL column `status` has `DEFAULT "created"` (see
[`sql/0.17.0.sql`](../sql/0.17.0.sql)). The list of accepted types is
validated by `model.ValidateType` in
[`model/validation.go`](../model/validation.go).

Valid values for filtering queries on the HTTP API are whitelisted in
[`controller/members.go`](../controller/members.go) (functions
`memberStatusListFromQuery` and `memberTypeListFromQuery`).

---

## Types

The `Type` field describes **what kind of person** the row represents. There
are four valid values.

| Constant              | Value     | French label (UI)   | Can log in | Has email | Receives registration email | Initial status        |
|-----------------------|-----------|---------------------|------------|-----------|-----------------------------|-----------------------|
| `MEMBERSTYPEADMIN`    | `admin`   | Administrateur      | yes        | yes       | yes                         | `created` → `active`  |
| `MEMBERSTYPEREGULAR`  | `member`  | Membre régulier     | yes        | yes       | yes                         | `created` → `active`  |
| `MEMBERSTYPEGUEST`    | `guest`   | Invité              | no         | no (forced to `""`) | no                | `active` (direct)     |
| `MEMBERSTYPECANALLA`  | `canalla` | Canalla (children)  | no         | no (forced to `""`) | no                | `active` (direct)     |

### `admin`

- Full HTTP API access via the `MEMBERSTYPEADMIN` token check in
  [`routes/api_v1.go`](../routes/api_v1.go) (create events, manage members,
  send reminders, etc.).
- Login grants permissions `["member", "admin"]` — see
  `getMemberPermissions` in [`controller/login.go`](../controller/login.go).
- Only `admin` subscribers receive the daily **summary email** the day
  before an event (`scheduler.go:160-169`).

### `member` (regular)

- Standard authenticated user. Can edit their own profile, answer events,
  and is the audience targeted by all reminder/notification emails.
- Login grants permissions `["member"]`.

### `guest`

- Used for occasional participants who are not part of the colla.
- `CreateMember` forces `Email = ""` for guests and skips the email
  uniqueness check (`controller/members.go:117-125`).
- Required fields exclude `email` (`missingRequiredFields`,
  `controller/members.go:405-409`).
- Created directly with `Status = active`; no welcome email is sent
  (`controller/members.go:172-178`).
- `SendRegistrationEmail` refuses to send to guests
  (`controller/members.go:373-376`).
- A guest can be **promoted** to `member`; on promotion the status is
  reset to `created` so the welcome email is triggered
  (`controller/members.go:296-305`).
- Asymmetry: a regular member **cannot** be demoted back to `guest`
  (`controller/members.go:229-233`).
- Cannot log in — `getMemberPermissions` returns `ERRORGUESTCANNOTLOGIN`
  for any type other than `admin` or `member`.

### `canalla`

- "Canalla" is the Catalan term for the children who climb to the top of
  a castell.
- Behaviour is **identical to `guest`** in every condition (the code
  systematically checks `Type == GUEST || Type == CANALLA` at lines 117,
  172, 298, 373 and 405 of `controller/members.go`).
- One small difference: when promoting `guest` → other type, the status
  reset to `created` is **skipped** if the target is `canalla` (see the
  comment at `controller/members.go:296-298`: *"Does not apply to
  canalla, they will stay activated and won't receive the welcome
  email"*).

---

## Statuses

The `Status` field describes **the lifecycle stage** of the account. It is
orthogonal to `Type`.

| Constant                  | Value     | French label (UI) | Visible in API queries | Triggered by                                                                 |
|---------------------------|-----------|-------------------|------------------------|-------------------------------------------------------------------------------|
| `MEMBERSSTATUSCREATED`    | `created` | Créé              | yes                    | SQL default; reset on guest → member promotion                                |
| `MEMBERSSTATUSACTIVATED`  | `active`  | Actif             | yes                    | `Credentials.ResetCredentials` (first password); direct creation for guest/canalla; reactivation via participation |
| `MEMBERSSTATUSPAUSED`     | `paused`  | En pause          | yes                    | `pauseAbsentMembers` scheduler task                                           |
| `MEMBERSSTATUSDELETED`    | `deleted` | Supprimé          | **no** (filtered out)  | `Member.DeleteMember` (soft delete)                                           |
| `MEMBERSSTATUSPURGED`     | `purged`  | (no UI label)     | **no** (filtered out)  | external operation only (no Go writer)                                       |

### `created`

- SQL default value on insert
  ([`sql/0.17.0.sql:36`](../sql/0.17.0.sql): `status TEXT NOT NULL DEFAULT "created"`).
- Indicates the account exists but the user has not yet set a password.
- `CreateMember` does not assign a status explicitly for `admin`/`member`,
  relying on the default.
- Re-applied when a `guest` is promoted to `member` so a fresh welcome
  email is triggered (`controller/members.go:296-305`).
- Transitions to `active` via `Credentials.ResetCredentials`
  ([`model/members.go:271-298`](../model/members.go)).

### `active`

- The normal, fully-enabled state.
- Set automatically:
  - by `Credentials.ResetCredentials` (`model/members.go:275`) when the
    user defines their password,
  - on creation of a `guest` or `canalla`
    (`controller/members.go:172-178`),
  - when a `paused` member answers *yes* or is marked *present* on an
    event within the inactivity window
    (`controller/participation.go:109-117` and `177-185`).
- Receives all reminder and summary emails (subject to the `subscribed`
  flag).

### `paused`

- The member is temporarily inactive but **still visible**.
- Set automatically by the scheduler task `pauseAbsentMembers`
  ([`controller/scheduler.go:385-405`](../controller/scheduler.go)) when
  the last participation date is older than the configuration value
  `inactive_delay_days`.
- Audience handling for manual reminders is defined in
  [`controller/reminders.go:118-146`](../controller/reminders.go):
  - default audience: `active` **and** `paused`
  - `noAnswerActive`: only `active`
  - `noAnswerActivePaused`: `active` **and** `paused`
- Automatically reverts to `active` on the next *yes* answer / *present*
  participation (see above).

### `deleted`

- Soft delete only. The row is preserved (history of participations).
- Set by `Member.DeleteMember`
  ([`model/members.go:237-250`](../model/members.go)) via
  `UPDATE … SET status='deleted'`.
- Filtered out of:
  - `Member.Get` (line 144: `AND status != 'deleted'`),
  - `Member.GetAll` (line 199: `status NOT IN ('deleted', 'purged')`),
  - manual reminder recipients
    (`controller/reminders.go:137-139`).

### `purged`

- Equivalent to `deleted` in every read path
  (`Member.GetAll` filter at line 199; manual reminder filter at
  `controller/reminders.go:137-139`).
- Intended for GDPR-style data purges (personal data scrubbed).
- **No Go code writes this status**: it is reserved for an external
  operation (admin tool, SQL script, future feature). It is, however,
  accepted as a valid filter value in
  `controller/members.go:476-489`.

---

## Lifecycle diagram

```
                 (CreateMember)
                       │
                       ▼
   admin / member ─► created ──ResetCredentials──► active ─┐
                                                           │
   guest / canalla ─► active  ◄──participation yes/present─┤
                                                           │
                              pauseAbsentMembers (cron)    │
                       active ─────────────────────────► paused
                                                           │
                              participation yes/present    │
                       active ◄───────────────────────── paused
                                                           │
                       any state ──DeleteMember──► deleted │
                                                           │
                       any state ──(external op)──► purged │
```

Type-specific transition:

```
   guest ──EditMember(type=member|admin)──► (created)  [welcome email]
   guest ──EditMember(type=canalla)──────► (status unchanged)
   member ──EditMember(type=guest)───────► REJECTED (400)
```

---

## Cross-reference: who is included in each email/audience

This is useful for understanding the visibility of each status.

| Email / audience                                   | Status filter                                  | Type filter        | Code reference                                  |
|----------------------------------------------------|------------------------------------------------|--------------------|--------------------------------------------------|
| **Daily event summary** (sent to admins)           | **none** → `GetAll([], [])` → effectively `status NOT IN ('deleted','purged')` | none — body filters to `admin && subscribed==1` | `controller/scheduler.go:130-170`                |
| Default reminder audience                          | `active` + `paused`                            | none               | `controller/reminders.go:120-122`                |
| `ManualReminderAudienceNoAnswerActive`             | `active`                                       | none               | `controller/reminders.go:123-124`                |
| `ManualReminderAudienceNoAnswerActivePaused`       | `active` + `paused`                            | none               | `controller/reminders.go:125-126`                |
| `ManualReminderAudienceMembers` (explicit UUIDs)   | exclude `deleted`, `purged`                    | n/a                | `controller/reminders.go:127-142`                |
| Event deleted / modified mass email                | `active` + `paused`, `subscribed==1`           | none               | `controller/scheduler.go:215-289`                |
| Pause scan (`pauseAbsentMembers`)                  | `active` only                                  | none               | `controller/scheduler.go:385-405`                |

### Note on the day-before summary email

The summary email lists **every member** except `deleted` and `purged` ones,
including members that have been `paused` for years and including `guest`
and `canalla` rows. The relevant call is:

```go
m := model.Member{}
members, err := m.GetAll(ctx, []string{}, []string{})
```

in `controller/scheduler.go:130-131`. With empty filter slices, `GetAll`
only applies its hard-coded `status NOT IN ('deleted', 'purged')` clause
([`model/members.go:199`](../model/members.go)). The filtered list is
then used both as the recipient list (further narrowed to
`Type == admin && Subscribed == 1` at line 161) **and** as the
`Participants` array embedded in the email body.
