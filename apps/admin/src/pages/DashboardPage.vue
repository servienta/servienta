<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";

const health = ref<{ ok: boolean; service: string; time: string } | null>(null);
const me = ref<string | null>(null);
const error = ref<string | null>(null);

onMounted(async () => {
  try {
    health.value = await api("/health");
    me.value = (await api<{ email: string }>("/me")).email;
  } catch (e) {
    error.value = (e as Error).message;
  }
});
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
      <div class="text-sm text-zinc-500">Signed in via Cloudflare Access</div>
      <div class="mt-1 font-medium">{{ me ?? "—" }}</div>
      <div v-if="error" class="mt-1 text-xs text-red-500">{{ error }}</div>
    </div>
  </div>
</template>
