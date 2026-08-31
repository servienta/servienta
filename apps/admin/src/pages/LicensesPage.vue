<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";
import { useCustomersStore } from "../stores/customers";
import { STANDS } from "../../shared/stands";
import { PLANS, planById } from "../../shared/plans";

interface LicenseRow {
  id: string; customerId: string; customerName: string; plan: string;
  stands: string[]; expiresAt: number; createdAt: number;
}
interface LicensePayload {
  v: number; jti: string; sub: string; name: string; stands: string[]; plan: string; iat: number; exp: number;
}

const store = useCustomersStore();
const rows = ref<LicenseRow[]>([]);
const customerId = ref("");
const plan = ref<string>("standard");
const stands = ref<string[]>([]);
const expires = ref("");
const error = ref<string | null>(null);
const viewing = ref<{ id: string; file: string; payload: LicensePayload } | null>(null);
const copied = ref(false);

function applyPlan() {
  const p = planById(plan.value);
  if (!p) return;
  if (p.id !== "custom") stands.value = [...p.stands];
  expires.value = new Date(Date.now() + p.termDays * 86400000).toISOString().slice(0, 10);
}
async function load() { rows.value = await api<LicenseRow[]>("/licenses"); }

onMounted(() => {
  applyPlan();
  Promise.all([load(), store.loaded ? Promise.resolve() : store.load()]).catch((e) => (error.value = e.message));
});

async function issue() {
  error.value = null;
  try {
    await api("/licenses", {
      method: "POST",
      body: JSON.stringify({
        customerId: customerId.value, plan: plan.value, stands: stands.value,
        expiresAt: new Date(expires.value + "T00:00:00Z").getTime(),
      }),
    });
    await load();
  } catch (e) { error.value = (e as Error).message; }
}
async function view(id: string) {
  error.value = null; copied.value = false;
  try {
    const file = await api<{ payload_b64: string; signature: string }>(`/licenses/${id}/file`);
    viewing.value = { id, file: JSON.stringify(file, null, 2), payload: JSON.parse(atob(file.payload_b64)) as LicensePayload };
  } catch (e) { error.value = (e as Error).message; }
}
async function copyFile() { if (!viewing.value) return; await navigator.clipboard.writeText(viewing.value.file); copied.value = true; }
function downloadFile() {
  if (!viewing.value) return;
  const url = URL.createObjectURL(new Blob([viewing.value.file], { type: "application/json" }));
  const a = document.createElement("a"); a.href = url; a.download = `servienta-license-${viewing.value.id}.json`; a.click(); URL.revokeObjectURL(url);
}
const cols = "1.1fr 100px 1.6fr 110px 110px 60px";
</script>

