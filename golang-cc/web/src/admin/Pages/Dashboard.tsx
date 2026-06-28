import { useEffect, useState } from "react";
import {
  Activity,
  Building2,
  CreditCard,
  Users,
} from "lucide-react";
import Head from "../../tool/Head";
import { DashboardSummary, getDashboard } from "../AdminApi";

const metrics: Record<
  string,
  { label: string; hint: string; icon: typeof Users; tone: string }
> = {
  members: {
    label: "會員",
    hint: "已建立的會員帳號",
    icon: Users,
    tone: "sage",
  },
  banks: {
    label: "合作銀行",
    hint: "卡片目錄中的銀行",
    icon: Building2,
    tone: "blue",
  },
  cards: {
    label: "信用卡",
    hint: "可供會員選擇的卡片",
    icon: CreditCard,
    tone: "amber",
  },
  active_activities: {
    label: "啟用方案",
    hint: "目前有效的回饋方案",
    icon: Activity,
    tone: "green",
  },
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
      <Head title="管理總覽" text="全域銀行卡片、回饋方案與會員狀態。" />
      {error && <div className="error">{error}</div>}
      {loading ? (
        <div className="skeleton-grid" aria-label="載入管理總覽中">
          {[0, 1, 2, 3].map((item) => <div className="skeleton-card" key={item} />)}
        </div>
      ) : (
        <div className="metric-grid">
          {Object.entries(summary).map(([key, value]) => (
            <Metric key={key} metricKey={key} value={value} />
          ))}
        </div>
      )}
      <section className="panel admin-welcome">
        <div>
          <span className="page-eyebrow">WORKSPACE</span>
          <h2>今天要維護什麼？</h2>
          <p className="muted">
            從左側導覽進入銀行、卡片與回饋方案。建議依序維護基礎資料，再建立卡片回饋方案。
          </p>
        </div>
        <div className="workflow">
          <span><b>1</b> 銀行</span>
          <i aria-hidden="true">→</i>
          <span><b>2</b> 卡片</span>
          <i aria-hidden="true">→</i>
          <span><b>3</b> 回饋方案</span>
        </div>
      </section>
    </>
  );
}

function Metric({
  metricKey,
  value,
}: {
  metricKey: string;
  value: number;
}) {
  const item = metrics[metricKey] || {
    label: metricKey,
    hint: "管理資料",
    icon: Activity,
    tone: "sage",
  };
  const Icon = item.icon;
  return (
    <article className={`metric-card tone-${item.tone}`}>
      <div className="metric-icon"><Icon size={20} /></div>
      <div>
        <strong>{value.toLocaleString()}</strong>
        <h2>{item.label}</h2>
        <p>{item.hint}</p>
      </div>
    </article>
  );
}
