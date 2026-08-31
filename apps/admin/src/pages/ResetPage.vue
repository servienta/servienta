<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";
import Mark from "../components/Mark.vue";

const route = useRoute();
const router = useRouter();
const next = ref("");
const error = ref<string | null>(null);

async function reset() {
  error.value = null;
  try {
    await api("/auth/reset", { method: "POST", body: JSON.stringify({ token: String(route.query.token ?? ""), next: next.value }) });
    router.push("/login");
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <div style="max-width:360px;margin:64px auto 0">
    <div class="card" style="border-radius:12px;padding:28px">
      <div style="display:flex;align-items:center;gap:9px">
        <Mark :size="4" />
        <h1 style="margin:0;font-size:17px;font-weight:600;letter-spacing:-0.02em">Set a new password</h1>
      </div>
      <form style="margin-top:20px;display:flex;flex-direction:column;gap:10px" @submit.prevent="reset">
        <input v-model="next" type="password" required minlength="10" placeholder="New password (10+ chars)" class="inp" style="width:100%;padding:10px 12px" />
        <button class="btn" style="width:100%;padding:10px">Save</button>
      </form>
      <p v-if="error" style="margin:14px 0 0;font-size:12.5px;color:var(--err)">{{ error }}</p>
    </div>
  </div>
</template>
