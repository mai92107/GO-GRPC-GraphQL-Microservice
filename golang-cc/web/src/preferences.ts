export type RewardPreference = {
  reward_unit_id: string;
  code: string;
  name: string;
  symbol: string;
  weight: string;
};

export type PreferenceTier = {
  id: string;
  units: RewardPreference[];
};

export type PreferenceWrite = {
  reward_unit_id: string;
  weight: string;
};

export function groupPreferences(preferences: RewardPreference[]): PreferenceTier[] {
  const groups = new Map<number, RewardPreference[]>();
  for (const preference of preferences) {
    const weight = Number(preference.weight);
    const group = groups.get(weight) ?? [];
    group.push(preference);
    groups.set(weight, group);
  }

  const tiers = [...groups.entries()]
    .sort(([left], [right]) => right - left)
    .map(([, units], index) => ({ id: `tier-${index + 1}`, units }));

  return tiers.length > 0 ? tiers : [{ id: "tier-1", units: [] }];
}

export function movePreference(tiers: PreferenceTier[], unitID: string, targetTierID: string): PreferenceTier[] {
  const unit = tiers.flatMap(tier => tier.units).find(item => item.reward_unit_id === unitID);
  if (!unit || !tiers.some(tier => tier.id === targetTierID)) return tiers;

  return tiers.map(tier => ({
    ...tier,
    units: tier.id === targetTierID
      ? [...tier.units.filter(item => item.reward_unit_id !== unitID), unit]
      : tier.units.filter(item => item.reward_unit_id !== unitID),
  }));
}

export function toPreferenceWrites(tiers: PreferenceTier[]): PreferenceWrite[] {
  return tiers.flatMap((tier, index) =>
    tier.units.map(unit => ({
      reward_unit_id: unit.reward_unit_id,
      weight: (1 + (tiers.length - index - 1) * 0.1).toFixed(1),
    })),
  );
}

export function tierLabel(index: number, tierCount: number): string {
  if (index === 0) return "我最愛的";
  if (index === tierCount - 1) return "這個我普普";
  return `第 ${index + 1} 優先`;
}
