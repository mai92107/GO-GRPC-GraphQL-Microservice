import type {
  MockActivityFlow,
  MockBenefitForm,
  MockComponentForm,
  MockGroupForm,
  MockRequirementForm,
} from "./mockFlowTypes";

const nextID = (prefix: string) =>
  `${prefix}-${Math.random().toString(36).slice(2, 8)}`;

export const emptyMockActivity = (): MockActivityFlow => ({
  activity: {
    id: "activity-sport-q3",
    bank_id: "bank-sinopac",
    bank_name: "永豐銀行",
    card_product_id: "card-sport",
    card_name: "SPORT 卡",
    title: "Q2 活動",
    description: "Mock 建檔流程，用於驗證 Activity / Group / Component / Requirement / Benefit。",
    source_url: "https://bank.example/activity/sport-q2",
    effective_from: "2026-04-01",
    effective_to: "2026-06-30",
    is_active: true,
  },
  reward_groups: [
    {
      id: "group-line-pay",
      activity_id: "activity-sport-q3",
      name: "LINE Pay",
      description: "行動支付回饋",
      display_order: 10,
      is_active: true,
      components: [
        {
          id: "component-linepay-base",
          reward_group_id: "group-line-pay",
          name: "LINE Pay 基本回饋",
          description: "使用 LINE Pay 付款取得基本回饋",
          layer: 1,
          stack_group: "BASE",
          stack_mode: "ADDITIVE",
          priority: 10,
          is_exclusive: false,
          is_best_only: false,
          effective_from: "2026-04-01",
          effective_to: "2026-06-30",
          is_active: true,
          requirements: [
            {
              id: "req-linepay",
              reward_component_id: "component-linepay-base",
              requirement_type: "PAYMENT_METHOD",
              operator: "IN",
              configuration_json: { payment_method_codes: ["LINE_PAY"] },
              description: "限 LINE Pay",
            },
          ],
          benefits: [
            {
              id: "benefit-linepay-base",
              reward_component_id: "component-linepay-base",
              benefit_type: "RATE_CASHBACK",
              value: "3",
              unit: "PERCENT",
              cap_amount: "300",
              cap_period: "MONTHLY",
              currency: "TWD",
              description: "3% 現金回饋，每月上限 300 元",
            },
          ],
        },
        {
          id: "component-visa-bonus",
          reward_group_id: "group-line-pay",
          name: "Visa / Mastercard 加碼",
          description: "卡組織限定加碼",
          layer: 2,
          stack_group: "NETWORK",
          stack_mode: "ADDITIVE",
          priority: 20,
          is_exclusive: false,
          is_best_only: false,
          effective_from: "2026-04-01",
          effective_to: "2026-06-30",
          is_active: true,
          requirements: [
            {
              id: "req-network",
              reward_component_id: "component-visa-bonus",
              requirement_type: "CARD_NETWORK",
              operator: "IN",
              configuration_json: { network_codes: ["VISA", "MASTERCARD"] },
              description: "限 Visa 或 Mastercard",
            },
            {
              id: "req-network-linepay",
              reward_component_id: "component-visa-bonus",
              requirement_type: "PAYMENT_METHOD",
              operator: "IN",
              configuration_json: { payment_method_codes: ["LINE_PAY"] },
              description: "限 LINE Pay",
            },
          ],
          benefits: [
            {
              id: "benefit-network",
              reward_component_id: "component-visa-bonus",
              benefit_type: "RATE_CASHBACK",
              value: "1",
              unit: "PERCENT",
              cap_amount: "100",
              cap_period: "MONTHLY",
              currency: "TWD",
              description: "Visa / Mastercard 加碼 1%",
            },
          ],
        },
      ],
    },
    {
      id: "group-wallet",
      activity_id: "activity-sport-q3",
      name: "行動支付擇優",
      description: "LINE Pay 與 Apple Pay 同組擇優",
      display_order: 20,
      is_active: true,
      components: [
        {
          id: "component-applepay-best",
          reward_group_id: "group-wallet",
          name: "Apple Pay 加碼",
          description: "同 PAYMENT stack group 只取最高",
          layer: 2,
          stack_group: "PAYMENT",
          stack_mode: "BEST_ONLY",
          priority: 30,
          is_exclusive: false,
          is_best_only: true,
          effective_from: "2026-04-01",
          effective_to: "2026-06-30",
          is_active: true,
          requirements: [
            {
              id: "req-applepay",
              reward_component_id: "component-applepay-best",
              requirement_type: "PAYMENT_METHOD",
              operator: "IN",
              configuration_json: { payment_method_codes: ["APPLE_PAY"] },
              description: "限 Apple Pay",
            },
          ],
          benefits: [
            {
              id: "benefit-applepay",
              reward_component_id: "component-applepay-best",
              benefit_type: "RATE_CASHBACK",
              value: "4",
              unit: "PERCENT",
              cap_amount: "200",
              cap_period: "MONTHLY",
              currency: "TWD",
              description: "Apple Pay 4%，PAYMENT 群組擇優",
            },
          ],
        },
        {
          id: "component-exclusive",
          reward_group_id: "group-wallet",
          name: "大型活動 10%",
          description: "不可與其他優惠併用",
          layer: 9,
          stack_group: "CAMPAIGN",
          stack_mode: "EXCLUSIVE",
          priority: 90,
          is_exclusive: true,
          is_best_only: false,
          effective_from: "2026-04-01",
          effective_to: "2026-06-30",
          is_active: true,
          requirements: [
            {
              id: "req-campaign",
              reward_component_id: "component-exclusive",
              requirement_type: "ACTION_REQUIRED",
              operator: "EQ",
              configuration_json: { action_codes: ["REGISTER"] },
              description: "需登錄，不可併用",
            },
          ],
          benefits: [
            {
              id: "benefit-exclusive",
              reward_component_id: "component-exclusive",
              benefit_type: "RATE_CASHBACK",
              value: "10",
              unit: "PERCENT",
              cap_amount: "500",
              cap_period: "CAMPAIGN",
              currency: "TWD",
              description: "大型活動 10%，不可與其他優惠併用",
            },
          ],
        },
      ],
    },
  ],
});

