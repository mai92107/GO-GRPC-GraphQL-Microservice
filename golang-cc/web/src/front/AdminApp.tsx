import { useState } from "react";

import { type User } from "../models"
import Dashboard from "../admin/Pages/Dashboard";
import Members from "../admin/Pages/Member";
import TelegramBindings from "../admin/Pages/Telegram";
import Cards from "../admin/Pages/Cards";
import Banks from "../admin/Pages/Banks";
import Activities from "../admin/Pages/Activities";
import ADSettings from "../admin/Pages/Settings";
import HeaderBar from "../tool/Header";
import Navi from "../tool/Navigator";

type Tab =
  | "dashboard"
  | "members"
  | "telegram"
  | "banks"
  | "cards"
  | "activities"
  | "system";

export default function AdminApp({
  user,
  onLogout,
}: {
  user: User;
  onLogout: () => void;
}) {
  const [tab, setTab] = useState<Tab>("dashboard");
  const pages = {
    dashboard: Dashboard,
    members: Members,
    telegram: TelegramBindings,
    banks: Banks,
    cards: Cards,
    activities: Activities,
    system: ADSettings,
  };
  const Page = pages[tab] || Dashboard;

  return (
    <div className="shell">
      <HeaderBar user={user} onLogout={onLogout}/>
      <main className="content">
        <Page />
      </main>
      <Navi user={user} tab={tab} setTab={setTab}/>
    </div>
  );
}
