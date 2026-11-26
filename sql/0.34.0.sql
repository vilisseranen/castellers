-- Cancelled events: an event can be cancelled rather than deleted. It stays in
-- the list of events, shown as cancelled, with the reason why. The reason is
-- one of 'weather', 'technical' or 'other'; the note is free text.
ALTER TABLE events ADD COLUMN cancelled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN cancel_reason TEXT;
ALTER TABLE events ADD COLUMN cancel_note TEXT;
