<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "./api";
import Mark from "./components/Mark.vue";
import ThemeToggle from "./components/ThemeToggle.vue";

const engineOk = ref<boolean | null>(null);
const version = ref<string | null>(null);
const endpoints = ref<Record<string, string>>({});
const error = ref<string | null>(null);
const received = ref<unknown[] | null>(null);
const service = ref("reference");
const run = ref("");
const resetMsg = ref<string | null>(null);
const license = ref<{ mode: string; stands: string[]; customer?: string; expires_at?: number; error?: string } | null>(null);
const licenseText = ref("");
const licenseMsg = ref<string | null>(null);
const uploading = ref(false);

async function refresh() {
  error.value = null;
  try {
    version.value = (await api<{ contract: string }>("/engine/version")).contract;
    endpoints.value = await api<Record<string, string>>("/engine/endpoints");
    license.value = await api("/engine/license");
    engineOk.value = true;
  } catch (e) {
    engineOk.value = false;
    error.value = (e as Error).message;
  }
}
async function loadReceived() {
  error.value = null;
  try {
    const q = run.value ? `?run=${encodeURIComponent(run.value)}` : "";
    received.value = await api<unknown[]>(`/engine/received/${service.value}${q}`);
  } catch (e) { error.value = (e as Error).message; }
}
async function reset() {
  resetMsg.value = null;
  try { await api("/engine/reset", { method: "POST" }); resetMsg.value = "Stand reset to a known state."; received.value = null; }
  catch (e) { error.value = (e as Error).message; }
}
async function applyLicense() {
  licenseMsg.value = null; uploading.value = true;
  try {
    const parsed = JSON.parse(licenseText.value);
    await api("/console/license", { method: "POST", body: JSON.stringify(parsed) });
    licenseMsg.value = "License applied — engine restarting…"; licenseText.value = "";
    await waitForEngine();
  } catch (e) { licenseMsg.value = (e as Error).message; } finally { uploading.value = false; }
}
async function removeLicense() {
  if (!confirm("Remove the license and revert to free mode?")) return;
  licenseMsg.value = null; uploading.value = true;
  try {
    await api("/console/license", { method: "DELETE" });
    licenseMsg.value = "License removed — engine restarting…";
    await waitForEngine();
  } catch (e) { licenseMsg.value = (e as Error).message; } finally { uploading.value = false; }
}
async function waitForEngine() {
  for (let i = 0; i < 40; i++) {
    await new Promise((r) => setTimeout(r, 500));
    try { await refresh(); if (engineOk.value) return; } catch {}
  }
}
onMounted(refresh);
</script>

