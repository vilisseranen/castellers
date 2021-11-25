CREATE TEMPORARY TABLE members_backup
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL,
    firstName BLOB NOT NULL,
    lastName BLOB NOT NULL,
    height BLOB NOT NULL,
    weight BLOB NOT NULL,
    extra BLOB NOT NULL,
    roles BLOB NOT NULL,
    type BLOB NOT NULL,
    email BLOB NOT NULL,
    contact BLOB NOT NULL,
    code TEXT NOT NULL,
    activated INTEGER NOT NULL DEFAULT 0,
    subscribed INTEGER NOT NULL DEFAULT 0,
    deleted INTEGER NOT NULL DEFAULT 0,
    language TEXT NOT NULL DEFAULT 'fr',
    CONSTRAINT uuid_unique UNIQUE (uuid)
);
INSERT INTO members_backup SELECT * FROM members;
DROP TABLE members;
CREATE TABLE members
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL,
    firstName BLOB NOT NULL,
    lastName BLOB NOT NULL,
    height BLOB NOT NULL,
    weight BLOB NOT NULL,
    extra BLOB NOT NULL,
    roles BLOB NOT NULL,
    type BLOB NOT NULL,
    email BLOB NOT NULL,
    contact BLOB NOT NULL,
    status TEXT NOT NULL DEFAULT "created",
    subscribed INTEGER NOT NULL DEFAULT 0,
    language TEXT NOT NULL DEFAULT 'fr',
    CONSTRAINT uuid_unique UNIQUE (uuid)
);

INSERT INTO members(id, uuid, firstName, lastName, height, weight, extra, roles, type, email, contact, subscribed, language) SELECT id, uuid, firstName, lastName, height, weight, extra, roles, type, email, contact, subscribed, language FROM members_backup;

UPDATE members SET status = 'active' WHERE id IN (SELECT id FROM members_backup WHERE activated = 1);
UPDATE members SET status = 'deleted' WHERE id IN (SELECT id FROM members_backup WHERE deleted = 1);

DROP TABLE members_backup;