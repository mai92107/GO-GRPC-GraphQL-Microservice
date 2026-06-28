import type { PaymentOption, Recommendation } from "../../../models";

export type Confirmation = {
  card: Recommendation;
  option: PaymentOption;
  paymentCode: string;
};
