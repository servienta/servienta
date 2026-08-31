// Shared light/dark theme with persistence. Admin defaults to light.
import { ref } from "vue";

const KEY = "servienta-console-theme";
export const theme = ref<"light" | "dark">("light");

export function initTheme() {
  try {
    const s = localStorage.getItem(KEY) as "light" | "dark" | null;
    if (s) theme.value = s;
  } catch {}
  apply();
}
export function toggleTheme() {
  theme.value = theme.value === "light" ? "dark" : "light";
  try { localStorage.setItem(KEY, theme.value); } catch {}
  apply();
}
function apply() {
  document.documentElement.setAttribute("data-theme", theme.value);
}
