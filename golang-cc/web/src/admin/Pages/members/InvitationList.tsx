import { Empty } from "../../../components";
import type { Invitation } from "../../AdminApi";

export function InvitationList({
  invites,
  loading,
}: {
  invites: Invitation[];
  loading: boolean;
}) {
  return (
    <section className="panel admin-side-column">
      <div className="section-title"><div><h2>等待接受</h2><p className="muted">{invites.length} 封有效邀請</p></div></div>
      <div className="list">
        {invites.map((invitation) => (
          <div className="compact-row" key={invitation.id}>
            <div><h3>{invitation.email}</h3><p>到期日 {invitation.expires_at.slice(0, 10)}</p></div>
            <span className="status-badge is-pending">待接受</span>
          </div>
        ))}
      </div>
      {!loading && invites.length === 0 && <Empty title="沒有待接受邀請" text="新的邀請會顯示在這裡。" />}
    </section>
  );
}
