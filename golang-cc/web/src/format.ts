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
  rate: string,
  paymentMethodCodes: string[],
  paymentMethods: { code: string; name: string }[],
): string {
  if (paymentMethodCodes.length === 0) return `${name} (${formatPercent(rate)})`;
  const names = paymentMethodCodes.map(code => paymentMethods.find(method => method.code === code)?.name ?? code);
  return `指定行動支付 (${names.join(", ")}) (${formatPercent(rate)})`;
}

export function formatScore(value: string): string {
  const number = Number(value);
  if (!Number.isFinite(number)) return value;
  return number.toLocaleString("zh-TW", { maximumFractionDigits: 3 });
}
