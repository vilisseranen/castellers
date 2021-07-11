/* Fix issue on notification */

CREATE TEMPORARY TABLE notifications_backup
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	notificationType TEXT NOT NULL,
	objectUUID TEXT NOT NULL, 
	sendDate INTEGER NOT NULL,
	delivered INTEGER NOT NULL DEFAULT 0,
    payload BLOB
);
INSERT INTO notifications_backup SELECT * FROM notifications;
DROP TABLE notifications;
CREATE TABLE notifications
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	notificationType TEXT NOT NULL,
	objectUUID TEXT, 
	sendDate INTEGER NOT NULL,
	delivered INTEGER NOT NULL DEFAULT 0,
    payload BLOB
);
INSERT INTO notifications SELECT * FROM notifications_backup;
DROP TABLE notifications_backup;


/* Create tables */
CREATE TABLE IF NOT EXISTS castell_types (name PRIMARY KEY);
CREATE TABLE IF NOT EXISTS castell_positions (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, column INTEGER NOT NULL, cordon INTEGER NOT NULL, part TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS castell_positions_in_castells (castell_type_name TEXT, castell_positions_id INTEGER, FOREIGN KEY(castell_type_name) REFERENCES castell_types(name), FOREIGN KEY(castell_positions_id) REFERENCES castell_positions(id), PRIMARY KEY(castell_type_name,castell_positions_id));
CREATE VIEW IF NOT EXISTS castell_types_view AS SELECT castell_types.name AS castell_name, castell_positions.id AS position_id, castell_positions.name AS position_name, castell_positions.column AS position_column, castell_positions.cordon AS position_cordon, castell_positions.part AS position_part FROM castell_positions INNER JOIN castell_positions_in_castells ON castell_positions.id = castell_positions_in_castells.castell_positions_id INNER JOIN castell_types ON castell_positions_in_castells.castell_type_name = castell_types.name;

/* Add positions in castells */
/*
| Castell | Column description                  | Column number |
| ------- | ----------------------------------- | ------------- |
| 2       | Column where the enxaneta climbs    | 1             |
| 2       | Column where the acotxador climbs   | 2             |
|         |                                     |               |
| 3       | Rengla                           	| 1             |
| 3       | Plena                               | 2             |
| 3       | Buida                               | 3             |
*/

INSERT INTO castell_positions(name, cordon, column, part) VALUES("enxaneta", 3, 1, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("acotxador", 2, 1, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("dos", 1, 1, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("dos", 1, 2, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("baix", 1, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("baix", 1, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("baix", 1, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("baix", 1, 4, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("segon", 2, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("segon", 2, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("segon", 2, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("segon", 2, 4, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("terç", 3, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("terç", 3, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("terç", 3, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("terç", 3, 4, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quart", 4, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quart", 4, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quart", 4, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quart", 4, 4, "tronc");

/* Add castells */
INSERT INTO castell_types(name) VALUES('1d4');
INSERT INTO castell_types(name) VALUES('1d5');
INSERT INTO castell_types(name) VALUES('2d5');
INSERT INTO castell_types(name) VALUES('2d6');
INSERT INTO castell_types(name) VALUES('3d5');
INSERT INTO castell_types(name) VALUES('3d6');
INSERT INTO castell_types(name) VALUES('4d5');
INSERT INTO castell_types(name) VALUES('4d6');


/* Build castells with positions */
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "1d4",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 4 AND column <= 1);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "1d5",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 1);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d5",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 2 AND column <= 2) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d6",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 2) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d5",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 2 AND column <= 3) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d6",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 3) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d5",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 2 AND column <= 4) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d6",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 4) OR (part = "pom");

/* Create castells with people */
CREATE TABLE IF NOT EXISTS castell_models(id INTEGER PRIMARY KEY AUTOINCREMENT, uuid TEXT NOT NULL, name TEXT NOT NULL, castell_type_name TEXT, deleted INTEGER NOT NULL DEFAULT 0,FOREIGN KEY(castell_type_name) REFERENCES castell_types(name), CONSTRAINT uuid_unique UNIQUE (uuid));
CREATE TABLE IF NOT EXISTS castell_members_positions(castell_model_id INTEGER NOT NULL, castell_position_id INTEGER NOT NULL, member_id INTEGER NOT NULL, FOREIGN KEY(castell_model_id) REFERENCES castell_models(id), FOREIGN KEY(castell_position_id) REFERENCES castell_positions(id), FOREIGN KEY(member_id) REFERENCES members(id), PRIMARY KEY(castell_model_id, castell_position_id, member_id));
CREATE VIEW IF NOT EXISTS castell_models_view AS SELECT model.uuid AS model_uuid, model.name AS model_name, model.castell_type_name AS model_type, model.deleted AS model_deleted, position_in_castell.name AS position_in_castell_name, position_in_castell.column AS position_in_castell_column, position_in_castell.cordon AS position_in_castell_cordon, position_in_castell.part AS position_in_castell_part, members.uuid AS member_uuid FROM castell_models AS model INNER JOIN castell_members_positions AS position ON position.castell_model_id = model.id INNER JOIN castell_positions AS position_in_castell ON position.castell_position_id = position_in_castell.id INNER JOIN members ON position.member_id = members.id;