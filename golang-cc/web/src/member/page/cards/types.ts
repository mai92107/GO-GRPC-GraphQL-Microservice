export type CardForm = {
  card_id: string;
  nickname: string;
  last_four: string;
  statement_day: string;
  payment_due_day: string;
  account_tier: string;
  is_active: boolean;
  card_network_id: string;
};

export const emptyForm = (cardID = ""): CardForm => ({
  card_id: cardID,
  nickname: "",
  last_four: "",
  statement_day: "",
  payment_due_day: "",
  account_tier: "",
  is_active: true,
  card_network_id: "",
});
