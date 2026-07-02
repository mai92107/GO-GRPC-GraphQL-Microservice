export type MockOption = { code: string; name: string };
export type MockCatalogCardOption = {
  bank_id: string;
  bank_name: string;
  card_product_id: string;
  card_name: string;
};

export const mockCatalogCards: MockCatalogCardOption[] = [
  {
    bank_id: "bank-sinopac",
    bank_name: "永豐銀行",
    card_product_id: "card-sport",
    card_name: "SPORT 卡",
  },
  {
    bank_id: "bank-line",
    bank_name: "LINE Bank",
    card_product_id: "card-linebank",
    card_name: "LINE Bank 聯名卡",
  },
  {
    bank_id: "bank-cathay",
    bank_name: "國泰世華",
    card_product_id: "card-cube",
    card_name: "CUBE 卡",
  },
];

export const requirementOperatorOptions: MockOption[] = [
  { code: "IN", name: "包含任一" },
  { code: "NOT_IN", name: "不包含" },
  { code: "EQ", name: "等於" },
  { code: "GTE", name: "大於等於" },
  { code: "LTE", name: "小於等於" },
  { code: "BETWEEN", name: "介於" },
];

export const benefitTypeOptions: MockOption[] = [
  { code: "RATE_CASHBACK", name: "百分比現金回饋" },
  { code: "FIXED_CASHBACK", name: "固定金額回饋" },
  { code: "POINT", name: "點數回饋" },
  { code: "MILE", name: "哩程回饋" },
  { code: "DISCOUNT", name: "折扣" },
  { code: "COUPON", name: "折價券" },
  { code: "GIFT", name: "贈品" },
  { code: "INSTALLMENT", name: "分期優惠" },
];

export const rewardUnitOptions: MockOption[] = [
  { code: "PERCENT", name: "百分比" },
  { code: "AMOUNT", name: "金額" },
  { code: "POINT", name: "點數" },
  { code: "POINT_PER_100", name: "每百元點數" },
  { code: "MILE", name: "哩程" },
];

export const defaultCapPeriodOptions: MockOption[] = [
  { code: "NONE", name: "無上限週期" },
  { code: "DAILY", name: "每日" },
  { code: "WEEKLY", name: "每週" },
  { code: "MONTHLY", name: "每月" },
  { code: "QUARTERLY", name: "每季" },
  { code: "YEARLY", name: "每年" },
  { code: "CAMPAIGN", name: "活動期間" },
];

const capPeriodStorageKey = "activity_mock_cap_period_options";

export function loadCapPeriodOptions() {
  if (typeof window === "undefined") return defaultCapPeriodOptions;
  const raw = window.localStorage.getItem(capPeriodStorageKey);
  if (!raw) return defaultCapPeriodOptions;
  try {
    const parsed = JSON.parse(raw) as MockOption[];
    return parsed.length ? parsed : defaultCapPeriodOptions;
  } catch {
    return defaultCapPeriodOptions;
  }
}

export function saveCapPeriodOptions(options: MockOption[]) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(capPeriodStorageKey, JSON.stringify(options));
}
