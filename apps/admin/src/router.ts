import { createRouter, createWebHistory } from "vue-router";
import DashboardPage from "./pages/DashboardPage.vue";
import CustomersPage from "./pages/CustomersPage.vue";
import LicensesPage from "./pages/LicensesPage.vue";
import LoginPage from "./pages/LoginPage.vue";
import ResetPage from "./pages/ResetPage.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "dashboard", component: DashboardPage },
    { path: "/customers", name: "customers", component: CustomersPage },
    { path: "/licenses", name: "licenses", component: LicensesPage },
    { path: "/login", name: "login", component: LoginPage, meta: { public: true } },
    { path: "/reset", name: "reset", component: ResetPage, meta: { public: true } },
  ],
});
