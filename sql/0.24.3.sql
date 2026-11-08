-- memberRegistration notifications used to store the whole member and author
-- (names, email, phone, height, weight...) in clear text. Keep only their UUIDs,
-- which is what the scheduler now reads.
UPDATE notifications
SET payload = CAST(json_object(
        'memberUuid', json_extract(CAST(payload AS TEXT), '$.member.uuid'),
        'authorUuid', json_extract(CAST(payload AS TEXT), '$.author.uuid')
    ) AS BLOB)
WHERE notificationType = 'memberRegistration'
  AND payload IS NOT NULL
  AND json_valid(CAST(payload AS TEXT))
  AND json_extract(CAST(payload AS TEXT), '$.member') IS NOT NULL;

-- Anything else left in those payloads cannot be read by the scheduler: drop it.
UPDATE notifications
SET payload = NULL
WHERE notificationType = 'memberRegistration'
  AND payload IS NOT NULL
  AND (NOT json_valid(CAST(payload AS TEXT))
       OR json_extract(CAST(payload AS TEXT), '$.memberUuid') IS NULL);
