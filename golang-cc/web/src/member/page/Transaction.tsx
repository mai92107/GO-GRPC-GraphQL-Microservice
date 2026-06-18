import { Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Empty } from "../../components";
import {
  deleteTransaction,
  getTransactions,
  TransactionSummary,
} from "../MemberApi";

export default function Transactions() {
  const [items, setItems] = useState<TransactionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [deletingID, setDeletingID] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setError("");
    try {
      setItems(await getTransactions());
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function remove(id: string) {
    if (!confirm("確定刪除這筆交易？回饋 cap 將會恢復。")) return;

    setDeletingID(id);
    setError("");
    try {
      await deleteTransaction(id);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setDeletingID(null);
    }
  }

  return (
    <>
      <div className="page-head">
        <div>
          <h1>交易紀錄</h1>
          <p>刪除或修改交易後，月度 cap 會重新計算。</p>
        </div>
      </div>

      {error && <div className="error">{error}</div>}

      {!loading && items.length > 0 && (
        <div className="list">
          {items.map((transaction) => (
            <article className="list-row" key={transaction.id}>
              <div>
                <h3>
                  NT$
                  {(transaction.amount_minor / 100).toLocaleString()} ·{" "}
                  {transaction.merchant_name || "其他店家"}
                </h3>
                <p>
                  {transaction.category_code} ·{" "}
                  {transaction.payment_method_name} ·{" "}
                  {transaction.transaction_date}
                  {transaction.note && ` · ${transaction.note}`}
                </p>
              </div>
              <button
                type="button"
                className="icon-button"
                aria-label="刪除交易"
                disabled={deletingID === transaction.id}
                onClick={() => void remove(transaction.id)}
              >
                <Trash2 size={16} />
              </button>
            </article>
          ))}
        </div>
      )}

      {loading && <p className="muted">載入交易紀錄中…</p>}

      {!loading && items.length === 0 && (
        <div className="panel">
          <Empty
            title="還沒有交易"
            text="在推薦結果確認使用卡片後，交易會出現在這裡。"
          />
        </div>
      )}
    </>
  );
}
