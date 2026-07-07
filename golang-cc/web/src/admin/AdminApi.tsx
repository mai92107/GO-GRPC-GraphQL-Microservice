import { api, del, patch, post, put } from "../api";
import type {
  AdminUser,
  Benefit,
  CatalogCard,
  Merchant,
  PaymentMethod,
  TelegramBinding,
  Unit,
} from "../models";

export type DashboardSummary = Record<string, number>;
export type Invitation = {
  id: string;
  email: string;
  expires_at: string;
  accepted_at: string | null;
  created_at: string;
};
export type Bank = {
  id: string;
  name: string;
  code: string;
  website_url: string;
  is_active: boolean;
};
export type Category = { id: string; name: string; is_active: boolean };
export type StackMode = "ADDITIVE" | "BEST_ONLY" | "EXCLUSIVE";
export type RequirementType =
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
  | "INSTALLMENT"
  | "ACCOUNT_TIER"
  | "USER_QUALIFICATION"
  | "DATE_RANGE"
  | "WEEKDAY"
  | "TIME_RANGE"
  | "ACTION_REQUIRED";
export type RequirementOperator =
  | "IN"
  | "NOT_IN"
  | "EQ"
  | "GTE"
  | "LTE"
  | "BETWEEN";
export type BenefitType =
  | "RATE_CASHBACK"
  | "FIXED_CASHBACK"
  | "POINT"
  | "MILE"
  | "DISCOUNT"
  | "COUPON"
  | "GIFT"
  | "INSTALLMENT";
export type ActivitySummary = {
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
  group_count: number;
  component_count: number;
  created_at: string;
  updated_at: string;
};
export type RewardRequirement = {
  id: string;
  reward_component_id: string;
  requirement_type: RequirementType;
  operator: RequirementOperator;
  configuration_json: Record<string, unknown>;
  description: string;
};
export type RewardBenefit = {
  id: string;
  reward_component_id: string;
  benefit_type: BenefitType;
  value: string;
  reward_unit_id: string;
  cap_amount: string | null;
  cap_formula: string | null;
  cap_period: string | null;
  description: string;
};
export type RewardComponent = {
  id: string;
  reward_group_ids: string[];
  name: string;
  description: string;
  layer: number;
  stack_group: string;
  stack_mode: StackMode;
  priority: number;
  effective_from: string;
  effective_to: string;
  is_active: boolean;
  requirements: RewardRequirement[];
  benefits: RewardBenefit[];
};
export type RewardGroup = {
  id: string;
  activity_id: string;
  name: string;
  description: string;
  display_order: number;
  is_active: boolean;
  components: RewardComponent[];
};
export type ActivityFlow = {
  activity: ActivitySummary;
  reward_groups: RewardGroup[];
};
export type RequirementTypeOption = {
  code: RequirementType;
  name: string;
  value_key: string;
  value_source: string;
};
export type ActivityRequirementOptions = {
  operators: { code: RequirementOperator; name: string }[];
  payment_methods: { code: string; name: string }[];
  merchants: { code: string; name: string }[];
  categories: { code: string; name: string }[];
  card_networks: { code: string; name: string }[];
  card_products: { code: string; name: string }[];
  card_plans: {
    id: string;
    card_product_id: string;
    plan_type: string;
    name: string;
  }[];
  account_tiers: { code: string; name: string }[];
  user_qualifications: { code: string; name: string }[];
  channels: { code: string; name: string }[];
  regions: { code: string; name: string }[];
  currencies: { code: string; name: string }[];
};
export type ActivityBenefitInput = Omit<
  RewardBenefit,
  "id" | "reward_component_id"
> & { id?: string };
export type LookupItem = { id: string; name: string; is_active: boolean };
export type RewardUnitInput = Omit<Unit, "id">;
export type MerchantInput = {
  name: string;
  aliases: string[];
  category_ids: string[];
};
type RewardVersionInput = {
  reward_unit_id: string;
  name: string;
  effect_type:
    | "ADD_RATE"
    | "SET_RATE"
    | "MULTIPLY_RATE"
    | "ADD_CASH"
    | "DISCOUNT";
  reward_value: string;
  effective_from: string;
  effective_to: string | null;
  announced_at: string | null;
  change_reason: string;
  display_change_until: string | null;
};

export const getDashboard = () => api<DashboardSummary>("/admin/dashboard");
export const getUsers = () => api<AdminUser[]>("/admin/users");
export const getInvitations = () => api<Invitation[]>("/admin/invitings");
export const invite = (email: string) =>
  post<{ id: string }>("/admin/invitations", { email });
