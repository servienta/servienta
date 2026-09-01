<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { api } from "./api";
import ThemeToggle from "./components/ThemeToggle.vue";
import { STANDS } from "../shared/stands";

const engineOk = ref<boolean | null>(null);
const version = ref<string | null>(null);
const endpoints = ref<Record<string, string>>({});
const error = ref<string | null>(null);
const received = ref<Record<string, unknown>[] | null>(null);
const service = ref("reference"); const run = ref("");
const resetMsg = ref<string | null>(null);
const license = ref<{ mode: string; stands: string[]; customer?: string; expires_at?: number; error?: string } | null>(null);
const licenseText = ref(""); const licenseMsg = ref<string | null>(null); const uploading = ref(false);
const view = ref<"stand" | "start">("stand");
const apiRef: [string, string, string][] = [
  ["GET", "/api/v1/endpoints", "Service addresses"],
  ["GET", "/api/v1/received/{svc}", "Read recorded messages"],
  ["PUT", "/api/v1/runs/{id}", "Declare a run (claim sources)"],
  ["PUT", "/api/v1/responses/{svc}", "Steer a reply"],
  ["PUT", "/api/v1/faults/receivers/{svc}", "Inject a receiver fault"],
  ["POST", "/api/v1/reset", "Reset the instance"],
];

const grantedSet = computed(() => new Set(license.value?.stands ?? []));

async function refresh() {
  error.value = null;
  try {
    version.value = (await api<{ contract: string }>("/engine/version")).contract;
    endpoints.value = await api<Record<string, string>>("/engine/endpoints");
    license.value = await api("/engine/license");
    engineOk.value = true;
  } catch (e) { engineOk.value = false; error.value = (e as Error).message; }
}
async function loadReceived() { error.value = null;
  try { const q = run.value ? `?run=${encodeURIComponent(run.value)}` : "";
    received.value = await api<Record<string, unknown>[]>(`/engine/received/${service.value}${q}`); }
  catch (e) { error.value = (e as Error).message; } }
async function reset() { resetMsg.value = null;
  try { await api("/engine/reset", { method: "POST" }); resetMsg.value = "reset to a known state"; received.value = null; }
  catch (e) { error.value = (e as Error).message; } }
async function applyLicense() { licenseMsg.value = null; uploading.value = true;
  try { const parsed = JSON.parse(licenseText.value);
    await api("/console/license", { method: "POST", body: JSON.stringify(parsed) });
    licenseMsg.value = "applied — engine restarting…"; licenseText.value = ""; await waitForEngine();
  } catch (e) { licenseMsg.value = (e as Error).message; } finally { uploading.value = false; } }
async function removeLicense() { if (!confirm("Remove the license and revert to free mode?")) return;
  licenseMsg.value = null; uploading.value = true;
  try { await api("/console/license", { method: "DELETE" }); licenseMsg.value = "removed — engine restarting…"; await waitForEngine(); }
  catch (e) { licenseMsg.value = (e as Error).message; } finally { uploading.value = false; } }
async function waitForEngine() { for (let i = 0; i < 40; i++) { await new Promise((r) => setTimeout(r, 500));
  try { await refresh(); if (engineOk.value) return; } catch {} } }
function tabStyle(v: "stand" | "start") {
  return {
    color: view.value === v ? "var(--ink)" : "var(--ink2)",
    paddingBottom: "2px",
    borderBottom: "1px solid " + (view.value === v ? "var(--ink)" : "transparent"),
    cursor: "pointer",
  };
}
onMounted(refresh);
</script>

