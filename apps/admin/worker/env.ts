export interface Env {
  DB: D1Database;
  // Cloudflare Access team domain, e.g. "example.cloudflareaccess.com".
  ACCESS_TEAM_DOMAIN?: string;
  // Audience tag of the Access application protecting admin.servienta.com.
  ACCESS_AUD?: string;
  // Local development only (.dev.vars); never set in production.
  DEV_ALLOW_INSECURE?: string;
  // Ed25519 private key, PKCS#8, base64 (wrangler secret put LICENSE_SIGNING_KEY).
  LICENSE_SIGNING_KEY?: string;
  // Ed25519 public key, SPKI, base64 — public by definition, lives in vars.
  LICENSE_PUBLIC_KEY?: string;
}
