<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";
import { useCustomersStore } from "../stores/customers";
import { STANDS } from "../../shared/stands";

interface LicenseRow {
  id: string;
  customerId: string;
  customerName: string;
  stands: string[];
  expiresAt: number;
  createdAt: number;
}

const store = useCustomersStore();
const rows = ref<LicenseRow[]>([]);
const customerId = ref("");
const stands = ref<string[]>([]);
const expires = ref("");
const issued = ref<string | null>(null);
const error = ref<string | null>(null);

interface LicensePayload {
  v: number;
  jti: string;
  sub: string;
  name: string;
  stands: string[];
  iat: number;
  exp: number;
}
const viewing = ref<{ id: string; file: string; payload: LicensePayload } | null>(null);
const copied = ref(false);

async function view(id: string) {
  error.value = null;
  copied.value = false;
  try {
    const file = await api<{ payload_b64: string; signature: string }>(`/licenses/${id}/file`);
    viewing.value = {
      id,
      file: JSON.stringify(file, null, 2),
      payload: JSON.parse(atob(file.payload_b64)) as LicensePayload,
    };
  } catch (e) {
    error.value = (e as Error).message;
  }
}

async function copyFile() {
  if (!viewing.value) return;
  await navigator.clipboard.writeText(viewing.value.file);
  copied.value = true;
}

function downloadFile() {
  if (!viewing.value) return;
  const url = URL.createObjectURL(new Blob([viewing.value.file], { type: "application/json" }));
  const a = document.createElement("a");
  a.href = url;
  a.download = `servienta-license-${viewing.value.id}.json`;
  a.click();
  URL.revokeObjectURL(url);
}

async function load() {
  rows.value = await api<LicenseRow[]>("/licenses");
}

onMounted(() => {
  Promise.all([load(), store.loaded ? Promise.resolve() : store.load()]).catch(
    (e) => (error.value = e.message),
  );
});

async function issue() {
  error.value = null;
  issued.value = null;
  try {
    const file = await api<{ payload_b64: string; signature: string }>("/licenses", {
      method: "POST",
      body: JSON.stringify({
        customerId: customerId.value,
        stands: stands.value,
        expiresAt: new Date(expires.value + "T00:00:00Z").getTime(),
      }),
    });
    issued.value = JSON.stringify(file, null, 2);
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <h1 class="text-xl font-semibold">Licenses</h1>

  <form class="mt-6 flex flex-wrap items-end gap-3" @submit.prevent="issue">
    <label class="block text-sm">
      <span class="text-zinc-500">Customer</span>
      <select v-model="customerId" required class="mt-1 block rounded-md border border-zinc-300 px-3 py-1.5">
        <option v-for="c in store.items" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
    </label>
    <fieldset class="block text-sm">
      <legend class="text-zinc-500">Stands</legend>
      <div class="mt-1 grid max-w-xl grid-cols-2 gap-x-4 gap-y-1 rounded-md border border-zinc-300 p-3 sm:grid-cols-3">
        <label v-for="s in STANDS" :key="s.id" class="flex items-center gap-2">
          <input v-model="stands" type="checkbox" :value="s.id" class="rounded border-zinc-300" />
          <span>{{ s.label }}</span>
        </label>
      </div>
    </fieldset>
    <label class="block text-sm">
      <span class="text-zinc-500">Expires</span>
      <input v-model="expires" required type="date" class="mt-1 block rounded-md border border-zinc-300 px-3 py-1.5" />
    </label>
    <button class="rounded-md bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700">Issue</button>
  </form>
  <p v-if="error" class="mt-2 text-sm text-red-600">{{ error }}</p>

  <div v-if="issued" class="mt-4 rounded-lg border border-emerald-200 bg-emerald-50 p-4">
    <div class="text-sm font-medium text-emerald-800">License file — hand this to the customer (mounted next to the engine):</div>
    <pre class="mt-2 overflow-x-auto text-xs text-zinc-700">{{ issued }}</pre>
  </div>

  <div class="mt-6 overflow-x-auto rounded-lg border border-zinc-200 bg-white">
    <table class="w-full text-sm">
      <thead class="border-b border-zinc-200 text-left text-zinc-500">
        <tr><th class="px-4 py-2">Customer</th><th class="px-4 py-2">Stands</th><th class="px-4 py-2">Expires</th><th class="px-4 py-2">Issued</th><th class="px-4 py-2"></th></tr>
      </thead>
      <tbody>
        <tr v-for="l in rows" :key="l.id" class="border-b border-zinc-100 last:border-0">
          <td class="px-4 py-2 font-medium">{{ l.customerName }}</td>
          <td class="px-4 py-2 text-zinc-600">{{ l.stands.join(", ") }}</td>
          <td class="px-4 py-2">{{ new Date(l.expiresAt).toISOString().slice(0, 10) }}</td>
          <td class="px-4 py-2 text-zinc-500">{{ new Date(l.createdAt).toISOString().slice(0, 10) }}</td>
          <td class="px-4 py-2 text-right">
            <button class="text-sm text-zinc-500 underline-offset-2 hover:text-zinc-900 hover:underline" @click="view(l.id)">View</button>
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td colspan="5" class="px-4 py-6 text-center text-zinc-400">No licenses issued</td>
        </tr>
      </tbody>
    </table>
  </div>

  <div v-if="viewing" class="mt-6 rounded-lg border border-zinc-200 bg-white p-4">
    <div class="flex items-center justify-between">
      <div class="text-sm font-medium">License {{ viewing.id }}</div>
      <div class="flex gap-3 text-sm">
        <button class="text-zinc-500 hover:text-zinc-900" @click="copyFile">{{ copied ? "Copied" : "Copy" }}</button>
        <button class="text-zinc-500 hover:text-zinc-900" @click="downloadFile">Download</button>
        <button class="text-zinc-500 hover:text-zinc-900" @click="viewing = null">Close</button>
      </div>
    </div>
    <dl class="mt-3 grid grid-cols-2 gap-x-6 gap-y-1 text-sm sm:grid-cols-3">
      <div><dt class="text-zinc-500">Customer</dt><dd class="font-medium">{{ viewing.payload.name }}</dd></div>
      <div class="col-span-2"><dt class="text-zinc-500">Stands</dt><dd class="font-medium">{{ viewing.payload.stands.join(", ") }}</dd></div>
      <div><dt class="text-zinc-500">Format</dt><dd class="font-medium">v{{ viewing.payload.v }}</dd></div>
      <div><dt class="text-zinc-500">Issued</dt><dd class="font-medium">{{ new Date(viewing.payload.iat).toISOString().slice(0, 10) }}</dd></div>
      <div><dt class="text-zinc-500">Expires</dt><dd class="font-medium">{{ new Date(viewing.payload.exp).toISOString().slice(0, 10) }}</dd></div>
      <div><dt class="text-zinc-500">Customer id</dt><dd class="font-mono text-xs">{{ viewing.payload.sub }}</dd></div>
    </dl>
    <div class="mt-3 text-xs text-zinc-500">License file — the customer mounts this next to the engine:</div>
    <pre class="mt-1 overflow-x-auto rounded-md bg-zinc-50 p-3 text-xs text-zinc-700">{{ viewing.file }}</pre>
  </div>
</template>
