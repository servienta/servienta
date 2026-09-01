<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";

const route = useRoute(); const router = useRouter();
const next = ref(""); const error = ref<string | null>(null);
async function reset() {
  error.value = null;
  try { await api("/auth/reset", { method: "POST", body: JSON.stringify({ token: String(route.query.token ?? ""), next: next.value }) }); router.push("/login"); }
  catch (e) { error.value = (e as Error).message; }
}
</script>

<template>
  <div style="max-width:400px;margin:72px auto 0">
    <div style="display:flex;align-items:baseline;gap:12px;padding-bottom:12px;border-bottom:2px solid var(--ink)">
      <span style="font-size:17px;font-weight:700;letter-spacing:0.16em;text-transform:uppercase">Servienta</span>
      <span class="mono" style="font-size:10.5px;letter-spacing:0.1em;text-transform:uppercase;color:var(--signal)">admin</span>
    </div>
    <form style="margin-top:24px;display:flex;flex-direction:column;gap:14px" @submit.prevent="reset">
      <label style="display:block"><span class="flabel">New password · 10+ chars</span>
        <input v-model="next" type="password" required minlength="10" class="field" style="margin-top:5px;display:block;width:100%;padding:10px 12px" /></label>
      <button class="btn" style="width:100%;padding:11px">Save</button>
    </form>
    <p v-if="error" class="mono" style="margin:14px 0 0;font-size:11.5px;color:var(--alert)">{{ error }}</p>
  </div>
</template>
