import type { Env } from "./env";
import { ulid } from "./ulid";

// Offline-verifiable license (D10/R12): the engine embeds the public key and
// verifies the exact payload bytes — the file carries base64(payload JSON) and
// the Ed25519 signature over those bytes.

export interface LicenseInput {
  customerId: string;
  customerName: string;
  edition: string;
  expiresAt: number; // epoch millis UTC
}

function toBytes(b64: string): Uint8Array {
  return Uint8Array.from(atob(b64), (ch) => ch.charCodeAt(0));
}

function toB64(buf: ArrayBuffer | Uint8Array): string {
  const bytes = buf instanceof Uint8Array ? buf : new Uint8Array(buf);
  let s = "";
  for (const b of bytes) s += String.fromCharCode(b);
  return btoa(s);
}

export async function issueLicense(env: Env, input: LicenseInput) {
  if (!env.LICENSE_SIGNING_KEY) {
    throw new Error("LICENSE_SIGNING_KEY secret is not set");
  }
  const payload = {
    v: 1,
    jti: ulid(),
    sub: input.customerId,
    name: input.customerName,
    edition: input.edition,
    iat: Date.now(),
    exp: input.expiresAt,
  };
  const payloadBytes = new TextEncoder().encode(JSON.stringify(payload));
  const key = await crypto.subtle.importKey(
    "pkcs8",
    toBytes(env.LICENSE_SIGNING_KEY),
    { name: "Ed25519" },
    false,
    ["sign"],
  );
  const signature = await crypto.subtle.sign("Ed25519", key, payloadBytes);
  return {
    id: payload.jti,
    payloadB64: toB64(payloadBytes),
    signature: toB64(signature),
  };
}
