import { FormEvent } from "react";
import { Dialog } from "../../../components";
import type { CardNetwork } from "../../../models";
import type { Bank } from "../../AdminApi";
import { CardFormFields } from "./CardFormFields";
import type { CardForm } from "./types";

type Props = {
  banks: Bank[];
  creating: boolean;
  form: CardForm;
  networks: CardNetwork[];
  onChange: (form: CardForm) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent) => void;
};

export function CreateCardDialog({
  banks,
  creating,
  form,
  networks,
  onChange,
  onClose,
  onSubmit,
}: Props) {
  const disabled =
    creating || !form.bank_id || !form.name.trim() || !form.network_ids.length;

  return (
    <Dialog title="新增卡片" onClose={onClose}>
      <form className="stack" onSubmit={onSubmit}>
        <CardFormFields
          autoFocus
          banks={banks}
          disabled={creating}
          form={form}
          networks={networks}
          onChange={onChange}
        />
        <div className="dialog-actions">
          <button type="button" className="button ghost" onClick={onClose}>
            取消
          </button>
          <button className="button" disabled={disabled}>
            {creating ? "新增中…" : "新增卡片"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}
