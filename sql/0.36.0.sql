-- A deleted member is kept for a while (90 days by default), so that an admin
-- can restore them. Their personal data is then erased: they become "purged".
-- deleted_at is when the member was deleted (Unix seconds), NULL otherwise.
ALTER TABLE members ADD COLUMN deleted_at INTEGER;

-- We do not know when the members already deleted were deleted: the delay
-- starts now for them.
UPDATE members SET deleted_at = CAST(strftime('%s', 'now') AS INTEGER) WHERE status = 'deleted';
