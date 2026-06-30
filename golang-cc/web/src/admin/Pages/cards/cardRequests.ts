import type { CatalogCard } from "../../../models";
import {
  type Bank,
  createCard,
  deleteCard,
  getBanks,
  getCard,
  getCardNetworks,
  getCards,
  updateCard,
} from "../../AdminApi";
import type { CardForm } from "./types";

export type { Bank };

export type CardPageData = {
  cards: CatalogCard[];
  banks: Bank[];
  networks: string[];
};

export async function loadCardPageData(): Promise<CardPageData> {
  const [cards, banks, networks] = await Promise.all([
    getCards(),
    getBanks(),
    getCardNetworks(),
  ]);

  return {
    cards: Array.isArray(cards) ? cards : [],
    banks: Array.isArray(banks) ? banks : [],
    networks: Array.isArray(networks) ? networks : [],
  };
}

export function loadCardForEdit(cardID: string) {
  return getCard(cardID);
}

export function createCardFromForm(form: CardForm) {
  return createCard(
    form.bank_id,
    form.name.trim(),
    form.qualified_type,
    form.selectable_type,
    form.networks,
  );
}

export function updateCardFromForm(
  cardID: string,
  form: CardForm,
  active: boolean,
) {
  return updateCard(
    cardID,
    form.bank_id,
    form.name.trim(),
    form.qualified_type,
    form.selectable_type,
    form.networks,
    active,
  );
}

export function deleteCardByID(cardID: string) {
  return deleteCard(cardID);
}
