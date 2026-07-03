export type MockStackMode = "ADDITIVE" | "BEST_ONLY" | "EXCLUSIVE";
export type MockRequirementType =
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
export type MockRequirementOperator =
  | "IN"
  | "NOT_IN"
  | "EQ"
  | "GTE"
  | "LTE"
  | "BETWEEN";
export type MockBenefitType =
  | "RATE_CASHBACK"
  | "FIXED_CASHBACK"
  | "POINT"
  | "MILE"
  | "DISCOUNT"
  | "COUPON"
  | "GIFT"
  | "INSTALLMENT";

export type MockRequirement = {
  id: string;
  reward_component_id: string;
  requirement_type: MockRequirementType;
  operator: MockRequirementOperator;
  configuration_json: Record<string, unknown>;
  description: string;
};

export type MockBenefit = {
  id: string;
  reward_component_id: string;
  benefit_type: MockBenefitType;
  value: string;
  unit: string;
  cap_amount: string | null;
  cap_period: string | null;
  currency: string;
  description: string;
};

export type MockRewardComponent = {
  id: string;
  reward_group_id: string;
  name: string;
  description: string;
  layer: number;
  stack_group: string;
  stack_mode: MockStackMode;
  priority: number;
  is_exclusive: boolean;
  is_best_only: boolean;
  effective_from: string;
  effective_to: string;
  is_active: boolean;
  requirements: MockRequirement[];
  benefits: MockBenefit[];
};

export type MockRewardGroup = {
  id: string;
  activity_id: string;
  name: string;
  description: string;
  display_order: number;
  is_active: boolean;
  components: MockRewardComponent[];
};

export type MockActivityFlow = {
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
  reward_groups: MockRewardGroup[];
};

export type MockGroupForm = Omit<
  MockRewardGroup,
  "id" | "activity_id" | "components"
>;

export type MockComponentForm = Omit<
  MockRewardComponent,
  "id" | "reward_group_id" | "requirements" | "benefits"
> & {
  reward_group_id: string;
};

export type MockRequirementForm = {
  reward_component_id: string;
  requirement_type: MockRequirementType;
  operator: MockRequirementOperator;
  values: string;
  description: string;
};

export type MockBenefitForm = Omit<MockBenefit, "id">;

export type ActivityMockViewMode = "overview" | "editor";

export type ActivityMockSelection =
  | { type: "activity"; id: string }
  | { type: "group"; id: string }
  | { type: "component"; id: string }
  | { type: "requirement"; id: string; componentID: string }
  | { type: "benefit"; id: string; componentID: string };

export type ActivityOverviewRow = MockActivityFlow;