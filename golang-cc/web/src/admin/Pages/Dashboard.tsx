import { useEffect, useState } from "react";
import Head from "../../tool/Head";
import { DashboardSummary, getDashboard } from "../AdminApi";

const labels: Record<string, string> = {
  members: "會員數",
  banks: "銀行數",
  cards: "卡片數",
  active_activities: "有效活動",
};

export default function Dashboard() {
  const [summary, setSummary] = useState<DashboardSummary>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getDashboard()
      .then(setSummary)
      .catch((requestError) => setError((requestError as Error).message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <>
      <Head title="管理總覽" text="全域銀行卡片、活動與會員狀態。" />
      {error && <div className="error">{error}</div>}
      {loading ? (
        <p className="muted">載入管理總覽中…</p>
      ) : (
        <div className="card-list">
          {Object.entries(summary).map(([key, value]) => (
            <div className="panel" key={key} style={{ textAlign: "center" }}>
              <h2>{value}</h2>
              <p className="muted">{labels[key] || key}</p>
            </div>
          ))}
        </div>
      )}
    </>
  );
}
