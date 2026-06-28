import type { RewardOverview } from "../../MemberApi";

type RewardGroup = RewardOverview["reward_groups"][number];

export function rewardLayers(overview: RewardOverview) {
  const grouped = new Map<
    string,
    {
      layer: string;
      total_rate: string;
      groups: RewardGroup[];
    }
  >();

  for (const group of overview.reward_groups) {
    const current = grouped.get(group.layer) || {
      layer: group.layer,
      total_rate: "0",
      groups: [],
    };
    current.groups.push(group);
    current.total_rate = String(
      current.groups.reduce((sum, item) => sum + Number(item.current_rate), 0),
    );
    grouped.set(group.layer, current);
  }

  return [...grouped.values()].map((layer) => ({
    ...layer,
    groups: [...layer.groups].sort((left, right) =>
      left.display_order !== right.display_order
        ? left.display_order - right.display_order
        : left.name.localeCompare(right.name, "zh-Hant"),
    ),
  }));
}
