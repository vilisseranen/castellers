CREATE TABLE IF NOT EXISTS members_credentials
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	username TEXT NOT NULL,
	password BLOB NOT NULL,
	FOREIGN KEY(uuid) REFERENCES members(uuid),
	UNIQUE(username)
);