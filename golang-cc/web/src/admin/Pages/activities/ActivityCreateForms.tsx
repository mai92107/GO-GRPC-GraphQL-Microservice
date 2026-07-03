import { CirclePlus, Plus } from "lucide-react";
import type { ActivityRequirementOptions, RequirementTypeOption } from "../../AdminApi";
import type { Unit } from "../../../models";
import { Field } from "../../../components";
import {
  emptyBenefitForm,
  emptyRequirementForm,
} from "./mockFlowHelpers";
import {
  BenefitFields,
  ComponentFields,
  ComponentSelect,
  FormHeading,
  RequirementFields,
} from "./ActivityFormControls";
import type {
  MockActivityFlow,
  MockBenefitForm,
  MockComponentForm,
  MockGroupForm,
  MockRequirementForm,
  MockRewardComponent,
} from "./mockFlowTypes";

type CapPeriodOption = { code: string; name: string };

export function CreateGroupForm({
  form,
  onAdd,
  onChange,
}: {
  form: MockGroupForm;
  onAdd: () => void;
  onChange: (form: MockGroupForm) => void;
}) {
  return (
    <section className="form-section compact-form-section">
      <FormHeading icon={<CirclePlus size={16} />} title="新增 Group" />
      <div className="three-col">
        <Field label="群組名稱">
          <input
            value={form.name}
            placeholder="例如：Apple Pay"
            onChange={(event) =>
              onChange({ ...form, name: event.target.value })
            }
          />
        </Field>
        <Field label="排序">
          <input
            type="number"
            value={form.display_order}
            onChange={(event) =>
              onChange({ ...form, display_order: Number(event.target.value) })
            }
          />
        </Field>
        <Field label="描述">
          <input
            value={form.description}
            onChange={(event) =>
              onChange({ ...form, description: event.target.value })
            }
          />
        </Field>
      </div>
      <button className="button secondary" type="button" onClick={onAdd}>
        <Plus size={16} /> 新增 Group
      </button>
    </section>
  );
}

export function CreateComponentForm({
  flow,
  form,
  onAdd,
  onChange,
}: {
  flow: MockActivityFlow;
  form: MockComponentForm;
  onAdd: () => void;
  onChange: (form: MockComponentForm) => void;
}) {
  return (
    <section className="form-section compact-form-section">
      <FormHeading icon={<CirclePlus size={16} />} title="新增 Component" />
      <ComponentFields flow={flow} form={form} onChange={onChange} />
      <button className="button secondary" type="button" onClick={onAdd}>
        <Plus size={16} /> 新增 Component
      </button>
    </section>
  );
}

export function CreateRequirementForm({
  components,
  form,
  requirementOptions,
  requirementTypes,
  onAdd,
  onChange,
}: {
  components: MockRewardComponent[];
  form: MockRequirementForm;
  requirementOptions: ActivityRequirementOptions | null;
  requirementTypes: RequirementTypeOption[];
  onAdd: () => void;
  onChange: (form: MockRequirementForm) => void;
}) {
  return (
    <section className="form-section compact-form-section">
      <FormHeading icon={<CirclePlus size={16} />} title="新增 Requirement" />
      <ComponentSelect
        components={components}
        value={form.reward_component_id}
        onChange={(componentID) => onChange(emptyRequirementForm(componentID))}
      />
      <RequirementFields form={form} requirementOptions={requirementOptions} requirementTypes={requirementTypes} onChange={onChange} />
      <button className="button secondary" type="button" onClick={onAdd}>
        <Plus size={16} /> 新增 Requirement
      </button>
    </section>
  );
}

export function CreateBenefitForm({
  capPeriodOptions,
  components,
  form,
  rewardUnits,
  onAdd,
  onChange,
}: {
  capPeriodOptions: CapPeriodOption[];
  components: MockRewardComponent[];
  form: MockBenefitForm;
  rewardUnits: Unit[];
  onAdd: () => void;
  onChange: (form: MockBenefitForm) => void;
}) {
  return (
    <section className="form-section compact-form-section">
      <FormHeading icon={<CirclePlus size={16} />} title="新增 Benefit" />
      <ComponentSelect
        components={components}
        value={form.reward_component_id}
        onChange={(componentID) => onChange(emptyBenefitForm(componentID))}
      />
      <BenefitFields
        capPeriodOptions={capPeriodOptions}
        form={form}
        rewardUnits={rewardUnits}
        onChange={onChange}
      />
      <button className="button secondary" type="button" onClick={onAdd}>
        <Plus size={16} /> 新增 Benefit
      </button>
    </section>
  );
}

