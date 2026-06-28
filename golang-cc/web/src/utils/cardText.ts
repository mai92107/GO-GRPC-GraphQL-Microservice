export const splitTags = (value = "") =>
  value.split(",").map((item) => item.trim()).filter(Boolean);

export const lastFourText = (lastFour: string) =>
  lastFour ? `•••• •••• •••• ${lastFour}` : "未設定末四碼";

export const previewLastFourText = (lastFour: string) =>
  lastFour ? ` · •••• ${lastFour}` : "";

export const activeStatusText = (active: boolean) => (active ? "啟用" : "停用");