<template>
  <div style="min-height:100vh">
    <header style="border-bottom:1px solid var(--rule)">
      <div style="max-width:1200px;margin:0 auto;padding:0 40px;height:58px;display:flex;align-items:center;gap:14px">
        <div style="display:flex;align-items:baseline;gap:12px">
          <span style="font-size:17px;font-weight:700;letter-spacing:0.16em;text-transform:uppercase">Servienta</span>
          <span class="mono" style="font-size:10.5px;letter-spacing:0.1em;text-transform:uppercase;color:var(--signal)">console</span>
        </div>
        <nav class="mono" style="display:flex;gap:22px;font-size:11px;letter-spacing:0.06em;text-transform:uppercase">
          <span class="ctab" :style="tabStyle('stand')" @click="view='stand'">Stand</span>
          <span class="ctab" :style="tabStyle('start')" @click="view='start'">Getting started</span>
        </nav>
        <div style="margin-left:auto;display:flex;align-items:center;gap:20px">
          <ThemeToggle />
          <span class="mono" style="display:inline-flex;align-items:center;gap:8px;font-size:11px;letter-spacing:0.06em;text-transform:uppercase;border-left:1px solid var(--rule);padding-left:20px"
            :style="{ color: engineOk ? 'var(--signal)' : 'var(--alert)' }">
            <span :style="{ width:'7px',height:'7px',background: engineOk ? 'var(--signal)' : 'var(--alert)' }"></span>
            {{ engineOk === null ? "…" : engineOk ? `engine up · contract v${version}` : "engine unreachable" }}
          </span>
        </div>
      </div>
    </header>

    <main style="max-width:1200px;margin:0 auto;padding:0 40px">
      <div style="border-left:1px solid var(--rule);border-right:1px solid var(--rule);min-height:calc(100vh - 58px);padding:40px 48px 80px">
        <div v-if="view === 'start'">
          <div style="display:flex;align-items:baseline;gap:16px">
            <span class="mono" style="font-size:12px;color:var(--signal)">§0</span>
            <h2 style="margin:0;font-size:30px;font-weight:600;letter-spacing:-0.01em">Getting started</h2>
            <a href="https://servienta.com/docs" target="_blank" class="mono" style="margin-left:auto;font-size:11px;color:var(--ink2)">full docs ↗</a>
          </div>
          <p style="margin:14px 0 0;max-width:660px;font-size:15px;line-height:1.6;color:var(--ink2)">This console manages a running engine over its versioned HTTP contract. You can drive the engine from the tabs above, or directly with <span class="mono" style="font-size:13px">curl</span> from your test suite — the calls below are exactly what this console makes.</p>

          <div style="margin-top:32px">
            <div class="flabel" style="padding-bottom:8px;border-bottom:2px solid var(--ink)">1 · Reach the engine</div>
            <p style="margin:12px 0 0;font-size:14px;line-height:1.6;color:var(--ink2)">In this compose stack the engine is not published to your host — only this console is, on <span class="mono">:8080</span>. To curl the engine directly, run it with published ports:</p>
            <pre class="mono" style="margin:12px 0 0;font-size:12px;line-height:2;color:var(--ink);overflow-x:auto">mkdir -p fixtures &amp;&amp; printf 'link down eth0\n' &gt; fixtures/hello.txt
docker run --rm -p 8080:8080 -p 8081:8081 -p 9000:9000 \
  -v "$PWD/fixtures:/fixtures:ro" ghcr.io/servienta/engine:latest</pre>
          </div>

          <div style="margin-top:36px">
            <div class="flabel" style="padding-bottom:8px;border-bottom:2px solid var(--ink)">2 · The core loop</div>
            <pre class="mono" style="margin:12px 0 0;font-size:12px;line-height:2;color:var(--ink);overflow-x:auto"><span style="color:var(--ink3)"># what's running, and where</span>
curl localhost:8080/api/v1/endpoints

<span style="color:var(--ink3)"># serve a mounted fixture, byte-for-byte</span>
curl localhost:8081/hello.txt

<span style="color:var(--ink3)"># send traffic, read back what arrived</span>
echo 'hello from my app' | nc localhost 9000
curl localhost:8080/api/v1/received/reference

