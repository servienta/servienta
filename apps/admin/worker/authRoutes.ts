import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import { drizzle } from "drizzle-orm/d1";
import { and, eq, isNull } from "drizzle-orm";
import { users, resetTokens } from "./db/schema";
import {
  verifyPassword,
  hashPassword,
  issueSession,
  clearSession,
  sessionEmail,
  sha256b64url,
  randomToken,
} from "./auth";
import type { Env } from "./env";

const RESET_TTL_MS = 15 * 60 * 1000;

export const authRoutes = new Hono<{ Bindings: Env }>();

authRoutes.post(
  "/login",
  zValidator("json", z.object({ email: z.string().email(), password: z.string().min(1) })),
  async (c) => {
    if (!c.env.SESSION_SECRET) return c.json({ error: "SESSION_SECRET is not set" }, 500);
    const { email, password } = c.req.valid("json");
    const db = drizzle(c.env.DB);
    const [user] = await db.select().from(users).where(eq(users.email, email.toLowerCase()));
    if (!user || !(await verifyPassword(password, user.passwordHash))) {
      return c.json({ error: "invalid email or password" }, 401);
    }
    await issueSession(c, user.email);
    return c.json({ email: user.email });
  },
);

authRoutes.post("/logout", async (c) => {
  clearSession(c);
  return c.body(null, 204);
});

authRoutes.post(
  "/change",
  zValidator("json", z.object({ current: z.string().min(1), next: z.string().min(12) })),
  async (c) => {
    const email = await sessionEmail(c);
    if (!email) return c.json({ error: "unauthorized" }, 401);
    const { current, next } = c.req.valid("json");
    const db = drizzle(c.env.DB);
    const [user] = await db.select().from(users).where(eq(users.email, email));
    if (!user || !(await verifyPassword(current, user.passwordHash))) {
      return c.json({ error: "current password is wrong" }, 403);
    }
    await db.update(users).set({ passwordHash: await hashPassword(next) }).where(eq(users.id, user.id));
    return c.body(null, 204);
  },
);

authRoutes.post(
  "/forgot",
  zValidator("json", z.object({ email: z.string().email() })),
  async (c) => {
    // Always 204: no account enumeration. Requires Resend to actually send.
    const { email } = c.req.valid("json");
    if (!c.env.RESEND_API_KEY) return c.body(null, 204);
    const db = drizzle(c.env.DB);
    const [user] = await db.select().from(users).where(eq(users.email, email.toLowerCase()));
    if (user) {
      const token = randomToken();
      await db.insert(resetTokens).values({
        tokenHash: await sha256b64url(token),
        userId: user.id,
        expiresAt: Date.now() + RESET_TTL_MS,
      });
      const link = `${new URL(c.req.url).origin}/reset?token=${token}`;
      await fetch("https://api.resend.com/emails", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${c.env.RESEND_API_KEY}`,
          "content-type": "application/json",
        },
        body: JSON.stringify({
          from: c.env.RESEND_FROM ?? "Servienta <onboarding@resend.dev>",
          to: [user.email],
          subject: "Servienta admin — password reset",
          text: `Reset link (valid 15 minutes): ${link}`,
        }),
      });
    }
    return c.body(null, 204);
  },
);

authRoutes.post(
  "/reset",
  zValidator("json", z.object({ token: z.string().min(1), next: z.string().min(12) })),
  async (c) => {
    const { token, next } = c.req.valid("json");
    const db = drizzle(c.env.DB);
    const hash = await sha256b64url(token);
    const [row] = await db
      .select()
      .from(resetTokens)
      .where(and(eq(resetTokens.tokenHash, hash), isNull(resetTokens.usedAt)));
    if (!row || row.expiresAt < Date.now()) return c.json({ error: "invalid or expired token" }, 403);
    await db.update(resetTokens).set({ usedAt: Date.now() }).where(eq(resetTokens.tokenHash, hash));
    await db.update(users).set({ passwordHash: await hashPassword(next) }).where(eq(users.id, row.userId));
    return c.body(null, 204);
  },
);
