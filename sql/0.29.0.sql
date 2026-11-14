-- Event templates: values proposed when an admin creates an event.
-- start_date and end_date are those of a sample event: only the time of day
-- and the duration are used.
CREATE TABLE IF NOT EXISTS event_templates
(
	uuid TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	name TEXT NOT NULL DEFAULT '',
	type TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	location_name TEXT NOT NULL DEFAULT '',
	lat REAL NOT NULL DEFAULT 0,
	lng REAL NOT NULL DEFAULT 0,
	start_date INTEGER NOT NULL DEFAULT 0,
	end_date INTEGER NOT NULL DEFAULT 0,
	uniform_required INTEGER NOT NULL DEFAULT 0
);

-- A first template for the weekly practice, copied from an existing practice:
-- one of a recurring series if there is any, and the closest to today. Nothing
-- is created when there is no practice yet. Admins can edit or remove it.
INSERT INTO event_templates (uuid, title, name, type, description, location_name, lat, lng, start_date, end_date, uniform_required)
SELECT
	lower(hex(randomblob(16))),
	'Entraînement hebdomadaire',
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
WHERE type = 'practice' AND deleted = 0
ORDER BY (recurringEvent IS NULL), ABS(startDate - CAST(strftime('%s', 'now') AS INTEGER))
LIMIT 1;
