import { Building2, Plus, Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Dialog, Empty, SearchField, StatusBadge } from "../../components";
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
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

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
      setShowCreate(false);
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
        bank.id,
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

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return keyword
      ? items.filter((bank) =>
          `${bank.name} ${bank.id}`.toLowerCase().includes(keyword),
        )
      : items;
  }, [items, query]);

  return (
    <>
      <Head
        title="銀行管理"
        text="建立與維護卡片所屬銀行。"
        actions={
          <button className="button" onClick={() => setShowCreate(true)}>
            <Plus size={17} /> 新增銀行
          </button>
        }
      />
      <section className="panel">
        {error && <div className="error">{error}</div>}
        <div className="collection-toolbar">
          <SearchField
            value={query}
            onChange={setQuery}
            placeholder="搜尋銀行名稱或代碼"
          />
          <span className="collection-count">{filtered.length} 家銀行</span>
        </div>

        {loading ? (
          <p className="muted">載入銀行資料中…</p>
        ) : items.length === 0 ? (
          <Empty title="尚未建立銀行" text="先建立銀行後才能新增卡片。" />
        ) : filtered.length === 0 ? (
          <Empty title="找不到銀行" text="請嘗試其他搜尋關鍵字。" />
        ) : (
          <div className="entity-grid">
            {filtered.map((bank) => {
              const processing = processingID === bank.id;
              return (
                <article className="entity-card" key={bank.id}>
                  <div className="entity-icon"><Building2 size={21} /></div>
                  <div className="entity-copy">
                    <h3>{bank.name}</h3>
                    <StatusBadge active={bank.is_active} />
                  </div>
                  <div className="entity-actions">
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
                </article>
              );
            })}
          </div>
        )}
      </section>
      {showCreate && (
        <Dialog title="新增銀行" onClose={() => !creating && setShowCreate(false)}>
          <form className="stack" onSubmit={create}>
            <label className="field">
              <span>銀行名稱</span>
              <input
                autoFocus
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="例如：玉山銀行"
                disabled={creating}
                required
              />
              <small>建立後可在卡片目錄中選擇這家銀行。</small>
            </label>
            <div className="dialog-actions">
              <button type="button" className="button ghost" onClick={() => setShowCreate(false)}>
                取消
              </button>
              <button className="button" disabled={creating || !name.trim()}>
                {creating ? "新增中…" : "新增銀行"}
              </button>
            </div>
          </form>
        </Dialog>
      )}
    </>
  );
}
