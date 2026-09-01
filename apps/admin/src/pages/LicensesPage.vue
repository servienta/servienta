<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";
import { useCustomersStore } from "../stores/customers";
import { STANDS } from "../../shared/stands";
import { PLANS, planById } from "../../shared/plans";

interface LicenseRow { id: string; customerId: string; customerName: string; plan: string; stands: string[]; expiresAt: number; createdAt: number; }
interface LicensePayload { v: number; jti: string; sub: string; name: string; stands: string[]; plan: string; iat: number; exp: number; }

const store = useCustomersStore();
const rows = ref<LicenseRow[]>([]);
const customerId = ref(""); const plan = ref<string>("standard"); const stands = ref<string[]>([]); const expires = ref("");
const error = ref<string | null>(null);
const viewing = ref<{ id: string; file: string; payload: LicensePayload } | null>(null);
const copied = ref(false);

function applyPlan() { const p = planById(plan.value); if (!p) return;
  if (p.id !== "custom") stands.value = [...p.stands];
  expires.value = new Date(Date.now() + p.termDays * 86400000).toISOString().slice(0, 10); }
async function load() { rows.value = await api<LicenseRow[]>("/licenses"); }
onMounted(() => { applyPlan(); Promise.all([load(), store.loaded ? Promise.resolve() : store.load()]).catch((e) => (error.value = e.message)); });
function toggleStand(id: string) {
  if (plan.value !== "custom") return;
  const i = stands.value.indexOf(id);
  if (i >= 0) stands.value.splice(i, 1);
  else stands.value.push(id);
}

async function issue() { error.value = null;
  try { await api("/licenses", { method: "POST", body: JSON.stringify({ customerId: customerId.value, plan: plan.value, stands: stands.value, expiresAt: new Date(expires.value + "T00:00:00Z").getTime() }) }); await load(); }
  catch (e) { error.value = (e as Error).message; } }
async function view(id: string) { error.value = null; copied.value = false;
  try { const file = await api<{ payload_b64: string; signature: string }>(`/licenses/${id}/file`);
    viewing.value = { id, file: JSON.stringify(file, null, 2), payload: JSON.parse(atob(file.payload_b64)) as LicensePayload }; }
  catch (e) { error.value = (e as Error).message; } }
async function copyFile() { if (!viewing.value) return; await navigator.clipboard.writeText(viewing.value.file); copied.value = true; }
function downloadFile() { if (!viewing.value) return;
  const url = URL.createObjectURL(new Blob([viewing.value.file], { type: "application/json" }));
  const a = document.createElement("a"); a.href = url; a.download = `servienta-license-${viewing.value.id}.json`; a.click(); URL.revokeObjectURL(url); }
const cols = "minmax(0,1.1fr) 110px minmax(0,1.5fr) 110px 110px 70px";
</script>

