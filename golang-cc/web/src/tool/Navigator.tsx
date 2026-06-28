import {
  Building2,
  CreditCard,
  Gift,
  History,
  LayoutDashboard,
  MessageCircle,
  Settings,
  Sparkles,
  Users,
} from "lucide-react";
import type { Dispatch, SetStateAction } from "react";
import { User } from "../models";

export default function Navi<T extends string>({
  user,
  tab,
  setTab,
}: {
  user: User;
  tab: T;
  setTab: Dispatch<SetStateAction<T>>;
}) {
  const adminMap = [
    { icon: <LayoutDashboard />, label: "總覽", name: "dashboard" },
    { icon: <Users />, label: "會員", name: "members" },
    { icon: <MessageCircle />, label: "TG", name: "telegram" },
    { icon: <Building2 />, label: "銀行", name: "banks" },
    { icon: <CreditCard />, label: "卡片", name: "cards" },
    { icon: <Gift />, label: "回饋", name: "activities" },
    { icon: <Settings />, label: "系統", name: "system" },
  ];

  const memberMap = [
    { icon: <Sparkles />, label: "推薦", name: "recommend" },
    { icon: <History />, label: "交易", name: "transactions" },
    { icon: <CreditCard />, label: "卡片", name: "cards" },
    { icon: <Settings />, label: "設定", name: "settings" },
  ]

  const map = user.role === "admin" ? adminMap : memberMap

  return (
    <nav className="bottom-nav" aria-label={user.role === "admin" ? "管理導覽" : "主要導覽"}>
      {map.map((item, key) => (
        <Nav
          key={item.name}
          icon={item.icon}
          label={item.label}
          active={tab === item.name}
          onClick={() => setTab(item.name as T)}
        />
      ))}
    </nav>
  );
}


function Nav(
  {icon, label, active, onClick}: 
  {
    icon: React.ReactNode;
    label: string;
    active: boolean;
    onClick: () => void;
  }) {
      return (
        <button
          className={active ? "active" : ""}
          onClick={onClick}
          aria-current={active ? "page" : undefined}
        >
          <span className="nav-icon">{icon}</span>
          <span>{label}</span>
        </button>
      );
    }
