import { Field } from "../../../components";
import type { Bank } from "../../AdminApi";
import type { CardForm } from "./types";
import { toggleNetwork } from "./types";

type Props = {
  banks: Bank[];
  disabled: boolean;
  form: CardForm;
  networks: string[];
  onChange: (form: CardForm) => void;
  autoFocus?: boolean;
};

export function CardFormFields({
  banks,
  disabled,
  form,
  networks,
  onChange,
  autoFocus = false,
}: Props) {
  return (
    <>
      <Field label="發卡銀行">
        <select
          value={form.bank_id}
          disabled={disabled || !banks.length}
          onChange={(event) => onChange({ ...form, bank_id: event.target.value })}
        >
          {banks.map((bank) => (
            <option value={bank.id} key={bank.id}>
              {bank.name}
            </option>
          ))}
        </select>
      </Field>
      <Field label="卡片名稱">
        <input
          autoFocus={autoFocus}
          value={form.name}
          disabled={disabled}
          onChange={(event) => onChange({ ...form, name: event.target.value })}
          placeholder="例如：U Bear 信用卡"
          required
        />
      </Field>
      <div className="field">
        <span>發卡別</span>
        <div className="choice-grid">
          {networks.map((network) => (
            <label className="choice-chip" key={network}>
              <input
                type="checkbox"
                checked={form.networks.includes(network)}
                disabled={disabled}
                onChange={(event) =>
                  onChange({
                    ...form,
                    networks: toggleNetwork(
                      form.networks,
                      network,
                      event.target.checked,
                    ),
                  })
                }
              />
              <span>{network}</span>
            </label>
          ))}
        </div>
        <small>會員新增持有卡片時，只能從這裡勾選的發卡別中選擇。</small>
      </div>
      <Field label="用戶須達標資格" hint="多個值請用逗號分隔。">
        <input
          value={form.qualified_type}
          disabled={disabled}
          onChange={(event) =>
            onChange({ ...form, qualified_type: event.target.value })
          }
          placeholder="例如：大戶Plus"
        />
      </Field>
      <Field label="用戶可選方案" hint="只作回饋提示，不給用戶點選。">
        <input
          value={form.selectable_type}
          disabled={disabled}
          onChange={(event) =>
            onChange({ ...form, selectable_type: event.target.value })
          }
          placeholder="例如：大大,大戶"
        />
      </Field>
    </>
  );
}
