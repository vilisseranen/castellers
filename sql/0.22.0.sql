-- Seed: série Évènements (events) et son premier badge, Montréal Complètement
-- Cirque 2026. Les noms et descriptions ne sont pas stockés : l'UI dérive les
-- clés i18n à partir du code (badges.series.<code>.name, badges.items.<code>.*).
INSERT INTO badge_series (uuid, code, position) VALUES
('000000000000000000000000000000005e100002', 'events', 2);

INSERT INTO badges (uuid, series_uuid, code, image, position) VALUES
('00000000000000000000000000000000bad00008', '000000000000000000000000000000005e100002', 'mcc2026', 'mcc2026.png', 1);
