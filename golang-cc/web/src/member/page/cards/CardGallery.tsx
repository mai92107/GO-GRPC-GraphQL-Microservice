import { ChevronDown } from "lucide-react";
import type { Card } from "../../../models";
import { previewLastFourText } from "../../../utils/cardText";
import { CardArtwork } from "./CardArtwork";

type Props = {
  cards: Card[];
  onOpen: (card: Card) => void;
};

export function CardGallery({ cards, onOpen }: Props) {
  return (
    <div className="wallet-card-gallery">
      {cards.map((card) => (
        <button
          type="button"
          className={`wallet-card-preview ${card.is_active ? "" : "inactive"}`}
          key={card.id}
          onClick={() => onOpen(card)}
        >
          <CardArtwork card={card} />
          <span className="wallet-card-preview-copy">
            <strong>{card.name}</strong>
            <small>
              {card.issuer}
              {previewLastFourText(card.last_four)}
            </small>
          </span>
          <ChevronDown className="wallet-card-enter-icon" size={20} />
        </button>
      ))}
    </div>
  );
}
