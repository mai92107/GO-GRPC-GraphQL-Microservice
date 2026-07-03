import { MapPin, Plus, UserCheck } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Dialog, Empty, Field, SearchField, StatusBadge } from "../../../components";
import type { LookupItem } from "../../AdminApi";

type LookupCollectionProps = {
  createItem: (id: string, name: string) => Promise<{ id: string }>;
  deleteItem: (id: string) => Promise<{ deleted: boolean }>;
  emptyText: string;
  icon: "region" | "qualification";
  itemLabel: string;
  loadItems: () => Promise<LookupItem[]>;
  searchPlaceholder: string;
  title: string;
  updateItem: (id: string, name: string, isActive: boolean) => Promise<{ updated: boolean }>;
};

const emptyForm = { id: "", name: "" };

export default function LookupCollection({
  createItem,
  deleteItem,
  emptyText,
  icon,
  itemLabel,
  loadItems,
  searchPlaceholder,
  title,
  updateItem,
}: LookupCollectionProps) {
  const [items, setItems] = useState<LookupItem[]>([]);
  const [form, setForm] = useState(emptyForm);
  const [editing, setEditing] = useState<LookupItem | null>(null);
  const [editName, setEditName] = useState("");
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const Icon = icon === "region" ? MapPin : UserCheck;

  const load = useCallback(async () => {
    try { setItems(await loadItems()); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setLoading(false); }
  }, [loadItems]);

  useEffect(() => void load(), [load]);

  async function create(event: FormEvent) {
    event.preventDefault();
    setProcessing(true); setError("");
    try { await createItem(form.id.trim(), form.name.trim()); setForm(emptyForm); setShowCreate(false); await load(); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  function openEditor(item: LookupItem) { setError(""); setEditing(item); setEditName(item.name); }

  async function persist(active: boolean) {
    if (!editing) return;
    setProcessing(true); setError("");
    try { await updateItem(editing.id, editName.trim(), active); setEditing(null); await load(); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  async function save(event: FormEvent) { event.preventDefault(); if (editing) await persist(editing.is_active); }

  async function remove() {
    if (!editing || !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)) return;
    setProcessing(true); setError("");
    try { await deleteItem(editing.id); setEditing(null); await load(); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return items.filter((item) => !keyword || `${item.id} ${item.name}`.toLowerCase().includes(keyword));
  }, [items, query]);

  return (
    <>
      <section className="panel setting-collection">
        <div className="section-title">
          <div><h2>{title}</h2><p className="muted">{emptyText}</p></div>
          <button className="button" onClick={() => setShowCreate(true)}><Plus size={16} /> 新增{itemLabel}</button>
        </div>
        {error && <div className="error preference-message">{error}</div>}
        <div className="collection-toolbar">
          <SearchField value={query} onChange={setQuery} placeholder={searchPlaceholder} />
          <span className="collection-count">{filtered.length} 筆</span>
        </div>
        {loading ? <p className="muted">載入中…</p> : !filtered.length ? (
          <Empty title={`找不到${itemLabel}`} text={items.length ? "請調整搜尋條件。" : `建立第一筆${itemLabel}。`} />
        ) : (
          <div className="entity-grid">
            {filtered.map((item) => (
              <article className="entity-card" key={item.id}>
                <div className="entity-icon"><Icon size={20} /></div>
                <div className="entity-copy">
                  <span className="entity-kicker">{item.id}</span>
                  <h3>{item.name}</h3>
                  <StatusBadge active={item.is_active} />
                </div>
                <button className="button ghost" onClick={() => openEditor(item)}>修改</button>
              </article>
            ))}
          </div>
        )}
      </section>

      {showCreate && (
        <Dialog title={`新增${itemLabel}`} onClose={() => !processing && setShowCreate(false)}>
          <form className="stack" onSubmit={create}>
            <Field label="ID"><input autoFocus required value={form.id} disabled={processing} placeholder="例如：TW" onChange={(e) => setForm({ ...form, id: e.target.value })} /></Field>
            <Field label="顯示名稱"><input required value={form.name} disabled={processing} onChange={(e) => setForm({ ...form, name: e.target.value })} /></Field>
            <div className="dialog-actions">
              <button type="button" className="button ghost" onClick={() => setShowCreate(false)}>取消</button>
              <button className="button" disabled={processing || !form.id.trim() || !form.name.trim()}>{processing ? "新增中…" : `新增${itemLabel}`}</button>
            </div>
          </form>
        </Dialog>
      )}

      {editing && (
        <Dialog title={`修改${itemLabel}`} onClose={() => !processing && setEditing(null)}>
          <form className="stack" onSubmit={save}>
            <Field label="ID"><input value={editing.id} disabled /></Field>
            <Field label="顯示名稱"><input autoFocus required value={editName} disabled={processing} onChange={(e) => setEditName(e.target.value)} /></Field>
            <p className="muted">目前狀態：{editing.is_active ? "啟用" : "停用"}</p>
            <div className="dialog-actions split-actions">
              <button type="button" className="button danger" disabled={processing} onClick={() => void remove()}>刪除{itemLabel}</button>
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
