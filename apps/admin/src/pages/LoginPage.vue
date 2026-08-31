<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "../api";
import Mark from "../components/Mark.vue";

const router = useRouter();
const email = ref("");
const password = ref("");
const error = ref<string | null>(null);
const notice = ref<string | null>(null);

async function login() {
  error.value = null;
  try {
    await api("/auth/login", { method: "POST", body: JSON.stringify({ email: email.value, password: password.value }) });
    router.push("/");
  } catch (e) {
    error.value = (e as Error).message;
  }
}
async function forgot() {
  error.value = null;
  if (!email.value) { error.value = "enter your email first"; return; }
  await api("/auth/forgot", { method: "POST", body: JSON.stringify({ email: email.value }) }).catch(() => {});
  notice.value = "If the account exists, a reset link was sent.";
}
</script>

<template>
  <div style="max-width:360px;margin:64px auto 0">
    <div class="card" style="border-radius:12px;padding:28px">
      <div style="display:flex;align-items:center;gap:9px">
        <Mark :size="4" />
        <h1 style="margin:0;font-size:17px;font-weight:600;letter-spacing:-0.02em">Servienta Admin</h1>
      </div>
      <form style="margin-top:20px;display:flex;flex-direction:column;gap:10px" @submit.prevent="login">
        <input v-model="email" type="email" required placeholder="Email" class="inp" style="width:100%;padding:10px 12px" />
        <input v-model="password" type="password" required placeholder="Password" class="inp" style="width:100%;padding:10px 12px" />
        <button class="btn" style="width:100%;padding:10px">Sign in</button>
      </form>
      <div class="link" style="margin-top:14px;font-size:12px" @click="forgot">Forgot password?</div>
    </div>
    <div v-if="notice" class="card" style="margin-top:16px;background:var(--notice-bg);padding:12px 14px">
      <div style="font-size:12.5px;color:var(--notice-fg)">{{ notice }}</div>
    </div>
    <div v-if="error" style="margin-top:10px;border:1px solid var(--err-bd);border-radius:10px;background:var(--err-bg);padding:12px 14px">
      <div style="font-size:12.5px;color:var(--err)">{{ error }}</div>
    </div>
  </div>
</template>
