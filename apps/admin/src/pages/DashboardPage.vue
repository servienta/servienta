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
  <h1 class="text-xl font-semibold">Dashboard</h1>
  <div class="mt-6 grid gap-4 sm:grid-cols-2">
    <div class="rounded-lg border border-zinc-200 bg-white p-4">
      <div class="text-sm text-zinc-500">API</div>
      <div class="mt-1 font-medium" :class="health?.ok ? 'text-emerald-600' : 'text-red-600'">
        {{ health?.ok ? "healthy" : "unreachable" }}
      </div>
      <div v-if="health" class="mt-1 text-xs text-zinc-400">{{ health.time }}</div>
    </div>
    <div class="rounded-lg border border-zinc-200 bg-white p-4">
      <div class="text-sm text-zinc-500">Signed in as</div>
      <div class="mt-1 font-medium">{{ me ?? "—" }}</div>
    </div>
  </div>

  <div class="mt-6 max-w-sm rounded-lg border border-zinc-200 bg-white p-4">
    <div class="text-sm font-medium">Change password</div>
    <form class="mt-3 space-y-2" @submit.prevent="changePassword">
      <input v-model="current" type="password" required placeholder="Current password" class="w-full rounded-md border border-zinc-300 px-3 py-1.5 text-sm" />
      <input v-model="next" type="password" required minlength="10" placeholder="New password (10+ chars)" class="w-full rounded-md border border-zinc-300 px-3 py-1.5 text-sm" />
      <button class="rounded-md bg-zinc-900 px-4 py-1.5 text-sm font-medium text-white hover:bg-zinc-700">Change</button>
    </form>
    <p v-if="changed" class="mt-2 text-sm text-emerald-700">Password changed.</p>
    <p v-if="error" class="mt-2 text-sm text-red-600">{{ error }}</p>
  </div>
</template>
