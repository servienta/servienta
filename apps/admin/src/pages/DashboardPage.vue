<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";

const health = ref<{ ok: boolean; service: string; time: string } | null>(null);
const me = ref<string | null>(null);
const error = ref<string | null>(null);
const current = ref(""); const next = ref(""); const changed = ref(false);

onMounted(async () => {
  try { health.value = await api("/health"); me.value = (await api<{ email: string }>("/me")).email; }
  catch (e) { error.value = (e as Error).message; }
});
async function changePassword() {
  error.value = null; changed.value = false;
  try { await api("/auth/change", { method: "POST", body: JSON.stringify({ current: current.value, next: next.value }) });
    changed.value = true; current.value = next.value = ""; }
  catch (e) { error.value = (e as Error).message; }
}
</script>

<template>
  <div style="display:flex;align-items:baseline;gap:16px">
    <span class="mono" style="font-size:12px;color:var(--signal)">§1</span>
    <h1 style="margin:0;font-size:34px;font-weight:600;letter-spacing:-0.01em">Dashboard</h1>
  </div>

  <div style="margin-top:32px;display:grid;grid-template-columns:repeat(2,minmax(0,1fr));border-left:1px solid var(--rule)">
    <div style="border-right:1px solid var(--rule);border-top:2px solid var(--ink);padding:14px 20px">
      <div class="flabel">API</div>
      <div style="margin-top:12px;display:flex;align-items:center;gap:8px">
        <span :style="{ width:'8px',height:'8px',background: health?.ok ? 'var(--signal)' : 'var(--alert)' }"></span>
        <span style="font-size:24px;font-weight:600;line-height:1" :style="{ color: health?.ok ? 'var(--signal)' : 'var(--alert)' }">{{ health?.ok ? "healthy" : "unreachable" }}</span>
      </div>
      <div v-if="health" class="mono" style="margin-top:8px;font-size:11px;color:var(--ink3)">{{ health.time }}</div>
    </div>
    <div style="border-right:1px solid var(--rule);border-top:2px solid var(--ink);padding:14px 20px">
      <div class="flabel">Signed in as</div>
      <div class="mono" style="margin-top:12px;font-size:15px;line-height:1.3">{{ me ?? "—" }}</div>
      <div class="mono" style="margin-top:8px;font-size:11px;color:var(--ink3)">single-user session · D14</div>
    </div>
  </div>

  <div style="margin-top:44px;max-width:420px">
    <div class="flabel" style="padding-bottom:8px;border-bottom:2px solid var(--ink)">Change password</div>
    <form style="margin-top:18px;display:flex;flex-direction:column;gap:14px" @submit.prevent="changePassword">
      <label style="display:block"><span class="flabel">Current</span>
        <input v-model="current" type="password" required class="field" style="margin-top:5px;display:block;width:100%" /></label>
      <label style="display:block"><span class="flabel">New · 10+ chars</span>
        <input v-model="next" type="password" required minlength="10" class="field" style="margin-top:5px;display:block;width:100%" /></label>
      <button class="btn" style="align-self:flex-start">Change</button>
    </form>
    <p v-if="changed" class="mono" style="margin:12px 0 0;font-size:11.5px;color:var(--signal)">password changed</p>
    <p v-if="error" class="mono" style="margin:12px 0 0;font-size:11.5px;color:var(--alert)">{{ error }}</p>
  </div>
</template>
