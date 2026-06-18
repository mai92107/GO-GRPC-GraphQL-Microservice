import { User } from "../models";
import { useState } from "react";
import HeaderBar from "../tool/Header";
import Navi from "../tool/Navigator";
import { Recommend } from "../member/page/Recommend";
import Cards from "../member/page/Cards";
import Transactions from "../member/page/Transaction";
import MemberSettings from "../member/page/MemberSettings";

type Tab = "recommend" | "transactions" | "cards" | "settings";

export default function MemberApp({
  user,
  onLogout,
}: {
  user: User;
  onLogout: () => void;
}) {
  const [tab, setTab] = useState<Tab>("recommend");
  const pages = {
    recommend: Recommend,
    transactions: Transactions,
    cards: Cards,
    settings: MemberSettings,
  };
  const Page = pages[tab] || Recommend;

  return (
    <div className="shell">
      <HeaderBar user={user} onLogout={onLogout} />
      <main className="content">
        <Page />
      </main>
      <Navi user={user} tab={tab} setTab={setTab}/>
    </div>
  );
}
