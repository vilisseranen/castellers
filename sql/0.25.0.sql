-- Pinya positions, for the pilar de 4 (1d4) only.
--
-- A pinya position uses the same four fields as a tronc or pom position:
--   part   = "pinya"
--   name   = agulla | contrafort | crossa | mans | vent | lateral
--   cordon = 0 for the nucli (agulla, contrafort, crosses),
--            1, 2... for the people in a rengla, counted from the baix outwards
--   column = which rengla of that name, as seen on the plan (front of the baix at the bottom):
--            agulla, contrafort: 1
--            crossa, vent:       1 = left, 2 = right
--            mans:               1 = front (behind the agulla), 2 = back (behind the contrafort)
--            lateral:            1 = front left, 2 = front right, 3 = back right, 4 = back left

-- Nucli
INSERT INTO castell_positions(name, cordon, column, part) VALUES("agulla", 0, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("contrafort", 0, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("crossa", 0, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("crossa", 0, 2, "pinya");

-- 8 rengles, 2 cordons
INSERT INTO castell_positions(name, cordon, column, part) VALUES("mans", 1, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("mans", 2, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("mans", 1, 2, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("mans", 2, 2, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("vent", 1, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("vent", 2, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("vent", 1, 2, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("vent", 2, 2, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 1, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 2, 1, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 1, 2, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 2, 2, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 1, 3, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 2, 3, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 1, 4, "pinya");
INSERT INTO castell_positions(name, cordon, column, part) VALUES("lateral", 2, 4, "pinya");

-- Only the pilar de 4 gets a pinya for now
INSERT INTO castell_positions_in_castells(castell_type_name, castell_positions_id) SELECT "1d4",id FROM castell_positions WHERE part = "pinya";
