import { Trash2 } from "lucide-react";
import type { Card } from "../../../models";
import { activeStatusText, lastFourText } from "../../../utils/cardText";
import type { RewardOverview } from "../../MemberApi";
import { CardArtwork } from "./CardArtwork";
import { QualificationList } from "./QualificationList";
import { RewardLayerList } from "./RewardLayerList";

type Props = {
  card: Card;
  deletingID: string | null;
  expanded: Record<string, boolean>;
  overview?: RewardOverview;
  overviewLoadingID: string;
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
  overviewLoadingID,
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

      {overviewLoadingID === card.id && (
        <section className="wallet-card-section">
          <p className="muted">載入卡片回饋中…</p>
        </section>
      )}
      {overview && (
        <>
          <QualificationList
            cardID={card.id}
            plans={overview.qualified_plans}
            savingKey={qualificationSaving}
            onToggle={onToggleQualification}
          />
          <RewardLayerList
            expanded={expanded}
            overview={overview}
            onToggle={onToggleGroup}
          />
        </>
      )}

      <footer className="wallet-card-actions">
        <button
          type="button"
          className="button danger"
          disabled={deletingID === card.id}
          onClick={() => onDelete(card.id)}
        >
          <Trash2 size={15} />
          {deletingID === card.id ? "移除中…" : "移除此卡片"}
        </button>
      </footer>
    </article>
  );
}