export const toggleActivate = (userId: string, status: AdminUser["status"]) =>
  patch<{ updated: boolean }>(`/admin/users/${userId}`, {
    status: status === "active" ? "disabled" : "active",
  });
export const toggleReset = (userId: string) =>
  post<void>(`/admin/users/${userId}/password-reset`);

export const getTelegramBindings = () =>
  api<TelegramBinding[]>("/admin/telegram-bindings");
export const postTelegramBinding = (chatID: number, userID: string) =>
  post<void>("/admin/telegram-bindings", { chat_id: chatID, user_id: userID });
export const deleteTelegramBinding = (chatID: number) =>
  del<{ deleted: boolean }>(`/admin/telegram-bindings/${chatID}`);

export const getBanks = () => api<Bank[]>("/admin/banks");
export const createBank = (name: string) =>
  post<{ id: string }>("/admin/banks", { name, is_active: true });
export const activateBank = (
  id: string,
  name: string,
  code: string,
  websiteURL: string,
  isActive: boolean,
) =>
  patch<{ updated: boolean }>(`/admin/banks/${id}`, {
    name,
    code,
    website_url: websiteURL,
    is_active: !isActive,
  });
export const deleteBank = (id: string) =>
  del<{ deleted: boolean }>(`/admin/banks/${id}`);

export const getCards = () => api<CatalogCard[]>("/admin/card-products");
export const getCard = (id: string) =>
  api<CatalogCard>(`/admin/card-products/${id}`);
export const getCardNetworks = () => api<string[]>("/admin/networks");
export const createCard = (
  bankID: string,
  cardName: string,
  qualifiedType: string,
  selectableType: string,
  networkIDs: string[],
) =>
  post<{ id: string }>("/admin/card-products", {
    bank_id: bankID,
    name: cardName,
    qualified_type: qualifiedType,
    selectable_type: selectableType,
    networks: networkIDs,
    is_active: true,
  });
export const updateCard = (
  id: string,
  bankID: string,
  cardName: string,
  qualifiedType: string,
  selectableType: string,
  networkIDs: string[],
  active: boolean,
) =>
  patch<{ updated: boolean }>(`/admin/card-products/${id}`, {
    bank_id: bankID,
    name: cardName,
    qualified_type: qualifiedType,
    selectable_type: selectableType,
    networks: networkIDs,
    is_active: active,
  });
export const deleteCard = (id: string) =>
  del<{ deleted: boolean }>(`/admin/card-products/${id}`);

export const getActivities = (filters?: {
  bank_id?: string;
  card_product_id?: string;
  is_active?: boolean;
}) => {
  const params = new URLSearchParams();
  if (filters?.bank_id) params.set("bank_id", filters.bank_id);
  if (filters?.card_product_id)
    params.set("card_product_id", filters.card_product_id);
  if (filters?.is_active !== undefined)
    params.set("is_active", String(filters.is_active));
  const query = params.toString();
  return api<ActivitySummary[]>(`/admin/activities${query ? `?${query}` : ""}`);
};
export const getActivity = (id: string) =>
  api<ActivityFlow>(`/admin/activities/${id}`);
export const getRequirementTypes = () =>
  api<RequirementTypeOption[]>("/admin/requirement-types");
export const getActivityRequirementOptions = (
  requirementType: RequirementType,
  cardProductID?: string,
) => {
  const params = new URLSearchParams({ requirement_type: requirementType });
  if (cardProductID) params.set("card_product_id", cardProductID);
  return api<ActivityRequirementOptions>(
    `/admin/activity-requirement-options?${params.toString()}`,
  );
};
export const createActivity = (flow: ActivityFlow) =>
  post<{ id: string }>("/admin/activities", activityCreatePayload(flow));
export const activateActivity = (activity: ActivitySummary) =>
  patch<{ updated: boolean }>(`/admin/activities/${activity.id}/status`, {
    is_active: !activity.is_active,
  });
export const updateActivity = (flow: ActivityFlow) =>
  put<{ updated: boolean }>(
    `/admin/activities/${flow.activity.id}`,
    activityDatePayload(flow),
  );
export const deleteActivity = (id: string) =>
  del<{ deleted: boolean }>(`/admin/activities/${id}`);

