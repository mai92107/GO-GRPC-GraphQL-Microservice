export type CardForm = {
  card_product_id: string;
  nickname: string;
  last_four: string;
  statement_day: string;
  payment_due_day: string;
  account_tier: string;
  is_active: boolean;
  card_network_id: string;
};

export const emptyForm = (cardProductID = ""): CardForm => ({
  card_product_id: cardProductID,
  nickname: "",
  last_four: "",
  statement_day: "",
  payment_due_day: "",
  account_tier: "",
  is_active: true,
  card_network_id: "",
});
