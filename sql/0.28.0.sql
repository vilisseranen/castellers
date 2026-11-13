-- Dues, second step.

-- An event at which dues are collected points to the collection.
ALTER TABLE events ADD COLUMN dues_collection_uuid TEXT;

-- Settings of the dues, shown to the members: for now the address to which
-- Interac transfers are sent.
CREATE TABLE IF NOT EXISTS dues_settings
(
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
