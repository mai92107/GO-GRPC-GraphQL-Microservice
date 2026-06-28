import type { CatalogCard } from "../../../models";
import type { Activity } from "../../AdminApi";

export function filterBankCards(
  cards: CatalogCard[],
  activeBank: string,
  query: string,
  cardActivities: Record<string, Activity[]>,
) {
  const keyword = query.trim().toLowerCase();
  return cards.filter((card) => {
    if (card.bank_name !== activeBank) return false;
    const loadedActivities = cardActivities[card.id] || [];
    const haystack = [
      card.bank_name,
      card.name,
      (card.networks || []).map((network) => network.name).join(" "),
      loadedActivities.map((activity) => activity.name).join(" "),
    ]
      .join(" ")
      .toLowerCase();
    return !keyword || haystack.includes(keyword);
  });
}
