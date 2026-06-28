import type { PaymentMethod } from "../models";
import type { Priority } from "../member/page/settings/types";

export const priorityFromWeight = (weight: string): Priority => {
  const value = Number(weight);
  if (value > 1.1) return "high";
  if (value < 1.1) return "low";
  return "normal";
};

export const isFixedPayment = (method: PaymentMethod) =>
  ["physical_card", "online_card"].includes(method.type || "");
