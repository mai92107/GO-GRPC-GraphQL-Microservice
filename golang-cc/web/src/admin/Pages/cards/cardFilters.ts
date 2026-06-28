import type { CatalogCard } from "../../../models";
import { splitTags } from "../../../utils/cardText";

export function filterCards(
  cards: CatalogCard[],
  query: string,
  bankFilter: string,
) {
  const keyword = query.trim().toLowerCase();

  return cards.filter((card) => {
    const matchesBank = bankFilter === "all" || card.bank_id === bankFilter;
    const haystack = [
      card.bank_name,
      card.name,
      card.account_tiers.join(" "),
      card.qualified_type,
      card.selectable_type,
      (card.networks || []).map((network) => network.name).join(" "),
    ]
      .join(" ")
      .toLowerCase();

    return matchesBank && (!keyword || haystack.includes(keyword));
  });
}

export { splitTags };
