export type User = { id: string; email: string; display_name: string; role: "admin" | "member" };
export type AdminUser = User & { status: "active" | "disabled"; created_at: string };
export type TelegramBinding = { chat_id:number;user_id:string;email:string;display_name:string;created_at:string;updated_at:string };
export type Card = { id:string;card_product_id:string;name:string;issuer:string;last_four:string;is_active:boolean;statement_day:number|null;payment_due_day:number|null;account_tier:string };
export type Benefit = {id:string;reward_unit_id:string;name:string;rate:string;monthly_cap:string|null;stack_group:string;priority:number;required_account_tiers:string[];action_required:"none"|"registration"|"app_switch"|"account_setup";action_message:string;payment_methods:string[];category_codes:string[];merchant_codes:string[]};
export type CatalogCard = { id:string;bank_id:string;bank_name:string;name:string;is_active:boolean;account_tiers:string[];activities:{id:string;name:string;start_date:string;end_date:string;is_active:boolean;source_url:string;verified_at:string|null;benefits:Benefit[]}[] };
export type Unit = { id: string; code: string; name: string; symbol: string; symbol_position:"prefix"|"suffix";twd_rate:string; precision: number };
export type PaymentMethod = { code:string;name:string;is_active?:boolean;is_system?:boolean };
export type Merchant = { code:string;name:string;aliases?:string[];category_codes?:string[];is_active?:boolean;is_system?:boolean };
export type Allocation = { rule_id: string; rule_name: string; activity_id:string;activity_name:string;stack_group:string;action_required:string;action_message:string;reward_unit: Unit; rate:string; allocated_reward: string; uncapped_reward: string; monthly_cap: string | null; remaining_before: string | null; preference_weight: string; score: string };
export type PaymentOption = { payment_method_code:string;payment_method_name:string;payment_methods:PaymentMethod[];total_score:string;total_unweighted:string;allocations:Allocation[];explanation:string[];reminders:string[] };
export type Recommendation = { rank: number; card_id: string; card_name: string; total_score: string; allocations: Allocation[]; explanation: string[];reminders:string[];payment_options:PaymentOption[] };

let csrf = "";
export const setCSRF = (value: string) => { csrf = value; };

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (options.method && options.method !== "GET" && csrf) headers.set("X-CSRF-Token", csrf);
  const response = await fetch(`/api${path}`, { ...options, headers, credentials: "include" });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.error?.message || "發生未預期錯誤");
  return payload.data as T;
}

export const post = <T,>(path: string, body?: unknown) => api<T>(path, { method: "POST", body: body === undefined ? undefined : JSON.stringify(body) });
export const patch = <T,>(path: string, body?: unknown) => api<T>(path, { method: "PATCH", body: body === undefined ? undefined : JSON.stringify(body) });
export const del = <T,>(path: string) => api<T>(path, { method: "DELETE" });

export const mutate = <T,>(path: string, method: string, body?: unknown) =>
  api<T>(path, { method, body: body === undefined ? undefined : JSON.stringify(body) });
