import { createRouter, createWebHistory } from "vue-router";
import DashboardPage from "./pages/DashboardPage.vue";
import CustomersPage from "./pages/CustomersPage.vue";
import LicensesPage from "./pages/LicensesPage.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "dashboard", component: DashboardPage },
    { path: "/customers", name: "customers", component: CustomersPage },
    { path: "/licenses", name: "licenses", component: LicensesPage },
  ],
});
