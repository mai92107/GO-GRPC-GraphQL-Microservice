export function formatDecimal(value: string, precision = 2): string {
  const number = Number(value);
  if (!Number.isFinite(number)) return value;
  return number.toLocaleString("zh-TW", {
    minimumFractionDigits: 0,
    maximumFractionDigits: precision,
  });
}

export function formatReward(value: string, unit: { symbol: string; symbol_position: "prefix" | "suffix"; precision: number }): string {
  const number = formatDecimal(value, unit.precision);
  return unit.symbol_position === "suffix" ? `${number}${unit.symbol}` : `${unit.symbol}${number}`;
}

export function formatPercent(rate: string): string {
  return `${formatDecimal(String(Number(rate) * 100), 6)}%`;
}

export function formatBenefitTitle(
  name: string,
  rewardValue: string,
  paymentMethodIDs: string[],
  paymentMethods: { id: string; name: string }[],
  effectType = "ADD_RATE",
): string {
  const value =
    effectType === "ADD_CASH" || effectType === "DISCOUNT"
      ? `NT$${formatDecimal(rewardValue, 2)}`
      : formatPercent(rewardValue);
  if (paymentMethodIDs.length === 0) return `${name} (${value})`;
  const names = paymentMethodIDs.map((id) => paymentMethods.find((method) => method.id === id)?.name ?? id);
  return `指定行動支付 (${names.join(", ")}) (${value})`;
}

export function formatScore(value: string): string {
  const number = Number(value);
  if (!Number.isFinite(number)) return value;
  return number.toLocaleString("zh-TW", { maximumFractionDigits: 3 });
}
