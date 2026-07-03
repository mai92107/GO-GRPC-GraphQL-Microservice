import { allComponents, removeComponentChild } from "./mockFlowHelpers";
import type {
  ActivityMockSelection,
  MockActivityFlow,
  MockBenefit,
  MockBenefitForm,
  MockRequirement,
  MockRequirementType,
  MockRewardComponent,
  MockRewardGroup,
} from "./mockFlowTypes";

export function deleteSelection(
  flow: MockActivityFlow,
  selection: ActivityMockSelection,
) {
  if (selection.type === "group")
    return {
      ...flow,
      reward_groups: flow.reward_groups.filter(
        (group) => group.id !== selection.id,
      ),
    };
  if (selection.type === "component")
    return {
      ...flow,
      reward_groups: flow.reward_groups.map((group) => ({
        ...group,
        components: group.components.filter(
          (component) => component.id !== selection.id,
        ),
      })),
    };
  if (selection.type === "requirement")
    return removeComponentChild(
      flow,
      selection.componentID,
      selection.id,
      "requirements",
    );
  if (selection.type === "benefit")
    return removeComponentChild(
      flow,
      selection.componentID,
      selection.id,
      "benefits",
    );
  return flow;
}

export function updateGroup(
  flow: MockActivityFlow,
  groupID: string,
  patch: Partial<MockRewardGroup>,
) {
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) =>
      group.id === groupID ? { ...group, ...patch } : group,
    ),
  };
}

export function updateComponent(
  flow: MockActivityFlow,
  componentID: string,
  patch: Partial<MockRewardComponent>,
) {
  const current = allComponents(flow).find(
    (component) => component.id === componentID,
  );
  if (!current) return flow;
  const nextGroupID = patch.reward_group_id || current.reward_group_id;
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) => {
      const existing = group.components.find(
        (component) => component.id === componentID,
      );
      if (existing && group.id !== nextGroupID)
        return {
          ...group,
          components: group.components.filter(
            (component) => component.id !== componentID,
          ),
        };
      if (group.id !== nextGroupID) return group;
      const nextComponent = {
        ...current,
        ...patch,
        reward_group_id: nextGroupID,
      };
      return {
        ...group,
        components: existing
          ? group.components.map((component) =>
              component.id === componentID ? nextComponent : component,
            )
          : [...group.components, nextComponent],
      };
    }),
  };
}

export function updateRequirement(
  flow: MockActivityFlow,
  requirementID: string,
  nextRequirement: MockRequirement,
) {
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) => ({
      ...group,
      components: group.components.map((component) => ({
        ...component,
        requirements: component.requirements.map((requirement) =>
          requirement.id === requirementID ? nextRequirement : requirement,
        ),
      })),
    })),
  };
}

export function updateBenefit(
  flow: MockActivityFlow,
  benefitID: string,
  nextBenefit: MockBenefit | MockBenefitForm,
) {
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) => ({
      ...group,
      components: group.components.map((component) => ({
        ...component,
        benefits: component.benefits.map((benefit) =>
          benefit.id === benefitID ? { ...benefit, ...nextBenefit } : benefit,
        ),
      })),
    })),
  };
}

export function displayRequirementValues(requirement: MockRequirement) {
  const config = requirement.configuration_json;
  const value =
    config.network_codes ||
    config.payment_method_codes ||
    config.merchant_ids ||
    config.tiers ||
    config.qualification_codes ||
    config.action_codes ||
    config.values;
  if (Array.isArray(value)) return value.join(", ");
  if (typeof config.amount === "number")
    return `${config.amount}, ${config.currency || "TWD"}`;
  return "";
}

export function configForRequirement(
  type: MockRequirementType,
  value: string,
) {
  const values = value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
  if (type === "CARD_NETWORK") return { network_codes: values };
  if (type === "PAYMENT_METHOD") return { payment_method_codes: values };
  if (type === "MERCHANT") return { merchant_ids: values };
  if (type === "AMOUNT")
    return { amount: Number(values[0] || 0), currency: values[1] || "TWD" };
  if (type === "ACCOUNT_TIER") return { tiers: values };
  if (type === "USER_QUALIFICATION") return { qualification_codes: values };
  if (type === "ACTION_REQUIRED") return { action_codes: values };
  return { values };
}

export function groupCalculationSummary(components: MockRewardComponent[]) {
  const exclusive = components.find(
    (component) => component.stack_mode === "EXCLUSIVE",
  );
  if (exclusive) return benefitSummary(exclusive.benefits);

  const bestOnly = components.filter(
    (component) => component.stack_mode === "BEST_ONLY",
  );
  if (bestOnly.length) {
    const best = [...bestOnly].sort(
      (a, b) => componentPercentValue(b) - componentPercentValue(a),
    )[0];
    return benefitSummary(best.benefits);
  }

  const values = components
    .map(componentPercentValue)
    .filter((value) => value > 0)
    .map(formatPercent);
  return values.length ? values.join(" + ") : "尚未設定可計算回饋";
}

export function benefitSummary(benefits: MockBenefit[]) {
  if (!benefits.length) return "尚未設定";
  return benefits
    .map((benefit) =>
      benefit.reward_unit_id === "PERCENT"
        ? `${benefit.value}%`
        : `${benefit.value} ${benefit.reward_unit_id}`,
    )
    .join(" + ");
}

export function requirementSummary(requirements: MockRequirement[]) {
  if (!requirements.length) return "不限條件";
  return requirements
    .map(
      (requirement) => requirement.description || requirement.requirement_type,
    )
    .join(" / ");
}

function componentPercentValue(component: MockRewardComponent) {
  return component.benefits.reduce((sum, benefit) => {
    if (benefit.benefit_type !== "RATE_CASHBACK" || benefit.reward_unit_id !== "PERCENT")
      return sum;
    const value = Number(benefit.value);
    return Number.isFinite(value) ? sum + value : sum;
  }, 0);
}

function formatPercent(value: number) {
  return `${Number(value.toFixed(4)).toLocaleString()}%`;
}


