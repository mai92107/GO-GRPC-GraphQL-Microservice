import { Empty, SearchField } from "../../../components";
import type { CatalogCard } from "../../../models";
import type { Activity } from "../../AdminApi";
import { ActivityBankTabs } from "./ActivityBankTabs";
import { ActivityCardList } from "./ActivityCardList";

type Props = {
  activeBank: string;
  bankOptions: string[];
  cardActivities: Record<string, Activity[]>;
  cardLoadingID: string;
  cards: CatalogCard[];
  expandedCards: string[];
  loading: boolean;
  processingID: string | null;
  query: string;
  selectedCards: CatalogCard[];
  onActiveBankChange: (bankName: string) => void;
  onEdit: (activity: Activity) => void;
  onQueryChange: (query: string) => void;
  onRemove: (activity: Activity) => void;
  onSchedule: (benefit: Activity["benefits"][number]) => void;
  onToggle: (activity: Activity) => void;
  onToggleCard: (card: CatalogCard) => void;
};

export function ActivitiesPanel(props: Props) {
  const activeCount = Object.values(props.cardActivities)
    .flat()
    .filter((activity) => activity.is_active).length;

  return (
    <section className="panel">
      <div className="collection-toolbar">
        <SearchField
          value={props.query}
          onChange={props.onQueryChange}
          placeholder="搜尋目前銀行的卡片或已載入方案"
        />
        <div className="summary-chips">
          <span>{props.bankOptions.length} 家銀行</span>
          <span>{props.cards.length} 張卡片</span>
          <span>{activeCount} 個已載入啟用方案</span>
        </div>
      </div>
      {props.loading ? (
        <p className="muted">載入回饋方案中…</p>
      ) : props.cards.length === 0 ? (
        <Empty title="尚未建立卡片" text="建立卡片後即可新增回饋方案。" />
      ) : (
        <>
          <ActivityBankTabs
            activeBank={props.activeBank}
            bankOptions={props.bankOptions}
            cardCount={(bankName) =>
              props.cards.filter((card) => card.bank_name === bankName).length
            }
            onChange={props.onActiveBankChange}
          />
          <ActivityCardList
            cardActivities={props.cardActivities}
            cardLoadingID={props.cardLoadingID}
            expandedCards={props.expandedCards}
            processingID={props.processingID}
            selectedCards={props.selectedCards}
            onEdit={props.onEdit}
            onRemove={props.onRemove}
            onSchedule={props.onSchedule}
            onToggle={props.onToggle}
            onToggleCard={props.onToggleCard}
          />
        </>
      )}
    </section>
  );
}
