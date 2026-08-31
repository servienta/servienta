<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";

const route = useRoute();
const router = useRouter();
const next = ref("");
const error = ref<string | null>(null);

async function reset() {
  error.value = null;
  try {
    await api("/auth/reset", {
      method: "POST",
      body: JSON.stringify({ token: String(route.query.token ?? ""), next: next.value }),
    });
    router.push("/login");
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <div class="mx-auto mt-24 max-w-sm rounded-lg border border-zinc-200 bg-white p-6">
    <h1 class="text-lg font-semibold">Set a new password</h1>
    <form class="mt-4 space-y-3" @submit.prevent="reset">
      <input v-model="next" type="password" required minlength="12" placeholder="New password (12+ chars)" class="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm" />
      <button class="w-full rounded-md bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700">Save</button>
    </form>
    <p v-if="error" class="mt-3 text-sm text-red-600">{{ error }}</p>
  </div>
</template>
