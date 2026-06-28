import { describe, expect, it } from "vitest";
import { groupPreferences, movePreference, tierLabel, toPreferenceWrites, type RewardPreference } from "./preferences";

const preference = (id: string, weight: string): RewardPreference => ({
  reward_unit_id: id,
  name: id,
  symbol: id,
  weight,
});

describe("preference tiers", () => {
  it("groups equal weights and orders tiers from highest to lowest", () => {
    const tiers = groupPreferences([preference("points", "1"), preference("cash", "2.0"), preference("miles", "1.000000")]);
    expect(tiers.map(tier => tier.units.map(unit => unit.reward_unit_id))).toEqual([["cash"], ["points", "miles"]]);
  });

  it("keeps at least one tier when there are no preferences", () => {
    expect(groupPreferences([])).toEqual([{ id: "tier-1", units: [] }]);
  });

  it("moves a preference between tiers without changing its peers", () => {
    const tiers = groupPreferences([preference("cash", "2"), preference("points", "1"), preference("miles", "1")]);
    const moved = movePreference(tiers, "points", tiers[0].id);
    expect(moved.map(tier => tier.units.map(unit => unit.reward_unit_id))).toEqual([["cash", "points"], ["miles"]]);
  });

  it("converts tiers into weights increasing by 0.1 from the last tier", () => {
    const tiers = groupPreferences([preference("cash", "4"), preference("points", "2"), preference("miles", "1")]);
    expect(toPreferenceWrites(tiers)).toEqual([
      { reward_unit_id: "cash", weight: "1.2" },
      { reward_unit_id: "points", weight: "1.1" },
      { reward_unit_id: "miles", weight: "1.0" },
    ]);
  });

  it("labels the first, middle, and last tiers", () => {
    expect([0, 1, 2].map(index => tierLabel(index, 3))).toEqual(["我最愛的", "第 2 優先", "這個我普普"]);
  });
});
