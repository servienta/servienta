import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    // `pnpm dev` (Vite) + `pnpm dev:worker` (wrangler on :8787) side by side.
    proxy: { "/api": "http://localhost:8787" },
  },
});
