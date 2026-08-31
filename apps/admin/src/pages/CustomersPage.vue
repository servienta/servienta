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
  try { await store.create(name.value, email.value); name.value = ""; email.value = ""; }
  catch (e) { error.value = (e as Error).message; }
}
function startEdit(id: string) {
  const c = store.items.find((c) => c.id === id);
  if (!c) return;
  editingId.value = id; editName.value = c.name; editEmail.value = c.email;
}
async function saveEdit() {
  if (!editingId.value) return;
  error.value = null;
  try { await store.update(editingId.value, editName.value, editEmail.value); editingId.value = null; }
  catch (e) { error.value = (e as Error).message; }
}
async function remove(id: string, customerName: string) {
  if (!confirm(`Delete ${customerName} and every license issued to them?`)) return;
  error.value = null;
  try { await store.remove(id); } catch (e) { error.value = (e as Error).message; }
}
const cols = "1.1fr 1.4fr 110px 150px";
</script>

<template>
  <h1 style="margin:0;font-size:24px;font-weight:600;letter-spacing:-0.03em">Customers</h1>

  <form style="margin-top:24px;display:flex;flex-wrap:wrap;align-items:flex-end;gap:12px" @submit.prevent="add">
    <label style="display:block;font-size:12.5px">
      <span class="label">Name</span>
      <input v-model="name" required class="inp" style="margin-top:5px;display:block;width:220px" />
    </label>
    <label style="display:block;font-size:12.5px">
      <span class="label">Email</span>
      <input v-model="email" required type="email" class="inp" style="margin-top:5px;display:block;width:260px" />
    </label>
    <button class="btn">Add</button>
  </form>
  <p v-if="error" style="margin:10px 0 0;font-size:13px;color:var(--err)">{{ error }}</p>

  <div class="card" style="margin-top:24px;overflow:hidden">
    <div class="label" :style="{ display:'grid', gridTemplateColumns:cols, gap:'16px', padding:'10px 16px', borderBottom:'1px solid var(--bd)' }">
      <span>Name</span><span>Email</span><span>Created</span><span></span>
    </div>
    <div v-for="c in store.items" :key="c.id" class="row" :style="{ display:'grid', gridTemplateColumns:cols, gap:'16px', alignItems:'center', padding:'11px 16px', borderBottom:'1px solid var(--bd-soft)' }">
      <template v-if="editingId === c.id">
        <input v-model="editName" class="inp" style="width:100%;padding:6px 9px;border-color:var(--acc-h)" />
        <input v-model="editEmail" type="email" class="inp" style="width:100%;padding:6px 9px" />
        <span class="mono" style="font-size:12px;color:var(--dim)">{{ new Date(c.createdAt).toISOString().slice(0,10) }}</span>
        <span style="display:flex;justify-content:flex-end;gap:14px">
          <span style="font-size:13px;font-weight:500;color:var(--fg);cursor:pointer" @click="saveEdit">Save</span>
          <span class="link" style="font-size:13px" @click="editingId = null">Cancel</span>
        </span>
      </template>
      <template v-else>
        <span style="font-size:13.5px;font-weight:500">{{ c.name }}</span>
        <span style="font-size:13.5px;color:var(--body)">{{ c.email }}</span>
        <span class="mono" style="font-size:12px;color:var(--dim)">{{ new Date(c.createdAt).toISOString().slice(0,10) }}</span>
        <span style="display:flex;justify-content:flex-end;gap:14px">
          <span class="link" style="font-size:13px" @click="startEdit(c.id)">Edit</span>
          <span style="font-size:13px;color:var(--err);cursor:pointer" @click="remove(c.id, c.name)">Delete</span>
        </span>
      </template>
    </div>
    <div v-if="store.loaded && store.items.length === 0" style="padding:24px;text-align:center;font-size:13px;color:var(--dim)">No customers yet</div>
  </div>
</template>

<style scoped>
.row:hover { background:var(--hover); }
</style>
