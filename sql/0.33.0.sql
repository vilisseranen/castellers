-- Musicians of the colla: a member can be a musician whatever their type.
-- The rehearsals of the musicians are events of type 'music'; the type of an
-- event is free text, so the events table does not change.
ALTER TABLE members ADD COLUMN musician INTEGER NOT NULL DEFAULT 0;
