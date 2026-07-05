import type { ActivitySummary } from "../../AdminApi";

export type ActivityStackMode = "ADDITIVE" | "BEST_ONLY" | "EXCLUSIVE";
export type ActivityRequirementType =
  | "none"
  | "PAYMENT_METHOD"
  | "CARD_NETWORK"
  | "CARD_PLAN"
  | "CARD_PRODUCT"
  | "MERCHANT"
  | "MERCHANT_CATEGORY"
  | "CONSUMPTION_CATEGORY"
  | "CHANNEL"
  | "REGION"
  | "CURRENCY"
  | "AMOUNT"
  | "ACCOUNT_TIER"
  | "USER_QUALIFICATION"
  | "DATE_RANGE"
  | "WEEKDAY"
  | "TIME_RANGE"
  | "ACTION_REQUIRED";
export type ActivityRequirementOperator =
  | "IN"
  | "NOT_IN"
  | "EQ"
  | "GTE"
  | "LTE"
  | "BETWEEN";
export type ActivityBenefitType =
  | "RATE_CASHBACK"
  | "FIXED_CASHBACK"
  | "POINT"
  | "MILE"
  | "DISCOUNT"
  | "COUPON"
  | "GIFT"
  | "INSTALLMENT";

export type ActivityRequirement = {
  id: string;
  reward_component_id: string;
  requirement_type: ActivityRequirementType;
  operator: ActivityRequirementOperator;
  configuration_json: Record<string, unknown>;
  description: string;
  is_active: boolean;
};

export type ActivityBenefit = {
  id: string;
  reward_component_id: string;
  benefit_type: ActivityBenefitType;
  value: string;
  reward_unit_id: string;
  cap_amount: string | null;
  cap_period: string | null;
  description: string;
  is_active: boolean;
};

export type ActivityRewardComponent = {
  id: string;
  reward_group_id: string;
  name: string;
  description: string;
  layer: number;
  stack_group: string;
  stack_mode: ActivityStackMode;
  priority: number;
  is_exclusive: boolean;
  is_best_only: boolean;
  effective_from: string;
  effective_to: string;
  is_active: boolean;
  requirements: ActivityRequirement[];
  benefits: ActivityBenefit[];
};

export type ActivityRewardGroup = {
  id: string;
  activity_id: string;
  name: string;
  description: string;
  display_order: number;
  is_active: boolean;
  components: ActivityRewardComponent[];
};

export type ActivityFlowModel = {
  activity: {
    id: string;
    bank_id: string;
    bank_name: string;
    card_product_id: string;
    card_name: string;
    title: string;
    description: string;
    source_url: string;
    effective_from: string;
    effective_to: string;
    is_active: boolean;
  };
  reward_groups: ActivityRewardGroup[];
};

export type ActivityGroupForm = Omit<
  ActivityRewardGroup,
  "id" | "activity_id" | "components"
>;

export type ActivityComponentForm = Omit<
  ActivityRewardComponent,
  "id" | "reward_group_id" | "requirements" | "benefits"
> & {
  reward_group_id: string;
};

export type ActivityRequirementForm = {
  reward_component_id: string;
  requirement_type: ActivityRequirementType;
  operator: ActivityRequirementOperator;
  values: string;
  description: string;
};

export type ActivityBenefitForm = Omit<ActivityBenefit, "id">;

export type ActivityFlowViewMode = "overview" | "editor";

export type ActivityFlowSelection =
  | { type: "activity"; id: string }
  | { type: "group"; id: string }
  | { type: "component"; id: string }
  | { type: "requirement"; id: string; componentID: string }
  | { type: "benefit"; id: string; componentID: string };

export type ActivityOverviewRow = ActivitySummary;
