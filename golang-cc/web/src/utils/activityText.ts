import type { CatalogCard } from "../models";
import type { Activity, ActivityBenefitInput } from "../admin/AdminApi";

export const effectOptions = [
  { value: "ADD_RATE", label: "加算回饋率" },
  { value: "SET_RATE", label: "覆蓋回饋率" },
  { value: "MULTIPLY_RATE", label: "倍率加成" },
  { value: "ADD_CASH", label: "固定回饋金" },
  { value: "DISCOUNT", label: "直接折抵" },
] as const;

export const layerOptions = [
  { value: 1, label: "Layer 1 · 卡片基本回饋" },
  { value: 2, label: "Layer 2 · 帳戶／會員加碼" },
  { value: 3, label: "Layer 3 · 指定通路／店家加碼" },
  { value: 4, label: "Layer 4 · 支付方式加碼" },
  { value: 5, label: "Layer 5 · 期間活動加碼" },
  { value: 6, label: "Layer 6 · 優惠券／折抵" },
  { value: 7, label: "Layer 7 · 手動調整／特殊規則" },
];

export const effectLabel = (effect: ActivityBenefitInput["effect_type"]) =>
  effectOptions.find((option) => option.value === effect)?.label || effect;

export const benefitModeLabel = (benefit: ActivityBenefitInput) => {
  if (benefit.qualified_type) return `資格限定 · ${benefit.qualified_type}`;
  if (benefit.action_required === "app_switch") {
    return `可切換方案${benefit.selectable_type ? ` · ${benefit.selectable_type}` : ""}`;
  }
  return `${benefit.layer} · ${effectLabel(benefit.effect_type)} · ${benefit.stack_group}`;
};

export const networkNames = (card: CatalogCard, activity: Activity) =>
  activity.network_ids
    .map((id) => (card.networks || []).find((network) => network.id === id)?.name)
    .filter(Boolean)
    .join("、") || "未指定發卡別";

export const percentText = (rate: string) => {
  const value = Number(rate) * 100;
  return Number.isFinite(value)
    ? `${Number(value.toFixed(4)).toLocaleString()}%`
    : rate;
};
