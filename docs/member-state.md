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
| `MEMBERSSTATUSACTIVATED`  | `active`  | Actif             | yes                    | `Credentials.ResetCredentials` (first password); direct creation for guest/canalla; reactivation via participation; manual admin reactivation |
| `MEMBERSSTATUSPAUSED`     | `paused`  | En pause          | yes                    | `pauseAbsentMembers` scheduler task; manual admin pause                       |
| `MEMBERSSTATUSDELETED`    | `deleted` | Supprimé          | only when asked for (`status=deleted`) | `Member.DeleteMember` (soft delete)                                           |
| `MEMBERSSTATUSPURGED`     | `purged`  | Ancien membre     | only when asked for (`status=purged`)  | `purgeDeletedMembers` scheduler task, 90 days after the deletion              |

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
- Set manually by an admin via `PUT /members/{uuid}/status` with
  `{"status":"active"}` (`controller.SetMemberStatus`). This goes through
  `Member.SetStatusManual`, which also stamps `last_activity_date = now`,
  so the inactivity scan treats the manual reactivation like a recent
  participation and will not pause the member again right away.
- Receives all reminder and summary emails (subject to the `subscribed`
  flag).

### `paused`

- The member is temporarily inactive but **still visible**.
- Set automatically by the scheduler task `pauseAbsentMembers`
  ([`controller/scheduler.go`](../controller/scheduler.go)) when the most
  recent of (last participated event, `last_activity_date`) is older than
  the configuration value `inactive_delay_days`.
- Set manually by an admin via `PUT /members/{uuid}/status` with
  `{"status":"paused"}`. A manual pause is stable: the scheduler never
  reactivates a member; only a participation (yes/present) or a manual
  admin reactivation does.
- Audience handling for manual reminders is defined in
  [`controller/reminders.go:118-146`](../controller/reminders.go):
  - default audience: `active` **and** `paused`
  - `noAnswerActive`: only `active`
  - `noAnswerActivePaused`: `active` **and** `paused`
- Automatically reverts to `active` on the next *yes* answer / *present*
  participation (see above).

### `deleted`

- Soft delete. The row and all its data are kept for `purge_delay_days`
  (90 days by default), so that an admin can restore the member.
- Set by `Member.DeleteMember`, which also stamps `deleted_at`.
- `POST /members/{uuid}/restore` (admins, `controller.RestoreMember`) brings
  the member back: `active` if they had chosen a password or are a guest or a
  canalla, `created` otherwise. It is refused with a 409 if their email was
  given to another member in the meantime.
- Not visible in `Member.Get`, nor in `Member.GetAll` unless `deleted` is in
  the list of statuses asked for. The list then gives `deletedAt` and
  `purgeAt`. The UI asks for them to keep writing the names in the past
  castells and to let the admins restore somebody.
- Never a recipient of an email.

### `purged`

- Set by the scheduler task `purgeDeletedMembers`
  ([`controller/scheduler.go`](../controller/scheduler.go)), every hour, for
  the members deleted more than `purge_delay_days` ago. There is no way back.
- `Member.PurgeMember` erases the name, email, contact, height, weight, roles,
  notes, shirt, musician flag and subscription, and removes the credentials,
  the links with dependents and the notes written about the dues of the member.
- The row stays, with its UUID, its type and its dates: participations,
  positions in castells, dues and badges still point to it, so the past
  events and castells still count somebody. The UI shows "Ancien membre".
- Not visible in `Member.Get`, nor in `Member.GetAll` unless `purged` is in
  the list of statuses asked for; all the personal fields are then empty.
- The backups keep the data of a purged member until they expire.

### Automatic deletion of inactive members

The scheduler task `deleteInactiveMembers`
([`controller/inactivity.go`](../controller/inactivity.go)) runs every hour,
unless `auto_delete_enabled` is false.

- **Who is due** (`controller.DeletionDue`): a `created` account that never
  answered nor came to an event, `pending_delete_delay_days` (180) after its
  creation; a `paused` member, or a `created` one who came anyway,
  `inactive_delete_delay_days` (730) after their last sign of life, the most
  recent of: the last event they answered or came to, `last_activity_date`
  and `created_at`. Never an `active` member, never an admin.
- **Notice**: `deletion_notice_at` is stamped and the member is told by email
  (`mail.SendDeletionNoticeEmail`); for a member without an email, the members
  responsible for them are told. A `created` account gets no email
  (`controller.ToldOfDeletion`): its owner never signed up. `deletionAt` in
  the list of members is the day of the deletion.