function activityCreatePayload(flow: ActivityFlow): ActivityFlow {
  const next = activityDatePayload(flow);
  return {
    ...next,
    activity: { ...next.activity, id: "" },
    reward_groups: next.reward_groups.map((group) => ({
      ...group,
      id: group.id,
      activity_id: "",
      components: group.components.map((component) => ({
        ...component,
        id: component.id,
        reward_group_ids: component.reward_group_ids,
        requirements: component.requirements.map((requirement) => ({
          ...requirement,
          id: "",
          reward_component_id: "",
        })),
        benefits: component.benefits.map((benefit) => ({
          ...benefit,
          id: "",
          reward_component_id: "",
        })),
      })),
    })),
  };
}

function activityDatePayload(flow: ActivityFlow): ActivityFlow {
  return {
    ...flow,
    activity: {
      ...flow.activity,
      effective_from: dateOnly(flow.activity.effective_from),
      effective_to: dateOnly(flow.activity.effective_to),
    },
    reward_groups: flow.reward_groups.map((group) => ({
      ...group,
      components: group.components.map((component) => ({
        ...component,
        effective_from: dateOnly(component.effective_from),
        effective_to: dateOnly(component.effective_to),
      })),
    })),
  };
}

function dateOnly(value: string) {
  return value.slice(0, 10);
}
export const publishRewardComponentVersion = (
  componentID: string,
  input: RewardVersionInput,
) =>
  post<{ id: string }>(
    `/admin/reward-components/${componentID}/versions`,
    input,
  );

export const getCategories = () => api<Category[]>("/admin/categories");
export const createCategory = (name: string) =>
  post<{ id: string }>("/admin/categories", { name });
export const updateCategory = (id: string, name: string, active: boolean) =>
  patch<{ updated: boolean }>(`/admin/categories/${id}`, {
    name,
    is_active: active,
  });
export const deleteCategory = (id: string) =>
  del<{ deleted: boolean }>(`/admin/categories/${id}`);

export const getRewardUnits = () => api<Unit[]>("/admin/reward-units");
export const createRewardUnit = (input: RewardUnitInput) =>
  post<{ id: string }>("/admin/reward-units", input);
export const updateRewardUnit = (id: string, input: RewardUnitInput) =>
  patch<{ updated: boolean }>(`/admin/reward-units/${id}`, input);
export const deleteRewardUnit = (id: string) =>
  del<{ deleted: boolean }>(`/admin/reward-units/${id}`);

export const getPaymentMethods = () =>
  api<PaymentMethod[]>("/admin/payment-methods");
export const createPaymentMethod = (name: string) =>
  post<{ id: string }>("/admin/payment-methods", { name });
export const updatePaymentMethod = (
  id: string,
  name: string,
  isActive: boolean,
) =>
  patch<{ updated: boolean }>(`/admin/payment-methods/${id}`, {
    name,
    is_active: isActive,
  });
export const deletePaymentMethod = (id: string) =>
  del<{ deleted: boolean }>(`/admin/payment-methods/${id}`);

export const getMerchants = () => api<Merchant[]>("/admin/merchants");
export const createMerchant = (input: MerchantInput) =>
  post<{ id: string }>("/admin/merchants", input);
export const updateMerchant = (merchant: Merchant, isActive: boolean) =>
  patch<{ updated: boolean }>(`/admin/merchants/${merchant.id}`, {
    name: merchant.name,
    aliases: merchant.aliases || [],
    category_ids: merchant.category_ids || [],
    is_active: isActive,
  });
export const deleteMerchant = (id: string) =>
  del<{ deleted: boolean }>(`/admin/merchants/${id}`);


export const getRegions = () => api<LookupItem[]>("/admin/regions");
export const createRegion = (id: string, name: string) =>
  post<{ id: string }>("/admin/regions", { id, name });
export const updateRegion = (id: string, name: string, isActive: boolean) =>
  patch<{ updated: boolean }>(`/admin/regions/${id}`, {
    id,
    name,
    is_active: isActive,
  });
export const deleteRegion = (id: string) =>
  del<{ deleted: boolean }>(`/admin/regions/${id}`);

export const getUserQualifications = () =>
  api<LookupItem[]>("/admin/user-qualifications");
export const createUserQualification = (id: string, name: string) =>
  post<{ id: string }>("/admin/user-qualifications", { id, name });
export const updateUserQualification = (
  id: string,
  name: string,
  isActive: boolean,
) =>
  patch<{ updated: boolean }>(`/admin/user-qualifications/${id}`, {
    id,
    name,
    is_active: isActive,
  });
export const deleteUserQualification = (id: string) =>
  del<{ deleted: boolean }>(`/admin/user-qualifications/${id}`);