<template>
  <div style="display:flex;align-items:baseline;gap:16px">
    <span class="mono" style="font-size:12px;color:var(--signal)">§3</span>
    <h1 style="margin:0;font-size:34px;font-weight:600;letter-spacing:-0.01em">Licenses</h1>
    <span class="mono" style="margin-left:auto;font-size:11px;color:var(--ink3)">Ed25519 · offline · D10</span>
  </div>

  <div style="margin-top:32px;border-top:2px solid var(--ink);padding-top:20px">
    <div class="flabel">Issue</div>
    <form style="margin-top:16px;display:flex;flex-wrap:wrap;align-items:flex-end;gap:16px" @submit.prevent="issue">
      <label style="display:block"><span class="flabel">Customer</span>
        <select v-model="customerId" required class="field" style="margin-top:5px;display:block;width:210px">
          <option v-for="c in store.items" :key="c.id" :value="c.id">{{ c.name }}</option></select></label>
      <label style="display:block"><span class="flabel">Plan</span>
        <select v-model="plan" class="field" style="margin-top:5px;display:block;width:160px" @change="applyPlan">
          <option v-for="pl in PLANS" :key="pl.id" :value="pl.id">{{ pl.id }}</option></select></label>
      <label style="display:block"><span class="flabel">Expires</span>
        <input v-model="expires" required type="date" class="field" style="margin-top:5px;display:block" /></label>
      <button class="btn">Issue</button>
    </form>

    <div style="margin-top:24px">
      <div class="flabel">Stands <span style="text-transform:none;letter-spacing:0" v-if="plan !== 'custom'">· set by plan</span></div>
      <div style="margin-top:10px;display:grid;grid-template-columns:repeat(5,minmax(0,1fr));border-top:1px solid var(--rule);border-left:1px solid var(--rule)">
        <div v-for="s in STANDS" :key="s.id" style="border-right:1px solid var(--rule);border-bottom:1px solid var(--rule);padding:9px 12px;display:flex;align-items:center;gap:9px"
          :style="{ background: stands.includes(s.id) ? 'rgba(47,122,82,0.05)' : 'transparent', cursor: plan === 'custom' ? 'pointer' : 'default' }"
          @click="toggleStand(s.id)">
          <span style="width:11px;height:11px;flex:none;border:1px solid"
            :style="{ borderColor: stands.includes(s.id) ? 'var(--signal)' : 'var(--rule)', background: stands.includes(s.id) ? 'var(--signal)' : 'transparent' }"></span>
          <span class="mono" style="font-size:11.5px;white-space:nowrap" :style="{ color: stands.includes(s.id) ? 'var(--ink)' : 'var(--ink3)' }">{{ s.id }}</span>
        </div>
      </div>
    </div>
  </div>
  <p v-if="error" class="mono" style="margin:12px 0 0;font-size:11.5px;color:var(--alert)">{{ error }}</p>

  <div style="margin-top:44px">
    <div class="flabel" :style="{ display:'grid', gridTemplateColumns:cols, gap:'16px', paddingBottom:'8px', borderBottom:'2px solid var(--ink)' }">
      <span>Customer</span><span>Plan</span><span>Stands</span><span>Expires</span><span>Issued</span><span></span>
    </div>
    <div v-for="l in rows" :key="l.id" class="rowh" :style="{ display:'grid', gridTemplateColumns:cols, gap:'16px', alignItems:'center', padding:'10px 0', borderBottom:'1px solid var(--rule2)' }">
      <span style="font-size:16px">{{ l.customerName }}</span>
      <span class="mono" style="font-size:11.5px" :style="{ color: l.plan === 'enterprise' ? 'var(--signal)' : 'var(--ink2)' }">{{ l.plan }}</span>
      <span class="mono" style="font-size:11px;color:var(--ink3);overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ l.stands.join(", ") }}</span>
      <span class="mono" style="font-size:12px;color:var(--ink2)">{{ new Date(l.expiresAt).toISOString().slice(0,10) }}</span>
      <span class="mono" style="font-size:12px;color:var(--ink3)">{{ new Date(l.createdAt).toISOString().slice(0,10) }}</span>
      <span class="act" style="text-align:right" @click="view(l.id)">View</span>
    </div>
    <div v-if="rows.length === 0" class="mono" style="padding:20px 0;font-size:12px;color:var(--ink3)">no licenses issued</div>
  </div>

  <div v-if="viewing" style="margin-top:44px;border-top:2px solid var(--ink);padding-top:20px">
    <div style="display:flex;align-items:baseline;justify-content:space-between">
      <div class="mono" style="font-size:12.5px">License {{ viewing.id }}</div>
      <div style="display:flex;gap:18px">
        <span class="act" @click="copyFile">{{ copied ? "Copied" : "Copy" }}</span>
        <span class="act" @click="downloadFile">Download</span>
        <span class="act" @click="viewing = null">Close</span>
      </div>
    </div>
    <div style="margin-top:20px;display:grid;grid-template-columns:repeat(3,minmax(0,1fr));border-left:1px solid var(--rule)">
      <div v-for="d in [['Customer', viewing.payload.name],['Plan', viewing.payload.plan],['Format', 'v'+viewing.payload.v],['Customer id', viewing.payload.sub],['Issued', new Date(viewing.payload.iat).toISOString().slice(0,10)],['Expires', new Date(viewing.payload.exp).toISOString().slice(0,10)]]" :key="d[0]"
        style="border-right:1px solid var(--rule);border-top:1px solid var(--rule);border-bottom:1px solid var(--rule);padding:10px 14px">
        <div class="flabel" style="font-size:10px">{{ d[0] }}</div>
        <div class="mono" style="margin-top:5px;font-size:12.5px">{{ d[1] }}</div>
      </div>
    </div>
    <div style="margin-top:18px;display:grid;grid-template-columns:180px minmax(0,1fr);gap:32px">
      <span class="flabel">Stands</span>
      <span class="mono" style="font-size:12.5px;line-height:1.8;color:var(--ink)">{{ viewing.payload.stands.join(", ") }}</span>
    </div>
    <div style="margin-top:20px;display:grid;grid-template-columns:180px minmax(0,1fr);gap:32px">
      <span class="flabel">License file</span>
      <pre class="mono" style="margin:0;font-size:11.5px;line-height:2;color:var(--ink2);overflow-x:auto">{{ viewing.file }}</pre>
    </div>
  </div>
</template>
