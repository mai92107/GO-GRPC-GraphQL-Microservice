import { Field } from "../../../components";
import type { Merchant, PaymentMethod } from "../../../models";
import type { ActivityBenefitInput } from "../../AdminApi";
import type { ActivityForm } from "./types";

type Props = {
  creating: boolean;
  form: ActivityForm;
  merchants: Merchant[];
  methods: PaymentMethod[];
  onChange: (form: ActivityForm) => void;
};

export function ActivityAdvancedFields({
  creating,
  form,
  merchants,
  methods,
  onChange,
}: Props) {
  return (
    <details className="advanced-fields">
      <summary>進階限制與操作條件</summary>
      <div className="stack">
        <div className="two-col">
          {form.benefit_mode === "standard" && (
            <Field label="其他前置操作">
              <select value={form.action_required} disabled={creating} onChange={(e) => onChange({ ...form, action_required: e.target.value as ActivityBenefitInput["action_required"] })}>
                <option value="none">無</option>
                <option value="registration">需登錄</option>
                <option value="account_setup">需帳戶設定</option>
              </select>
            </Field>
          )}
        </div>
        <details className="option-disclosure">
          <summary><span>限定支付方式</span><small>{form.payment_methods.length ? `已選 ${form.payment_methods.length} 項` : "不限支付方式"}</small></summary>
          <div className="choice-grid">
            {methods.map((method) => (
              <label className="choice-chip" key={method.id}>
                <input type="checkbox" checked={form.payment_methods.includes(method.id)} disabled={creating} onChange={(e) => onChange({ ...form, payment_methods: e.target.checked ? [...form.payment_methods, method.id] : form.payment_methods.filter((id) => id !== method.id) })} />
                <span>{method.name}</span>
              </label>
            ))}
          </div>
        </details>
        <details className="option-disclosure">
          <summary><span>限定店家</span><small>{form.merchant_ids.length ? `已選 ${form.merchant_ids.length} 家` : "不限店家"}</small></summary>
          <div className="choice-grid merchant-choices">
            {merchants.map((merchant) => (
              <label className="choice-chip" key={merchant.id}>
                <input type="checkbox" checked={form.merchant_ids.includes(merchant.id!)} disabled={creating} onChange={(e) => onChange({ ...form, merchant_ids: e.target.checked ? [...form.merchant_ids, merchant.id!] : form.merchant_ids.filter((id) => id !== merchant.id!) })} />
                <span>{merchant.name}</span>
              </label>
            ))}
          </div>
        </details>
        <Field label="操作提醒">
          <input value={form.action_message} disabled={creating} placeholder="例如：每月需至 App 登錄" onChange={(e) => onChange({ ...form, action_message: e.target.value })} />
        </Field>
      </div>
    </details>
  );
}
