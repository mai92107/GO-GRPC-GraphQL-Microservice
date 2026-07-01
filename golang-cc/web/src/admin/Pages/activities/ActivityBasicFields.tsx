import { Field } from "../../../components";
import type { CatalogCard } from "../../../models";
import type { ActivityForm } from "./types";

type Props = {
  availableCards: CatalogCard[];
  bankOptions: string[];
  creating: boolean;
  form: ActivityForm;
  selectedCard?: CatalogCard;
  selectedNetworks: string[];
  allNetworksSelected: boolean;
  onChange: (form: ActivityForm) => void;
};

export function ActivityBasicFields({
  allNetworksSelected,
  availableCards,
  bankOptions,
  creating,
  form,
  onChange,
  selectedCard,
  selectedNetworks,
}: Props) {
  const selectCard = (cardID: string) => {
    const card = availableCards.find((item) => item.id === cardID);
    onChange({
      ...form,
      card_product_id: cardID,
      networks: card?.networks || [],
      qualified_type: "",
      selectable_type: "",
    });
  };

  return (
    <div className="form-section">
      <div className="form-section-title">
        <span>1</span>
        <div>
          <h3>方案基本資料</h3>
          <p>選擇卡片並設定方案期間</p>
        </div>
      </div>
      <div className="two-col">
        <Field label="銀行">
          <select
            value={form.bank_name}
            disabled={creating}
            onChange={(e) =>
              onChange({
                ...form,
                bank_name: e.target.value,
                card_product_id: "",
                networks: [],
                qualified_type: "",
                selectable_type: "",
              })
            }
          >
            {bankOptions.map((bankName) => (
              <option value={bankName} key={bankName}>
                {bankName}
              </option>
            ))}
          </select>
        </Field>
        <Field label="卡片">
          <select
            value={form.card_product_id}
            disabled={creating || !availableCards.length}
            onChange={(e) => selectCard(e.target.value)}
          >
            {availableCards.map((card) => (
              <option value={card.id} key={card.id}>
                {card.name}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <div className="two-col">
        <Field label="方案名稱">
          <input
            autoFocus
            required
            value={form.name}
            disabled={creating}
            placeholder="例如：2026 上半年核心權益"
            onChange={(e) => onChange({ ...form, name: e.target.value })}
          />
        </Field>
      </div>
      <div className="two-col">
        <Field label="開始日期">
          <input
            type="date"
            required
            value={form.start_date}
            disabled={creating}
            onChange={(e) => onChange({ ...form, start_date: e.target.value })}
          />
        </Field>
        <Field label="結束日期">
          <input
            type="date"
            required
            value={form.end_date}
            disabled={creating}
            onChange={(e) => onChange({ ...form, end_date: e.target.value })}
          />
        </Field>
      </div>
      <Field label="官方來源" hint="建議填入銀行回饋方案頁，方便日後查核。">
        <input
          type="url"
          value={form.source_url}
          disabled={creating}
          placeholder="https://"
          onChange={(e) => onChange({ ...form, source_url: e.target.value })}
        />
      </Field>
      <div className="field">
        <span>適用發卡別</span>
        <div className="choice-grid">
          <label className="choice-chip">
            <input
              type="checkbox"
              checked={allNetworksSelected}
              disabled={creating || !selectedNetworks.length}
              onChange={(event) =>
                onChange({
                  ...form,
                  networks: event.target.checked ? selectedNetworks : [],
                })
              }
            />
            <span>全選</span>
          </label>
          {(selectedCard?.networks || []).map((network, id) => (
            <label className="choice-chip" key={id}>
              <input
                type="checkbox"
                checked={form.networks.includes(network)}
                disabled={creating}
                onChange={(event) =>
                  onChange({
                    ...form,
                    networks: event.target.checked
                      ? [...new Set([...form.networks, network])]
                      : form.networks.filter((value) => value !== network),
                  })
                }
              />
              <span>{network}</span>
            </label>
          ))}
        </div>
        <small>方案只會推薦給持有相同發卡別的會員卡片。</small>
      </div>
    </div>
  );
}
