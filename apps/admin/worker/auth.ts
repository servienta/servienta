import type { Context, Next } from "hono";
import { createRemoteJWKSet, jwtVerify } from "jose";
import type { Env } from "./env";

// Cloudflare Access IS the authentication (D13): the Worker verifies the JWT
// that Access injects. Deny by default — an unconfigured deployment refuses
// every request instead of standing open; local dev opts out explicitly via
// DEV_ALLOW_INSECURE in .dev.vars.
export async function requireAccess(c: Context<{ Bindings: Env; Variables: { accessEmail: string } }>, next: Next) {
  const env = c.env;
  if (!env.ACCESS_TEAM_DOMAIN || !env.ACCESS_AUD) {
    if (env.DEV_ALLOW_INSECURE === "true") {
      c.set("accessEmail", "dev@localhost");
      return next();
    }
    return c.json(
      { error: "Cloudflare Access is not configured (ACCESS_TEAM_DOMAIN, ACCESS_AUD); refusing by default." },
      403,
    );
  }
  const token = c.req.header("Cf-Access-Jwt-Assertion");
  if (!token) return c.json({ error: "missing Cloudflare Access JWT" }, 403);
  try {
    const issuer = `https://${env.ACCESS_TEAM_DOMAIN}`;
    const jwks = createRemoteJWKSet(new URL(`${issuer}/cdn-cgi/access/certs`));
    const { payload } = await jwtVerify(token, jwks, { issuer, audience: env.ACCESS_AUD });
    c.set("accessEmail", String(payload.email ?? payload.sub));
    return next();
  } catch {
    return c.json({ error: "invalid Cloudflare Access JWT" }, 403);
  }
}
