import { Check, Plus, ShieldCheck, SlidersHorizontal, Trash2 } from "lucide-react";
import type React from "react";
import { Field } from "../../../components";
import type { Merchant, PaymentMethod, Unit } from "../../../models";
import { benefitModeLabel, effectLabel, percentText } from "../../../utils/activityText";
import type { ActivityBenefitInput, Category } from "../../AdminApi";
import { effectOptions, layerOptions } from "./options";
import { ActivityAdvancedFields } from "./ActivityAdvancedFields";
import type { ActivityForm, BenefitEffectType, BenefitMode } from "./types";

type Props = {
  benefits: ActivityBenefitInput[];
  categories: Category[];
  creating: boolean;
  form: ActivityForm;
  merchants: Merchant[];
  methods: PaymentMethod[];
  qualifiedTypes: string[];
  selectableTypes: string[];
  unavailableTypes: Set<string>;
  units: Unit[];
  onAdd: () => void;
  onChange: (form: ActivityForm) => void;
  onModeChange: (mode: BenefitMode) => void;
  onRemove: (index: number) => void;
};

export function ActivityBenefitFields(props: Props) {
  const { benefits, categories, creating, form, onChange } = props;
  const disableAdd = creating || !form.benefit_name.trim() || (form.benefit_mode === "selectable" && (!form.selectable_type || props.unavailableTypes.has(form.selectable_type)));

  return (
    <div className="form-section">
      <div className="form-section-title"><span>2</span><div><h3>優惠條件</h3><p>一個方案可加入多項回饋</p></div></div>
      {benefits.length > 0 && (
        <div className="staged-benefits">
          {benefits.map((benefit, index) => (
            <div className="compact-row" key={`${benefit.name}-${index}`}>
              <div><h3>{benefit.name}</h3><p>{effectLabel(benefit.effect_type)} {benefit.reward_value} · {benefitModeLabel(benefit)}</p></div>
              <button type="button" className="icon-button" aria-label={`移除 ${benefit.name}`} disabled={creating} onClick={() => props.onRemove(index)}><Trash2 size={16} /></button>
            </div>
          ))}
        </div>
      )}
      <div className="two-col">
        <Field label="優惠名稱"><input value={form.benefit_name} disabled={creating} placeholder="例如：網路消費加碼" onChange={(e) => onChange({ ...form, benefit_name: e.target.value })} /></Field>
        <Field label="回饋單位"><select value={form.reward_unit_id} disabled={creating} onChange={(e) => onChange({ ...form, reward_unit_id: e.target.value })}>{props.units.map((unit) => <option value={unit.id} key={unit.id}>{unit.name}</option>)}</select></Field>
      </div>
      <div className="three-col">
        <Field label="效果類型" hint="加碼回饋、折扣或固定金額"><select value={form.effect_type} disabled={creating} onChange={(e) => onChange({ ...form, effect_type: e.target.value as BenefitEffectType })}>{effectOptions.map((option) => <option value={option.value} key={option.value}>{option.label}</option>)}</select></Field>
        <Field label="效果值" hint={form.effect_type === "ADD_CASH" || form.effect_type === "DISCOUNT" ? "固定金額，例如 100" : `目前顯示為 ${percentText(form.reward_value)}`}><input inputMode="decimal" value={form.reward_value} disabled={creating} onChange={(e) => onChange({ ...form, reward_value: e.target.value })} /></Field>
        <Field label="Layer" hint="同層回饋會擇優計算"><select value={form.layer} disabled={creating} onChange={(e) => onChange({ ...form, layer: e.target.value })}>{layerOptions.map((option) => <option value={option.value} key={option.value}>{option.label}</option>)}</select></Field>
      </div>
      <div className="three-col">
        <Field label="消費類別" hint="留白代表無限制"><select value={form.category_id} disabled={creating} onChange={(e) => onChange({ ...form, category_id: e.target.value })}><option value="">無限制</option>{categories.map((category) => <option value={category.id} key={category.id}>{category.name}</option>)}</select></Field>
        <Field label="每月上限" hint="留白代表無上限"><input type="number" min="0" step="0.000001" value={form.monthly_cap} disabled={creating} onChange={(e) => onChange({ ...form, monthly_cap: e.target.value })} /></Field>
        <Field label="顯示順序" hint="同層排序，數字小在前"><input type="number" value={form.display_order} disabled={creating} onChange={(e) => onChange({ ...form, display_order: Number(e.target.value) })} /></Field>
      </div>
      <ModePicker {...props} />
      <ActivityAdvancedFields creating={creating} form={form} merchants={props.merchants} methods={props.methods} onChange={onChange} />
      <button type="button" className="button secondary add-benefit-button" disabled={disableAdd} onClick={props.onAdd}><Plus size={16} /> 加入這項優惠</button>
    </div>
  );
}

