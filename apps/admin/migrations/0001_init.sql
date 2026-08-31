-- Initial schema; matches worker/db/schema.ts (D9: common SQLite∩Postgres subset).
CREATE TABLE customers (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  created_at INTEGER NOT NULL
);

CREATE TABLE licenses (
  id TEXT PRIMARY KEY,
  customer_id TEXT NOT NULL REFERENCES customers(id),
  edition TEXT NOT NULL,
  expires_at INTEGER NOT NULL,
  payload_b64 TEXT NOT NULL,
  signature TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE INDEX licenses_customer_id ON licenses(customer_id);
