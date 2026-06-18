import { useEffect, useState } from "react";
import { setCSRF } from "./api";
import { type User } from "./models";
import { Logo } from "./components";
import { getCurrentUser, logout } from "./auth/AuthApi";

import AdminApp from "./front/AdminApp";
import MemberApp from "./front/MemberApp";
import Auth from "./front/AuthApp";

export default function App() {
  const
    [user, setUser] = useState<User | null>(null),
    [loading, setLoading] = useState(true);

  useEffect(() => {
    getCurrentUser()
      .then((x) => {
        setUser(x.user);
        setCSRF(x.csrf_token);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading)
    return (
      <div className="auth-page">
        <Logo />
      </div>
    );

  if (!user)
    return (
      <Auth
        onLogin={(u, c) => {
          setCSRF(c);
          setUser(u);
        }}
      />
    );

  if (user.role === "admin")
    return (
      <AdminApp
        user={user}
        onLogout={() =>
          logout().finally(() => setUser(null))
        }
      />
    );

  if (user.role === "member")
    return (
      <MemberApp
        user={user}
        onLogout={() =>
          logout().finally(() => setUser(null))
        }
      />
    );
}
