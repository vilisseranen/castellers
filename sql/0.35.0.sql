-- Shirt of the colla: not every member has one. For those who do, the admins
-- note its size (an even number from 28 to 44) and the day they got it
-- (YYYY-MM-DD). Both are NULL when they are not known.
ALTER TABLE members ADD COLUMN shirt_size INTEGER;
ALTER TABLE members ADD COLUMN shirt_since TEXT;
