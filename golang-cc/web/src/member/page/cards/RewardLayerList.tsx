import { ChevronDown, Layers3 } from "lucide-react";
import { formatPercent } from "../../../format";
import { zhDate } from "../../../utils/dateText";
import type { RewardOverview } from "../../MemberApi";
import { rewardLayers } from "./rewardLayers";

type Props = {
  expanded: Record<string, boolean>;
  overview: RewardOverview;
  onToggle: (key: string, nextValue: boolean) => void;
};

function RewardGroupDetail({
  group,
}: {
  group: RewardOverview["reward_groups"][number];
}) {
  return (
    <div className="reward-group-detail">
      <div>
        {group.requirements.map((requirement, id) => (
          <p key={id}>・{requirement}</p>
        ))}
        {group.reminders.map((reminder, id) => (
          <p className="warning-text" key={id}>
            ・{reminder}
          </p>
        ))}
        {group.change_effective_at && (
          <p>
            ・{zhDate(group.change_effective_at)} 生效
          </p>
        )}
      </div>
    </div>
  );
}

export function RewardLayerList({ expanded, overview, onToggle }: Props) {
  if (!overview.reward_groups.length) return null;

  return (
    <section className="wallet-card-section reward-section">
      <div className="wallet-section-heading">
        <h4>卡片回饋</h4>
        <p>實際回饋內容與比例仍以銀行最新公告為準。</p>
      </div>
      <div className="reward-overview layer-overview">
        {rewardLayers(overview).map((layer) => (
          <section className="wallet-reward-layer" key={layer.layer}>
            <div className="wallet-reward-layer-head">
              <span className="reward-group-icon">
                <Layers3 size={19} />
              </span>
              <span>
                <strong>{layer.layer}</strong>
              </span>
              <strong>{formatPercent(layer.total_rate)}</strong>
            </div>
            {layer.groups.map((group, index) => {
              const key = `${overview.card.id}:${group.component_id}`;
              const isOpen = expanded[key] ?? index === 0;
              const summary =
                group.cap?.spendable && Number(group.cap.spendable) > 0
                  ? `上限約 NT$ ${Number(group.cap.spendable).toFixed(0)}`
                  : "無上限";

              return (
                <div
                  className={`reward-group ${isOpen ? "is-open" : ""}`}
                  key={group.component_id}
                >
                  <button
                    type="button"
                    className="reward-overview-toggle"
                    aria-expanded={isOpen}
                    onClick={() => onToggle(key, !isOpen)}
                  >
                    <span className="reward-group-copy">
                      <strong>{group.name}</strong>
                      <small>{summary}</small>
                    </span>
                    <span className="reward-group-rate">
                      <strong>
                        回饋
                        {group.show_previous_as_strikethrough &&
                          group.previous_rate && (
                            <del>{formatPercent(group.previous_rate)}</del>
                          )}
                        {formatPercent(group.current_rate)}
                      </strong>
                    </span>
                    <ChevronDown className={isOpen ? "rotated" : ""} size={20} />
                  </button>
                  {isOpen && <RewardGroupDetail group={group} />}
                </div>
              );
            })}
          </section>
        ))}
      </div>
    </section>
  );
}
