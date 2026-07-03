import { Layers3 } from "lucide-react";
import { Field } from "../../../components";
import type { ActivityRequirementOptions, RequirementTypeOption } from "../../AdminApi";
import type { Unit } from "../../../models";
import {
  updateBenefit,
  updateComponent,
  updateGroup,
  updateRequirement,
} from "./activityFlowHelpers";
import {
  BenefitFields,
  ComponentFields,
  FormHeading,
  ActiveToggle,
  RequirementDirectFields,
} from "./ActivityFormControls";
import type {
  MockActivityFlow,
  MockBenefit,
  MockComponentForm,
  MockRequirement,
  MockRewardComponent,
  MockRewardGroup,
} from "./mockFlowTypes";

type CapPeriodOption = { code: string; name: string };
type CatalogCardOption = { bank_id: string; bank_name: string; card_product_id: string; card_name: string };

export function ActivityEditor({
  activityCardOptions,
  bankOptions,
  flow,
  onBankChange,
  onCardChange,
  onChange,
}: {
  activityCardOptions: CatalogCardOption[];
  bankOptions: CatalogCardOption[];
  flow: MockActivityFlow;
  onBankChange: (bankID: string) => void;
  onCardChange: (cardProductID: string) => void;
  onChange: (flow: MockActivityFlow) => void;
}) {
  return (
    <section className="form-section selected-editor">
      <div className="form-section-title">
        <span>2</span>
        <div>
          <h3>Activity 編輯</h3>
          <p>活動容器，不放回饋比例與判斷條件</p>
        </div>
      </div>
      <div className="two-col">
        <Field label="銀行">
          <select value={flow.activity.bank_id} onChange={(event) => onBankChange(event.target.value)}>
            {bankOptions.map((bank) => (
              <option value={bank.bank_id} key={bank.bank_id}>
                {bank.bank_name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="卡別">
          <select value={flow.activity.card_product_id} onChange={(event) => onCardChange(event.target.value)}>
            {activityCardOptions.map((card) => (
              <option value={card.card_product_id} key={card.card_product_id}>
                {card.card_name}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <Field label="活動名稱" hint={`右側 demo 顯示為：${flow.activity.bank_name} ${flow.activity.card_name} ${flow.activity.title}`}>
        <input value={flow.activity.title} onChange={(event) => onChange({ ...flow, activity: { ...flow.activity, title: event.target.value } })} />
      </Field>
      <div className="two-col">
        <Field label="開始日期">
          <input type="date" value={flow.activity.effective_from} onChange={(event) => onChange({ ...flow, activity: { ...flow.activity, effective_from: event.target.value } })} />
        </Field>
        <Field label="結束日期">
          <input type="date" value={flow.activity.effective_to} onChange={(event) => onChange({ ...flow, activity: { ...flow.activity, effective_to: event.target.value } })} />
        </Field>
      </div>
      <Field label="狀態">
        <ActiveToggle checked={flow.activity.is_active} onChange={(is_active) => onChange({ ...flow, activity: { ...flow.activity, is_active } })} />
      </Field>
    </section>
  );
}

export function GroupEditor({
  flow,
  group,
  onChange,
}: {
  flow: MockActivityFlow;
  group: MockRewardGroup;
  onChange: (flow: MockActivityFlow) => void;
}) {
  return (
    <section className="form-section selected-editor">
      <FormHeading icon={<Layers3 size={16} />} title="修改 Group" />
      <div className="three-col">
        <Field label="群組名稱">
          <input
            value={group.name}
            onChange={(event) =>
              onChange(
                updateGroup(flow, group.id, { name: event.target.value }),
              )
            }
          />
        </Field>
        <Field label="排序">
          <input
            type="number"
            value={group.display_order}
            onChange={(event) =>
              onChange(
                updateGroup(flow, group.id, {
                  display_order: Number(event.target.value),
                }),
              )
            }
          />
        </Field>
        <Field label="描述">
          <input
            value={group.description}
            onChange={(event) =>
              onChange(
                updateGroup(flow, group.id, {
                  description: event.target.value,
                }),
              )
            }
          />
        </Field>
      </div>
      <Field label="狀態">
        <ActiveToggle
          checked={group.is_active}
          onChange={(is_active) =>
            onChange(updateGroup(flow, group.id, { is_active }))
          }
        />
      </Field>
    </section>
  );
}

export function ComponentEditor({
  component,
  flow,
  onChange,
}: {
  component: MockRewardComponent;
  flow: MockActivityFlow;
  onChange: (flow: MockActivityFlow) => void;
}) {
  const form: MockComponentForm = {
    reward_group_id: component.reward_group_id,
    name: component.name,
    description: component.description,
    layer: component.layer,
    stack_group: component.stack_group,
    stack_mode: component.stack_mode,
    priority: component.priority,
    is_exclusive: component.is_exclusive,
    is_best_only: component.is_best_only,
    effective_from: component.effective_from,
    effective_to: component.effective_to,
    is_active: component.is_active,
  };

  return (
    <section className="form-section selected-editor">
      <FormHeading icon={<Layers3 size={16} />} title="修改 Component" />
      <ComponentFields
        flow={flow}
        form={form}
        onChange={(next) => onChange(updateComponent(flow, component.id, next))}
      />
    </section>
  );
}

export function RequirementEditor({
  flow,
  onChange,
  requirement,
  requirementOptions,
  requirementTypes,
}: {
  flow: MockActivityFlow;
  requirement: MockRequirement;
  requirementOptions: ActivityRequirementOptions | null;
  requirementTypes: RequirementTypeOption[];
  onChange: (flow: MockActivityFlow) => void;
}) {
  return (
    <section className="form-section selected-editor">
      <FormHeading icon={<Layers3 size={16} />} title="修改 Requirement" />
      <RequirementDirectFields
        requirement={requirement}
        requirementOptions={requirementOptions}
        requirementTypes={requirementTypes}
        onChange={(next) =>
          onChange(updateRequirement(flow, requirement.id, next))
        }
      />
    </section>
  );
}

export function BenefitEditor({
  benefit,
  capPeriodOptions,
  flow,
  onChange,
  rewardUnits,
}: {
  benefit: MockBenefit;
  capPeriodOptions: CapPeriodOption[];
  flow: MockActivityFlow;
  rewardUnits: Unit[];
  onChange: (flow: MockActivityFlow) => void;
}) {
  return (
    <section className="form-section selected-editor">
      <FormHeading icon={<Layers3 size={16} />} title="修改 Benefit" />
      <BenefitFields
        capPeriodOptions={capPeriodOptions}
        form={benefit}
        rewardUnits={rewardUnits}
        onChange={(next) => onChange(updateBenefit(flow, benefit.id, next))}
      />
    </section>
  );
}
