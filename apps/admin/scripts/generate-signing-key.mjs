// Generates the Ed25519 license-signing keypair (D10/R12).
// Private key -> `wrangler secret put LICENSE_SIGNING_KEY` (paste the pkcs8 line).
// Public key  -> LICENSE_PUBLIC_KEY var in wrangler.jsonc, and embedded in the engine.
const pair = await crypto.subtle.generateKey("Ed25519", true, ["sign", "verify"]);
const pkcs8 = Buffer.from(await crypto.subtle.exportKey("pkcs8", pair.privateKey)).toString("base64");
const spki = Buffer.from(await crypto.subtle.exportKey("spki", pair.publicKey)).toString("base64");
console.log("PRIVATE (pkcs8, base64) — secret, never commit:\n" + pkcs8);
console.log("\nPUBLIC (spki, base64) — goes to vars and into the engine:\n" + spki);
