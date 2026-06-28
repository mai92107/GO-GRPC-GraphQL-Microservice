import {
  CalendarDays,
  ChevronDown,
  ChevronRight,
  CreditCard,
  ExternalLink,
  Pencil,
  Trash2,
} from "lucide-react";
import { Empty, StatusBadge } from "../../../components";
import type { CatalogCard } from "../../../models";
import { benefitModeLabel, effectLabel, networkNames } from "../../../utils/activityText";
import type { Activity } from "../../AdminApi";

type Props = {
  cardActivities: Record<string, Activity[]>;
  cardLoadingID: string;
  expandedCards: string[];
  processingID: string | null;
  selectedCards: CatalogCard[];
  onEdit: (activity: Activity) => void;
  onRemove: (activity: Activity) => void;
  onSchedule: (benefit: Activity["benefits"][number]) => void;
  onToggle: (activity: Activity) => void;
  onToggleCard: (card: CatalogCard) => void;
};

function ActivityArticle({
  activity,
  card,
  processingID,
  onEdit,
  onRemove,
  onSchedule,
  onToggle,
}: Omit<Props, "selectedCards" | "expandedCards" | "cardActivities" | "cardLoadingID" | "onToggleCard"> & {
  activity: Activity;
  card: CatalogCard;
}) {
  return (
    <article className="activity-admin-card">
      <div className="activity-card-head">
        <div>
          <span className="entity-kicker">{card.bank_name}</span>
          <h3>{activity.name}</h3>
          <p><CalendarDays size={14} />{activity.start_date} – {activity.end_date}</p>
          <p>{networkNames(card, activity)}</p>
        </div>
        <div className="activity-head-actions">
          <StatusBadge active={activity.is_active} />
          {activity.source_url && (
            <a className="icon-button" href={activity.source_url} target="_blank" rel="noreferrer" aria-label="開啟官方來源">
              <ExternalLink size={15} />
            </a>
          )}
        </div>
      </div>
      <div className="benefit-admin-list">
        {activity.benefits.map((benefit) => (
          <button
            type="button"
            className="benefit-admin-row"
            key={benefit.id}
            disabled={processingID === benefit.id}
            onClick={() => onSchedule(benefit)}
          >
            <span>
              <strong>{benefit.name}</strong>
              <small>{benefitModeLabel(benefit)} · 點擊建立新版本</small>
            </span>
            <b>{effectLabel(benefit.effect_type)} {benefit.reward_value}</b>
            <ChevronRight size={17} />
          </button>
        ))}
      </div>
      <footer>
        <span className="muted">{activity.benefits.length} 項回饋條件</span>
        <div className="toolbar">
          <button type="button" className="button secondary" disabled={processingID === activity.id} onClick={() => onEdit(activity)}>
            <Pencil size={16} />編輯層級
          </button>
          <button type="button" className="button ghost" disabled={processingID === activity.id} onClick={() => onToggle(activity)}>
            {activity.is_active ? "停用方案" : "重新啟用"}
          </button>
          <button type="button" className="icon-button danger-icon" aria-label={`刪除 ${activity.name}`} disabled={processingID === activity.id} onClick={() => onRemove(activity)}>
            <Trash2 size={16} />
          </button>
        </div>
      </footer>
    </article>
  );
}

export function ActivityCardList(props: Props) {
  if (!props.selectedCards.length) {
    return <Empty title="找不到卡片" text="請嘗試其他搜尋關鍵字。" />;
  }

  return (
    <div className="bank-card-list">
      {props.selectedCards.map((card) => {
        const expanded = props.expandedCards.includes(card.id);
        const activities = props.cardActivities[card.id] || [];
        const isLoading = props.cardLoadingID === card.id;
        return (
          <section className="bank-card-panel" key={card.id}>
            <button type="button" className="bank-card-head" aria-expanded={expanded} onClick={() => props.onToggleCard(card)}>
              <span className="bank-monogram"><CreditCard size={18} /></span>
              <span>
                <strong>{card.name}</strong>
                <small>{(card.networks || []).map((network) => network.name).join("、") || "未設定發卡別"}{props.cardActivities[card.id] && `・${activities.length} 個方案`}</small>
              </span>
              {expanded ? <ChevronDown size={20} /> : <ChevronRight size={20} />}
            </button>
            {expanded && (
              <div className="bank-card-body">
                {isLoading ? (
                  <p className="muted">載入卡片回饋方案中…</p>
                ) : activities.length === 0 ? (
                  <Empty title="尚未建立回饋方案" text="這張卡片目前沒有回饋方案。" />
                ) : (
                  activities.map((activity) => (
                    <ActivityArticle
                      key={activity.id}
                      activity={activity}
                      card={card}
                      processingID={props.processingID}
                      onEdit={props.onEdit}
                      onRemove={props.onRemove}
                      onSchedule={props.onSchedule}
                      onToggle={props.onToggle}
                    />
                  ))
                )}
              </div>
            )}
          </section>
        );
      })}
    </div>
  );
}
