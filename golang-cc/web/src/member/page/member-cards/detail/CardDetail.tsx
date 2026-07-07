import { FormEvent, useEffect, useState } from "react";
import { Pencil, Save, Trash2, X } from "lucide-react";
import { Field } from "../../../../components";
import type { Card } from "../../../../models";
import { activeStatusText, lastFourText } from "../../../../utils/cardText";
import type { RewardOverview as RewardOverviewData } from "../../../MemberApi";
import { CardArtwork } from "../../cards/CardArtwork";
import { Qualification } from "./Qualification";
import { RewardOverview } from "./RewardOverview";

type Props = {
  card: Card;
  deletingID: string | null;
  expanded: Record<string, boolean>;
  overview?: RewardOverviewData;
  overviewLoading: boolean;
  qualificationSaving: string;
  creditLimitSaving: boolean;
  onDelete: (id: string) => void;
  onUpdateCreditLimit: (id: string, creditLimit: string) => Promise<void>;
  onToggleGroup: (key: string, nextValue: boolean) => void;
  onToggleQualification: (
    cardID: string,
    planID: string,
    nextValue: boolean,
  ) => void;
};

export function CardDetail({
  card,
  deletingID,
  expanded,
  overview,
  overviewLoading,
  qualificationSaving,
  creditLimitSaving,
  onDelete,
  onUpdateCreditLimit,
  onToggleGroup,
  onToggleQualification,
}: Props) {
  const [editingLimit, setEditingLimit] = useState(false);
  const [creditLimit, setCreditLimit] = useState(card.credit_limit || "");

  useEffect(() => {
    setCreditLimit(card.credit_limit || "");
    setEditingLimit(false);
  }, [card.member_card_id, card.credit_limit]);

  async function submitCreditLimit(event: FormEvent) {
    event.preventDefault();
    await onUpdateCreditLimit(card.member_card_id, creditLimit);
    setEditingLimit(false);
  }

  return (
    <article
      className={`credit-card reward-wallet-card wallet-card-detail ${
        card.is_active ? "" : "inactive"
      }`}
    >
      <div className="wallet-detail-artwork">
        <CardArtwork card={card} />
        <div>
          <span>{card.issuer}</span>
          <h3>{card.name}</h3>
          <small>{lastFourText(card.last_four)}</small>
          <small>{activeStatusText(card.is_active)}</small>
        </div>
      </div>

      <section className="reward-group">
        <header>
          <div>
            <span className="page-eyebrow">CREDIT LIMIT</span>
            <h4>信用額度</h4>
          </div>
          {!editingLimit && (
            <button
              type="button"
              className="icon-button"
              aria-label="修改信用額度"
              onClick={() => setEditingLimit(true)}
            >
              <Pencil size={16} />
            </button>
          )}
        </header>
        {editingLimit ? (
          <form className="stack" onSubmit={submitCreditLimit}>
            <Field label="目前信用額度">
              <input
                required
                type="number"
                min="1"
                step="1"
                inputMode="decimal"
                value={creditLimit}
                disabled={creditLimitSaving}
                onChange={(event) => setCreditLimit(event.target.value)}
              />
            </Field>
            <div className="wallet-card-actions">
              <button
                type="button"
                className="button ghost"
                disabled={creditLimitSaving}
                onClick={() => {
                  setCreditLimit(card.credit_limit || "");
                  setEditingLimit(false);
                }}
              >
                <X size={15} />
                取消
              </button>
              <button className="button" disabled={creditLimitSaving}>
                <Save size={15} />
                {creditLimitSaving ? "儲存中…" : "儲存額度"}
              </button>
            </div>
          </form>
        ) : (
          <p className="muted">{card.credit_limit ? `NT$ ${card.credit_limit}` : "尚未設定"}</p>
        )}
      </section>

      {overview && (
        <Qualification
          cardID={card.member_card_id}
          plans={overview.qualified_plans || []}
          savingKey={qualificationSaving}
          onToggle={onToggleQualification}
        />
      )}
      <RewardOverview
        expanded={expanded}
        loading={overviewLoading}
        overview={overview}
        onToggle={onToggleGroup}
      />

      <footer className="wallet-card-actions">
        <button
          type="button"
          className="button danger"
          disabled={deletingID === card.member_card_id}
          onClick={() => onDelete(card.member_card_id)}
        >
          <Trash2 size={15} />
          {deletingID === card.member_card_id ? "移除中…" : "移除此卡片"}
        </button>
      </footer>
    </article>
  );
}
