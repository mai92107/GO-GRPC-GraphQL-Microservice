import { LogOut } from "lucide-react";
import { Logo } from "../components";
import { User } from "../models";

export default function HeaderBar({
  user,
  onLogout,
}: {
  user: User;
  onLogout: () => void;
}) {
  return (
    <header className="topbar">
      <Logo />
      <div className="user">
        {user.role === "admin" ? (
          <span>
            管理後台
            <br />
            {user.email}
          </span>
        ) : (
          <>
            <span>
              {user.display_name}
              <br />
              {user.email}
            </span>
            <span className="avatar">{user.display_name.slice(0, 1)}</span>
          </>
        )}
        <span className="avatar">管</span>
        <button className="icon-button" aria-label="登出" onClick={onLogout}>
          <LogOut size={17} />
        </button>
      </div>
    </header>
  );
}
