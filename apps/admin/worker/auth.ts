// Single-user email+password auth (D14): PBKDF2 hash in Cloudflare D1,
// stateless HMAC-signed session cookie, reset tokens hashed at rest.
import type { Context, Next } from "hono";
import { getCookie, setCookie, deleteCookie } from "hono/cookie";
import type { Env } from "./env";

const PBKDF2_ITERS = 100_000;
const SESSION_TTL_S = 7 * 24 * 3600;
const COOKIE = "servienta_session";

const enc = new TextEncoder();

function b64url(bytes: Uint8Array): string {
  let s = "";
  for (const b of bytes) s += String.fromCharCode(b);
  return btoa(s).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
}
function fromB64url(s: string): Uint8Array {
  const b = s.replaceAll("-", "+").replaceAll("_", "/");
  return Uint8Array.from(atob(b), (ch) => ch.charCodeAt(0));
}

export async function pbkdf2(password: string, salt: Uint8Array, iters = PBKDF2_ITERS): Promise<Uint8Array> {
  const key = await crypto.subtle.importKey("raw", enc.encode(password), "PBKDF2", false, ["deriveBits"]);
  const bits = await crypto.subtle.deriveBits(
    { name: "PBKDF2", hash: "SHA-256", salt: salt as BufferSource, iterations: iters },
    key,
    256,
  );
  return new Uint8Array(bits);
}

// Stored format: pbkdf2$<iters>$<salt b64url>$<hash b64url>
export async function hashPassword(password: string): Promise<string> {
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const hash = await pbkdf2(password, salt);
  return `pbkdf2$${PBKDF2_ITERS}$${b64url(salt)}$${b64url(hash)}`;
}

export async function verifyPassword(password: string, stored: string): Promise<boolean> {
  const [scheme, iters, salt, hash] = stored.split("$");
  if (scheme !== "pbkdf2") return false;
  const got = await pbkdf2(password, fromB64url(salt), Number(iters));
  const want = fromB64url(hash);
  if (got.length !== want.length) return false;
  let diff = 0;
  for (let i = 0; i < got.length; i++) diff |= got[i] ^ want[i];
  return diff === 0;
}

async function hmac(secret: string, data: string): Promise<string> {
  const key = await crypto.subtle.importKey("raw", enc.encode(secret), { name: "HMAC", hash: "SHA-256" }, false, ["sign"]);
  return b64url(new Uint8Array(await crypto.subtle.sign("HMAC", key, enc.encode(data))));
}

export async function issueSession(c: Context<{ Bindings: Env }>, email: string): Promise<void> {
  const payload = b64url(enc.encode(JSON.stringify({ email, exp: Math.floor(Date.now() / 1000) + SESSION_TTL_S })));
  const sig = await hmac(c.env.SESSION_SECRET!, payload);
  setCookie(c, COOKIE, `${payload}.${sig}`, {
    httpOnly: true,
    secure: true,
    sameSite: "Lax",
    path: "/",
    maxAge: SESSION_TTL_S,
  });
}

export function clearSession(c: Context<{ Bindings: Env }>): void {
  deleteCookie(c, COOKIE, { path: "/" });
}

export async function sessionEmail(c: Context<{ Bindings: Env }>): Promise<string | null> {
  if (!c.env.SESSION_SECRET) return null;
  const raw = getCookie(c, COOKIE);
  if (!raw) return null;
  const [payload, sig] = raw.split(".");
  if (!payload || !sig || (await hmac(c.env.SESSION_SECRET, payload)) !== sig) return null;
  try {
    const data = JSON.parse(new TextDecoder().decode(fromB64url(payload)));
    if (typeof data.email !== "string" || data.exp * 1000 < Date.now()) return null;
    return data.email;
  } catch {
    return null;
  }
}

export async function requireSession(
  c: Context<{ Bindings: Env; Variables: { userEmail: string } }>,
  next: Next,
) {
  const email = await sessionEmail(c);
  if (!email) return c.json({ error: "unauthorized" }, 401);
  c.set("userEmail", email);
  return next();
}

export async function sha256b64url(s: string): Promise<string> {
  return b64url(new Uint8Array(await crypto.subtle.digest("SHA-256", enc.encode(s))));
}

export function randomToken(): string {
  return b64url(crypto.getRandomValues(new Uint8Array(32)));
}
