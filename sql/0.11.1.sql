INSERT INTO castell_positions(name, cordon, column, part) VALUES("quint", 5, 1, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quint", 5, 2, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quint", 5, 3, "tronc");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("quint", 5, 4, "tronc");

INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "1d5",id FROM castell_positions WHERE (part = "tronc" AND cordon = 5 AND column <= 1);