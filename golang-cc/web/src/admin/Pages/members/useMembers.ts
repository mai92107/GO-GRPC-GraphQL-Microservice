import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import type { AdminUser } from "../../../models";
import { getInvitations, getUsers, Invitation, invite, toggleActivate, toggleReset } from "../../AdminApi";

export function useMembers() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [invites, setInvites] = useState<Invitation[]>([]);
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(true);
  const [inviting, setInviting] = useState(false);
  const [processingID, setProcessingID] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [query, setQuery] = useState("");
  const [showInvite, setShowInvite] = useState(false);

  const load = useCallback(async () => {
    try {
      const [nextUsers, nextInvites] = await Promise.all([getUsers(), getInvitations()]);
      setUsers(nextUsers); setInvites(nextInvites);
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { void load(); }, [load]);
  const members = useMemo(() => users.filter((user) => user.role === "member"), [users]);
  const filteredMembers = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return keyword ? members.filter((user) => `${user.display_name} ${user.email}`.toLowerCase().includes(keyword)) : members;
  }, [members, query]);

  async function inviteMember(event: FormEvent) {
    event.preventDefault(); setInviting(true); setError(""); setNotice("");
    try {
      await invite(email.trim());
      setEmail(""); setShowInvite(false); setNotice("邀請已寄出。");
      await load();
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setInviting(false); }
  }

  async function resetPassword(user: AdminUser) {
    setProcessingID(user.id); setError(""); setNotice("");
    try { await toggleReset(user.id); setNotice(`已寄送 ${user.display_name} 的密碼重設信。`); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessingID(null); }
  }

  async function toggleStatus(user: AdminUser) {
    setProcessingID(user.id); setError(""); setNotice("");
    try { await toggleActivate(user.id, user.status); await load(); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessingID(null); }
  }

  return {
    email, error, filteredMembers, invites, inviting, loading, members, notice,
    processingID, query, showInvite, inviteMember, resetPassword, setEmail,
    setQuery, setShowInvite, toggleStatus,
  };
}
