<script setup lang="ts">
import { RouterLink, RouterView } from "vue-router";
</script>

<template>
  <div class="min-h-screen bg-zinc-50 text-zinc-900">
    <header class="border-b border-zinc-200 bg-white">
      <div class="mx-auto flex max-w-5xl items-center gap-8 px-6 py-4">
        <span class="font-semibold tracking-tight">Servienta Admin</span>
        <nav class="flex gap-4 text-sm">
          <RouterLink to="/" class="text-zinc-500 hover:text-zinc-900" active-class="text-zinc-900 font-medium">Dashboard</RouterLink>
          <RouterLink to="/customers" class="text-zinc-500 hover:text-zinc-900" active-class="text-zinc-900 font-medium">Customers</RouterLink>
          <RouterLink to="/licenses" class="text-zinc-500 hover:text-zinc-900" active-class="text-zinc-900 font-medium">Licenses</RouterLink>
        </nav>
      </div>
    </header>
    <main class="mx-auto max-w-5xl px-6 py-8">
      <RouterView />
    </main>
  </div>
</template>
