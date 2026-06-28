export type Priority = "high" | "normal" | "low";
export type SettingsView = "overview" | "preferences" | "payments";

export const priorityOptions: {
  id: Priority;
  label: string;
  description: string;
  weight: string;
}[] = [
  { id: "high", label: "優先", description: "推薦時更重視", weight: "1.2" },
  { id: "normal", label: "一般", description: "維持均衡考量", weight: "1.1" },
  { id: "low", label: "較少", description: "有需要再考慮", weight: "1.0" },
];
