import { defineStore } from "pinia";
import { api } from "../api";

export interface Customer {
  id: string;
  name: string;
  email: string;
  createdAt: number;
}

export const useCustomersStore = defineStore("customers", {
  state: () => ({ items: [] as Customer[], loaded: false }),
  actions: {
    async load() {
      this.items = await api<Customer[]>("/customers");
      this.loaded = true;
    },
    async create(name: string, email: string) {
      const row = await api<Customer>("/customers", {
        method: "POST",
        body: JSON.stringify({ name, email }),
      });
      this.items.unshift(row);
    },
  },
});
