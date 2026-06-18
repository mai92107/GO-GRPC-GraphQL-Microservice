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
import { User } from "../models";

export default function Navi({ user, tab, setTab }: { user: User, tab: string; setTab: any }) {
  const adminMap = [
    { icon: <LayoutDashboard />, label: "總覽", name: "dashboard" },
    { icon: <Users />, label: "會員", name: "members" },
    { icon: <MessageCircle />, label: "Telegram", name: "telegram" },
    { icon: <Building2 />, label: "銀行", name: "banks" },
    { icon: <CreditCard />, label: "卡片", name: "cards" },
    { icon: <Gift />, label: "活動", name: "activities" },
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
    <nav className="bottom-nav" aria-label="管理導覽">
      {map.map((item, key) => (
        <Nav
          key={key}
          icon={item.icon}
          label={item.label}
          active={tab === item.name}
          onClick={() => setTab(item.name)}
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
        <button className={active ? "active" : ""} onClick={onClick}>
          {icon}
          {label}
        </button>
      );
    }