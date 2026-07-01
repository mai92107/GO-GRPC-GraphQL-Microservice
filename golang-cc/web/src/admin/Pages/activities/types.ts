import type { ActivityBenefitInput } from "../../AdminApi";

export type BenefitMode = "standard" | "selectable" | "qualified";
export type BenefitEffectType = ActivityBenefitInput["effect_type"];

export type ActivityForm = {
  card_product_id: string;
  bank_name: string;
  name: string;
  start_date: string;
  end_date: string;
  source_url: string;
  networks: string[];
  reward_unit_id: string;
  benefit_name: string;
  layer: string;
  display_order: number;
  effect_type: BenefitEffectType;
  reward_value: string;
  monthly_cap: string;
  stack_group: string;
  category_id: string;
  benefit_mode: BenefitMode;
  qualified_type: string;
  selectable_type: string;
  action_required: ActivityBenefitInput["action_required"];
  action_message: string;
  payment_methods: string[];
  merchant_ids: string[];
};

export const emptyActivityForm = (): ActivityForm => ({
  card_product_id: "",
  bank_name: "",
  name: "",
  start_date: "2026-07-01",
  end_date: "2026-12-31",
  source_url: "",
  networks: [],
  reward_unit_id: "",
  benefit_name: "",
  layer: "1",
  display_order: 0,
  effect_type: "ADD_RATE",
  reward_value: "0.01",
  monthly_cap: "",
  stack_group: "base",
  category_id: "",
  benefit_mode: "standard",
  qualified_type: "",
  selectable_type: "",
  action_required: "none",
  action_message: "",
  payment_methods: [],
  merchant_ids: [],
});
