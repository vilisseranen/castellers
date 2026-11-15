-- A template for the summer practices, which take place outdoors: copied from
-- the practice of June 25th 2026 at parc La Fontaine.
--
-- The event is the practice that started on that day in Montreal
-- (2026-06-25 00:00 to 2026-06-26 00:00, UTC-4). If there are several, the one
-- whose location mentions La Fontaine comes first. Nothing is created when the
-- event is not found or when a template already has this title. Admins can
-- edit or remove the template.
INSERT INTO event_templates (uuid, title, name, type, description, location_name, lat, lng, start_date, end_date, uniform_required)
SELECT
	lower(hex(randomblob(16))),
	'Entraînement d''été',
	name,
	type,
	COALESCE(description, ''),
	COALESCE(locationName, ''),
	COALESCE(lat, 0),
	COALESCE(lng, 0),
	startDate,
	endDate,
	COALESCE(uniformRequired, 0)
FROM events
WHERE type = 'practice'
  AND deleted = 0
  AND startDate >= 1782360000
  AND startDate < 1782446400
  AND NOT EXISTS (SELECT 1 FROM event_templates WHERE title = 'Entraînement d''été' COLLATE NOCASE)
ORDER BY (COALESCE(locationName, '') NOT LIKE '%fontaine%'), startDate
LIMIT 1;
