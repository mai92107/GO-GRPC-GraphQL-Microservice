import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import type { AdminUser, TelegramBinding } from "../../../models";
import { deleteTelegramBinding, getTelegramBindings, getUsers, postTelegramBinding } from "../../AdminApi";

export function useTelegramBindings() {
  const [bindings, setBindings] = useState<TelegramBinding[]>([]);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [chatID, setChatID] = useState("");
  const [userID, setUserID] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [deletingChatID, setDeletingChatID] = useState<number | null>(null);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const availableFrom = (nextUsers: AdminUser[], nextBindings: TelegramBinding[]) =>
    nextUsers.filter((user) => user.role === "member" && user.status === "active" && !nextBindings.some((binding) => binding.user_id === user.id));

  const load = useCallback(async () => {
    try {
      const [nextBindings, nextUsers] = await Promise.all([getTelegramBindings(), getUsers()]);
      setBindings(nextBindings); setUsers(nextUsers);
      const available = availableFrom(nextUsers, nextBindings);
      setUserID((current) => available.some((user) => user.id === current) ? current : available[0]?.id || "");
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { void load(); }, [load]);
  const availableUsers = useMemo(() => availableFrom(users, bindings), [bindings, users]);

  async function create(event: FormEvent) {
    event.preventDefault(); setError(""); setNotice("");
    const parsedChatID = Number(chatID);
    if (!Number.isSafeInteger(parsedChatID)) { setError("Chat ID 格式無效"); return; }
    setCreating(true);
    try {
      await postTelegramBinding(parsedChatID, userID);
      setChatID(""); setShowCreate(false); setNotice("Telegram Chat 綁定完成");
      await load();
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setCreating(false); }
  }

  async function remove(binding: TelegramBinding) {
    if (!confirm(`確定解除 ${binding.display_name} 的 Telegram 綁定？`)) return;
    setDeletingChatID(binding.chat_id); setError(""); setNotice("");
    try { await deleteTelegramBinding(binding.chat_id); setNotice("Telegram Chat 綁定已解除"); await load(); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setDeletingChatID(null); }
  }

  return {
    availableUsers, bindings, chatID, creating, deletingChatID, error, loading,
    notice, showCreate, userID, create, remove, setChatID, setShowCreate,
    setUserID,
  };
}
