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
      published_at: null,
      published_by: null,
      publish_status: "draft",
      group_count: 0,
      component_count: 0,
      created_at: "",
      updated_at: "",
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

export const emptyComponentForm = (
  groupID = "",
  effectiveFrom = "",
  effectiveTo = "",
): ActivityComponentForm => ({
  reward_group_ids: groupID ? [groupID] : [],
  name: "",
  description: "",
  layer: 1,
  stack_group: "BASE",
  stack_mode: "ADDITIVE",
  priority: 10,
  is_exclusive: false,
  is_best_only: false,
  effective_from: effectiveFrom,
  effective_to: effectiveTo,
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
  cap_formula: "",
  cap_period: null,
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
      form.reward_group_ids.includes(group.id)
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
    cap_formula: form.cap_formula || null,
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
  const seen = new Set<string>();
  return flow.reward_groups.flatMap((group) =>
    group.components.filter((component) => {
      if (seen.has(component.id)) return false;
      seen.add(component.id);
      return true;
    }),
  );
}

export function validateActivityFlow(flow: ActivityFlowModel) {
  const components = allComponents(flow);
  const requirements = components.flatMap((component) => component.requirements);
  const benefits = components.flatMap((component) => component.benefits);
  const checks = [
    {
      label: "Activity 已選銀行與卡別",
      pass: Boolean(flow.activity.bank_id && flow.activity.card_product_id),
    },
    {
      label: "Activity 日期與名稱完整",
      pass: Boolean(
        flow.activity.title &&
          flow.activity.effective_from &&
          flow.activity.effective_to &&
          !dateBefore(flow.activity.effective_to, flow.activity.effective_from),
      ),
    },
    { label: "至少 1 個 Group", pass: flow.reward_groups.length > 0 },
    { label: "至少 1 個 Component", pass: components.length > 0 },
    { label: "至少 1 個 Requirement", pass: requirements.length > 0 },
    { label: "至少 1 個 Benefit", pass: benefits.length > 0 },
  ];

  components.forEach((component, index) => {
    const name = component.name.trim() || `#${index + 1}`;
    checks.push({
      label: `Component「${name}」已填名稱與所屬 Group`,
      pass: Boolean(component.name.trim() && component.reward_group_ids.length),
    });
    checks.push({
      label: `Component「${name}」起訖日完整`,
      pass: Boolean(component.effective_from && component.effective_to),
    });
    checks.push({
      label: `Component「${name}」結束日期不可早於開始日期`,
      pass: Boolean(
        component.effective_from &&
          component.effective_to &&
          !dateBefore(component.effective_to, component.effective_from),
      ),
    });
    checks.push({
      label: `Component「${name}」至少 1 個 Requirement`,
      pass: component.requirements.length > 0,
    });
    checks.push({
      label: `Component「${name}」至少 1 個 Benefit`,
      pass: component.benefits.length > 0,
    });
  });

  return checks;
}

function dateBefore(left: string, right: string) {
  return Boolean(left && right && left < right);
}

function configurationForRequirement(form: ActivityRequirementForm) {
  if (isUnconditionalRequirement(form.requirement_type)) return {};

  const values = form.values
    .split(",")
    .map((value) => value.trim())
    .filter(Boolean);

  switch (form.requirement_type) {
    case "CARD_NETWORK":
      return { networks: values };
    case "PAYMENT_METHOD":
      return { payment_method_codes: values.length ? values : ["any_payment"] };
    case "CARD_PLAN":
      return { card_plan_ids: values };
    case "CARD_PRODUCT":
      return { card_product_ids: values };
    case "MERCHANT":
      return { merchant_ids: values };
    case "MERCHANT_CATEGORY":
    case "CONSUMPTION_CATEGORY":
      return { category_ids: values };
    case "AMOUNT":
      return { amount: Number(values[0] || 0), currency: values[1] || "TWD" };
    case "INSTALLMENT":
      return { is_installment: values[0] !== "false" };
    case "ACCOUNT_TIER":
      return { tiers: values };
    case "USER_QUALIFICATION":
      return { qualification_codes: values };
    case "CHANNEL":
      return { channels: values };
    case "REGION":
      return { regions: values };
    case "CURRENCY":
      return { currency_codes: values };
    case "WEEKDAY":
      return { weekdays: values };
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
