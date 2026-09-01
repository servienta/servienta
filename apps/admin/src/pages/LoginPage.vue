<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "../api";

const router = useRouter();
const email = ref(""); const password = ref("");
const error = ref<string | null>(null); const notice = ref<string | null>(null);

async function login() {
  error.value = null;
  try { await api("/auth/login", { method: "POST", body: JSON.stringify({ email: email.value, password: password.value }) }); router.push("/"); }
  catch (e) { error.value = (e as Error).message; }
}
async function forgot() {
  error.value = null;
  if (!email.value) { error.value = "enter your email first"; return; }
  await api("/auth/forgot", { method: "POST", body: JSON.stringify({ email: email.value }) }).catch(() => {});
  notice.value = "If the account exists, a reset link was sent.";
}
</script>

<template>
  <div style="max-width:400px;margin:72px auto 0">
    <div style="display:flex;align-items:baseline;gap:12px;padding-bottom:12px;border-bottom:2px solid var(--ink)">
      <span style="font-size:17px;font-weight:700;letter-spacing:0.16em;text-transform:uppercase">Servienta</span>
      <span class="mono" style="font-size:10.5px;letter-spacing:0.1em;text-transform:uppercase;color:var(--signal)">admin</span>
    </div>
    <form style="margin-top:24px;display:flex;flex-direction:column;gap:14px" @submit.prevent="login">
      <label style="display:block"><span class="flabel">Email</span>
        <input v-model="email" type="email" required class="field" style="margin-top:5px;display:block;width:100%;padding:10px 12px" /></label>
      <label style="display:block"><span class="flabel">Password</span>
        <input v-model="password" type="password" required class="field" style="margin-top:5px;display:block;width:100%;padding:10px 12px" /></label>
      <button class="btn" style="width:100%;padding:11px">Sign in</button>
    </form>
    <div class="act" style="margin-top:16px" @click="forgot">Forgot password?</div>
    <div v-if="notice" style="margin-top:32px;border-left:2px solid var(--signal);padding:10px 14px;background:var(--band)">
      <div class="mono" style="font-size:11.5px;color:var(--signal)">{{ notice }}</div>
    </div>
    <div v-if="error" style="margin-top:10px;border-left:2px solid var(--alert);padding:10px 14px;background:var(--band)">
      <div class="mono" style="font-size:11.5px;color:var(--alert)">{{ error }}</div>
    </div>
  </div>
</template>
