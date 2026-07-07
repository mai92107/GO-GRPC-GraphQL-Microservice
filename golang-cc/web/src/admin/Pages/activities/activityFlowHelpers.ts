import { allComponents, removeComponentChild } from "./activityFlowFactory";
import type {
  ActivityFlowSelection,
  ActivityFlowModel,
  ActivityBenefit,
  ActivityBenefitForm,
  ActivityRequirement,
  ActivityRequirementType,
  ActivityRewardComponent,
  ActivityRewardGroup,
} from "./activityFlowTypes";
import { isUnconditionalRequirement } from "./requirementHelpers";

export function deleteSelection(
  flow: ActivityFlowModel,
  selection: ActivityFlowSelection,
) {
  if (selection.type === "group")
    return {
      ...flow,
      reward_groups: flow.reward_groups.filter(
        (group) => group.id !== selection.id,
      ).map((group) => ({
        ...group,
        components: group.components
          .map((component) => ({
            ...component,
            reward_group_ids: component.reward_group_ids.filter(
              (groupID) => groupID !== selection.id,
            ),
          }))
          .filter((component) => component.reward_group_ids.length > 0),
      })),
    };
  if (selection.type === "component")
    return removeComponentFromGroup(flow, selection.id, selection.groupID);
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
  flow: ActivityFlowModel,
  groupID: string,
  patch: Partial<ActivityRewardGroup>,
) {
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) =>
      group.id === groupID ? { ...group, ...patch } : group,
    ),
  };
}

export function updateComponent(
  flow: ActivityFlowModel,
  componentID: string,
  patch: Partial<ActivityRewardComponent>,
) {
  const current = allComponents(flow).find(
    (component) => component.id === componentID,
  );
  if (!current) return flow;
  const nextGroupIDs = patch.reward_group_ids || current.reward_group_ids;
  const nextComponent = {
    ...current,
    ...patch,
    reward_group_ids: nextGroupIDs,
  };
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) => {
      const existing = group.components.find(
        (component) => component.id === componentID,
      );
      if (existing && !nextGroupIDs.includes(group.id))
        return {
          ...group,
          components: group.components.filter(
            (component) => component.id !== componentID,
          ),
        };
      if (!nextGroupIDs.includes(group.id)) return group;
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

function removeComponentFromGroup(
  flow: ActivityFlowModel,
  componentID: string,
  groupID?: string,
) {
  const current = allComponents(flow).find(
    (component) => component.id === componentID,
  );
  if (!current) return flow;
  const nextGroupIDs = groupID
    ? current.reward_group_ids.filter((id) => id !== groupID)
    : [];
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) => ({
      ...group,
      components: group.components
        .filter(
          (component) =>
            component.id !== componentID || nextGroupIDs.includes(group.id),
        )
        .map((component) =>
          component.id === componentID
            ? { ...component, reward_group_ids: nextGroupIDs }
            : component,
        ),
    })),
  };
}

export function updateRequirement(
  flow: ActivityFlowModel,
  requirementID: string,
  nextRequirement: ActivityRequirement,
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
  flow: ActivityFlowModel,
  benefitID: string,
  nextBenefit: ActivityBenefit | ActivityBenefitForm,
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

export function displayRequirementValues(requirement: ActivityRequirement) {
  if (isUnconditionalRequirement(requirement.requirement_type)) return "";

  const config = requirement.configuration_json;
  const value =
    config.network_codes ||
    config.payment_method_codes ||
    config.card_plan_ids ||
    config.card_product_ids ||
    config.merchant_ids ||
    config.category_ids ||
    config.channels ||
    config.regions ||
    config.tiers ||
    config.qualification_codes ||
    config.action_codes ||
    config.weekdays ||
    config.values;
  if (Array.isArray(value)) return value.join(", ");
  if (typeof config.is_installment === "boolean")
    return config.is_installment ? "true" : "false";
  if (typeof config.amount === "number")
    return `${config.amount}, ${config.currency || "TWD"}`;
  return "";
}

export function configForRequirement(
  type: ActivityRequirementType,
  value: string,
) {
  if (isUnconditionalRequirement(type)) return {};

  const values = value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
  if (type === "CARD_NETWORK") return { network_codes: values };
  if (type === "PAYMENT_METHOD") return { payment_method_codes: values };
  if (type === "CARD_PLAN") return { card_plan_ids: values };
  if (type === "CARD_PRODUCT") return { card_product_ids: values };
  if (type === "MERCHANT") return { merchant_ids: values };
  if (type === "MERCHANT_CATEGORY" || type === "CONSUMPTION_CATEGORY")
    return { category_ids: values };
  if (type === "AMOUNT")
    return { amount: Number(values[0] || 0), currency: values[1] || "TWD" };
  if (type === "INSTALLMENT") return { is_installment: values[0] !== "false" };
  if (type === "ACCOUNT_TIER") return { tiers: values };
  if (type === "USER_QUALIFICATION") return { qualification_codes: values };
  if (type === "CHANNEL") return { channels: values };
  if (type === "REGION") return { regions: values };
  if (type === "WEEKDAY") return { weekdays: values };
  if (type === "ACTION_REQUIRED") return { action_codes: values };
  return { values };
}

export function groupCalculationSummary(components: ActivityRewardComponent[]) {
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

export function benefitSummary(benefits: ActivityBenefit[]) {
  if (!benefits.length) return "尚未設定";
  return benefits
    .map((benefit) =>
      benefit.reward_unit_id === "PERCENT"
        ? `${benefit.value}%`
        : `${benefit.value} ${benefit.reward_unit_id}`,
    )
    .join(" + ");
}

export function requirementSummary(requirements: ActivityRequirement[]) {
  if (!requirements.length) return "不限條件";
  return requirements
    .map(
      (requirement) => requirement.description || requirement.requirement_type,
    )
    .join(" / ");
}

function componentPercentValue(component: ActivityRewardComponent) {
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
