import { CreditCard } from "lucide-react";
import { Empty, SearchField } from "../../../components";
import type { CatalogCard } from "../../../models";
import type { Bank } from "../../AdminApi";
import { splitTags } from "./cardFilters";

type Props = {
  banks: Bank[];
  cards: CatalogCard[];
  editLoadingID: string;
  itemsCount: number;
  loading: boolean;
  onBankFilterChange: (value: string) => void;
  onEdit: (card: CatalogCard) => void;
  onQueryChange: (value: string) => void;
  query: string;
  bankFilter: string;
};

function CardTags({ card }: { card: CatalogCard }) {
  const networks = card.networks || [];
  const ruleTags = [
    ...splitTags(card.qualified_type),
    ...splitTags(card.selectable_type),
  ];

  return (
    <>
      <div className="tag-row">
        {networks.map((network) => (
          <span className="tag" key={network.id}>
            {network.name}
          </span>
        ))}
      </div>
      <div className="tag-row">
        {ruleTags.map((tag) => (
          <span className="tag" key={tag}>
            {tag}
          </span>
        ))}
      </div>
    </>
  );
}

export function CardCatalogList({
  banks,
  cards,
  editLoadingID,
  itemsCount,
  loading,
  onBankFilterChange,
  onEdit,
  onQueryChange,
  query,
  bankFilter,
}: Props) {
  return (
    <section className="panel">
      <div className="collection-toolbar">
        <SearchField
          value={query}
          onChange={onQueryChange}
          placeholder="搜尋銀行、卡片或帳戶等級"
        />
        <select
          className="filter-select"
          value={bankFilter}
          onChange={(event) => onBankFilterChange(event.target.value)}
        >
          <option value="all">全部銀行</option>
          {banks.map((bank) => (
            <option value={bank.id} key={bank.id}>
              {bank.name}
            </option>
          ))}
        </select>
        <span className="collection-count">{cards.length} 張卡片</span>
      </div>

      {loading ? (
        <p className="muted">載入卡片目錄中…</p>
      ) : itemsCount === 0 ? (
        <Empty title="尚未建立卡片" text="建立銀行後即可新增卡片。" />
      ) : (
        <div className="entity-grid">
          {cards.map((card) => (
            <article className="entity-card card-entity" key={card.id}>
              <div className="entity-icon">
                <CreditCard size={21} />
              </div>
              <div className="entity-copy">
                <span className="entity-kicker">{card.bank_name}</span>
                <h3>{card.name}</h3>
                <CardTags card={card} />
              </div>
              <button
                type="button"
                className="button ghost"
                disabled={editLoadingID === card.id}
                onClick={() => onEdit(card)}
              >
                {editLoadingID === card.id ? "載入中…" : "修改"}
              </button>
            </article>
          ))}
          {!cards.length && (
            <Empty title="找不到卡片" text="請調整搜尋或銀行篩選。" />
          )}
        </div>
      )}
    </section>
  );
}
