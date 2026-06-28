import { MailPlus } from "lucide-react";
import Head from "../../tool/Head";
import { InvitationList } from "./members/InvitationList";
import { InviteDialog } from "./members/InviteDialog";
import { MemberList } from "./members/MemberList";
import { useMembers } from "./members/useMembers";

export default function Members() {
  const members = useMembers();

  return (
    <>
      <Head
        title="會員與邀請"
        text="管理會員狀態、邀請與帳號安全。"
        actions={
          <button className="button" onClick={() => members.setShowInvite(true)}>
            <MailPlus size={17} /> 邀請會員
          </button>
        }
      />
      {members.error && <div className="error preference-message">{members.error}</div>}
      {members.notice && <div className="notice preference-message">{members.notice}</div>}
      <div className="admin-columns">
        <MemberList
          filteredMembers={members.filteredMembers}
          loading={members.loading}
          members={members.members}
          onQueryChange={members.setQuery}
          onReset={(user) => void members.resetPassword(user)}
          onToggle={(user) => void members.toggleStatus(user)}
          processingID={members.processingID}
          query={members.query}
        />
        <InvitationList invites={members.invites} loading={members.loading} />
      </div>
      {members.showInvite && (
        <InviteDialog
          email={members.email}
          inviting={members.inviting}
          onChange={members.setEmail}
          onClose={() => !members.inviting && members.setShowInvite(false)}
          onSubmit={members.inviteMember}
        />
      )}
    </>
  );
}
