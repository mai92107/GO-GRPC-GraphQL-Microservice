import { Pencil } from "lucide-react";
import { Empty, Field } from "../../../components";
import type { MockCatalogCardOption } from "./activityMockSettings";
import { ActivityClientPreview } from "./ActivityClientPreview";
import type { ActivityOverviewRow } from "./mockFlowTypes";

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
  bankOptions: MockCatalogCardOption[];
  cardFilter: string;
  cardOptions: MockCatalogCardOption[];
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
        {rows.map((row, id) => (
          <ActivityClientPreview
            flow={row}
            key={id}
            actions={
              <button
                className="button ghost"
                type="button"
                onClick={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  onEdit(row.activity.id);
                }}
              >
                <Pencil size={15} /> 編輯
              </button>
            }
          />
        ))}
        {!rows.length && (
          <Empty title="找不到活動" text="請調整銀行或信用卡篩選。" />
        )}
      </div>
    </section>
  );
}