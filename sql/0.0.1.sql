CREATE TABLE IF NOT EXISTS events
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	startDate INTEGER NOT NULL,
	endDate INTEGER NOT NULL,
	description TEXT,
	uuid TEXT NOT NULL,
	recurringEvent TEXT,
	type TEXT NOT NULL,
	locationName TEXT,
	lat REAL,
	lng REAL,
	CONSTRAINT uuid_unique UNIQUE (uuid),
	FOREIGN KEY(recurringEvent) REFERENCES recurring_events(uuid)
);
CREATE TABLE IF NOT EXISTS members
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
CREATE TABLE IF NOT EXISTS participation
(
	member_uuid INTEGER NOT NULL,
	event_uuid INTEGER NOT NULL,
    answer TEXT,
	presence TEXT,
	PRIMARY KEY (member_uuid, event_uuid)
);
CREATE TABLE IF NOT EXISTS recurring_events
(
	uuid TEXT PRIMARY KEY,
	name TEXT NOT NULL,
    description TEXT,
	interval TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS notifications
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	notificationType TEXT NOT NULL,
	authorUUID TEXT NOT NULL,
	objectUUID TEXT NOT NULL, 
	sendDate INTEGER NOT NULL,
	delivered INTEGER NOT NULL DEFAULT 0
);