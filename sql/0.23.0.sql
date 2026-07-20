-- Indicate whether the casteller uniform is expected for an event.
ALTER TABLE events ADD COLUMN uniformRequired INTEGER NOT NULL DEFAULT 0;
