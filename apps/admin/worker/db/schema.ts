import { sqliteTable, text, integer } from "drizzle-orm/sqlite-core";

// D9 portability rules: text ULIDs for ids, integer epoch-millis UTC
// timestamps, JSON as text. Keep every column inside the SQLite∩Postgres
// common subset.

export const customers = sqliteTable("customers", {
  id: text("id").primaryKey(),
  name: text("name").notNull(),
  email: text("email").notNull().unique(),
  createdAt: integer("created_at").notNull(),
});

export const licenses = sqliteTable("licenses", {
  id: text("id").primaryKey(),
  customerId: text("customer_id")
    .notNull()
    .references(() => customers.id),
  edition: text("edition").notNull(),
  expiresAt: integer("expires_at").notNull(),
  // The exact signed bytes (base64) and their Ed25519 signature (base64).
  // The engine verifies the bytes, so canonicalization never matters (D10).
  payloadB64: text("payload_b64").notNull(),
  signature: text("signature").notNull(),
  createdAt: integer("created_at").notNull(),
});
