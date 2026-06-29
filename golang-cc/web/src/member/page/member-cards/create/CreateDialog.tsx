import { FormEvent } from "react";
import { Dialog, Field } from "../../../../components";
import type { CatalogCard } from "../../../../models";
import type { CardForm } from "../../cards/types";

type Props = {
  catalog: CatalogCard[];
  error: string;
  form: CardForm;
  formForCard: (cardID: string) => CardForm;
  onChange: (form: CardForm) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent) => void;
  saving: boolean;
  selectedCard?: CatalogCard;
};

export function CreateDialog({
  catalog,
  error,
  form,
  formForCard,
  onChange,
  onClose,
  onSubmit,
  saving,
  selectedCard,
}: Props) {
  return (
    <Dialog title="加入卡片夾" onClose={onClose}>
      <form className="stack" onSubmit={onSubmit}>
        {error && <div className="error">{error}</div>}
        <Field label="銀行卡片">
          <select
            value={form.card_id}
            disabled={saving}
            onChange={(event) =>
              onChange({
                ...formForCard(event.target.value),
                nickname: form.nickname,
                last_four: form.last_four,
                statement_day: form.statement_day,
                payment_due_day: form.payment_due_day,
                is_active: form.is_active,
              })
            }
          >
            {catalog.map((card) => (
              <option value={card.id} key={card.id}>
                {card.bank_name} · {card.name}
              </option>
            ))}
          </select>
        </Field>
        {selectedCard && selectedCard.networks.length > 0 && (
          <Field label="卡組織">
            <select
              required
              value={form.card_network_id}
              disabled={saving}
              onChange={(event) =>
                onChange({ ...form, card_network_id: event.target.value })
              }
            >
              {selectedCard.networks.map((network) => (
                <option value={network} key={network}>
                  {network}
                </option>
              ))}
            </select>
          </Field>
        )}
        <Field label="自訂暱稱">
          <input
            value={form.nickname}
            disabled={saving}
            onChange={(event) =>
              onChange({ ...form, nickname: event.target.value })
            }
          />
        </Field>
        <Field label="卡號末四碼">
          <input
            inputMode="numeric"
            pattern="[0-9]{4}"
            value={form.last_four}
            disabled={saving}
            onChange={(event) =>
              onChange({ ...form, last_four: event.target.value })
            }
          />
        </Field>
        <div className="two-col">
          <Field label="結帳日">
            <input
              type="number"
              min="1"
              max="31"
              value={form.statement_day}
              disabled={saving}
              onChange={(event) =>
                onChange({ ...form, statement_day: event.target.value })
              }
            />
          </Field>
          <Field label="繳款日">
            <input
              type="number"
              min="1"
              max="31"
              value={form.payment_due_day}
              disabled={saving}
              onChange={(event) =>
                onChange({ ...form, payment_due_day: event.target.value })
              }
            />
          </Field>
        </div>
        <label className="choice-chip">
          <input
            type="checkbox"
            checked={form.is_active}
            disabled={saving}
            onChange={(event) =>
              onChange({ ...form, is_active: event.target.checked })
            }
          />
          <span>啟用卡片</span>
        </label>
        <button className="button" disabled={saving || !form.card_id}>
          {saving ? "加入中…" : "加入卡片夾"}
        </button>
      </form>
    </Dialog>
  );
}
