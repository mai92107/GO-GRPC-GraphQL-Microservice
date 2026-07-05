import { Pencil } from "lucide-react";
import { Empty, Field, StatusBadge } from "../../../components";
import type { ActivityCatalogCardOption } from "./activityFlowSettings";
import type { ActivityOverviewRow } from "./activityFlowTypes";

export function ActivityOverviewList({
  bankFilter,
  bankOptions,
  cardFilter,
  cardOptions,
  onBankFilterChange,
  onCardFilterChange,
  onEdit,
  rows,
}: {
  bankFilter: string;
  bankOptions: ActivityCatalogCardOption[];
  cardFilter: string;
  cardOptions: ActivityCatalogCardOption[];
  rows: ActivityOverviewRow[];
  onBankFilterChange: (bankID: string) => void;
  onCardFilterChange: (cardID: string) => void;
  onEdit: (activityID: string) => void;
}) {
  return (
    <section className="activity-overview">
      <div className="activity-overview-filters">
        <Field label="銀行">
          <select
            value={bankFilter}
            onChange={(event) => onBankFilterChange(event.target.value)}
          >
            <option value="all">全部銀行</option>
            {bankOptions.map((bank) => (
              <option value={bank.bank_id} key={bank.bank_id}>
                {bank.bank_name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="信用卡">
          <select
            value={cardFilter}
            onChange={(event) => onCardFilterChange(event.target.value)}
          >
            <option value="all">全部信用卡</option>
            {cardOptions.map((card) => (
              <option value={card.card_product_id} key={card.card_product_id}>
                {card.card_name}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <div className="summary-chips">
        <span>{rows.length} 個活動</span>
      </div>
      <div className="activity-overview-list">
        {rows.map((row) => (
          <article
            className={`activity-overview-row ${row.is_active ? "is-active" : "is-inactive"}`}
            key={row.id}
          >
            <div>
              <span className="page-eyebrow">ACTIVITY</span>
              <h3>{`${row.bank_name} ${row.card_name} ${row.title}`}</h3>
              <p>{row.effective_from} - {row.effective_to}</p>
              <small>{row.group_count} groups · {row.component_count} components</small>
            </div>
            <div className="activity-head-actions">
              <StatusBadge active={row.is_active} />
              <button
                className="button ghost"
                type="button"
                onClick={() => onEdit(row.id)}
              >
                <Pencil size={15} /> 展開
              </button>
            </div>
          </article>
        ))}
        {!rows.length && (
          <Empty title="找不到活動" text="請調整銀行或信用卡篩選。" />
        )}
      </div>
    </section>
  );
}
