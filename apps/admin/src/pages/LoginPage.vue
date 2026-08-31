<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "../api";

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
  if (!email.value) {
    error.value = "enter your email first";
    return;
  }
  await api("/auth/forgot", { method: "POST", body: JSON.stringify({ email: email.value }) }).catch(() => {});
  notice.value = "If the account exists, a reset link was sent.";
}
</script>

<template>
  <div class="mx-auto mt-24 max-w-sm rounded-lg border border-zinc-200 bg-white p-6">
    <h1 class="text-lg font-semibold">Servienta Admin</h1>
    <form class="mt-4 space-y-3" @submit.prevent="login">
      <input v-model="email" type="email" required placeholder="Email" class="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm" />
      <input v-model="password" type="password" required placeholder="Password" class="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm" />
      <button class="w-full rounded-md bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700">Sign in</button>
    </form>
    <button class="mt-3 text-xs text-zinc-500 hover:text-zinc-800" @click="forgot">Forgot password?</button>
    <p v-if="error" class="mt-3 text-sm text-red-600">{{ error }}</p>
    <p v-if="notice" class="mt-3 text-sm text-emerald-700">{{ notice }}</p>
  </div>
</template>
