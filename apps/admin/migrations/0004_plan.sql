-- D17: a license records the plan it was sold under (accounting); stands stay
-- the source of truth for what the engine enables.
ALTER TABLE licenses ADD COLUMN plan TEXT NOT NULL DEFAULT 'custom';