<template>
  <h1 style="margin:0;font-size:24px;font-weight:600;letter-spacing:-0.03em">Licenses</h1>

  <div class="card" style="margin-top:24px;padding:18px">
    <form style="display:flex;flex-wrap:wrap;align-items:flex-end;gap:14px" @submit.prevent="issue">
      <label style="display:block">
        <span class="label">Customer</span>
        <select v-model="customerId" required class="inp" style="margin-top:5px;display:block;width:200px">
          <option v-for="c in store.items" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </label>
      <label style="display:block">
        <span class="label">Plan</span>
        <select v-model="plan" class="inp" style="margin-top:5px;display:block;width:150px" @change="applyPlan">
          <option v-for="pl in PLANS" :key="pl.id" :value="pl.id">{{ pl.label }}</option>
        </select>
      </label>
      <label style="display:block">
        <span class="label">Expires</span>
        <input v-model="expires" required type="date" class="inp" style="margin-top:5px;display:block" />
      </label>
      <button class="btn" style="padding:9px 20px">Issue</button>
    </form>

    <fieldset style="margin:18px 0 0;padding:0;border:none">
      <legend class="label" style="padding:0">Stands <span v-if="plan !== 'custom'" style="color:var(--dim);text-transform:none;letter-spacing:0">(set by plan)</span></legend>
      <div style="margin-top:8px;display:grid;grid-template-columns:repeat(4,1fr);gap:8px 20px;border:1px solid var(--bd-soft);border-radius:8px;padding:14px;background:var(--input-bg)">
        <label v-for="s in STANDS" :key="s.id" style="display:flex;align-items:center;gap:8px;font-size:13px" :style="{ color: stands.includes(s.id) ? 'var(--fg)' : 'var(--dim)' }">
          <input v-model="stands" type="checkbox" :value="s.id" :disabled="plan !== 'custom'" style="accent-color:#8b6cff" />
          <span>{{ s.label }}</span>
        </label>
      </div>
    </fieldset>
  </div>
  <p v-if="error" style="margin:10px 0 0;font-size:13px;color:var(--err)">{{ error }}</p>

  <div class="card" style="margin-top:16px;overflow:hidden">
    <div class="label" :style="{ display:'grid', gridTemplateColumns:cols, gap:'14px', padding:'10px 16px', borderBottom:'1px solid var(--bd)' }">
      <span>Customer</span><span>Plan</span><span>Stands</span><span>Expires</span><span>Issued</span><span></span>
    </div>
    <div v-for="l in rows" :key="l.id" class="row" :style="{ display:'grid', gridTemplateColumns:cols, gap:'14px', alignItems:'center', padding:'11px 16px', borderBottom:'1px solid var(--bd-soft)' }">
      <span style="font-size:13.5px;font-weight:500">{{ l.customerName }}</span>
      <span><span class="chip">{{ l.plan }}</span></span>
      <span class="mono" style="font-size:11.5px;color:var(--dim);overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ l.stands.join(", ") }}</span>
      <span class="mono" style="font-size:12px;color:var(--body)">{{ new Date(l.expiresAt).toISOString().slice(0,10) }}</span>
      <span class="mono" style="font-size:12px;color:var(--dim)">{{ new Date(l.createdAt).toISOString().slice(0,10) }}</span>
      <span class="link" style="text-align:right;font-size:13px" @click="view(l.id)">View</span>
    </div>
    <div v-if="rows.length === 0" style="padding:24px;text-align:center;font-size:13px;color:var(--dim)">No licenses issued</div>
  </div>

  <div v-if="viewing" class="card" style="margin-top:16px;padding:18px">
    <div style="display:flex;align-items:center;justify-content:space-between">
      <div class="mono" style="font-size:12.5px;font-weight:500">License {{ viewing.id }}</div>
      <div style="display:flex;gap:16px;font-size:13px">
        <span class="link" @click="copyFile">{{ copied ? "Copied" : "Copy" }}</span>
        <span class="link" @click="downloadFile">Download</span>
        <span class="link" @click="viewing = null">Close</span>
      </div>
    </div>
    <div style="margin-top:16px;display:grid;grid-template-columns:repeat(3,1fr);gap:6px 24px">
      <div><div class="label">Customer</div><div style="margin-top:3px;font-size:13.5px;font-weight:500">{{ viewing.payload.name }}</div></div>
      <div style="grid-column:span 2"><div class="label">Stands</div><div class="mono" style="margin-top:3px;font-size:12.5px">{{ viewing.payload.stands.join(", ") }}</div></div>
      <div><div class="label">Plan</div><div style="margin-top:3px;font-size:13.5px;font-weight:500">{{ viewing.payload.plan }}</div></div>
      <div><div class="label">Format</div><div style="margin-top:3px;font-size:13.5px;font-weight:500">v{{ viewing.payload.v }}</div></div>
      <div><div class="label">Expires</div><div style="margin-top:3px;font-size:13.5px;font-weight:500">{{ new Date(viewing.payload.exp).toISOString().slice(0,10) }}</div></div>
    </div>
    <div style="margin-top:16px;font-size:12px;color:var(--dim)">License file — the customer mounts this next to the engine:</div>
    <pre class="mono" style="margin:6px 0 0;border-radius:8px;background:var(--code-bg);border:1px solid var(--bd-soft);padding:14px;font-size:11.5px;line-height:1.8;color:var(--body);overflow-x:auto">{{ viewing.file }}</pre>
  </div>
</template>

<style scoped>
.row:hover { background:var(--hover); }
</style>
