<script setup lang="ts">
import { theme, toggleTheme } from "../theme";
</script>
<template>
  <button class="tg" :title="theme === 'dark' ? 'Switch to light' : 'Switch to dark'" @click="toggleTheme">
    <span class="ring"><span class="knob" :class="theme"></span></span>
  </button>
</template>
<style scoped>
.tg { display:flex; align-items:center; justify-content:center; width:24px; height:24px;
  color:var(--mu); cursor:pointer; background:none; border:none; padding:0; }
.tg:hover { color:var(--acc-h); }
.ring { position:relative; display:block; width:14px; height:14px; border-radius:999px;
  border:1.5px solid currentColor; overflow:hidden; }
.knob { position:absolute; top:0; bottom:0; background:currentColor; }
.knob.dark { left:0; right:50%; }
.knob.light { left:50%; right:0; }
</style>
