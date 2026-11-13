-- Date at which a member was added (Unix seconds), 0 when it is not known.
ALTER TABLE members ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;

-- Best effort for the members who are already there: the registration email
-- is queued when a member is added, so the oldest one gives the date. Members
-- added without an email keep 0.
UPDATE members
SET created_at = COALESCE((
	SELECT MIN(n.sendDate)
	FROM notifications n
	WHERE n.notificationType = 'memberRegistration'
	  AND n.sendDate > 0
	  AND (
		n.objectUUID = members.uuid
		OR (
			n.payload IS NOT NULL
			AND json_valid(CAST(n.payload AS TEXT))
			AND json_extract(CAST(n.payload AS TEXT), '$.memberUuid') = members.uuid
		)
	  )
), 0);
