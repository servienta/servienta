<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from "vue-router";
import { api } from "./api";
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
    <header v-if="!route.meta.public" style="border-bottom:1px solid var(--rule)">
      <div style="max-width:1200px;margin:0 auto;padding:0 40px;height:58px;display:flex;align-items:center;gap:28px">
        <div style="display:flex;align-items:baseline;gap:12px">
          <span style="font-size:17px;font-weight:700;letter-spacing:0.16em;text-transform:uppercase">Servienta</span>
          <span class="mono" style="font-size:10.5px;letter-spacing:0.1em;text-transform:uppercase;color:var(--signal)">admin</span>
        </div>
        <nav class="mono" style="display:flex;gap:22px;font-size:11px;letter-spacing:0.06em;text-transform:uppercase">
          <RouterLink to="/" class="tab" active-class="tab-on">Dashboard</RouterLink>
          <RouterLink to="/customers" class="tab" active-class="tab-on">Customers</RouterLink>
          <RouterLink to="/licenses" class="tab" active-class="tab-on">Licenses</RouterLink>
        </nav>
        <div style="margin-left:auto;display:flex;align-items:center;gap:20px">
          <ThemeToggle />
          <span class="act" @click="signOut">Sign out</span>
        </div>
      </div>
    </header>
    <main style="max-width:1200px;margin:0 auto;padding:0 40px">
      <div :style="route.meta.public ? '' : 'border-left:1px solid var(--rule);border-right:1px solid var(--rule);min-height:calc(100vh - 58px);padding:40px 48px 80px'">
        <RouterView />
      </div>
    </main>
  </div>
</template>

<style scoped>
.tab { color:var(--ink2); padding-bottom:2px; border-bottom:1px solid transparent; }
.tab:hover { color:var(--ink); }
.tab-on { color:var(--ink); border-bottom-color:var(--ink); }
</style>
