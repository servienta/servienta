<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";

const health = ref<{ ok: boolean; service: string; time: string } | null>(null);
const me = ref<string | null>(null);
const error = ref<string | null>(null);
const current = ref("");
const next = ref("");
const changed = ref(false);

onMounted(async () => {
  try {
    health.value = await api("/health");
    me.value = (await api<{ email: string }>("/me")).email;
  } catch (e) {
    error.value = (e as Error).message;
  }
});

async function changePassword() {
  error.value = null;
  changed.value = false;
  try {
    await api("/auth/change", { method: "POST", body: JSON.stringify({ current: current.value, next: next.value }) });
    changed.value = true;
    current.value = next.value = "";
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <h1 style="margin:0;font-size:24px;font-weight:600;letter-spacing:-0.03em">Dashboard</h1>
  <p v-if="error" style="margin:12px 0 0;font-size:13px;color:var(--err)">{{ error }}</p>

  <div style="margin-top:24px;display:grid;grid-template-columns:repeat(2,1fr);gap:16px">
    <div class="card" style="padding:18px">
      <div class="label">API</div>
      <div style="margin-top:10px;display:flex;align-items:center;gap:8px">
        <span :style="{ width:'7px',height:'7px',borderRadius:'999px',background: health?.ok ? 'var(--ok-dot)' : 'var(--err)' }"></span>
        <span style="font-size:17px;font-weight:500" :style="{ color: health?.ok ? 'var(--ok)' : 'var(--err)' }">{{ health?.ok ? "healthy" : "unreachable" }}</span>
      </div>
      <div v-if="health" class="mono" style="margin-top:6px;font-size:11.5px;color:var(--dim)">{{ health.time }}</div>
    </div>
    <div class="card" style="padding:18px">
      <div class="label">Signed in as</div>
      <div style="margin-top:10px;font-size:17px;font-weight:500">{{ me ?? "—" }}</div>
      <div class="mono" style="margin-top:6px;font-size:11.5px;color:var(--dim)">single-user session · D14</div>
    </div>
  </div>

  <div class="card" style="margin-top:16px;max-width:400px;padding:18px">
    <div style="font-size:14px;font-weight:600">Change password</div>
    <form style="margin-top:14px;display:flex;flex-direction:column;gap:8px" @submit.prevent="changePassword">
      <input v-model="current" type="password" required placeholder="Current password" class="inp" style="width:100%" />
      <input v-model="next" type="password" required minlength="10" placeholder="New password (10+ chars)" class="inp" style="width:100%" />
      <button class="btn" style="align-self:flex-start;margin-top:4px">Change</button>
    </form>
    <p v-if="changed" style="margin:10px 0 0;font-size:13px;color:var(--ok)">Password changed.</p>
    <p v-if="error" style="margin:10px 0 0;font-size:13px;color:var(--err)">{{ error }}</p>
  </div>
</template>
