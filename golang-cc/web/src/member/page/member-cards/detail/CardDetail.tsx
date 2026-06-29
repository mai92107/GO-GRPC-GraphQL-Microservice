import { Trash2 } from "lucide-react";
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
  onDelete: (id: string) => void;
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
  onDelete,
  onToggleGroup,
  onToggleQualification,
}: Props) {
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