<span style="color:var(--ink3)"># steer a reply (licensed services), inject a fault, reset</span>
curl -X PUT  localhost:8080/api/v1/responses/dns   -d '{"outcome":"nxdomain"}'
curl -X PUT  localhost:8080/api/v1/faults/receivers/syslog -d '{"mode":"drop"}'
curl -X POST localhost:8080/api/v1/reset</pre>
          </div>

          <div style="margin-top:36px">
            <div class="flabel" style="padding-bottom:8px;border-bottom:2px solid var(--ink)">3 · Run the walkthrough</div>
            <p style="margin:12px 0 0;font-size:14px;line-height:1.6;color:var(--ink2)">A one-file script that runs every step above against your engine and prints the result in your terminal — <span class="mono" style="font-size:13px">scripts/walkthrough.sh</span> in the repo.</p>
            <pre class="mono" style="margin:12px 0 0;font-size:12px;line-height:2;color:var(--ink);overflow-x:auto">./scripts/walkthrough.sh</pre>
          </div>

          <div style="margin-top:36px">
            <div class="flabel" style="padding-bottom:8px;border-bottom:2px solid var(--ink)">4 · Reference</div>
            <div class="mono" style="margin-top:12px;display:grid;grid-template-columns:56px minmax(0,1fr) minmax(0,1fr);gap:16px;padding-bottom:8px;border-bottom:1px solid var(--rule);font-size:10.5px;letter-spacing:0.1em;text-transform:uppercase;color:var(--ink3)"><span>Verb</span><span>Path</span><span>Purpose</span></div>
            <div v-for="e in apiRef" :key="e[1]" class="mono" style="display:grid;grid-template-columns:56px minmax(0,1fr) minmax(0,1fr);gap:16px;align-items:center;padding:8px 0;border-bottom:1px solid var(--rule2);font-size:11.5px">
              <span :style="{ color: e[0]==='GET' ? 'var(--signal)' : e[0]==='DEL' ? 'var(--alert)' : e[0]==='POST' ? 'var(--ink)' : 'var(--warn)' }">{{ e[0] }}</span>
              <span style="color:var(--ink)">{{ e[1] }}</span>
              <span style="color:var(--ink2)">{{ e[2] }}</span>
            </div>
          </div>
        </div>

        <div v-if="view === 'start'"></div>
        <template v-else>
        <div v-if="engineOk === false" class="mono" style="margin-bottom:36px;border-left:2px solid var(--alert);background:var(--band);padding:12px 16px;font-size:11.5px;color:var(--alert)">{{ error }}</div>

        <section v-if="license">
          <div style="display:flex;align-items:baseline;gap:16px">
            <span class="mono" style="font-size:12px;color:var(--signal)">§1</span>
            <h2 style="margin:0;font-size:30px;font-weight:600;letter-spacing:-0.01em">License</h2>
            <span class="mono" style="margin-left:auto;display:inline-flex;align-items:center;gap:8px;font-size:11px;letter-spacing:0.08em;text-transform:uppercase"
              :style="{ color: license.mode === 'licensed' ? 'var(--signal)' : 'var(--warn)' }">
              <span :style="{ width:'7px',height:'7px',background: license.mode === 'licensed' ? 'var(--signal)' : 'var(--warn)' }"></span>{{ license.mode }}
            </span>
          </div>
          <div style="margin-top:20px;display:grid;grid-template-columns:repeat(4,minmax(0,1fr));border-left:1px solid var(--rule)">
            <div v-for="d in [['Licensed to', license.customer ?? '—'],['Expires', license.expires_at ? new Date(license.expires_at).toISOString().slice(0,10) : '—'],['Stands', license.stands.length + ' of ' + STANDS.length],['Error', license.error ?? 'none']]" :key="d[0]"
              style="border-right:1px solid var(--rule);border-top:2px solid var(--ink);padding:12px 16px">
              <div class="flabel" style="font-size:10px">{{ d[0] }}</div>
              <div class="mono" style="margin-top:5px;font-size:12.5px" :style="{ color: d[0]==='Error' && license.error ? 'var(--alert)' : 'var(--ink)' }">{{ d[1] }}</div>
            </div>
          </div>
          <div style="margin-top:20px;display:grid;grid-template-columns:repeat(13,minmax(0,1fr));border-left:1px solid var(--rule)">
            <div v-for="s in STANDS" :key="s.id" style="border-right:1px solid var(--rule);border-top:1px solid var(--rule);border-bottom:1px solid var(--rule);padding:8px 6px;text-align:center"
              :style="{ background: grantedSet.has(s.id) ? 'rgba(47,122,82,0.05)' : 'transparent' }">
              <div style="width:8px;height:8px;margin:0 auto;border:1px solid" :style="{ borderColor: grantedSet.has(s.id) ? 'var(--signal)' : 'var(--rule)', background: grantedSet.has(s.id) ? 'var(--signal)' : 'transparent' }"></div>
              <div class="mono" style="margin-top:7px;font-size:9.5px;overflow:hidden;text-overflow:ellipsis" :style="{ color: grantedSet.has(s.id) ? 'var(--ink)' : 'var(--ink3)' }">{{ s.id.replace('snmp-traps','snmp') }}</div>
            </div>
          </div>
          <div style="margin-top:28px;display:grid;grid-template-columns:200px minmax(0,1fr);gap:32px">
            <div>
              <div class="flabel">Apply a license</div>
              <p style="margin:8px 0 0;font-size:14px;line-height:1.6;color:var(--ink2)">Paste the license file issued in the admin panel. The engine restarts to apply it.</p>
            </div>
            <div style="min-width:0">
              <textarea v-model="licenseText" rows="4" placeholder='{"payload_b64":"…","signature":"…"}' class="mono field" style="width:100%;line-height:1.8;resize:vertical;font-size:12px"></textarea>
              <div style="margin-top:12px;display:flex;align-items:center;gap:20px">
                <button class="btn" :disabled="uploading || !licenseText" @click="applyLicense">Apply</button>
                <span v-if="license.mode === 'licensed'" class="mono" style="font-size:11px;letter-spacing:0.06em;text-transform:uppercase;color:var(--alert);cursor:pointer" @click="removeLicense">Remove license</span>
                <span v-if="uploading" class="mono" style="font-size:11.5px;color:var(--ink3)">working…</span>
                <span v-if="licenseMsg" class="mono" style="font-size:11.5px;color:var(--signal)">{{ licenseMsg }}</span>
              </div>
            </div>
          </div>
        </section>

        <section style="margin-top:52px">
          <div style="display:flex;align-items:baseline;gap:16px">
            <span class="mono" style="font-size:12px;color:var(--signal)">§2</span>
            <h2 style="margin:0;font-size:30px;font-weight:600;letter-spacing:-0.01em">Stand endpoints</h2>
            <span class="act" style="margin-left:auto" @click="refresh">Refresh</span>
          </div>
          <div style="margin-top:20px">
            <div class="flabel" style="display:grid;grid-template-columns:180px minmax(0,1fr) 100px;gap:20px;padding-bottom:8px;border-bottom:2px solid var(--ink)"><span>Service</span><span>Address</span><span>State</span></div>
            <div v-for="(addr, name) in endpoints" :key="name" class="rowh" style="display:grid;grid-template-columns:180px minmax(0,1fr) 100px;gap:20px;align-items:center;padding:8px 0;border-bottom:1px solid var(--rule2)">
              <span class="mono" style="font-size:12.5px">{{ name }}</span>
              <span class="mono" style="font-size:11.5px;color:var(--ink2)">{{ addr }}</span>
              <span class="mono" style="display:flex;align-items:center;gap:7px;font-size:11px;color:var(--signal)"><span style="width:6px;height:6px;background:var(--signal)"></span>up</span>
            </div>
            <div v-if="Object.keys(endpoints).length === 0" class="mono" style="padding:14px 0;font-size:12px;color:var(--ink3)">—</div>
          </div>
        </section>

        <section style="margin-top:52px">
          <div style="display:flex;align-items:baseline;gap:16px">
            <span class="mono" style="font-size:12px;color:var(--signal)">§3</span>
            <h2 style="margin:0;font-size:30px;font-weight:600;letter-spacing:-0.01em">Received messages</h2>
            <span class="mono" style="margin-left:auto;font-size:11px;color:var(--ink3)">GET /received/{service}?run={id}</span>
          </div>
          <div style="margin-top:20px;display:flex;align-items:flex-end;gap:16px">
            <label style="display:block"><span class="flabel">Receiver</span><input v-model="service" class="mono field" style="margin-top:5px;display:block;width:190px" /></label>
            <label style="display:block"><span class="flabel">Run</span><input v-model="run" placeholder="(all)" class="mono field" style="margin-top:5px;display:block;width:140px" /></label>
            <button class="btn" @click="loadReceived">Load</button>
            <span v-if="received" class="mono" style="margin-left:auto;font-size:11px;color:var(--ink3)">{{ received.length }} messages</span>
          </div>
          <pre v-if="received" class="mono" style="margin:18px 0 0;max-height:320px;overflow:auto;border-top:2px solid var(--ink);padding-top:14px;font-size:11.5px;line-height:1.9;color:var(--ink2)">{{ JSON.stringify(received, null, 2) }}</pre>
        </section>

        <section style="margin-top:52px">
          <div style="display:flex;align-items:baseline;gap:16px">
            <span class="mono" style="font-size:12px;color:var(--signal)">§4</span>
            <h2 style="margin:0;font-size:30px;font-weight:600;letter-spacing:-0.01em">Reset</h2>
            <span class="mono" style="margin-left:auto;font-size:11px;color:var(--ink3)">POST /reset</span>
          </div>
          <div style="margin-top:20px;border-top:2px solid var(--ink);padding-top:18px;display:grid;grid-template-columns:200px minmax(0,1fr);gap:32px;align-items:start">
            <div class="flabel">Instance-wide</div>
            <div>
              <p style="margin:0;max-width:560px;font-size:15px;line-height:1.6;color:var(--ink2)">Clear recorded messages, faults, and run declarations across the whole stand.</p>
              <div style="margin-top:16px;display:flex;align-items:center;gap:20px">
                <button class="btn-ghost" @click="reset">Reset stand</button>
                <span v-if="resetMsg" class="mono" style="font-size:11.5px;color:var(--signal)">{{ resetMsg }}</span>
              </div>
            </div>
          </div>
        </section>
        </template>
      </div>
    </main>
  </div>
</template>
