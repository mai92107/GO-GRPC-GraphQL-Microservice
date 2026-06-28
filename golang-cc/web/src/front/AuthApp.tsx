import { useState, FormEvent } from "react";
import {
  acceptInvitation,
  login,
  requestPasswordReset,
  resetPassword,
} from "../auth/AuthApi";
import { Logo, Field } from "../components";
import { User } from "../models";

export default function Auth({ onLogin }: { onLogin: (user: User, csrf: string) => void }) {
  const params = new URLSearchParams(location.search),
    token = params.get("token") || "",
    path = location.pathname;
  const mode = path.includes("accept-invitation")
    ? "accept"
    : path.includes("reset-password")
      ? "reset"
      : "login";
  const [email, setEmail] = useState(""),
    [password, setPassword] = useState(""),
    [name, setName] = useState(""),
    [error, setError] = useState(""),
    [notice, setNotice] = useState("");

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      if (mode === "accept") {
        await acceptInvitation(token, name, password);
        setNotice("帳號已啟用，現在可以登入");
        history.replaceState(null, "", "/");
      } else if (mode === "reset") {
        await resetPassword(token, password);
        setNotice("密碼已更新，現在可以登入");
        history.replaceState(null, "", "/");
      } else {
        const data = await login(email, password);
        onLogin(data.user, data.csrf_token);
      }
    } catch (e) {
      setError((e as Error).message);
    }
  }
  async function reset() {
    setError("");
    await requestPasswordReset(email);
    setNotice("若帳號存在，重設連結已寄至信箱");
  }
  return (
    <main className="auth-page">
      <section className="auth-card">
        <Logo />
        <h1>
          {mode === "accept"
            ? "啟用帳號"
            : mode === "reset"
              ? "設定新密碼"
              : "歡迎回來"}
        </h1>
        <p>
          {mode === "login"
            ? "登入後，幾秒內找到最適合這次消費的卡。"
            : "安全連結僅能使用一次。"}
        </p>
        {error && <div className="error">{error}</div>}
        {notice && <div className="notice">{notice}</div>}
        <form className="stack" onSubmit={submit}>
          {mode === "accept" && (
            <Field label="顯示名稱">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </Field>
          )}
          {mode === "login" && (
            <Field label="Email">
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
              />
            </Field>
          )}
          <Field
            label={mode === "login" ? "密碼" : "新密碼"}
            hint={mode !== "login" ? "至少 12 個字元" : undefined}
          >
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={12}
              autoComplete={
                mode === "login" ? "current-password" : "new-password"
              }
            />
          </Field>
          <button className="button" type="submit">
            {mode === "login"
              ? "登入"
              : mode === "accept"
                ? "啟用帳號"
                : "更新密碼"}
          </button>
          {mode === "login" && (
            <button className="link-button" type="button" onClick={reset}>
              寄送密碼重設連結
            </button>
          )}
        </form>
      </section>
    </main>
  );
}
