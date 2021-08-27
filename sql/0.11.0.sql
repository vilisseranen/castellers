/* Recreate pom using column with the column where they climb */
DELETE FROM castell_positions_in_castells WHERE castell_positions_id IN (SELECT id FROM castell_positions WHERE part = "pom");

/* 1 */
DELETE FROM castell_positions_in_castells WHERE castell_type_name = "1d5";
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "1d5",id FROM castell_positions WHERE (part = "tronc" AND cordon <= 5 AND column <= 1);


/* 2 */
INSERT INTO castell_positions(name, cordon, column, part) VALUES("acotxador", 2, 2, "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d5",id FROM castell_positions WHERE (name = "dos" AND (column = 1 OR column = 2));
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d5",id FROM castell_positions WHERE (name = "acotxador" AND column = 2);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d5",id FROM castell_positions WHERE (name = "enxaneta" AND column = 1);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d6",id FROM castell_positions WHERE (name = "dos" AND (column = 1 OR column = 2));
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d6",id FROM castell_positions WHERE (name = "acotxador" AND column = 2);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "2d6",id FROM castell_positions WHERE (name = "enxaneta" AND column = 1);

/* 3 */
INSERT INTO castell_positions(name, cordon, column, part) VALUES("acotxador", 2, 3, "pom");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("enxaneta", 3, 2, "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d5",id FROM castell_positions WHERE (name = "dos" AND (column = 1 OR column = 2));
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d5",id FROM castell_positions WHERE (name = "acotxador" AND column = 3);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d5",id FROM castell_positions WHERE (name = "enxaneta" AND column = 2);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d6",id FROM castell_positions WHERE (name = "dos" AND (column = 1 OR column = 2));
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d6",id FROM castell_positions WHERE (name = "acotxador" AND column = 3);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "3d6",id FROM castell_positions WHERE (name = "enxaneta" AND column = 2);

/* 4 */
INSERT INTO castell_positions(name, cordon, column, part) VALUES("dos", 1, 4, "pom");
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d5",id FROM castell_positions WHERE (name = "dos" AND (column = 2 OR column = 4));
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d5",id FROM castell_positions WHERE (name = "acotxador" AND column = 3);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d5",id FROM castell_positions WHERE (name = "enxaneta" AND column = 1);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d6",id FROM castell_positions WHERE (name = "dos" AND (column = 2 OR column = 4));
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d6",id FROM castell_positions WHERE (name = "acotxador" AND column = 3);
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "4d6",id FROM castell_positions WHERE (name = "enxaneta" AND column = 1);
