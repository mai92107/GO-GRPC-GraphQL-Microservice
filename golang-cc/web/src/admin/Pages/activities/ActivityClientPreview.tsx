import type { ReactNode } from "react";
import {
  benefitSummary,
  groupCalculationSummary,
  requirementSummary,
} from "./activityFlowHelpers";
import type { MockActivityFlow } from "./mockFlowTypes";

export function ActivityClientPreview({
  actions,
  eyebrow = "MEMBER PREVIEW",
  flow,
}: {
  actions?: ReactNode;
  eyebrow?: string;
  flow: MockActivityFlow;
}) {
  const sortedGroups = [...flow.reward_groups].sort(
    (a, b) => a.display_order - b.display_order,
  );

  return (
    <section className="client-preview">
      <details className="client-preview-activity">
        <summary className={`client-preview-card ${flow.activity.is_active ? "is-active" : "is-inactive"}`}>
          <div>
            <span className="page-eyebrow">{eyebrow}</span>
            <h3>{`${flow.activity.bank_name} ${flow.activity.card_name} ${flow.activity.title}`}</h3>
            <p>
              {flow.activity.effective_from} - {flow.activity.effective_to}
            </p>
          </div>
          {actions && <div className="client-preview-actions">{actions}</div>}
        </summary>
        <div className="client-preview-groups">
          {sortedGroups.map((group) => (
            <article className={`client-reward-group ${group.is_active ? "is-active" : "is-inactive"}`} key={group.id}>
              <header>
                <span>
                  <strong>{group.name}</strong>
                  <small>{group.description}</small>
                </span>
                <strong>{groupCalculationSummary(group.components)}</strong>
              </header>
              {group.components.map((component) => (
                <div
                  className={`client-component ${component.stack_mode.toLowerCase()} ${component.is_active ? "is-active" : "is-inactive"}`}
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
            </article>
          ))}
        </div>
      </details>
    </section>
  );
}