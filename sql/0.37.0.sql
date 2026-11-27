-- Members who have not come for a long time, or who never activated their
-- account, are deleted automatically after a notice.
-- deletion_notice_at is when the member was told (Unix seconds), NULL when no
-- deletion is planned. The deletion happens 30 days later by default.
ALTER TABLE members ADD COLUMN deletion_notice_at INTEGER;

-- admin_notified_at is when the admins were told about the planned deletion
-- of the member, or about the coming purge of a deleted member. NULL when they
-- were not told yet.
ALTER TABLE members ADD COLUMN admin_notified_at INTEGER;
