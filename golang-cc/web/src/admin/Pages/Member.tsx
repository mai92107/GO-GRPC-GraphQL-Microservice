import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Empty, Field } from "../../components";
import type { AdminUser } from "../../models";
import Head from "../../tool/Head";
import {
  getInvitations,
  getUsers,
  Invitation,
  invite,
  toggleActivate,
  toggleReset,
} from "../AdminApi";

export default function Members() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [invites, setInvites] = useState<Invitation[]>([]);
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(true);
  const [inviting, setInviting] = useState(false);
  const [processingID, setProcessingID] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    try {
      const [nextUsers, nextInvites] = await Promise.all([
        getUsers(),
        getInvitations(),
      ]);
      setUsers(nextUsers);
      setInvites(nextInvites);
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const members = useMemo(
    () => users.filter((user) => user.role === "member"),
    [users],
  );

  async function inviteMember(event: FormEvent) {
    event.preventDefault();
    setInviting(true);
    setError("");
    setNotice("");
    try {
      await invite(email.trim());
      setEmail("");
      setNotice("邀請已寄出。");
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setInviting(false);
    }
  }

  async function resetPassword(user: AdminUser) {
    setProcessingID(user.id);
    setError("");
    setNotice("");
    try {
      await toggleReset(user.id);
      setNotice(`已寄送 ${user.display_name} 的密碼重設信。`);
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }

  async function toggleStatus(user: AdminUser) {
    setProcessingID(user.id);
    setError("");
    setNotice("");
    try {
      await toggleActivate(user.id, user.status);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }

  return (
    <>
      <Head
        title="會員與邀請"
        text="只有管理員能邀請、停用會員與寄送密碼重設信。"
      />
      {error && <div className="error">{error}</div>}
      {notice && <div className="notice">{notice}</div>}
      <div className="settings-grid">
        <section className="panel">
          <h2>邀請會員</h2>
          <form className="stack" onSubmit={inviteMember}>
            <Field label="Email">
              <input
                type="email"
                value={email}
                disabled={inviting}
                required
                onChange={(event) => setEmail(event.target.value)}
              />
            </Field>
            <button className="button" disabled={inviting}>
              {inviting ? "寄送中…" : "寄送啟用連結"}
            </button>
          </form>
          <div className="list">
            {invites.map((invitation) => (
              <div className="list-row" key={invitation.id}>
                <div>
                  <h3>{invitation.email}</h3>
                  <p>等待接受 · 到期日 {invitation.expires_at.slice(0, 10)}</p>
                </div>
              </div>
            ))}
          </div>
          {!loading && invites.length === 0 && (
            <p className="muted">目前沒有等待接受的邀請。</p>
          )}
        </section>

        <section className="panel">
          <h2>當前會員</h2>
          {loading ? (
            <p className="muted">載入會員資料中…</p>
          ) : members.length === 0 ? (
            <Empty title="尚無會員" text="寄送邀請以新增會員。" />
          ) : (
            <div className="list">
              {members.map((user) => {
                const processing = processingID === user.id;
                return (
                  <div className="list-row" key={user.id}>
                    <div>
                      <h3>{user.display_name}</h3>
                      <p>
                        {user.email} ·{" "}
                        {user.status === "active" ? "啟用中" : "已停用"}
                      </p>
                    </div>
                    <div className="toolbar">
                      <button
                        type="button"
                        className="button ghost"
                        disabled={processing}
                        onClick={() => void resetPassword(user)}
                      >
                        重設密碼
                      </button>
                      <button
                        type="button"
                        className="button secondary"
                        disabled={processing}
                        onClick={() => void toggleStatus(user)}
                      >
                        {user.status === "active" ? "停用" : "啟用"}
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>
      </div>
    </>
  );
}