function ModePicker({ creating, form, onChange, onModeChange, qualifiedTypes, selectableTypes, unavailableTypes }: Props) {
  return (
    <div className="eligibility-builder">
      <div className="eligibility-heading"><div><strong>這項回饋如何取得？</strong><small>選擇最符合銀行回饋規則的方式</small></div><span>必要設定</span></div>
      <div className="benefit-mode-grid" role="radiogroup" aria-label="回饋取得方式">
        <ModeButton active={form.benefit_mode === "standard"} icon={<Check size={17} />} label="一般回饋" text="持卡即可參加，可另外設定登錄等操作" onClick={() => onModeChange("standard")} />
        <ModeButton active={form.benefit_mode === "selectable"} disabled={!selectableTypes.length} icon={<SlidersHorizontal size={17} />} label="可切換方案" text={selectableTypes.length ? "會員需在銀行 App 切換，推薦時會主動提醒" : "請先到卡片管理建立類別方案"} onClick={() => onModeChange("selectable")} />
        <ModeButton active={form.benefit_mode === "qualified"} disabled={!qualifiedTypes.length} icon={<ShieldCheck size={17} />} label="資格限定" text={qualifiedTypes.length ? "只有符合指定卡片資格者可獲得" : "請先到卡片管理建立類別資格"} onClick={() => onModeChange("qualified")} />
      </div>
      {form.benefit_mode === "selectable" && <SelectableOptions creating={creating} form={form} onChange={onChange} selectableTypes={selectableTypes} unavailableTypes={unavailableTypes} />}
      {form.benefit_mode === "qualified" && <QualifiedOptions creating={creating} form={form} onChange={onChange} qualifiedTypes={qualifiedTypes} />}
    </div>
  );
}

function ModeButton({ active, disabled, icon, label, onClick, text }: { active: boolean; disabled?: boolean; icon: React.ReactNode; label: string; onClick: () => void; text: string }) {
  return <button type="button" className={`benefit-mode-card ${active ? "selected" : ""}`} role="radio" aria-checked={active} disabled={disabled} onClick={onClick}><span className="benefit-mode-icon">{icon}</span><span><strong>{label}</strong><small>{text}</small></span></button>;
}

function SelectableOptions({ creating, form, onChange, selectableTypes, unavailableTypes }: Pick<Props, "creating" | "form" | "onChange" | "selectableTypes" | "unavailableTypes">) {
  return <div className="mode-followup"><div className="field"><span>綁定方案</span><div className="choice-grid">{selectableTypes.map((type) => { const unavailable = unavailableTypes.has(type) && form.selectable_type !== type; return <label className={`choice-chip ${unavailable ? "disabled" : ""}`} key={type}><input type="radio" name="selectable-type" checked={form.selectable_type === type} disabled={creating || unavailable} onChange={() => onChange({ ...form, selectable_type: type })} /><span>{type}{unavailable ? "（期間內已存在）" : ""}</span></label>; })}</div><small>單一回饋只能綁定一個方案；同期間已存在的方案請修改原回饋方案，不可重複新增。</small></div><Field label="切換提醒" hint="推薦這張卡時，會員會看到這段文字。"><input value={form.action_message} disabled={creating} placeholder="例如：請先至銀行 App 切換為「玩旅刷」" onChange={(e) => onChange({ ...form, action_message: e.target.value })} /></Field></div>;
}

function QualifiedOptions({ creating, form, onChange, qualifiedTypes }: Pick<Props, "creating" | "form" | "onChange" | "qualifiedTypes">) {
  return <div className="mode-followup"><div className="field"><span>適用會員資格</span><div className="choice-grid">{qualifiedTypes.map((tier) => <label className="choice-chip" key={tier}><input type="radio" name="qualified-type" checked={form.qualified_type === tier} disabled={creating} onChange={() => onChange({ ...form, qualified_type: tier })} /><span>{tier}</span></label>)}</div><small>這裡只會顯示卡片需符合資格的方案。</small></div></div>;
}
