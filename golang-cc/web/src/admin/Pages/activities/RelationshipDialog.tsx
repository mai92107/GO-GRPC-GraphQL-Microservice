import { FormEvent } from "react";
import { Dialog, Field } from "../../../components";
import { benefitModeLabel, effectLabel } from "../../../utils/activityText";
import type { Activity, ActivityBenefitInput } from "../../AdminApi";
import { layerOptions } from "./options";

type Props = {
  activity: Activity;
  benefits: ActivityBenefitInput[];
  processingID: string | null;
  onChange: (benefits: ActivityBenefitInput[]) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent) => void;
};

export function RelationshipDialog({
  activity,
  benefits,
  processingID,
  onChange,
  onClose,
  onSubmit,
}: Props) {
  const saving = processingID === activity.id;

  return (
    <Dialog title="編輯回饋層級" onClose={onClose}>
      <form className="stack activity-form relationship-form" onSubmit={onSubmit}>
        <div className="form-section">
          <div className="form-section-title"><span>1</span><div><h3>{activity.name}</h3><p>調整方案層級與同群疊加/互斥關係</p></div></div>
          <div className="stack">
            {benefits.map((benefit, index) => (
              <RelationshipRow
                benefit={benefit}
                disabled={saving}
                index={index}
                key={benefit.id || index}
                onChange={onChange}
                rows={benefits}
              />
            ))}
          </div>
        </div>
        <div className="dialog-actions">
          <button type="button" className="button ghost" disabled={saving} onClick={onClose}>取消</button>
          <button className="button" disabled={saving}>{saving ? "儲存中…" : "儲存層級關係"}</button>
        </div>
      </form>
    </Dialog>
  );
}

function RelationshipRow({
  benefit,
  disabled,
  index,
  onChange,
  rows,
}: {
  benefit: ActivityBenefitInput;
  disabled: boolean;
  index: number;
  onChange: (benefits: ActivityBenefitInput[]) => void;
  rows: ActivityBenefitInput[];
}) {
  const update = (next: Partial<ActivityBenefitInput>) =>
    onChange(rows.map((item, itemIndex) => (itemIndex === index ? { ...item, ...next } : item)));

  return (
    <div className="relationship-editor">
      <div className="relationship-editor-head">
        <div><h3>{benefit.name}</h3><p className="muted">{effectLabel(benefit.effect_type)} {benefit.reward_value}</p></div>
        <span>{benefitModeLabel(benefit)}</span>
      </div>
      <div className="two-col">
        <Field label="Layer">
          <select value={benefit.layer} disabled={disabled} onChange={(event) => update({ layer: event.target.value })}>
            {layerOptions.map((option) => <option value={option.value} key={option.value}>{option.label}</option>)}
          </select>
        </Field>
        <Field label="水平關係">
          <select value={benefit.stack_group} disabled={disabled} onChange={(event) => update({ stack_group: event.target.value })}>
            <option value="stack">可疊加</option>
            <option value="best_of_group">同群擇優</option>
            <option value="exclusive">互斥</option>
          </select>
        </Field>
      </div>
      <div className="two-col">
        <Field label="累計層"><input value={benefit.stack_group} disabled={disabled} onChange={(event) => update({ stack_group: event.target.value })} /></Field>
      </div>
    </div>
  );
}
