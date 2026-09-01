// The product's stands — the units of licensing (D15): a license grants a
// subset of these; the engine starts only what is granted (R12, phase 5).
// File-serving transports are licensed individually.
export const STANDS = [
  { id: "http", label: "Files: HTTP" },
  { id: "https", label: "Files: HTTPS" },
  { id: "ftp", label: "Files: FTP" },
  { id: "tftp", label: "Files: TFTP" },
  { id: "scp", label: "Files: SCP" },
  { id: "syslog", label: "Syslog" },
  { id: "snmp-traps", label: "SNMP traps" },
  { id: "radius", label: "RADIUS" },
  { id: "tacacs", label: "TACACS+" },
  { id: "dns", label: "DNS" },
  { id: "ntp", label: "NTP" },
  { id: "kafka", label: "Kafka" },
  { id: "ipfix", label: "IPFIX" },
] as const;

export type StandId = (typeof STANDS)[number]["id"];
export const STAND_IDS: string[] = STANDS.map((s) => s.id);
