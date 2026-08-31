<script setup lang="ts">
defineProps<{ size?: number }>();
</script>
<template>
  <span class="mark" :style="{ '--u': (size ?? 3) + 'px' }">
    <i></i><i></i><i></i><i></i><i class="live"></i><i></i><i></i><i></i><i></i>
  </span>
</template>
<style scoped>
.mark { display:grid; grid-template-columns:repeat(3,var(--u)); grid-template-rows:repeat(3,var(--u)); gap:var(--u); }
.mark i { background:var(--dot); }
.mark i.live { background:#8b6cff; }
</style>
