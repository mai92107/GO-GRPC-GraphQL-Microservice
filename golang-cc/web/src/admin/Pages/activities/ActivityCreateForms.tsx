import { CirclePlus, Plus } from "lucide-react";
import type { ActivityRequirementOptions, RequirementTypeOption } from "../../AdminApi";
import type { Unit } from "../../../models";
import { Field } from "../../../components";
import {
  emptyBenefitForm,
  emptyRequirementForm,
} from "./activityFlowFactory";
import {
  BenefitFields,
  ComponentFields,
  ComponentSelect,
  FormHeading,
  RequirementFields,
} from "./ActivityFormControls";
import type {
  ActivityFlowModel,
  ActivityBenefitForm,
  ActivityComponentForm,
  ActivityGroupForm,
  ActivityRequirementForm,
  ActivityRewardComponent,
} from "./activityFlowTypes";

type CapPeriodOption = { code: string; name: string };

export function CreateGroupForm({
  form,
  onAdd,
  onChange,
}: {
  form: ActivityGroupForm;
  onAdd: () => void;
  onChange: (form: ActivityGroupForm) => void;
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
  flow: ActivityFlowModel;
  form: ActivityComponentForm;
  onAdd: () => void;
  onChange: (form: ActivityComponentForm) => void;
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
  components: ActivityRewardComponent[];
  form: ActivityRequirementForm;
  requirementOptions: ActivityRequirementOptions | null;
  requirementTypes: RequirementTypeOption[];
  onAdd: () => void;
  onChange: (form: ActivityRequirementForm) => void;
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
  components: ActivityRewardComponent[];
  form: ActivityBenefitForm;
  rewardUnits: Unit[];
  onAdd: () => void;
  onChange: (form: ActivityBenefitForm) => void;
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

