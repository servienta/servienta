-- D15: a license enumerates stands; the edition placeholder goes away.
ALTER TABLE licenses DROP COLUMN edition;
ALTER TABLE licenses ADD COLUMN stands TEXT NOT NULL DEFAULT '[]';
