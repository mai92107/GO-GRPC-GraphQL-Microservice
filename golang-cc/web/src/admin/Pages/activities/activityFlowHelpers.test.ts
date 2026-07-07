import { describe, expect, it } from "vitest";
import {
  configForRequirement,
  displayRequirementValues,
} from "./activityFlowHelpers";
import {
  emptyActivityFlow,
  emptyComponentForm,
  validateActivityFlow,
} from "./activityFlowFactory";
import { normalizeRequirementFormType } from "./requirementHelpers";
import type { ActivityRequirement } from "./activityFlowTypes";

describe("activity flow requirement helpers", () => {
  it("maps installment requirement values to boolean config", () => {
    expect(configForRequirement("INSTALLMENT", "true")).toEqual({
      is_installment: true,
    });
    expect(configForRequirement("INSTALLMENT", "false")).toEqual({
      is_installment: false,
    });
  });

  it("displays installment requirement values as stable booleans", () => {
    const requirement: ActivityRequirement = {
      id: "requirement",
      reward_component_id: "component",
      requirement_type: "INSTALLMENT",
      operator: "EQ",
      configuration_json: { is_installment: true },
      description: "",
      is_active: true,
    };

    expect(displayRequirementValues(requirement)).toBe("true");
  });

  it("defaults installment requirement forms to equality and true", () => {
    expect(normalizeRequirementFormType("INSTALLMENT")).toEqual({
      requirement_type: "INSTALLMENT",
      operator: "EQ",
      values: "true",
      description: "",
    });
  });

  it("creates component forms with caller-provided activity dates", () => {
    expect(
      emptyComponentForm("group", "2026-07-01", "2026-09-30"),
    ).toMatchObject({
      reward_group_ids: ["group"],
      effective_from: "2026-07-01",
      effective_to: "2026-09-30",
    });
  });

  it("reports invalid component dates in flow validation", () => {
    const flow = emptyActivityFlow();
    flow.activity = {
      ...flow.activity,
      bank_id: "bank",
      card_product_id: "card",
      title: "Q3",
      effective_from: "2026-07-01",
      effective_to: "2026-09-30",
    };
    flow.reward_groups = [
      {
        id: "group",
        activity_id: flow.activity.id,
        name: "Group",
        description: "",
        display_order: 10,
        is_active: true,
        components: [
          {
            ...emptyComponentForm("group", "2026-10-01", "2026-09-30"),
            id: "component",
            name: "UP 選 Q3",
            requirements: [
              {
                id: "requirement",
                reward_component_id: "component",
                requirement_type: "none",
                operator: "EQ",
                configuration_json: {},
                description: "",
                is_active: true,
              },
            ],
            benefits: [
              {
                id: "benefit",
                reward_component_id: "component",
                benefit_type: "RATE_CASHBACK",
                value: "3",
                reward_unit_id: "unit",
                cap_amount: null,
                cap_formula: null,
                cap_period: null,
                description: "",
                is_active: true,
              },
            ],
          },
        ],
      },
    ];

    expect(validateActivityFlow(flow)).toContainEqual({
      label: "Component「UP 選 Q3」結束日期不可早於開始日期",
      pass: false,
    });
  });
});
