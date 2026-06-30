import type { CatalogCard } from "../../../models";
import type { CardForm } from "./types";

export function cardFormFromCatalogCard(card: CatalogCard): CardForm {
  return {
    bank_id: card.bank_id,
    name: card.name,
    qualified_type: card.qualified_type || "",
    selectable_type: card.selectable_type || "",
    networks: card.networks || [],
  };
}

export function selectedNetworksOrDefault(
  selected: string[],
  available: string[],
) {
  return selected.length ? selected : available;
}

export function requestMessage(error: unknown) {
  return error instanceof Error ? error.message : "操作失敗";
}
