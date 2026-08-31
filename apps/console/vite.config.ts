import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

// SPA builds into internal/webui/dist so the Go binary can embed it.
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: { outDir: "internal/webui/dist", emptyOutDir: true },
  server: { proxy: { "/api": "http://localhost:8080" } },
});
