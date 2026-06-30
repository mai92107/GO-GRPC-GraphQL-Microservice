import { FormEvent } from "react";
import { Dialog } from "../../../components";
import type { CatalogCard } from "../../../models";
import type { Bank } from "../../AdminApi";
import { CardFormFields } from "./CardFormFields";
import type { CardForm } from "./types";

type Props = {
  banks: Bank[];
  card: CatalogCard;
  form: CardForm;
  networks: string[];
  onChange: (form: CardForm) => void;
  onClose: () => void;
  onDelete: () => void;
  onSave: (event: FormEvent) => void;
  onToggleActive: () => void;
  saving: boolean;
};

export function EditCardDialog({
  banks,
  card,
  form,
  networks,
  onChange,
  onClose,
  onDelete,
  onSave,
  onToggleActive,
  saving,
}: Props) {
  return (
    <Dialog title="修改卡片資訊" onClose={onClose}>
      <form className="stack" onSubmit={onSave}>
        <CardFormFields
          banks={banks}
          disabled={saving}
          form={form}
          networks={networks}
          onChange={onChange}
        />
        <p className="muted">目前狀態：{card.is_active ? "啟用" : "停用"}</p>
        <div className="dialog-actions split-actions">
          <button
            type="button"
            className="button danger"
            disabled={saving}
            onClick={onDelete}
          >
            刪除卡片
          </button>
          <span />
          <button
            className="button"
            disabled={saving || !form.networks.length}
          >
            {saving ? "處理中…" : "儲存"}
          </button>
          <button
            type="button"
            className="button secondary"
            disabled={saving}
            onClick={onToggleActive}
          >
            {card.is_active ? "停用" : "啟用"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}
