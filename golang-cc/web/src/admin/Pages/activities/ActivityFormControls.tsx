import type { ReactNode } from "react";
import { Field } from "../../../components";
import {
  benefitTypeOptions,
  requirementOperatorOptions,
  rewardUnitOptions,
} from "./activityMockSettings";
import {
  configForRequirement,
  displayRequirementValues,
} from "./activityFlowHelpers";
import type {
  MockActivityFlow,
  MockBenefit,
  MockBenefitForm,
  MockBenefitType,
  MockComponentForm,
  MockRequirement,
  MockRequirementForm,
  MockRequirementOperator,
  MockRequirementType,
  MockRewardComponent,
  MockStackMode,
} from "./mockFlowTypes";

const stackModes: MockStackMode[] = ["ADDITIVE", "BEST_ONLY", "EXCLUSIVE"];
const stackGroups = [
  "BASE",
  "PAYMENT",
  "NETWORK",
  "MERCHANT",
  "NEW_USER",
  "ACCOUNT_TIER",
  "CAMPAIGN",
];
const requirementTypes: MockRequirementType[] = [
  "PAYMENT_METHOD",
  "CARD_NETWORK",
  "MERCHANT",
  "AMOUNT",
  "ACCOUNT_TIER",
  "USER_QUALIFICATION",
  "ACTION_REQUIRED",
];

type CapPeriodOption = { code: string; name: string };

