-- Dues registry (cotisations).
--
-- The application does not handle payments: an admin only records which dues
-- a member has settled.
--
-- A season follows the membership year of the Cercle culturel catalan: it
-- starts on October 1st of the year `season` and ends on September 30th of the
-- following year. It holds at most four collections (cueillettes):
--   cercle = annual membership of the Cercle, collected in the autumn
--   winter, spring, summer = the three dues of the colla
CREATE TABLE IF NOT EXISTS dues_collections
(
	uuid TEXT PRIMARY KEY,
	season INTEGER NOT NULL,
	kind TEXT NOT NULL,
	-- amount in cents
	amount INTEGER NOT NULL DEFAULT 0,
	due_date INTEGER NOT NULL DEFAULT 0,
	CONSTRAINT season_kind_unique UNIQUE (season, kind)
);

-- Journal of who ticked or unticked what, and when. Rows are only ever added:
-- the current state of a member for a collection is the last row of that pair.
--   action = paid | unpaid | exempt
--   note   = optional, encrypted like the personal data of the members
CREATE TABLE IF NOT EXISTS dues_log
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	member_uuid TEXT NOT NULL,
	collection_uuid TEXT NOT NULL,
	action TEXT NOT NULL,
	actor_uuid TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	note TEXT,
	FOREIGN KEY(member_uuid) REFERENCES members(uuid),
	FOREIGN KEY(collection_uuid) REFERENCES dues_collections(uuid),
	FOREIGN KEY(actor_uuid) REFERENCES members(uuid)
);

CREATE INDEX IF NOT EXISTS dues_log_member_collection ON dues_log (member_uuid, collection_uuid, id);
