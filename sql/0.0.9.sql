/* Create tables */
CREATE TABLE IF NOT EXISTS castell_types (name PRIMARY KEY);
CREATE TABLE IF NOT EXISTS castell_positions (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, column INTEGER NOT NULL, cordon INTEGER NOT NULL, part TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS castell_positions_in_castells (castell_type_name TEXT, castell_positions_id INTEGER, FOREIGN KEY(castell_type_name) REFERENCES castell_types(name), FOREIGN KEY(castell_positions_id) REFERENCES castell_positions(id), PRIMARY KEY(castell_type_name,castell_positions_id));
CREATE VIEW castell_types_view AS SELECT castell_types.name AS castell_name, castell_positions.id AS position_id, castell_positions.name AS position_name, castell_positions.column AS position_column, castell_positions.cordon AS position_cordon, castell_positions.part AS position_part FROM castell_positions INNER JOIN castell_positions_in_castells ON castell_positions.id = castell_positions_in_castells.castell_positions_id INNER JOIN castell_types ON castell_positions_in_castells.castell_type_name = castell_types.name;

/* Add positions in castells */
/*
| Castell | Column description               | Column number |
| ------- | -------------------------------- | ------------- |
| 2       | Column where acotxador climbs    | 1             |
| 2       | Column where the enxaneta climbs | 2             |
|         |                                  |               |
| 3       | Rengla                           | 1             |
| 3       | Plena                            | 2             |
| 3       | Buida                            | 3             |
*/

INSERT INTO castell_positions(name, cordon, column, part) VALUES("enxaneta", 3, 1, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("acotxador", 2, 1, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("dos", 1, 1, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("dos", 1, 2, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("baix", 1, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("baix", 1, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("baix", 1, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("segon", 2, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("segon", 2, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("segon", 2, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("terç", 3, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("terç", 3, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("terç", 3, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quart", 4, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quart", 4, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quart", 4, 3, "tronc");

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
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="1d4"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 1) OR (part = "pom" AND cordon = 3 AND column <= 1);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="1d5"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 4 AND column <= 1) OR (part = "pom" AND cordon = 3 AND column <= 1);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="2d5"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 2 AND column <= 2) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="2d6"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 2) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="3d5"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 2 AND column <= 3) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="3d6"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 3) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="4d5"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 2 AND column <= 4) OR (part = "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT (SELECT name FROM castell_types WHERE name="4d6"),id FROM castell_positions WHERE (part = "tronc" AND cordon <= 3 AND column <= 4) OR (part = "pom");

/* Create castells with people */
CREATE TABLE IF NOT EXISTS castell_models(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, castell_type_name TEXT, FOREIGN KEY(castell_type_name) REFERENCES castell_types(name));
CREATE TABLE IF NOT EXISTS castell_members_positions(castell_model_id INTEGER, castell_position_id INTEGER, member_id INTEGER, FOREIGN KEY(castell_model_id) REFERENCES castell_models(id), FOREIGN KEY(castell_position_id) REFERENCES castell_positions(id), FOREIGN KEY(member_id) REFERENCES members(id), PRIMARY KEY(castell_model_id, castell_position_id, member_id));