- **Deletion**: `deletion_notice_days` (30) after the notice, if the member is
  still due, `Member.DeleteMember`. From there, see `deleted` and `purged`.
- **Cancelled** when the member is not due anymore: they answered or came to
  an event, logged in (`Member.KeepMemberIfNoticed`), chose a password, or an
  admin reactivated them or kept them (`POST /members/{uuid}/keep`,
  `Member.KeepMember`, which stamps `last_activity_date`).
- **Admins**: `sendDeletionDigest`, every Monday at 9:00 in Montréal, sends
  one email to the subscribed admins with the members whose deletion was
  planned and the deleted members purged within `deletion_notice_days`.
  `admin_notified_at` remembers who was announced, so each one is announced
  once.

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
                  pauseAbsentMembers (cron) / admin pause │
                       active ─────────────────────────► paused
                                                           │
              participation yes/present / admin reactivate │
                       active ◄───────────────────────── paused
                                                           │
                       any state ──DeleteMember──► deleted │
                                                           │
              created / active ◄──admin restore── deleted  │
                                                           │
              deleted ──purgeDeletedMembers (90 days)──► purged
```

Manual admin transitions go through `PUT /members/{uuid}/status`
(`controller.SetMemberStatus`, admin-only, `active` <-> `paused` only).
A manual reactivation stamps `last_activity_date`, which
`pauseAbsentMembers` treats as a recent activity (counter reset).

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
| **Daily event summary** — recipients               | **none** → `GetAll([], [])` → effectively `status NOT IN ('deleted','purged')` | body filters to `admin && subscribed==1` | `controller/scheduler.go:130-170`                |
| **Daily event summary** — printed list             | `active` **OR** `Participation != ""` (any RSVP) | none             | `controller/scheduler.go:159-167`                |
| Default reminder audience                          | `active` + `paused`                            | none               | `controller/reminders.go:120-122`                |
| `ManualReminderAudienceNoAnswerActive`             | `active`                                       | none               | `controller/reminders.go:123-124`                |
| `ManualReminderAudienceNoAnswerActivePaused`       | `active` + `paused`                            | none               | `controller/reminders.go:125-126`                |
| `ManualReminderAudienceMembers` (explicit UUIDs)   | exclude `deleted`, `purged`                    | n/a                | `controller/reminders.go:127-142`                |
| Event deleted / modified mass email                | `active` + `paused`, `subscribed==1`           | none               | `controller/scheduler.go:215-289`                |
| Pause scan (`pauseAbsentMembers`)                  | `active` only                                  | none               | `controller/scheduler.go:385-405`                |

### Note on the rehearsals of the musicians

For an event of type `music`, every audience above is narrowed down to the
members whose `musician` flag is 1 (see `EventAudience` in
`controller/events.go`): creation, change, deletion, automatic reminder and
manual reminders. The two exceptions are a manual reminder to explicit UUIDs,
which goes to all the members picked, and an event that changes type from or
to `music`, about which everybody is told. In the printed list of the
summary, `active` becomes `active` and musician.

### Note on the day-before summary email

The summary email is **sent to** every subscribed admin (regardless of
status, as long as the row is not `deleted` or `purged`). It **lists in
its body** only members who are either `active` **or** who have given a
participation answer (`Participation != ""`, meaning they answered
`yes`, `no` or `maybe`).

The recipient list is built from:

```go
m := model.Member{}
members, err := m.GetAll(ctx, []string{}, []string{})
```

in `controller/scheduler.go:130-131`. With empty filter slices, `GetAll`
only applies its hard-coded `status NOT IN ('deleted', 'purged')` clause
([`model/members.go:199`](../model/members.go)).

Each member's participation is then fetched, and a separate slice
`participantsForEmail` is built (see `controller/scheduler.go:159-167`)
by keeping only members where:

```go
member.Status == model.MEMBERSSTATUSACTIVATED || member.Participation != ""
```

That filtered slice is the one passed as `Participants` in the email
payload, while the unfiltered slice continues to drive the recipient
iteration. The intent is to keep the body actionable: long-paused
members who never answered, and guests/canallas who were never invited
to RSVP, are omitted; but anyone who explicitly replied — even with
"no" or "maybe" — remains visible.
