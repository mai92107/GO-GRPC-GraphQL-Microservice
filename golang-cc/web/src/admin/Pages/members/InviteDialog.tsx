import { FormEvent } from "react";
import { Dialog, Field } from "../../../components";

type Props = {
  email: string;
  inviting: boolean;
  onChange: (email: string) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent) => void;
};

export function InviteDialog({ email, inviting, onChange, onClose, onSubmit }: Props) {
  return (
    <Dialog title="邀請新會員" onClose={onClose}>
      <form className="stack" onSubmit={onSubmit}>
        <Field label="會員 Email" hint="系統會寄送一次性的帳號啟用連結。">
          <input autoFocus type="email" value={email} placeholder="name@example.com" disabled={inviting} required onChange={(event) => onChange(event.target.value)} />
        </Field>
        <div className="dialog-actions">
          <button type="button" className="button ghost" onClick={onClose}>取消</button>
          <button className="button" disabled={inviting || !email.trim()}>{inviting ? "寄送中…" : "寄送啟用連結"}</button>
        </div>
      </form>
    </Dialog>
  );
}
