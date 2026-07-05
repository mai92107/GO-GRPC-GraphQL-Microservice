import type {
  ActivityFlowModel,
  ActivityBenefitForm,
  ActivityComponentForm,
  ActivityGroupForm,
  ActivityRequirementForm,
} from "./activityFlowTypes";
import {
  isUnconditionalRequirement,
  unconditionalRequirementOperator,
} from "./requirementHelpers";

const nextID = (prefix: string) =>
  `${prefix}-${Math.random().toString(36).slice(2, 8)}`;

export const emptyActivityFlow = (): ActivityFlowModel => {
  const activityID = nextID("activity");
  return {
    activity: {
      id: activityID,
      bank_id: "",
      bank_name: "",
      card_product_id: "",
      card_name: "",
      title: "",
      description: "",
      source_url: "",
      effective_from: "",
      effective_to: "",
      is_active: true,
    },
    reward_groups: [],
  };
};

export const emptyGroupForm = (): ActivityGroupForm => ({
  name: "",
  description: "",
  display_order: 30,
  is_active: true,
});

export const emptyComponentForm = (groupID = ""): ActivityComponentForm => ({
  reward_group_id: groupID,
  name: "",
  description: "",
  layer: 1,
  stack_group: "BASE",
  stack_mode: "ADDITIVE",
  priority: 10,
  is_exclusive: false,
  is_best_only: false,
  effective_from: "2026-04-01",
  effective_to: "2026-06-30",
  is_active: true,
});

export const emptyRequirementForm = (componentID = ""): ActivityRequirementForm => ({
  reward_component_id: componentID,
  requirement_type: "PAYMENT_METHOD",
  operator: "IN",
  values: "LINE_PAY",
  description: "限 LINE Pay",
});

export const emptyBenefitForm = (componentID = ""): ActivityBenefitForm => ({
  reward_component_id: componentID,
  benefit_type: "RATE_CASHBACK",
  value: "1",
  reward_unit_id: "",
  cap_amount: "",
  cap_period: "MONTHLY",
  description: "",
  is_active: true,
});

export function withGroup(flow: ActivityFlowModel, form: ActivityGroupForm) {
  const group = {
    ...form,
    id: nextID("group"),
    activity_id: flow.activity.id,
    components: [],
  };
  return {
    ...flow,
    reward_groups: [...flow.reward_groups, group],
  };
}

export function withComponent(flow: ActivityFlowModel, form: ActivityComponentForm) {
  const component = {
    ...form,
    id: nextID("component"),
    requirements: [],
    benefits: [],
  };
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) =>
      group.id === form.reward_group_id
        ? { ...group, components: [...group.components, component] }
        : group,
    ),
  };
}

export function withRequirement(
  flow: ActivityFlowModel,
  form: ActivityRequirementForm,
) {
  const requirement = {
    id: nextID("req"),
    reward_component_id: form.reward_component_id,
    requirement_type: form.requirement_type,
    operator: isUnconditionalRequirement(form.requirement_type)
      ? unconditionalRequirementOperator
      : form.operator,
    configuration_json: configurationForRequirement(form),
    description: form.description,
    is_active: true,
  };
  return updateComponent(flow, form.reward_component_id, (component) => ({
    ...component,
    requirements: [...component.requirements, requirement],
  }));
}

export function withBenefit(flow: ActivityFlowModel, form: ActivityBenefitForm) {
  const benefit = {
    ...form,
    id: nextID("benefit"),
    cap_amount: form.cap_amount || null,
    cap_period: form.cap_period || null,
  };
  return updateComponent(flow, form.reward_component_id, (component) => ({
    ...component,
    benefits: [...component.benefits, benefit],
  }));
}

export function removeComponentChild(
  flow: ActivityFlowModel,
  componentID: string,
  childID: string,
  childType: "requirements" | "benefits",
) {
  return updateComponent(flow, componentID, (component) => ({
    ...component,
    [childType]: component[childType].filter((item) => item.id !== childID),
  }));
}

export function allComponents(flow: ActivityFlowModel) {
  return flow.reward_groups.flatMap((group) => group.components);
}

export function validateActivityFlow(flow: ActivityFlowModel) {
  const components = allComponents(flow);
  const requirements = components.flatMap((component) => component.requirements);
  const benefits = components.flatMap((component) => component.benefits);

  return [
    { label: "Activity 已選銀行與卡別", pass: Boolean(flow.activity.bank_id && flow.activity.card_product_id) },
    { label: "Activity 日期與名稱完整", pass: Boolean(flow.activity.title && flow.activity.effective_from && flow.activity.effective_to) },
    { label: "至少 1 個 Group", pass: flow.reward_groups.length > 0 },
    { label: "至少 1 個 Component", pass: components.length > 0 },
    { label: "至少 1 個 Requirement", pass: requirements.length > 0 },
    { label: "至少 1 個 Benefit", pass: benefits.length > 0 },
  ];
}

function configurationForRequirement(form: ActivityRequirementForm) {
  if (isUnconditionalRequirement(form.requirement_type)) return {};

  const values = form.values
    .split(",")
    .map((value) => value.trim())
    .filter(Boolean);

  switch (form.requirement_type) {
    case "CARD_NETWORK":
      return { network_codes: values };
    case "PAYMENT_METHOD":
      return { payment_method_codes: values.length ? values : ["any_payment"] };
    case "MERCHANT":
      return { merchant_ids: values };
    case "AMOUNT":
      return { amount: Number(values[0] || 0), currency: values[1] || "TWD" };
    case "ACCOUNT_TIER":
      return { tiers: values };
    case "USER_QUALIFICATION":
      return { qualification_codes: values };
    case "ACTION_REQUIRED":
      return { action_codes: values };
    default:
      return { values };
  }
}

function updateComponent(
  flow: ActivityFlowModel,
  componentID: string,
  updater: (component: ReturnType<typeof allComponents>[number]) => ReturnType<typeof allComponents>[number],
) {
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) => ({
      ...group,
      components: group.components.map((component) =>
        component.id === componentID ? updater(component) : component,
      ),
    })),
  };
}