<template>
  <div style="min-height:100vh">
    <div style="border-bottom:1px solid var(--bd);background:var(--surface)">
      <div style="max-width:1000px;margin:0 auto;padding:0 24px;height:56px;display:flex;align-items:center;gap:14px">
        <div style="display:flex;align-items:center;gap:9px">
          <Mark :size="3" />
          <span style="font-size:15px;font-weight:600;letter-spacing:-0.02em">Servienta</span>
          <span class="mono" style="font-size:10.5px;letter-spacing:0.1em;text-transform:uppercase;color:#8b6cff;border:1px solid rgba(139,108,255,0.4);border-radius:4px;padding:2px 6px">console</span>
        </div>
        <div style="margin-left:auto;display:flex;align-items:center;gap:12px">
          <ThemeToggle />
          <span class="mono" style="display:inline-flex;align-items:center;gap:7px;border-radius:999px;padding:4px 11px;font-size:11.5px;font-weight:500"
            :style="{ background: engineOk ? 'rgba(16,185,129,0.13)' : 'rgba(239,68,68,0.13)', color: engineOk ? 'var(--ok)' : 'var(--err)' }">
            <span :style="{ width:'6px',height:'6px',borderRadius:'999px',background: engineOk ? 'var(--ok-dot)' : 'var(--err)' }"></span>
            {{ engineOk === null ? "…" : engineOk ? `engine up · contract v${version}` : "engine unreachable" }}
          </span>
        </div>
      </div>
    </div>

    <main style="max-width:1000px;margin:0 auto;padding:28px 24px 80px;display:flex;flex-direction:column;gap:16px">
      <div v-if="engineOk === false" class="mono" style="border:1px solid var(--err-bd);background:var(--err-bg);border-radius:10px;padding:12px 16px;font-size:12.5px;color:var(--err)">{{ error }}</div>

      <section v-if="license" class="card" style="overflow:hidden">
        <div style="display:flex;align-items:center;justify-content:space-between;padding:14px 18px;border-bottom:1px solid var(--bd-soft)">
          <h2 style="margin:0;font-size:15px;font-weight:600;letter-spacing:-0.02em">License</h2>
          <span style="border-radius:999px;padding:3px 10px;font-size:11px;font-weight:500" class="mono"
            :style="{ background: license.mode === 'licensed' ? 'rgba(16,185,129,0.13)' : 'rgba(217,119,6,0.13)', color: license.mode === 'licensed' ? 'var(--ok)' : '#d97706' }">{{ license.mode }}</span>
        </div>
        <div style="padding:18px">
          <p v-if="license.customer" style="margin:0;font-size:13.5px;color:var(--body)">Licensed to <span style="font-weight:500;color:var(--fg)">{{ license.customer }}</span><span v-if="license.expires_at"> · expires {{ new Date(license.expires_at).toISOString().slice(0,10) }}</span></p>
          <p v-if="license.error" style="margin:0;font-size:13px;color:var(--err)">License rejected: {{ license.error }} — running in free mode.</p>
          <div style="margin-top:14px;display:flex;flex-wrap:wrap;gap:6px">
            <span v-for="s in license.stands" :key="s" class="chip">{{ s }}</span>
          </div>
          <div style="margin-top:20px;padding-top:18px;border-top:1px solid var(--bd-soft)">
            <div style="font-size:13.5px;font-weight:600">Apply a license</div>
            <p style="margin:5px 0 0;font-size:12.5px;line-height:1.6;color:var(--mu)">Paste the license file issued in the admin panel (the JSON with payload_b64 and signature). The engine restarts to apply it.</p>
            <textarea v-model="licenseText" rows="4" placeholder='{"payload_b64":"…","signature":"…"}' class="mono inp" style="margin-top:10px;width:100%;border-radius:8px;font-size:12px;line-height:1.7;resize:vertical"></textarea>
            <div style="margin-top:10px;display:flex;align-items:center;gap:14px">
              <button class="btn" style="padding:9px 20px" :disabled="uploading || !licenseText" @click="applyLicense">Apply</button>
              <span v-if="license.mode === 'licensed'" style="font-size:13px;color:var(--err);cursor:pointer" @click="removeLicense">Remove license</span>
              <span v-if="uploading" style="font-size:13px;color:var(--dim)">working…</span>
            </div>
            <p v-if="licenseMsg" style="margin:10px 0 0;font-size:13px;color:var(--ok)">{{ licenseMsg }}</p>
          </div>
        </div>
      </section>

      <section class="card" style="overflow:hidden">
        <div style="display:flex;align-items:center;justify-content:space-between;padding:14px 18px;border-bottom:1px solid var(--bd-soft)">
          <h2 style="margin:0;font-size:15px;font-weight:600;letter-spacing:-0.02em">Stand endpoints</h2>
          <span class="link" style="font-size:13px" @click="refresh">Refresh</span>
        </div>
        <div v-for="(addr, name) in endpoints" :key="name" class="row" style="display:grid;grid-template-columns:170px 1fr 90px;gap:16px;align-items:center;padding:9px 18px;border-bottom:1px solid var(--bd-soft)">
          <span style="font-size:13.5px;font-weight:500">{{ name }}</span>
          <span class="mono" style="font-size:12.5px;color:var(--body)">{{ addr }}</span>
          <span class="mono" style="justify-self:end;font-size:11px;color:var(--ok)">up</span>
        </div>
        <div v-if="Object.keys(endpoints).length === 0" style="padding:18px;font-size:13px;color:var(--dim)">—</div>
      </section>

      <section class="card" style="overflow:hidden">
        <div style="padding:14px 18px;border-bottom:1px solid var(--bd-soft)">
          <h2 style="margin:0;font-size:15px;font-weight:600;letter-spacing:-0.02em">Received messages</h2>
        </div>
        <div style="padding:18px">
          <div style="display:flex;align-items:flex-end;gap:12px">
            <label style="display:block">
              <span class="label">Receiver</span>
              <input v-model="service" class="mono inp" style="margin-top:5px;display:block;width:180px;font-size:13px" />
            </label>
            <label style="display:block">
              <span class="label">Run</span>
              <input v-model="run" placeholder="(all)" class="mono inp" style="margin-top:5px;display:block;width:130px;font-size:13px" />
            </label>
            <button class="btn" style="padding:9px 20px" @click="loadReceived">Load</button>
          </div>
          <pre v-if="received" class="mono" style="margin:14px 0 0;max-height:280px;overflow:auto;border:1px solid var(--bd-soft);border-radius:8px;background:var(--code-bg);padding:14px;font-size:11.5px;line-height:1.85;color:var(--body)">{{ JSON.stringify(received, null, 2) }}</pre>
        </div>
      </section>

      <section class="card" style="padding:18px">
        <h2 style="margin:0;font-size:15px;font-weight:600;letter-spacing:-0.02em">Reset</h2>
        <p style="margin:5px 0 0;font-size:13.5px;color:var(--mu)">Clear recorded messages, faults, and run declarations across the whole stand.</p>
        <button class="btn-ghost" style="margin-top:14px" @click="reset">Reset stand</button>
        <p v-if="resetMsg" style="margin:10px 0 0;font-size:13px;color:var(--ok)">{{ resetMsg }}</p>
      </section>
    </main>
  </div>
</template>

<style scoped>
.row:hover { background:var(--hover); }
</style>
