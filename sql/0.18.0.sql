CREATE TABLE IF NOT EXISTS members_dependent
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	responsible_uuid TEXT NOT NULL,
    dependent_uuid TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (responsible_uuid, dependent_uuid),
	FOREIGN KEY(responsible_uuid) REFERENCES members(uuid),
    FOREIGN KEY(dependent_uuid) REFERENCES members(uuid)
);