export const emptyGroupForm = (): MockGroupForm => ({
  name: "",
  description: "",
  display_order: 30,
  is_active: true,
});

export const emptyComponentForm = (groupID = ""): MockComponentForm => ({
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

export const emptyRequirementForm = (componentID = ""): MockRequirementForm => ({
  reward_component_id: componentID,
  requirement_type: "PAYMENT_METHOD",
  operator: "IN",
  values: "LINE_PAY",
  description: "限 LINE Pay",
});

export const emptyBenefitForm = (componentID = ""): MockBenefitForm => ({
  reward_component_id: componentID,
  benefit_type: "RATE_CASHBACK",
  value: "1",
  unit: "PERCENT",
  cap_amount: "",
  cap_period: "MONTHLY",
  currency: "TWD",
  description: "",
});

export function withGroup(flow: MockActivityFlow, form: MockGroupForm) {
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

export function withComponent(flow: MockActivityFlow, form: MockComponentForm) {
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
  flow: MockActivityFlow,
  form: MockRequirementForm,
) {
  const requirement = {
    id: nextID("req"),
    reward_component_id: form.reward_component_id,
    requirement_type: form.requirement_type,
    operator: form.operator,
    configuration_json: configurationForRequirement(form),
    description: form.description,
  };
  return updateComponent(flow, form.reward_component_id, (component) => ({
    ...component,
    requirements: [...component.requirements, requirement],
  }));
}

export function withBenefit(flow: MockActivityFlow, form: MockBenefitForm) {
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
  flow: MockActivityFlow,
  componentID: string,
  childID: string,
  childType: "requirements" | "benefits",
) {
  return updateComponent(flow, componentID, (component) => ({
    ...component,
    [childType]: component[childType].filter((item) => item.id !== childID),
  }));
}

export function allComponents(flow: MockActivityFlow) {
  return flow.reward_groups.flatMap((group) => group.components);
}

export function validateMockFlow(flow: MockActivityFlow) {
  const components = allComponents(flow);
  const requirements = components.flatMap((component) => component.requirements);
  const benefits = components.flatMap((component) => component.benefits);
  const paymentCodes = requirements.flatMap((requirement) =>
    Array.isArray(requirement.configuration_json.payment_method_codes)
      ? requirement.configuration_json.payment_method_codes
      : [],
  );
  const networkCodes = requirements.flatMap((requirement) =>
    Array.isArray(requirement.configuration_json.network_codes)
      ? requirement.configuration_json.network_codes
      : [],
  );
  const stackModes = new Set(components.map((component) => component.stack_mode));
  const hasBase = components.some(
    (component) => component.layer === 1 || component.stack_group === "BASE",
  );
  const hasBonus = components.some(
    (component) => component.layer > 1 || component.stack_group !== "BASE",
  );

  return [
    {
      label: "五層資料結構",
      pass:
        Boolean(flow.activity.title) &&
        flow.reward_groups.length > 0 &&
        components.length > 0 &&
        requirements.length > 0 &&
        benefits.length > 0,
    },
    {
      label: "Visa / Mastercard 限定",
      pass: networkCodes.includes("VISA") || networkCodes.includes("MASTERCARD"),
    },
    {
      label: "LINE Pay / Apple Pay",
      pass: paymentCodes.includes("LINE_PAY") && paymentCodes.includes("APPLE_PAY"),
    },
    {
      label: "基本回饋 + 加碼回饋",
      pass: hasBase && hasBonus,
    },
    {
      label: "互斥 / 取最佳 / 可疊加",
      pass:
        stackModes.has("EXCLUSIVE") &&
        stackModes.has("BEST_ONLY") &&
        stackModes.has("ADDITIVE"),
    },
  ];
}

function configurationForRequirement(form: MockRequirementForm) {
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
  flow: MockActivityFlow,
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



