import { UserRound } from "lucide-react";
import { Empty, SearchField, StatusBadge } from "../../../components";
import type { AdminUser } from "../../../models";

type Props = {
  filteredMembers: AdminUser[];
  loading: boolean;
  members: AdminUser[];
  processingID: string | null;
  query: string;
  onQueryChange: (query: string) => void;
  onReset: (user: AdminUser) => void;
  onToggle: (user: AdminUser) => void;
};

export function MemberList(props: Props) {
  return (
    <section className="panel admin-main-column">
      <div className="section-title"><div><h2>當前會員</h2><p className="muted">{props.members.length} 位會員</p></div></div>
      <SearchField value={props.query} onChange={props.onQueryChange} placeholder="搜尋姓名或 Email" />
      {props.loading ? <p className="muted">載入會員資料中…</p> : <MemberListBody {...props} />}
    </section>
  );
}

function MemberListBody({ filteredMembers, members, processingID, onReset, onToggle }: Props) {
  if (members.length === 0) return <Empty title="尚無會員" text="寄送邀請以新增會員。" />;
  if (filteredMembers.length === 0) return <Empty title="找不到會員" text="請嘗試其他搜尋關鍵字。" />;
  return (
    <div className="list spacious-list">
      {filteredMembers.map((user) => {
        const processing = processingID === user.id;
        return (
          <div className="list-row" key={user.id}>
            <div className="member-identity">
              <span className="member-avatar"><UserRound size={19} /></span>
              <div><h3>{user.display_name}</h3><p>{user.email}</p><StatusBadge active={user.status === "active"} /></div>
            </div>
            <div className="toolbar">
              <button type="button" className="button ghost" disabled={processing} onClick={() => onReset(user)}>重設密碼</button>
              <button type="button" className="button secondary" disabled={processing} onClick={() => onToggle(user)}>{user.status === "active" ? "停用" : "啟用"}</button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
