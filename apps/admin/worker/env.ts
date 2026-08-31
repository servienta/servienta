export interface Env {
  DB: D1Database;
  ASSETS: Fetcher;
  // HMAC key for session cookies (wrangler secret put SESSION_SECRET).
  SESSION_SECRET?: string;
  // Resend API key for password-reset emails; forgot-password is disabled until set.
  RESEND_API_KEY?: string;
  // From address for reset emails.
  RESEND_FROM?: string;
  // Ed25519 private key, PKCS#8, base64 (wrangler secret put LICENSE_SIGNING_KEY).
  LICENSE_SIGNING_KEY?: string;
  // Ed25519 public key, SPKI, base64 — public by definition, lives in vars.
  LICENSE_PUBLIC_KEY?: string;
}
