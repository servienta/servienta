<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from "vue-router";
import { api } from "./api";
import Mark from "./components/Mark.vue";
import ThemeToggle from "./components/ThemeToggle.vue";

const route = useRoute();
const router = useRouter();

async function signOut() {
  await api("/auth/logout", { method: "POST" }).catch(() => {});
  router.push("/login");
}
</script>

<template>
  <div style="min-height:100vh">
    <div v-if="!route.meta.public" style="border-bottom:1px solid var(--bd);background:var(--surface)">
      <div style="max-width:1000px;margin:0 auto;padding:0 24px;height:56px;display:flex;align-items:center;gap:28px">
        <div style="display:flex;align-items:center;gap:9px">
          <Mark :size="3" />
          <span style="font-size:15px;font-weight:600;letter-spacing:-0.02em">Servienta</span>
          <span class="mono" style="font-size:10.5px;letter-spacing:0.1em;text-transform:uppercase;color:#8b6cff;border:1px solid rgba(139,108,255,0.4);border-radius:4px;padding:2px 6px">admin</span>
        </div>
        <nav style="display:flex;gap:4px">
          <RouterLink to="/" class="tab" active-class="tab-on">Dashboard</RouterLink>
          <RouterLink to="/customers" class="tab" active-class="tab-on">Customers</RouterLink>
          <RouterLink to="/licenses" class="tab" active-class="tab-on">Licenses</RouterLink>
        </nav>
        <div style="margin-left:auto;display:flex;align-items:center;gap:14px">
          <ThemeToggle />
          <span class="link" style="font-size:13.5px" @click="signOut">Sign out</span>
        </div>
      </div>
    </div>
    <main style="max-width:1000px;margin:0 auto;padding:32px 24px 80px">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.tab { font-size:13.5px; padding:5px 11px; border-radius:6px; color:var(--mu); font-weight:400; }
.tab:hover { color:var(--fg); }
.tab-on { color:var(--fg); font-weight:500; background:var(--chip-bg); }
</style>
