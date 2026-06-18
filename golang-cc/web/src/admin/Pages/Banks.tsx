import { Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { Empty } from "../../components";
import Head from "../../tool/Head";
import {
  activateBank,
  Bank,
  createBank,
  deleteBank,
  getBanks,
} from "../AdminApi";

export default function Banks() {
  const [items, setItems] = useState<Bank[]>([]);
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [processingID, setProcessingID] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems(await getBanks());
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function create(event: FormEvent) {
    event.preventDefault();
    setCreating(true);
    setError("");
    try {
      await createBank(name.trim());
      setName("");
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function toggle(bank: Bank) {
    setProcessingID(bank.id);
    setError("");
    try {
      await activateBank(
        bank.id,
        bank.name,
        bank.code,
        bank.website_url,
        bank.is_active,
      );
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }

  async function remove(bank: Bank) {
    if (!confirm(`確定刪除「${bank.name}」？`)) return;
    setProcessingID(bank.id);
    setError("");
    try {
      await deleteBank(bank.id);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }

  return (
    <>
      <Head title="銀行管理" text="建立與維護卡片所屬銀行。" />
      <section className="panel">
        {error && <div className="error">{error}</div>}
        <form className="toolbar" onSubmit={create}>
          <input
            className="inputBlock"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="銀行名稱"
            disabled={creating}
            required
          />
          <button className="button" disabled={creating}>
            {creating ? "新增中…" : "新增銀行"}
          </button>
        </form>

        {loading ? (
          <p className="muted">載入銀行資料中…</p>
        ) : items.length === 0 ? (
          <Empty title="尚未建立銀行" text="先建立銀行後才能新增卡片。" />
        ) : (
          <div className="list" style={{ marginTop: "1em" }}>
            {items.map((bank) => {
              const processing = processingID === bank.id;
              return (
                <div className="list-row" key={bank.id}>
                  <div>
                    <h3>{bank.name}</h3>
                    <p>
                      {bank.code || "無代碼"} ·{" "}
                      {bank.is_active ? "啟用" : "停用"}
                    </p>
                  </div>
                  <div className="toolbar">
                    <button
                      type="button"
                      className="button ghost"
                      disabled={processing}
                      onClick={() => void toggle(bank)}
                    >
                      {bank.is_active ? "停用" : "啟用"}
                    </button>
                    <button
                      type="button"
                      className="icon-button"
                      aria-label={`刪除 ${bank.name}`}
                      disabled={processing}
                      onClick={() => void remove(bank)}
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>
    </>
  );
}
