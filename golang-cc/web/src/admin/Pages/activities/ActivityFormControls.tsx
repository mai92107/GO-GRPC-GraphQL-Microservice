import type { ReactNode } from "react";
import type { ActivityRequirementOptions, RequirementTypeOption } from "../../AdminApi";
import type { Unit } from "../../../models";
import { Field } from "../../../components";
import { benefitTypeOptions, requirementOperatorOptions } from "./activityFlowSettings";
import {
  configForRequirement,
  displayRequirementValues,
} from "./activityFlowHelpers";
import {
  isUnconditionalRequirement,
  normalizeRequirementFormType,
  unconditionalRequirementOperator,
} from "./requirementHelpers";
import type {
  ActivityFlowModel,
  ActivityBenefit,
  ActivityBenefitForm,
  ActivityBenefitType,
  ActivityComponentForm,
  ActivityRequirement,
  ActivityRequirementForm,
  ActivityRequirementOperator,
  ActivityRequirementType,
  ActivityRewardComponent,
  ActivityStackMode,
} from "./activityFlowTypes";

const stackModes: ActivityStackMode[] = ["ADDITIVE", "BEST_ONLY", "EXCLUSIVE"];
const stackGroups = [
  "BASE",
  "PAYMENT",
  "NETWORK",
  "MERCHANT",
  "NEW_USER",
  "ACCOUNT_TIER",
  "CAMPAIGN",
];

type CapPeriodOption = { code: string; name: string };

type RequirementOptionProps = {
  requirementOptions?: ActivityRequirementOptions | null;
  requirementTypes?: RequirementTypeOption[];
};

function operatorOptions(options?: ActivityRequirementOptions | null) {
  return options?.operators?.length ? options.operators : requirementOperatorOptions;
}

export function ActiveToggle({
  checked,
  onChange,
}: {
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <label className="activity-active-toggle">
      <input
        type="checkbox"
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
      />
      <span>{checked ? "啟用" : "停用"}</span>
    </label>
  );
}

export function ComponentFields({
  flow,
  form,
  onChange,
}: {
  flow: ActivityFlowModel;
  form: ActivityComponentForm;
  onChange: (form: ActivityComponentForm) => void;
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
            onChange={(event) =>
              onChange({
                ...form,
                stack_mode: event.target.value as ActivityStackMode,
              })
            }
          >
            {stackModes.map((mode) => (
              <option value={mode} key={mode}>
                {mode}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <div className="two-col">
        <Field label="描述">
          <input
            value={form.description}
            onChange={(event) =>
              onChange({ ...form, description: event.target.value })
            }
          />
        </Field>
        <Field label="狀態">
          <ActiveToggle
            checked={form.is_active}
            onChange={(is_active) => onChange({ ...form, is_active })}
          />
        </Field>
      </div>
    </>
  );
}

export function RequirementFields({
  form,
  onChange,
  requirementOptions,
  requirementTypes,
}: {
  form: ActivityRequirementForm;
  onChange: (form: ActivityRequirementForm) => void;
} & RequirementOptionProps) {
  const isUnconditional = isUnconditionalRequirement(form.requirement_type);

  return (
    <>
      <div className="three-col">
        <Field label="條件類型">
          <select
            value={form.requirement_type}
            onChange={(event) => {
              const requirement_type = event.target.value as ActivityRequirementType;
              onChange({
                ...form,
                ...normalizeRequirementFormType(requirement_type),
              });
            }}
          >
            {(requirementTypes || []).map((type) => (
              <option value={type.code} key={type.code}>
                {type.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="運算子">
          <select
            value={
              isUnconditional ? unconditionalRequirementOperator : form.operator
            }
            disabled={isUnconditional}
            onChange={(event) =>
              onChange({
                ...form,
                operator: event.target.value as ActivityRequirementOperator,
              })
            }
          >
            {operatorOptions(requirementOptions).map((operator) => (
              <option value={operator.code} key={operator.code}>
                {operator.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="值" hint="多值用逗號，例如 VISA, MASTERCARD">
          <input
            value={form.values}
            disabled={isUnconditional}
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
  requirementOptions,
  requirementTypes,
}: {
  requirement: ActivityRequirement;
  onChange: (requirement: ActivityRequirement) => void;
} & RequirementOptionProps) {
  const isUnconditional = isUnconditionalRequirement(
    requirement.requirement_type,
  );

  return (
    <>
      <div className="three-col">
        <Field label="條件類型">
          <select
            value={requirement.requirement_type}
            onChange={(event) => {
              const requirement_type = event.target.value as ActivityRequirementType;
              const nextIsUnconditional =
                isUnconditionalRequirement(requirement_type);
              onChange({
                ...requirement,
                requirement_type,
                operator: nextIsUnconditional
                  ? unconditionalRequirementOperator
                  : requirement.operator,
                configuration_json: {},
                description:
                  isUnconditional || nextIsUnconditional
                    ? ""
                    : requirement.description,
              });
            }}
          >
            {(requirementTypes || []).map((type) => (
              <option value={type.code} key={type.code}>
                {type.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="運算子">
          <select
            value={
              isUnconditional
                ? unconditionalRequirementOperator
                : requirement.operator
            }
            disabled={isUnconditional}
            onChange={(event) =>
              onChange({
                ...requirement,
                operator: event.target.value as ActivityRequirementOperator,
              })
            }
          >
            {operatorOptions(requirementOptions).map((operator) => (
              <option value={operator.code} key={operator.code}>
                {operator.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="值">
          <input
            value={displayRequirementValues(requirement)}
            disabled={isUnconditional}
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
      <div className="two-col">
        <Field label="描述">
          <input
            value={requirement.description}
            onChange={(event) =>
              onChange({ ...requirement, description: event.target.value })
            }
          />
        </Field>
        <Field label="狀態">
          <ActiveToggle
            checked={requirement.is_active}
            onChange={(is_active) => onChange({ ...requirement, is_active })}
          />
        </Field>
      </div>
    </>
  );
}

export function BenefitFields<T extends ActivityBenefit | ActivityBenefitForm>({
  capPeriodOptions,
  form,
  onChange,
  rewardUnits,
}: {
  capPeriodOptions: CapPeriodOption[];
  form: T;
  onChange: (form: T) => void;
  rewardUnits: Unit[];
}) {
  const units = rewardUnits.length
    ? rewardUnits.map((unit) => ({ code: unit.id, name: unit.name }))
    : [{ code: "PERCENT", name: "百分比" }];
  return (
    <>
      <div className="three-col">
        <Field label="回饋類型">
          <select
            value={form.benefit_type}
            onChange={(event) =>
              onChange({
                ...form,
                benefit_type: event.target.value as ActivityBenefitType,
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
        <Field label="回饋單位">
          <select
            value={form.reward_unit_id}
            onChange={(event) =>
              onChange({ ...form, reward_unit_id: event.target.value })
            }
          >
            {units.map((unit) => (
              <option value={unit.code} key={unit.code}>
                {unit.name}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <div className="two-col">
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
      </div>
      <div className="two-col">
        <Field label="描述">
          <input
            value={form.description}
            onChange={(event) =>
              onChange({ ...form, description: event.target.value })
            }
          />
        </Field>
        <Field label="狀態">
          <ActiveToggle
            checked={form.is_active}
            onChange={(is_active) => onChange({ ...form, is_active })}
          />
        </Field>
      </div>
    </>
  );
}

export function ComponentSelect({
  components,
  onChange,
  value,
}: {
  components: ActivityRewardComponent[];
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
