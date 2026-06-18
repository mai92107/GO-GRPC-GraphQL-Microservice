import { api, del, patch, post } from "../api";
import type {
  Card,
  CatalogCard,
  Merchant,
  PaymentMethod,
  Recommendation,
  Unit,
} from "../models";
import type {
  PreferenceWrite,
  RewardPreference,
} from "../preferences";

export type Category = {
  code: string;
  name: string;
};

export type MemberCardInput = {
  card_product_id?: string;
  nickname: string;
  last_four: string;
  statement_day: number | null;
  payment_due_day: number | null;
  account_tier: string;
  is_active: boolean;
};

export type RecommendationInput = {
  amount_minor: number;
  category_code: string;
  merchant_code: string;
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
  category_code: string;
  merchant_code: string;
  merchant_name: string;
  payment_method_code: string;
  transaction_date: string;
  note?: string;
  recommendation_summary: RecommendationSummary[];
};

export type TransactionSummary = {
  id: string;
  card_id: string;
  amount_minor: number;
  category_code: string;
  merchant_code: string;
  merchant_name: string;
  payment_method_code: string;
  payment_method_name: string;
  transaction_date: string;
  note: string;
};

export type Transaction = TransactionSummary & {
  user_id: string;
  allocations: Recommendation["allocations"];
};

export const getCatalogCards = () =>
  api<CatalogCard[]>("/member/catalog/cards");

export const getCards = () => api<Card[]>("/member/cards");

export const getCard = (id: string) =>
  api<Card>(`/member/cards/${id}`);

export const createCard = (input: MemberCardInput) =>
  post<{ id: string; card_product_id: string }>("/member/cards", input);

export const updateCard = (id: string, input: MemberCardInput) =>
  patch<{ updated: boolean }>(`/member/cards/${id}`, input);

export const deleteCard = (id: string) =>
  del<{ deleted: boolean }>(`/member/cards/${id}`);

export const getRewardUnits = () =>
  api<Unit[]>("/member/reward-units");

export const getCategories = () =>
  api<Category[]>("/member/categories");

export const getPaymentMethods = () =>
  api<PaymentMethod[]>("/member/payment-methods");

export const getMerchants = (categoryCode: string) =>
  api<Merchant[]>(
    `/member/merchants?category_code=${encodeURIComponent(categoryCode)}`,
  );

export const getRewardPreferences = () =>
  api<RewardPreference[]>("/member/reward-preferences");

export const updateRewardPreferences = (
  preferences: PreferenceWrite[],
) =>
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

export const updateTransaction = (
  id: string,
  input: TransactionInput,
) =>
  patch<Transaction>(`/member/transactions/${id}`, input);

export const deleteTransaction = (id: string) =>
  del<{ deleted: boolean }>(`/member/transactions/${id}`);
