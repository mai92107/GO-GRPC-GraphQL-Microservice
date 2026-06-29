import type { QualificationStatus } from "../../../MemberApi";

type Props = {
  cardID: string;
  plans: QualificationStatus[];
  savingKey: string;
  onToggle: (cardID: string, planID: string, nextValue: boolean) => void;
};

export function Qualification({
  cardID,
  plans,
  savingKey,
  onToggle,
}: Props) {
  if (!plans.length) return null;

  return (
    <section className="wallet-card-section">
      <div className="wallet-section-heading">
        <h4>優惠資格</h4>
        <p>可複選；若目前都不符合，也可以全部取消。</p>
      </div>
      <div
        className="qualification-list"
        role="group"
        aria-label="優惠資格"
        style={{
          gridTemplateColumns: `repeat(${plans.length}, minmax(0, 1fr))`,
        }}
      >
        {plans.map((plan) => {
          const key = `${cardID}:${plan.plan_id}`;
          return (
            <button
              type="button"
              className={`qualification-toggle ${plan.is_qualified ? "selected" : ""}`}
              key={plan.plan_id}
              disabled={savingKey === key}
              aria-pressed={plan.is_qualified}
              onClick={() => onToggle(cardID, plan.plan_id, !plan.is_qualified)}
            >
              <span>
                <strong>{plan.name}</strong>
              </span>
            </button>
          );
        })}
      </div>
    </section>
  );
}
