<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useCustomersStore } from "../stores/customers";

const store = useCustomersStore();
const name = ref("");
const email = ref("");
const error = ref<string | null>(null);

const editingId = ref<string | null>(null);
const editName = ref("");
const editEmail = ref("");

onMounted(() => store.load().catch((e) => (error.value = e.message)));

async function add() {
  error.value = null;
  try {
    await store.create(name.value, email.value);
    name.value = "";
    email.value = "";
  } catch (e) {
    error.value = (e as Error).message;
  }
}

function startEdit(id: string) {
  const c = store.items.find((c) => c.id === id);
  if (!c) return;
  editingId.value = id;
  editName.value = c.name;
  editEmail.value = c.email;
}

async function saveEdit() {
  if (!editingId.value) return;
  error.value = null;
  try {
    await store.update(editingId.value, editName.value, editEmail.value);
    editingId.value = null;
  } catch (e) {
    error.value = (e as Error).message;
  }
}

async function remove(id: string, customerName: string) {
  if (!confirm(`Delete ${customerName} and every license issued to them?`)) return;
  error.value = null;
  try {
    await store.remove(id);
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <h1 class="text-xl font-semibold">Customers</h1>

  <form class="mt-6 flex flex-wrap items-end gap-3" @submit.prevent="add">
    <label class="block text-sm">
      <span class="text-zinc-500">Name</span>
      <input v-model="name" required class="mt-1 block rounded-md border border-zinc-300 px-3 py-1.5" />
    </label>
    <label class="block text-sm">
      <span class="text-zinc-500">Email</span>
      <input v-model="email" required type="email" class="mt-1 block rounded-md border border-zinc-300 px-3 py-1.5" />
    </label>
    <button class="rounded-md bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700">Add</button>
  </form>
  <p v-if="error" class="mt-2 text-sm text-red-600">{{ error }}</p>

  <div class="mt-6 overflow-x-auto rounded-lg border border-zinc-200 bg-white">
    <table class="w-full text-sm">
      <thead class="border-b border-zinc-200 text-left text-zinc-500">
        <tr><th class="px-4 py-2">Name</th><th class="px-4 py-2">Email</th><th class="px-4 py-2">Created</th><th class="px-4 py-2"></th></tr>
      </thead>
      <tbody>
        <tr v-for="c in store.items" :key="c.id" class="border-b border-zinc-100 last:border-0">
          <template v-if="editingId === c.id">
            <td class="px-4 py-2"><input v-model="editName" class="w-full rounded-md border border-zinc-300 px-2 py-1" /></td>
            <td class="px-4 py-2"><input v-model="editEmail" type="email" class="w-full rounded-md border border-zinc-300 px-2 py-1" /></td>
            <td class="px-4 py-2 text-zinc-500">{{ new Date(c.createdAt).toISOString().slice(0, 10) }}</td>
            <td class="px-4 py-2 text-right whitespace-nowrap">
              <button class="text-sm font-medium text-zinc-900 hover:underline" @click="saveEdit">Save</button>
              <button class="ml-3 text-sm text-zinc-500 hover:text-zinc-900" @click="editingId = null">Cancel</button>
            </td>
          </template>
          <template v-else>
            <td class="px-4 py-2 font-medium">{{ c.name }}</td>
            <td class="px-4 py-2">{{ c.email }}</td>
            <td class="px-4 py-2 text-zinc-500">{{ new Date(c.createdAt).toISOString().slice(0, 10) }}</td>
            <td class="px-4 py-2 text-right whitespace-nowrap">
              <button class="text-sm text-zinc-500 hover:text-zinc-900 hover:underline" @click="startEdit(c.id)">Edit</button>
              <button class="ml-3 text-sm text-red-500 hover:text-red-700 hover:underline" @click="remove(c.id, c.name)">Delete</button>
            </td>
          </template>
        </tr>
        <tr v-if="store.loaded && store.items.length === 0">
          <td colspan="4" class="px-4 py-6 text-center text-zinc-400">No customers yet</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
