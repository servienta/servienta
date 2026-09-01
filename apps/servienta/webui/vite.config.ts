import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

// SPA builds into the engine module (internal/webui/dist) for go:embed.
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: { outDir: "../internal/webui/dist", emptyOutDir: true },
  server: { proxy: { "/api": "http://localhost:8080" } },
});
