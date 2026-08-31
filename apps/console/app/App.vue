<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "./api";

const engineOk = ref<boolean | null>(null);
const version = ref<string | null>(null);
const endpoints = ref<Record<string, string>>({});
const error = ref<string | null>(null);
const received = ref<unknown[] | null>(null);
const service = ref("reference");
const resetMsg = ref<string | null>(null);

async function refresh() {
  error.value = null;
  try {
    version.value = (await api<{ contract: string }>("/engine/version")).contract;
    endpoints.value = await api<Record<string, string>>("/engine/endpoints");
    engineOk.value = true;
  } catch (e) {
    engineOk.value = false;
    error.value = (e as Error).message;
  }
}

async function loadReceived() {
  error.value = null;
  try {
    received.value = await api<unknown[]>(`/engine/received/${service.value}`);
  } catch (e) {
    error.value = (e as Error).message;
  }
}

async function reset() {
  resetMsg.value = null;
  try {
    await api("/engine/reset", { method: "POST" });
    resetMsg.value = "Stand reset to a known state.";
    received.value = null;
  } catch (e) {
    error.value = (e as Error).message;
  }
}

onMounted(refresh);
</script>

<template>
  <div class="min-h-screen bg-zinc-50 text-zinc-900">
    <header class="border-b border-zinc-200 bg-white">
      <div class="mx-auto flex max-w-5xl items-center gap-4 px-6 py-4">
        <span class="font-semibold tracking-tight">Servienta Console</span>
        <span
          class="ml-auto rounded-full px-2 py-0.5 text-xs font-medium"
          :class="engineOk === null ? 'bg-zinc-100 text-zinc-500' : engineOk ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'"
        >
          engine {{ engineOk === null ? "…" : engineOk ? `up · contract ${version}` : "unreachable" }}
        </span>
      </div>
    </header>

    <main class="mx-auto max-w-5xl space-y-6 px-6 py-8">
      <p v-if="error" class="rounded-md bg-red-50 px-4 py-2 text-sm text-red-700">{{ error }}</p>

      <section class="rounded-lg border border-zinc-200 bg-white p-4">
        <div class="flex items-center justify-between">
          <h2 class="font-semibold">Stand endpoints</h2>
          <button class="text-sm text-zinc-500 hover:text-zinc-900" @click="refresh">Refresh</button>
        </div>
        <table class="mt-3 w-full text-sm">
          <tbody>
            <tr v-for="(addr, name) in endpoints" :key="name" class="border-b border-zinc-100 last:border-0">
              <td class="py-1.5 font-medium">{{ name }}</td>
              <td class="py-1.5 font-mono text-zinc-600">{{ addr }}</td>
            </tr>
          </tbody>
        </table>
      </section>

      <section class="rounded-lg border border-zinc-200 bg-white p-4">
        <h2 class="font-semibold">Received messages</h2>
        <div class="mt-3 flex items-end gap-3">
          <label class="text-sm">
            <span class="text-zinc-500">Receiver</span>
            <input v-model="service" class="mt-1 block rounded-md border border-zinc-300 px-3 py-1.5" />
          </label>
          <button class="rounded-md bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700" @click="loadReceived">Load</button>
        </div>
        <pre v-if="received" class="mt-3 max-h-72 overflow-auto rounded-md bg-zinc-50 p-3 text-xs text-zinc-700">{{ JSON.stringify(received, null, 2) }}</pre>
      </section>

      <section class="rounded-lg border border-zinc-200 bg-white p-4">
        <h2 class="font-semibold">Reset</h2>
        <p class="mt-1 text-sm text-zinc-500">Clear recorded messages, faults, and run declarations across the whole stand.</p>
        <button class="mt-3 rounded-md border border-zinc-300 px-4 py-2 text-sm hover:border-zinc-500" @click="reset">Reset stand</button>
        <p v-if="resetMsg" class="mt-2 text-sm text-emerald-700">{{ resetMsg }}</p>
      </section>
    </main>
  </div>
</template>
