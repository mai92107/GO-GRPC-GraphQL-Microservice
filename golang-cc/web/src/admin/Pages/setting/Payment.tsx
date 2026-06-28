import { Plus, Smartphone } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Dialog, Empty, Field, SearchField, StatusBadge } from "../../../components";
import type { PaymentMethod } from "../../../models";
import {
  createPaymentMethod,
  deletePaymentMethod,
  getPaymentMethods,
  updatePaymentMethod,
} from "../../AdminApi";

const emptyForm = { name: "" };

export default function PaymentMethods() {
  const [items, setItems] = useState<PaymentMethod[]>([]);
  const [form, setForm] = useState(emptyForm);
  const [editing, setEditing] = useState<PaymentMethod | null>(null);
  const [editName, setEditName] = useState("");
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    try { setItems(await getPaymentMethods()); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => void load(), [load]);

  async function create(event: FormEvent) {
    event.preventDefault();
    setProcessing(true); setError("");
    try {
      await createPaymentMethod(form.name.trim());
      setForm(emptyForm); setShowCreate(false); await load();
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  function openEditor(method: PaymentMethod) {
    setError(""); setEditing(method); setEditName(method.name);
  }

  async function persist(active: boolean) {
    if (!editing) return;
    setProcessing(true); setError("");
    try {
      await updatePaymentMethod(editing.id, editName.trim(), active);
      setEditing(null); await load();
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (editing) await persist(Boolean(editing.is_active));
  }

  async function remove() {
    if (!editing || editing.is_system || !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)) return;
    setProcessing(true); setError("");
    try { await deletePaymentMethod(editing.id); setEditing(null); await load(); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return items.filter((item) => !keyword || `${item.name} ${item.id}`.toLowerCase().includes(keyword));
  }, [items, query]);

  return (
    <>
      <section className="panel setting-collection">
        <div className="section-title">
          <div><h2>支付方式</h2><p className="muted">管理實體卡、行動支付與電子票證選項。</p></div>
          <button className="button" onClick={() => setShowCreate(true)}><Plus size={16} /> 新增支付方式</button>
        </div>
        {error && <div className="error preference-message">{error}</div>}
        <div className="collection-toolbar">
          <SearchField value={query} onChange={setQuery} placeholder="搜尋支付方式" />
          <span className="collection-count">{filtered.length} 種支付方式</span>
        </div>
        {loading ? <p className="muted">載入支付方式中…</p> : !filtered.length ? (
          <Empty title="找不到支付方式" text={items.length ? "請調整搜尋條件。" : "建立第一種支付方式。"} />
        ) : (
          <div className="entity-grid">
            {filtered.map((method) => (
              <article className="entity-card" key={method.id}>
                <div className="entity-icon"><Smartphone size={20} /></div>
                <div className="entity-copy">
                  <span className="entity-kicker">{method.is_system ? "系統預設" : "自訂支付方式"}</span>
                  <h3>{method.name}</h3>
                  <StatusBadge active={Boolean(method.is_active)} />
                </div>
                <button className="button ghost" onClick={() => openEditor(method)}>修改</button>
              </article>
            ))}
          </div>
        )}
      </section>

      {showCreate && (
        <Dialog title="新增支付方式" onClose={() => !processing && setShowCreate(false)}>
          <form className="stack" onSubmit={create}>
            <Field label="顯示名稱">
              <input autoFocus required value={form.name} disabled={processing} placeholder="例如：LINE Pay" onChange={(e) => setForm({ ...form, name: e.target.value })} />
            </Field>
            <div className="dialog-actions">
              <button type="button" className="button ghost" onClick={() => setShowCreate(false)}>取消</button>
              <button className="button" disabled={processing || !form.name.trim()}>{processing ? "新增中…" : "新增支付方式"}</button>
            </div>
          </form>
        </Dialog>
      )}

      {editing && (
        <Dialog title="修改支付方式" onClose={() => !processing && setEditing(null)}>
          <form className="stack" onSubmit={save}>
            <Field label="顯示名稱"><input autoFocus required value={editName} disabled={processing} onChange={(e) => setEditName(e.target.value)} /></Field>
            {editing.is_system && <div className="notice">這是系統預設支付方式，可以調整名稱與狀態，但不能刪除。</div>}
            <p className="muted">目前狀態：{editing.is_active ? "啟用" : "停用"}</p>
            <div className="dialog-actions split-actions">
              <button type="button" className="button danger" disabled={processing || editing.is_system} onClick={() => void remove()}>刪除支付方式</button>
              <span />
              <button className="button" disabled={processing || !editName.trim()}>{processing ? "處理中…" : "儲存"}</button>
              <button type="button" className="button secondary" disabled={processing} onClick={() => void persist(!editing.is_active)}>{editing.is_active ? "停用" : "啟用"}</button>
            </div>
          </form>
        </Dialog>
      )}
    </>
  );
}
