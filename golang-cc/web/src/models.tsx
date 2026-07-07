export type User = {
  id: string;
  email: string;
  display_name: string;
  role: "admin" | "member";
};
export type AdminUser = User & {
  status: "active" | "disabled";
  created_at: string;
};
export type TelegramBinding = {
  chat_id: number;
  user_id: string;
  email: string;
  display_name: string;
  created_at: string;
  updated_at: string;
};
export type Card = {
  member_card_id: string;
  name: string;
  nickname?: string;
  issuer: string;
  last_four: string;
  credit_limit: string;
  card_network_id?: string;
  is_active: boolean;
  statement_day: number | null;
  payment_due_day: number | null;
  account_tier: string;
  card_image_url: string;
  primary_color: string;
  qualified_type: string;
  selectable_type: string;
  network: string;
};

export type Benefit = {
  id: string;
  reward_unit_id: string;
  name: string;
  display_order: number;
  effect_type:
    | "ADD_RATE"
    | "SET_RATE"
    | "MULTIPLY_RATE"
    | "ADD_CASH"
    | "DISCOUNT";
  reward_value: string;
  monthly_cap: string | null;
  layer: string;
  stack_group: string;
  priority: number;
  qualified_type: string;
  selectable_type: string;
  action_required: "none" | "registration" | "app_switch" | "account_setup";
  action_message: string;
  payment_methods: string[];
  category_ids: string[];
  merchant_ids: string[];
};
export type CatalogCard = {
  id: string;
  bank_id: string;
  bank_name: string;
  name: string;
  card_image_url: string;
  primary_color: string;
  is_active: boolean;
  account_tiers: string[];
  qualified_type: string;
  selectable_type: string;
  networks: string[];
  activities: {
    id: string;
    name: string;
    start_date: string;
    end_date: string;
    is_active: boolean;
    source_url: string;
    verified_at: string | null;
    network_ids: string[];
    benefits: Benefit[];
  }[];
};
export type Unit = {
  id: string;
  name: string;
  symbol: string;
  symbol_position: "prefix" | "suffix";
  twd_rate: string;
  precision: number;
};
export type PaymentMethod = {
  id: string;
  name: string;
  is_active?: boolean;
  is_system?: boolean;
  type?:
    | "physical_card"
    | "online_card"
    | "mobile_payment"
    | "electronic_ticket";
  is_available?: boolean;
};
export type Merchant = {
  id: string;
  name: string;
  aliases?: string[];
  category_ids?: string[];
  is_active?: boolean;
  is_system?: boolean;
};
export type Allocation = {
  rule_id: string;
  rule_name: string;
  activity_id: string;
  activity_name: string;
  layer: string;
  stack_group: string;
  display_order: number;
  effect_type:
    | "ADD_RATE"
    | "SET_RATE"
    | "MULTIPLY_RATE"
    | "ADD_CASH"
    | "DISCOUNT";
  reward_value: string;
  action_required: string;
  action_message: string;
  reward_unit: Unit;
  reward_rate: string;
  allocated_reward: string;
  uncapped_reward: string;
  monthly_cap: string | null;
  remaining_before: string | null;
  preference_weight: string;
  score: string;
  suggested_card_plan_id?: string;
  suggested_plan_name?: string;
};
export type LayerBenefit = {
  benefit_id: string;
  name: string;
  activity_id: string;
  activity_name: string;
  effect_type:
    | "ADD_RATE"
    | "SET_RATE"
    | "MULTIPLY_RATE"
    | "ADD_CASH"
    | "DISCOUNT";
  reward_value: string;
  reward_rate: string;
  reward_amount: string;
  stack_group: string;
  monthly_cap: string | null;
  remaining_before: string | null;
  is_applied: boolean;
  exclude_reason?: string;
};
export type RewardLayer = {
  layer: number;
  reward_rate: string;
  reward_amount: string;
  benefits: LayerBenefit[];
};
export type ExcludedBenefit = {
  benefit_id: string;
  name: string;
  activity_id: string;
  activity_name: string;
  reason: string;
  stack_group: string;
  layer: string;
};
export type PaymentOption = {
  payment_method_id: string;
  payment_method_name: string;
  payment_methods: PaymentMethod[];
  total_score: string;
  total_unweighted: string;
  allocations: Allocation[];
  layers: RewardLayer[];
  excluded_benefits: ExcludedBenefit[];
  explanation: string[];
  reminders: string[];
};
export type Recommendation = {
  rank: number;
  card_id: string;
  card_name: string;
  total_score: string;
  allocations: Allocation[];
  layers: RewardLayer[];
  excluded_benefits: ExcludedBenefit[];
  explanation: string[];
  reminders: string[];
  payment_options: PaymentOption[];
};
