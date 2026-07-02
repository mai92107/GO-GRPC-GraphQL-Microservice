import {
  benefitSummary,
  groupCalculationSummary,
  requirementSummary,
} from "./activityFlowHelpers";
import type { MockActivityFlow } from "./mockFlowTypes";

export function ActivityClientPreview({ flow }: { flow: MockActivityFlow }) {
  const sortedGroups = [...flow.reward_groups].sort(
    (a, b) => a.display_order - b.display_order,
  );

  return (
    <section className="client-preview">
      <div className="client-preview-card">
        <span className="page-eyebrow">MEMBER PREVIEW</span>
        <h3>{`${flow.activity.bank_name} ${flow.activity.card_name} ${flow.activity.title}`}</h3>
        <p>
          {flow.activity.effective_from} - {flow.activity.effective_to}
        </p>
      </div>
      <div className="client-preview-groups">
        {sortedGroups.map((group) => (
          <details className="client-reward-group" key={group.id} open>
            <summary>
              <span>
                <strong>{group.name}</strong>
                <small>{group.description}</small>
              </span>
              <strong>{groupCalculationSummary(group.components)}</strong>
            </summary>
            {group.components.map((component) => (
              <div
                className={`client-component ${component.stack_mode.toLowerCase()}`}
                key={component.id}
              >
                <div>
                  <strong>{component.name}</strong>
                  <small>
                    Layer {component.layer} · {component.stack_mode}
                  </small>
                </div>
                <span>{benefitSummary(component.benefits)}</span>
                <p>{requirementSummary(component.requirements)}</p>
              </div>
            ))}
          </details>
        ))}
      </div>
    </section>
  );
}