export function ComponentFields({
  flow,
  form,
  onChange,
}: {
  flow: MockActivityFlow;
  form: MockComponentForm;
  onChange: (form: MockComponentForm) => void;
}) {
  return (
    <>
      <div className="two-col">
        <Field label="所屬 Group">
          <select
            value={form.reward_group_id}
            onChange={(event) =>
              onChange({ ...form, reward_group_id: event.target.value })
            }
          >
            {flow.reward_groups.map((group) => (
              <option value={group.id} key={group.id}>
                {group.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Component 名稱">
          <input
            value={form.name}
            placeholder="例如：新戶加碼"
            onChange={(event) =>
              onChange({ ...form, name: event.target.value })
            }
          />
        </Field>
      </div>
      <div className="three-col">
        <Field label="Layer">
          <input
            type="number"
            min="1"
            value={form.layer}
            onChange={(event) =>
              onChange({ ...form, layer: Number(event.target.value) })
            }
          />
        </Field>
        <Field label="Stack Group">
          <select
            value={form.stack_group}
            onChange={(event) =>
              onChange({ ...form, stack_group: event.target.value })
            }
          >
            {stackGroups.map((group) => (
              <option value={group} key={group}>
                {group}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Stack Mode">
          <select
            value={form.stack_mode}
            onChange={(event) => {
              const stackMode = event.target.value as MockStackMode;
              onChange({
                ...form,
                stack_mode: stackMode,
                is_exclusive: stackMode === "EXCLUSIVE",
                is_best_only: stackMode === "BEST_ONLY",
              });
            }}
          >
            {stackModes.map((mode) => (
              <option value={mode} key={mode}>
                {mode}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <Field label="描述">
        <input
          value={form.description}
          onChange={(event) =>
            onChange({ ...form, description: event.target.value })
          }
        />
      </Field>
    </>
  );
}

export function RequirementFields({
  form,
  onChange,
}: {
  form: MockRequirementForm;
  onChange: (form: MockRequirementForm) => void;
}) {
  return (
    <>
      <div className="three-col">
        <Field label="條件類型">
          <select
            value={form.requirement_type}
            onChange={(event) =>
              onChange({
                ...form,
                requirement_type: event.target.value as MockRequirementType,
              })
            }
          >
            {requirementTypes.map((type) => (
              <option value={type} key={type}>
                {type}
              </option>
            ))}
          </select>
        </Field>
        <Field label="運算子">
          <select
            value={form.operator}
            onChange={(event) =>
              onChange({
                ...form,
                operator: event.target.value as MockRequirementOperator,
              })
            }
          >
            {requirementOperatorOptions.map((operator) => (
              <option value={operator.code} key={operator.code}>
                {operator.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="值" hint="多值用逗號，例如 VISA, MASTERCARD">
          <input
            value={form.values}
            onChange={(event) =>
              onChange({ ...form, values: event.target.value })
            }
          />
        </Field>
      </div>
      <Field label="描述">
        <input
          value={form.description}
          onChange={(event) =>
            onChange({ ...form, description: event.target.value })
          }
        />
      </Field>
    </>
  );
}

export function RequirementDirectFields({
  onChange,
  requirement,
}: {
  requirement: MockRequirement;
  onChange: (requirement: MockRequirement) => void;
}) {
  return (
    <>
      <div className="three-col">
        <Field label="條件類型">
          <select
            value={requirement.requirement_type}
            onChange={(event) =>
              onChange({
                ...requirement,
                requirement_type: event.target.value as MockRequirementType,
              })
            }
          >
            {requirementTypes.map((type) => (
              <option value={type} key={type}>
                {type}
              </option>
            ))}
          </select>
        </Field>
        <Field label="運算子">
          <select
            value={requirement.operator}
            onChange={(event) =>
              onChange({
                ...requirement,
                operator: event.target.value as MockRequirementOperator,
              })
            }
          >
            {requirementOperatorOptions.map((operator) => (
              <option value={operator.code} key={operator.code}>
                {operator.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="值">
          <input
            value={displayRequirementValues(requirement)}
            onChange={(event) =>
              onChange({
                ...requirement,
                configuration_json: configForRequirement(
                  requirement.requirement_type,
                  event.target.value,
                ),
              })
            }
          />
        </Field>
      </div>
      <Field label="描述">
        <input
          value={requirement.description}
          onChange={(event) =>
            onChange({ ...requirement, description: event.target.value })
          }
        />
      </Field>
    </>
  );
}

export function BenefitFields<T extends MockBenefit | MockBenefitForm>({
  capPeriodOptions,
  form,
  onChange,
}: {
  capPeriodOptions: CapPeriodOption[];
  form: T;
  onChange: (form: T) => void;
}) {
  return (
    <>
      <div className="three-col">
        <Field label="回饋類型">
          <select
            value={form.benefit_type}
            onChange={(event) =>
              onChange({
                ...form,
                benefit_type: event.target.value as MockBenefitType,
              })
            }
          >
            {benefitTypeOptions.map((type) => (
              <option value={type.code} key={type.code}>
                {type.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="數值">
          <input
            inputMode="decimal"
            value={form.value}
            onChange={(event) =>
              onChange({ ...form, value: event.target.value })
            }
          />
        </Field>
        <Field label="單位">
          <select
            value={form.unit}
            onChange={(event) =>
              onChange({ ...form, unit: event.target.value })
            }
          >
            {rewardUnitOptions.map((unit) => (
              <option value={unit.code} key={unit.code}>
                {unit.name}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <div className="three-col">
        <Field label="上限金額">
          <input
            inputMode="decimal"
            value={form.cap_amount || ""}
            onChange={(event) =>
              onChange({ ...form, cap_amount: event.target.value })
            }
          />
        </Field>
        <Field label="上限週期">
          <select
            value={form.cap_period || "NONE"}
            onChange={(event) =>
              onChange({
                ...form,
                cap_period:
                  event.target.value === "NONE" ? null : event.target.value,
              })
            }
          >
            {capPeriodOptions.map((period) => (
              <option value={period.code} key={period.code}>
                {period.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="幣別">
          <input
            value={form.currency}
            onChange={(event) =>
              onChange({ ...form, currency: event.target.value })
            }
          />
        </Field>
      </div>
      <Field label="描述">
        <input
          value={form.description}
          onChange={(event) =>
            onChange({ ...form, description: event.target.value })
          }
        />
      </Field>
    </>
  );
}

export function ComponentSelect({
  components,
  onChange,
  value,
}: {
  components: MockRewardComponent[];
  value: string;
  onChange: (componentID: string) => void;
}) {
  return (
    <Field label="所屬 Component">
      <select value={value} onChange={(event) => onChange(event.target.value)}>
        {components.map((component) => (
          <option value={component.id} key={component.id}>
            {component.name}
          </option>
        ))}
      </select>
    </Field>
  );
}

export function FormHeading({ icon, title }: { icon: ReactNode; title: string }) {
  return (
    <div className="inline-form-heading">
      {icon}
      <strong>{title}</strong>
    </div>
  );
}
