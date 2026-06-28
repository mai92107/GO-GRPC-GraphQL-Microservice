import type { CatalogCard } from "../../../models";
import type { Activity, ActivityBenefitInput } from "../../AdminApi";
import type { ActivityForm, BenefitMode } from "./types";

export const splitCardTypes = (value = "") =>
  value.split(",").map((item) => item.trim()).filter(Boolean);

export const rangesOverlap = (
  startA: string,
  endA: string,
  startB: string,
  endB: string,
) => startA <= endB && startB <= endA;

export const mapCardActivities = (card: CatalogCard): Activity[] =>
  (card.activities || []).map((activity) => ({
    ...activity,
    card_product_id: card.id,
    shared_monthly_caps: {},
  }));

export function unavailableSelectableTypes(
  items: Activity[],
  benefits: ActivityBenefitInput[],
  form: ActivityForm,
) {
  const unavailable = new Set<string>();
  const samePeriod = items.filter(
    (activity) =>
      activity.card_product_id === form.card_product_id &&
      rangesOverlap(activity.start_date, activity.end_date, form.start_date, form.end_date),
  );
  [...samePeriod.flatMap((activity) => activity.benefits), ...benefits].forEach(
    (benefit) => {
      if (benefit.action_required === "app_switch" && benefit.selectable_type) {
        unavailable.add(benefit.selectable_type);
      }
    },
  );
  return unavailable;
}

export function benefitPayload(form: ActivityForm): ActivityBenefitInput {
  return {
    reward_unit_id: form.reward_unit_id,
    name: form.benefit_name,
    display_order: form.display_order,
    effect_type: form.effect_type,
    reward_value: form.reward_value,
    monthly_cap: form.monthly_cap || null,
    layer: form.layer,
    stack_group: form.stack_group,
    priority: 100,
    qualified_type: form.benefit_mode === "qualified" ? form.qualified_type : "",
    selectable_type: form.benefit_mode === "selectable" ? form.selectable_type : "",
    action_required: form.benefit_mode === "selectable" ? "app_switch" : form.action_required,
    action_message: form.action_message,
    payment_methods: form.payment_methods,
    category_ids: [form.category_id],
    merchant_ids: form.merchant_ids,
  };
}

export function validateBenefitForm(
  form: ActivityForm,
  unavailableTypes: Set<string>,
) {
  if (!form.benefit_name.trim() || !form.reward_unit_id) {
    return "優惠名稱、回饋單位為必填。";
  }
  if (form.benefit_mode === "qualified" && !form.qualified_type.trim()) {
    return "資格限定優惠需要選擇一項會員資格。";
  }
  if (
    form.benefit_mode === "selectable" &&
    (!form.selectable_type ||
      unavailableTypes.has(form.selectable_type) ||
      !form.action_message.trim())
  ) {
    return "可切換方案需要選擇未重複的方案，並填寫會員在 App 中的切換提醒。";
  }
  return "";
}

export function nextFormForMode(
  form: ActivityForm,
  mode: BenefitMode,
  selectableTypes: string[],
  unavailableTypes: Set<string>,
) {
  return {
    ...form,
    benefit_mode: mode,
    qualified_type: mode === "qualified" ? form.qualified_type : "",
    selectable_type:
      mode === "selectable"
        ? selectableTypes.find((type) => !unavailableTypes.has(type)) || ""
        : "",
    action_required: mode === "standard" ? form.action_required : "none",
    action_message: mode === "standard" ? form.action_message : "",
  };
}
