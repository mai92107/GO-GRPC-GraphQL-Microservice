import { useEffect, useState } from "react";
import { api } from "../../api";
import Head from "../Head";

export default function Dashboard() {
  const [d, setD] = useState<Record<string, number>>({});

  useEffect(() => {
    api<Record<string, number>>("/admin/dashboard").then(setD);
  }, []);

  return (
    <>
      <Head title="管理總覽" text="全域銀行卡片、活動與會員狀態。" />
      <div className="card-list">
        {Object.entries(d).map(([k, v]) => (
          <div className="panel" key={k} style={{ textAlign: "center" }}>
            <h2>{v}</h2>
            <p className="muted">
              {
                {
                  members: "會員數",
                  banks: "銀行數",
                  cards: "卡片數",
                  active_activities: "有效活動",
                }[k]
              }
            </p>
          </div>
        ))}
      </div>
    </>
  );
}