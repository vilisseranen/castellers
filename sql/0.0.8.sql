
CREATE TEMPORARY TABLE notifications_backup
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	notificationType TEXT NOT NULL,
	objectUUID TEXT NOT NULL, 
	sendDate INTEGER NOT NULL,
	delivered INTEGER NOT NULL DEFAULT 0,
    payload BLOB
);
INSERT INTO notifications_backup SELECT id,notificationType,objectUUID,sendDate,delivered,payload FROM notifications;
DROP TABLE notifications;
CREATE TABLE notifications
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	notificationType TEXT NOT NULL,
	objectUUID TEXT NOT NULL, 
	sendDate INTEGER NOT NULL,
	delivered INTEGER NOT NULL DEFAULT 0,
    payload BLOB
);
INSERT INTO notifications SELECT id,notificationType,objectUUID,sendDate,delivered,payload FROM notifications_backup;
DROP TABLE notifications_backup;
