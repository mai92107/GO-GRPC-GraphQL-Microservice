import type { PaymentMethod } from "../../../models";

export function PaymentPill({
  method,
  selected,
  onClick,
}: {
  method: PaymentMethod;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button type="button" className={`payment-pill ${selected ? "selected" : ""}`} aria-pressed={selected} onClick={onClick}>
      {method.name}
    </button>
  );
}
