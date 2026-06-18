import { Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Field } from "../../components";
import type { AdminUser, TelegramBinding } from "../../models";
import Head from "../../tool/Head";
import {
  deleteTelegramBinding,
  getTelegramBindings,
  getUsers,
  postTelegramBinding,
} from "../AdminApi";

export default function TelegramBindings() {
  const [bindings, setBindings] = useState<TelegramBinding[]>([]);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [chatID, setChatID] = useState("");
  const [userID, setUserID] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [deletingChatID, setDeletingChatID] = useState<number | null>(null);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      const [nextBindings, nextUsers] = await Promise.all([
        getTelegramBindings(),
        getUsers(),
      ]);
      setBindings(nextBindings);
      setUsers(nextUsers);
      const available = nextUsers.filter(
        (user) =>
          user.role === "member" &&
          user.status === "active" &&
          !nextBindings.some((binding) => binding.user_id === user.id),
      );
      setUserID((current) =>
        available.some((user) => user.id === current)
          ? current
          : available[0]?.id || "",
      );
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const availableUsers = useMemo(
    () =>
      users.filter(
        (user) =>
          user.role === "member" &&
          user.status === "active" &&
          !bindings.some((binding) => binding.user_id === user.id),
      ),
    [bindings, users],
  );

  async function create(event: FormEvent) {
    event.preventDefault();
    setError("");
    setNotice("");
    const parsedChatID = Number(chatID);
    if (!Number.isSafeInteger(parsedChatID)) {
      setError("Chat ID 格式無效");
      return;
    }

    setCreating(true);
    try {
      await postTelegramBinding(parsedChatID, userID);
      setChatID("");
      setNotice("Telegram Chat 綁定完成");
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function remove(binding: TelegramBinding) {
    if (!confirm(`確定解除 ${binding.display_name} 的 Telegram 綁定？`)) {
      return;
    }
    setDeletingChatID(binding.chat_id);
    setError("");
    setNotice("");
    try {
      await deleteTelegramBinding(binding.chat_id);
      setNotice("Telegram Chat 綁定已解除");
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setDeletingChatID(null);
    }
  }

  return (
    <>
      <Head
        title="Telegram 綁定"
        text="將會員與 Telegram Chat ID 綁定後，會員即可透過 Bot 取得卡片推薦。"
      />
      {error && <div className="error preference-message">{error}</div>}
      {notice && <div className="notice preference-message">{notice}</div>}
      <div className="settings-grid">
        <section className="panel">
          <h2>新增綁定</h2>
          <p className="muted">
            會員對 Bot 輸入 /recommand，首次會回覆 Chat ID。
          </p>
          <form className="stack" onSubmit={create}>
            <Field label="Telegram Chat ID">
              <input
                inputMode="numeric"
                pattern="-?[0-9]+"
                placeholder="例如：123456789"
                value={chatID}
                disabled={creating}
                required
                onChange={(event) => setChatID(event.target.value.trim())}
              />
            </Field>
            <Field label="啟用會員">
              <select
                value={userID}
                disabled={loading || creating || !availableUsers.length}
                required
                onChange={(event) => setUserID(event.target.value)}
              >
                <option value="">
                  {availableUsers.length
                    ? "請選擇會員"
                    : "沒有可綁定的啟用會員"}
                </option>
                {availableUsers.map((user) => (
                  <option value={user.id} key={user.id}>
                    {user.display_name} · {user.email}
                  </option>
                ))}
              </select>
            </Field>
            <button
              className="button"
              disabled={creating || !userID || !availableUsers.length}
            >
              {creating ? "建立中…" : "建立綁定"}
            </button>
          </form>
        </section>

        <section className="panel">
          <h2>目前綁定</h2>
          {loading ? (
            <p className="muted">載入 Telegram 綁定中…</p>
          ) : (
            <div className="list">
              {bindings.map((binding) => (
                <div className="list-row" key={binding.chat_id}>
                  <div>
                    <h3>{binding.display_name}</h3>
                    <p>
                      {binding.email}
                      <br />
                      Chat ID：{binding.chat_id}
                    </p>
                  </div>
                  <button
                    type="button"
                    className="icon-button"
                    aria-label={`解除 ${binding.display_name} Telegram 綁定`}
                    disabled={deletingChatID === binding.chat_id}
                    onClick={() => void remove(binding)}
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              ))}
            </div>
          )}
          {!loading && bindings.length === 0 && (
            <p className="muted">目前沒有 Telegram 綁定。</p>
          )}
        </section>
      </div>
    </>
  );
}
