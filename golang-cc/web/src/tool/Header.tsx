import { LogOut, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import { Logo } from "../components";
import { User } from "../models";

export default function HeaderBar({
  user,
  onLogout,
}: {
  user: User;
  onLogout: () => void;
}) {
  const [theme, setTheme] = useState<"light" | "dark">(() => {
    const saved = localStorage.getItem("treecard-theme");
    if (saved === "light" || saved === "dark") return saved;
    return matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  });

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    document.documentElement.style.colorScheme = theme;
    localStorage.setItem("treecard-theme", theme);
    document
      .querySelector('meta[name="theme-color"]')
      ?.setAttribute("content", theme === "dark" ? "#071c14" : "#496858");
  }, [theme]);

  return (
    <header className="topbar">
      <Logo />
      <div className="user">
        <span className="user-copy">
          <strong>{user.role === "admin" ? "管理後台" : user.display_name}</strong>
          <small>{user.email}</small>
        </span>
        <span className="avatar">
          {user.role === "admin" ? "管" : user.display_name.slice(0, 1)}
        </span>
        <button
          className="icon-button theme-toggle"
          aria-label={theme === "dark" ? "切換為日間模式" : "切換為夜間模式"}
          aria-pressed={theme === "dark"}
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
        >
          {theme === "dark" ? <Sun size={17} /> : <Moon size={17} />}
        </button>
        <button className="icon-button" aria-label="登出" onClick={onLogout}>
          <LogOut size={17} />
        </button>
      </div>
    </header>
  );
}
