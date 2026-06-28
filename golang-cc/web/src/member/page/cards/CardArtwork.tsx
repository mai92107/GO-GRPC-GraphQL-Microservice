import { Leaf } from "lucide-react";
import type { Card } from "../../../models";

export function CardArtwork({ card }: { card: Card }) {
  const background = card.primary_color || "#245fcb";

  return (
    <div
      className="wallet-card-artwork"
      style={{ backgroundColor: background }}
    >
      {card.card_image_url ? (
        <img
          src={card.card_image_url}
          alt={`${card.issuer} ${card.name} 卡面`}
        />
      ) : (
        <>
          <span>{card.issuer}</span>
          <small>{card.network?.name || "Credit Card"}</small>
          <Leaf aria-hidden="true" />
        </>
      )}
    </div>
  );
}
