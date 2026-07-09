import { api, del, patch, post } from "../api";
import type {
  Card,
  CatalogCard,
  Merchant,
  PaymentMethod,
  Recommendation,
  Unit,
} from "../models";
import type { PreferenceWrite, RewardPreference } from "../preferences";

export type Category = {
  id: string;
  name: string;
};

export type MemberCardInput = {
  card_id: string;
  nickname: string;
  last_four: string;
  statement_day: number | null;
  payment_due_day: number | null;
  account_tier: string;
  is_active: boolean;
  network: string;
  credit_limit: string;
};

export type QualificationStatus = {
  plan_id: string;
  name: string;
  is_qualified: boolean;
  effective_from: string;
};

export type RewardOverview = {
  card: Card;
  qualified_plans: QualificationStatus[];
  reward_groups: {
    component_id: string;
    name: string;
    current_rate: string;
    layer: string;
    display_order: number;
    effect_type: string;
    reward_value: string;
    previous_rate: string | null;
    next_rate: string | null;
    change_effective_at: string | null;
    show_previous_as_strikethrough: boolean;
    requirements: string[];
    reminders: string[];
    cap: {
      type: string;
      limit: string;
      period: string;
      spendable: string;
    } | null;
  }[];
};

export type RecommendationInput = {
  amount_minor: number;
  category_id: string;
  merchant_id: string;
  merchant_name: string;
  date: string;
};

export type RecommendationResult = {
  input: RecommendationInput;
  recommendations: Recommendation[];
  empty_reason: string;
};

export type RecommendationSummary = {
  rule_id: string;
  allocated_reward: string;
};

export type TransactionInput = {
  card_id: string;
  amount_minor: number;
  category_id: string;
  merchant_id: string;
  merchant_name: string;
  payment_method_id: string;
  transaction_date: string;
  note?: string;
  recommendation_summary: RecommendationSummary[];
};

export type TransactionSummary = {
  id: string;
  card_id: string;
  amount_minor: number;
  category_id: string;
  merchant_id: string;
  merchant_name: string;
  payment_method_id: string;
  payment_method_name: string;
  transaction_date: string;
  note: string;
};

export type Transaction = TransactionSummary & {
  user_id: string;
  allocations: Recommendation["allocations"];
};

export const getCatalogCards = () =>
  api<CatalogCard[]>("/member/catalog/cards").then((cards) =>
    (Array.isArray(cards) ? cards : []).map(normalizeCatalogCard),
  );

export const getCatalogCard = (id: string) =>
  api<CatalogCard>(`/member/catalog/cards/${id}`).then(normalizeCatalogCard);

const stringArray = (value: unknown): string[] =>
  Array.isArray(value)
    ? value.filter((item): item is string => typeof item === "string")
    : [];

export const normalizeCatalogCard = (card: CatalogCard): CatalogCard => ({
  ...card,
  account_tiers: stringArray(card.account_tiers),
  networks: stringArray(card.networks),
  activities: (Array.isArray(card.activities) ? card.activities : []).map(
    (activity) => ({
      ...activity,
      networks: stringArray(activity.networks),
      benefits: (Array.isArray(activity.benefits)
        ? activity.benefits
        : []
      ).map((benefit) => ({
        ...benefit,
        payment_methods: stringArray(benefit.payment_methods),
        category_ids: stringArray(benefit.category_ids),
        merchant_ids: stringArray(benefit.merchant_ids),
      })),
    }),
  ),
});

export const getCards = () => api<Card[]>("/member/cards");

export const getCard = (id: string) => api<Card>(`/member/cards/${id}`);

export const getRewardOverview = (id: string, at = new Date().toISOString()) =>
  api<RewardOverview>(
    `/member/cards/${id}/reward-overview?at=${encodeURIComponent(at)}`,
  );

export const createCard = (input: MemberCardInput) =>
  post<{ id: string; card_product_id: string }>("/member/cards", input);

export const updateCard = (id: string, input: MemberCardInput) =>
  patch<{ updated: boolean }>(`/member/cards/${id}`, input);

export const deleteCard = (id: string) =>
  del<{ deleted: boolean }>(`/member/cards/${id}`);

export const getRewardUnits = () => api<Unit[]>("/member/reward-units");

export const getCategories = () => api<Category[]>("/member/categories");

export const getPaymentMethods = () =>
  api<PaymentMethod[]>("/member/payment-methods");

export const updatePaymentMethods = (paymentMethodIDs: string[]) =>
  api<{ updated: boolean }>("/member/payment-methods", {
    method: "PUT",
    body: JSON.stringify({ payment_method_ids: paymentMethodIDs }),
  });

export const setQualificationStatus = (
  cardID: string,
  planID: string,
  isQualified: boolean,
  effectiveFrom = new Date().toISOString(),
) =>
  api<{ updated: boolean }>(
    `/member/cards/${cardID}/qualifications/${planID}`,
    {
      method: "PUT",
      body: JSON.stringify({
        is_qualified: isQualified,
        effective_from: effectiveFrom,
      }),
    },
  );

export const getMerchants = (categoryCode: string) =>
  api<Merchant[]>(
    `/member/merchants?category_id=${encodeURIComponent(categoryCode)}`,
  );

export const getRewardPreferences = () =>
  api<RewardPreference[]>("/member/reward-preferences");

export const updateRewardPreferences = (preferences: PreferenceWrite[]) =>
  patch<{ updated: boolean }>("/member/reward-preferences", {
    preferences,
  });

export const getRecommendations = (input: RecommendationInput) =>
  post<RecommendationResult>("/member/recommendations", input);

export const getTransactions = () =>
  api<TransactionSummary[]>("/member/transactions");

export const getTransaction = (id: string) =>
  api<TransactionSummary>(`/member/transactions/${id}`);

export const createTransaction = (input: TransactionInput) =>
  post<{
    transaction: Transaction;
    recommendation_changed: boolean;
  }>("/member/transactions", input);

export const updateTransaction = (id: string, input: TransactionInput) =>
  patch<Transaction>(`/member/transactions/${id}`, input);

export const deleteTransaction = (id: string) =>
  del<{ deleted: boolean }>(`/member/transactions/${id}`);
