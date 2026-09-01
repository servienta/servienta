<script setup lang="ts">
import { theme, toggleTheme } from "../theme";
</script>
<template>
  <button class="tg" :title="theme === 'dark' ? 'Switch to light' : 'Switch to dark'" @click="toggleTheme">
    <span class="ring"><span class="knob" :class="theme"></span></span>
  </button>
</template>
<style scoped>
.tg { display:flex; align-items:center; justify-content:center; width:20px; height:20px; color:var(--ink2); cursor:pointer; background:none; border:none; padding:0; }
.tg:hover { color:var(--signal); }
.ring { position:relative; display:block; width:13px; height:13px; border:1px solid currentColor; overflow:hidden; }
.knob { position:absolute; top:0; bottom:0; background:currentColor; }
.knob.dark { left:0; right:50%; }
.knob.light { left:50%; right:0; }
</style>
