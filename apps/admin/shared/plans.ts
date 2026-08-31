import { STAND_IDS } from "./stands";

// Pricing plans (D17): presets over the stand set. "Free" is not an issued
// license (no file needed, D16); the rest are issuable, with Custom leaving
// the stand selection freeform. termDays is the default subscription length.
export interface Plan {
  id: string;
  label: string;
  stands: string[]; // [] for custom (chosen by hand)
  termDays: number;
}

const FILE_STANDS = ["http", "https", "ftp", "tftp", "scp"];
const STANDARD_ADD = ["syslog", "snmp-traps", "dns", "ntp"];

export const PLANS: Plan[] = [
  { id: "files", label: "Files", stands: FILE_STANDS, termDays: 365 },
  { id: "standard", label: "Standard", stands: [...FILE_STANDS, ...STANDARD_ADD], termDays: 365 },
  { id: "enterprise", label: "Enterprise", stands: [...STAND_IDS], termDays: 365 },
  { id: "custom", label: "Custom", stands: [], termDays: 365 },
];

export const PLAN_IDS: string[] = PLANS.map((p) => p.id);
export const planById = (id: string): Plan | undefined => PLANS.find((p) => p.id === id);
