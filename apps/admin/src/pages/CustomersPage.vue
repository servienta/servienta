<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useCustomersStore } from "../stores/customers";

const store = useCustomersStore();
const name = ref(""); const email = ref(""); const error = ref<string | null>(null);
const editingId = ref<string | null>(null); const editName = ref(""); const editEmail = ref("");

onMounted(() => store.load().catch((e) => (error.value = e.message)));
async function add() { error.value = null;
  try { await store.create(name.value, email.value); name.value = ""; email.value = ""; } catch (e) { error.value = (e as Error).message; } }
function startEdit(id: string) { const c = store.items.find((c) => c.id === id); if (!c) return; editingId.value = id; editName.value = c.name; editEmail.value = c.email; }
async function saveEdit() { if (!editingId.value) return; error.value = null;
  try { await store.update(editingId.value, editName.value, editEmail.value); editingId.value = null; } catch (e) { error.value = (e as Error).message; } }
async function remove(id: string, n: string) { if (!confirm(`Delete ${n} and every license issued to them?`)) return; error.value = null;
  try { await store.remove(id); } catch (e) { error.value = (e as Error).message; } }
const cols = "minmax(0,1.1fr) minmax(0,1.4fr) 120px 160px";
</script>

<template>
  <div style="display:flex;align-items:baseline;gap:16px">
    <span class="mono" style="font-size:12px;color:var(--signal)">§2</span>
    <h1 style="margin:0;font-size:34px;font-weight:600;letter-spacing:-0.01em">Customers</h1>
    <span class="mono" style="margin-left:auto;font-size:11px;color:var(--ink3)">{{ store.items.length }} records</span>
  </div>

  <form style="margin-top:32px;display:flex;flex-wrap:wrap;align-items:flex-end;gap:16px" @submit.prevent="add">
    <label style="display:block"><span class="flabel">Name</span><input v-model="name" required class="field" style="margin-top:5px;display:block;width:240px" /></label>
    <label style="display:block"><span class="flabel">Email</span><input v-model="email" required type="email" class="field" style="margin-top:5px;display:block;width:280px" /></label>
    <button class="btn">Add</button>
  </form>
  <p v-if="error" class="mono" style="margin:12px 0 0;font-size:11.5px;color:var(--alert)">{{ error }}</p>

  <div style="margin-top:36px">
    <div class="flabel" :style="{ display:'grid', gridTemplateColumns:cols, gap:'20px', paddingBottom:'8px', borderBottom:'2px solid var(--ink)' }">
      <span>Name</span><span>Email</span><span>Created</span><span></span>
    </div>
    <div v-for="c in store.items" :key="c.id" class="rowh" :style="{ display:'grid', gridTemplateColumns:cols, gap:'20px', alignItems:'center', padding:'10px 0', borderBottom:'1px solid var(--rule2)' }">
      <template v-if="editingId === c.id">
        <input v-model="editName" class="field" style="width:100%;padding:6px 9px;border-color:var(--ink)" />
        <input v-model="editEmail" type="email" class="field" style="width:100%;padding:6px 9px" />
        <span class="mono" style="font-size:12px;color:var(--ink3)">{{ new Date(c.createdAt).toISOString().slice(0,10) }}</span>
        <span style="display:flex;justify-content:flex-end;gap:16px">
          <span class="act" style="color:var(--ink)" @click="saveEdit">Save</span>
          <span class="act" @click="editingId = null">Cancel</span>
        </span>
      </template>
      <template v-else>
        <span style="font-size:16px">{{ c.name }}</span>
        <span class="mono" style="font-size:12.5px;color:var(--ink2)">{{ c.email }}</span>
        <span class="mono" style="font-size:12px;color:var(--ink3)">{{ new Date(c.createdAt).toISOString().slice(0,10) }}</span>
        <span style="display:flex;justify-content:flex-end;gap:16px">
          <span class="act" @click="startEdit(c.id)">Edit</span>
          <span class="act" style="color:var(--alert)" @click="remove(c.id, c.name)">Delete</span>
        </span>
      </template>
    </div>
    <div v-if="store.loaded && store.items.length === 0" class="mono" style="padding:20px 0;font-size:12px;color:var(--ink3)">no records</div>
  </div>
</template>
