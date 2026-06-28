import { FormEvent } from "react";
import { Dialog, Field } from "../../../components";
import type { AdminUser } from "../../../models";

type Props = {
  availableUsers: AdminUser[];
  chatID: string;
  creating: boolean;
  loading: boolean;
  userID: string;
  onChatIDChange: (chatID: string) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent) => void;
  onUserIDChange: (userID: string) => void;
};

export function CreateTelegramDialog(props: Props) {
  return (
    <Dialog title="新增 Telegram 綁定" onClose={props.onClose}>
      <form className="stack" onSubmit={props.onSubmit}>
        <Field label="Telegram Chat ID">
          <input autoFocus inputMode="numeric" pattern="-?[0-9]+" placeholder="例如：123456789" value={props.chatID} disabled={props.creating} required onChange={(event) => props.onChatIDChange(event.target.value.trim())} />
        </Field>
        <Field label="啟用會員">
          <select value={props.userID} disabled={props.loading || props.creating || !props.availableUsers.length} required onChange={(event) => props.onUserIDChange(event.target.value)}>
            <option value="">請選擇會員</option>
            {props.availableUsers.map((user) => <option value={user.id} key={user.id}>{user.display_name} · {user.email}</option>)}
          </select>
        </Field>
        <div className="dialog-actions">
          <button type="button" className="button ghost" onClick={props.onClose}>取消</button>
          <button className="button" disabled={props.creating || !props.userID || !props.chatID}>{props.creating ? "建立中…" : "建立綁定"}</button>
        </div>
      </form>
    </Dialog>
  );
}
