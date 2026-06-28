import { describe, expect, it } from "vitest";
import { formatBenefitTitle, formatDecimal, formatScore } from "./format";

describe("reward formatting", () => {
  it("uses the reward unit precision", () => {
    expect(formatDecimal("10.000000", 2)).toBe("10");
    expect(formatDecimal("8.500000", 0)).toBe("9");
  });

  it("places reward symbols before or after the value", async () => {
    const { formatReward } = await import("./format");
    expect(formatReward("10.00", { symbol: "NT$", symbol_position: "prefix", precision: 2 })).toBe("NT$10");
    expect(formatReward("10.00", { symbol: "點", symbol_position: "suffix", precision: 2 })).toBe("10點");
  });

  it("keeps recommendation scores concise", () => {
    expect(formatScore("12.345600")).toBe("12.346");
    expect(formatScore("0.000000")).toBe("0");
  });

  it("describes restricted payment benefits using their payment methods", () => {
    expect(formatBenefitTitle("原始名稱", "0.045", ["line_pay", "jkopay"], [
      { id: "line_pay", name: "LINE Pay" },
      { id: "jkopay", name: "街口支付" },
    ])).toBe("指定行動支付 (LINE Pay, 街口支付) (4.5%)");
  });
});
