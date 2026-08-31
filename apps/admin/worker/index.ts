import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import { drizzle } from "drizzle-orm/d1";
import { desc, eq } from "drizzle-orm";
import { customers, licenses } from "./db/schema";
import { requireSession, sessionEmail } from "./auth";
import { authRoutes } from "./authRoutes";
import { issueLicense } from "./license";
import { STAND_IDS } from "../shared/stands";
import { ulid } from "./ulid";
import type { Env } from "./env";

// TODO(stack.md): move route definitions to @hono/zod-openapi once the
// surface settles, so the OpenAPI document is generated, not hand-kept.

const app = new Hono<{ Bindings: Env; Variables: { userEmail: string } }>().basePath("/api");

app.get("/health", (c) =>
  c.json({ ok: true, service: "servienta-admin", time: new Date().toISOString() }),
);

app.route("/auth", authRoutes);

app.get("/me", async (c) => {
  const email = await sessionEmail(c);
  if (!email) return c.json({ error: "unauthorized" }, 401);
  return c.json({ email });
});

app.use("*", requireSession);

app.get("/customers", async (c) => {
  const db = drizzle(c.env.DB);
  const rows = await db.select().from(customers).orderBy(desc(customers.createdAt));
  return c.json(rows);
});

app.post(
  "/customers",
  zValidator("json", z.object({ name: z.string().min(1), email: z.string().email() })),
  async (c) => {
    const body = c.req.valid("json");
    const db = drizzle(c.env.DB);
    const row = { id: ulid(), name: body.name, email: body.email, createdAt: Date.now() };
    await db.insert(customers).values(row);
    return c.json(row, 201);
  },
);

app.put(
  "/customers/:id",
  zValidator("json", z.object({ name: z.string().min(1), email: z.string().email() })),
  async (c) => {
    const body = c.req.valid("json");
    const db = drizzle(c.env.DB);
    const id = c.req.param("id");
    const [existing] = await db.select().from(customers).where(eq(customers.id, id));
    if (!existing) return c.json({ error: "not found" }, 404);
    try {
      await db.update(customers).set({ name: body.name, email: body.email }).where(eq(customers.id, id));
    } catch {
      return c.json({ error: "email already in use" }, 409);
    }
    return c.json({ ...existing, name: body.name, email: body.email });
  },
);

// Deletes the customer AND every license issued to them.
app.delete("/customers/:id", async (c) => {
  const db = drizzle(c.env.DB);
  const id = c.req.param("id");
  const [existing] = await db.select().from(customers).where(eq(customers.id, id));
  if (!existing) return c.json({ error: "not found" }, 404);
  await db.delete(licenses).where(eq(licenses.customerId, id));
  await db.delete(customers).where(eq(customers.id, id));
  return c.body(null, 204);
});

app.get("/licenses", async (c) => {
  const db = drizzle(c.env.DB);
  const rows = await db
    .select({
      id: licenses.id,
      customerId: licenses.customerId,
      customerName: customers.name,
      stands: licenses.stands,
      expiresAt: licenses.expiresAt,
      createdAt: licenses.createdAt,
    })
    .from(licenses)
    .innerJoin(customers, eq(customers.id, licenses.customerId))
    .orderBy(desc(licenses.createdAt));
  return c.json(rows.map((r) => ({ ...r, stands: JSON.parse(r.stands) as string[] })));
});

app.post(
  "/licenses",
  zValidator(
    "json",
    z.object({
      customerId: z.string().min(1),
      stands: z.array(z.string()).min(1).refine((s) => s.every((id) => STAND_IDS.includes(id)), {
        message: "unknown stand id",
      }),
      expiresAt: z.number().int().positive(),
    }),
  ),
  async (c) => {
    const body = c.req.valid("json");
    const db = drizzle(c.env.DB);
    const [customer] = await db.select().from(customers).where(eq(customers.id, body.customerId));
    if (!customer) return c.json({ error: "unknown customer" }, 404);
    const signed = await issueLicense(c.env, {
      customerId: customer.id,
      customerName: customer.name,
      stands: body.stands,
      expiresAt: body.expiresAt,
    });
    await db.insert(licenses).values({
      id: signed.id,
      customerId: customer.id,
      stands: JSON.stringify(body.stands),
      expiresAt: body.expiresAt,
      payloadB64: signed.payloadB64,
      signature: signed.signature,
      createdAt: Date.now(),
    });
    // The license file the customer mounts next to the engine (R12).
    return c.json({ payload_b64: signed.payloadB64, signature: signed.signature }, 201);
  },
);

app.get("/licenses/:id/file", async (c) => {
  const db = drizzle(c.env.DB);
  const [row] = await db.select().from(licenses).where(eq(licenses.id, c.req.param("id")));
  if (!row) return c.json({ error: "not found" }, 404);
  return c.json({ payload_b64: row.payloadB64, signature: row.signature });
});

export default app;
