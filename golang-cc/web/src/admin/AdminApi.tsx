import { api, del, patch, post } from "../api";
import type { AdminUser, Benefit, CatalogCard, CardNetwork, Merchant, PaymentMethod, TelegramBinding, Unit } from "../models";

export type DashboardSummary = Record<string, number>;
export type Invitation = { id: string; email: string; expires_at: string; accepted_at: string | null; created_at: string };
export type Bank = { id: string; name: string; code: string; website_url: string; is_active: boolean };
export type Category = { id: string; name: string; is_active: boolean };
export type Activity = {
  id: string; card_product_id: string; name: string; start_date: string;
  end_date: string; is_active: boolean; source_url: string;
  verified_at: string | null; network_ids: string[];
  shared_monthly_caps: Record<string, string>; benefits: Benefit[];
};
export type ActivityBenefitInput = Omit<Benefit, "id"> & { id?: string };
export type RewardUnitInput = Omit<Unit, "id">;
export type MerchantInput = { name: string; aliases: string[]; category_ids: string[] };
type RewardVersionInput = {
  reward_unit_id: string; name: string;
  effect_type: "ADD_RATE" | "SET_RATE" | "MULTIPLY_RATE" | "ADD_CASH" | "DISCOUNT";
  reward_value: string; effective_from: string; effective_to: string | null;
  announced_at: string | null; change_reason: string;
  display_change_until: string | null;
};

export const getDashboard = () => api<DashboardSummary>("/admin/dashboard");
export const getUsers = () => api<AdminUser[]>("/admin/users");
export const getInvitations = () => api<Invitation[]>("/admin/invitings");
export const invite = (email: string) => post<{ id: string }>("/admin/invitations", { email });
export const toggleActivate = (userId: string, status: AdminUser["status"]) =>
  patch<{ updated: boolean }>(`/admin/users/${userId}`, {
    status: status === "active" ? "disabled" : "active",
  });
export const toggleReset = (userId: string) => post<void>(`/admin/users/${userId}/password-reset`);

export const getTelegramBindings = () => api<TelegramBinding[]>("/admin/telegram-bindings");
export const postTelegramBinding = (chatID: number, userID: string) =>
  post<void>("/admin/telegram-bindings", { chat_id: chatID, user_id: userID });
export const deleteTelegramBinding = (chatID: number) =>
  del<{ deleted: boolean }>(`/admin/telegram-bindings/${chatID}`);

export const getBanks = () => api<Bank[]>("/admin/banks");
export const createBank = (name: string) => post<{ id: string }>("/admin/banks", { name, is_active: true });
export const activateBank = (id: string, name: string, code: string, websiteURL: string, isActive: boolean) =>
  patch<{ updated: boolean }>(`/admin/banks/${id}`, {
    name, code, website_url: websiteURL, is_active: !isActive,
  });
export const deleteBank = (id: string) => del<{ deleted: boolean }>(`/admin/banks/${id}`);

export const getCards = () => api<CatalogCard[]>("/admin/card-products");
export const getCard = (id: string, includeActivities = false) =>
  api<CatalogCard>(`/admin/card-products/${id}${includeActivities ? "?include_activities=1" : ""}`);
export const getCardNetworks = () => api<CardNetwork[]>("/admin/card-networks");
export const createCard = (bankID: string, cardName: string, qualifiedType: string, selectableType: string, networkIDs: string[]) =>
  post<{ id: string }>("/admin/card-products", {
    bank_id: bankID, name: cardName, qualified_type: qualifiedType,
    selectable_type: selectableType, network_ids: networkIDs, is_active: true,
  });
export const updateCard = (id: string, bankID: string, cardName: string, qualifiedType: string, selectableType: string, networkIDs: string[], active: boolean) =>
  patch<{ updated: boolean }>(`/admin/card-products/${id}`, {
    bank_id: bankID, name: cardName, qualified_type: qualifiedType,
    selectable_type: selectableType, network_ids: networkIDs, is_active: active,
  });
export const deleteCard = (id: string) => del<{ deleted: boolean }>(`/admin/card-products/${id}`);

export const getActivities = () => api<Activity[]>("/admin/activities");
export const createActivity = (cardId: string, eventName: string, startAt: string, endAt: string, sourceUrl: string, networkIDs: string[], benefits: ActivityBenefitInput[]) =>
  post<{ id: string }>("/admin/activities", {
    card_product_id: cardId, name: eventName, start_date: startAt,
    end_date: endAt, source_url: sourceUrl, network_ids: networkIDs,
    verified_at: new Date().toISOString().slice(0, 10),
    shared_monthly_caps: {}, is_active: true, benefits,
  });
export const activateActivity = (activity: Activity) =>
  patch<{ updated: boolean }>(`/admin/activities/${activity.id}`, {
    ...activity, is_active: !activity.is_active,
  });
export const updateActivity = (activity: Activity) =>
  patch<{ updated: boolean }>(`/admin/activities/${activity.id}`, activity);
export const deleteActivity = (id: string) => del<{ deleted: boolean }>(`/admin/activities/${id}`);
export const publishRewardComponentVersion = (componentID: string, input: RewardVersionInput) =>
  post<{ id: string }>(`/admin/reward-components/${componentID}/versions`, input);

export const getCategories = () => api<Category[]>("/admin/categories");
export const createCategory = (name: string) => post<{ id: string }>("/admin/categories", { name });
export const updateCategory = (id: string, name: string, active: boolean) =>
  patch<{ updated: boolean }>(`/admin/categories/${id}`, { name, is_active: active });
export const deleteCategory = (id: string) => del<{ deleted: boolean }>(`/admin/categories/${id}`);

export const getRewardUnits = () => api<Unit[]>("/admin/reward-units");
export const createRewardUnit = (input: RewardUnitInput) => post<{ id: string }>("/admin/reward-units", input);
export const updateRewardUnit = (id: string, input: RewardUnitInput) =>
  patch<{ updated: boolean }>(`/admin/reward-units/${id}`, input);
export const deleteRewardUnit = (id: string) => del<{ deleted: boolean }>(`/admin/reward-units/${id}`);

export const getPaymentMethods = () => api<PaymentMethod[]>("/admin/payment-methods");
export const createPaymentMethod = (name: string) => post<{ id: string }>("/admin/payment-methods", { name });
export const updatePaymentMethod = (id: string, name: string, isActive: boolean) =>
  patch<{ updated: boolean }>(`/admin/payment-methods/${id}`, {
    name, is_active: isActive,
  });
export const deletePaymentMethod = (id: string) => del<{ deleted: boolean }>(`/admin/payment-methods/${id}`);

export const getMerchants = () => api<Merchant[]>("/admin/merchants");
export const createMerchant = (input: MerchantInput) => post<{ id: string }>("/admin/merchants", input);
export const updateMerchant = (merchant: Merchant, isActive: boolean) =>
  patch<{ updated: boolean }>(`/admin/merchants/${merchant.id}`, {
    name: merchant.name, aliases: merchant.aliases || [],
    category_ids: merchant.category_ids || [], is_active: isActive,
  });
export const deleteMerchant = (id: string) => del<{ deleted: boolean }>(`/admin/merchants/${id}`);
