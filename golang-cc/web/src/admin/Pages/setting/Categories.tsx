import { Shapes, Plus } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Dialog, Empty, Field, SearchField, StatusBadge } from "../../../components";
import {
  Category,
  createCategory,
  deleteCategory,
  getCategories,
  updateCategory,
} from "../../AdminApi";

const emptyForm = { name: "" };

export default function Categories() {
  const [items, setItems] = useState<Category[]>([]);
  const [form, setForm] = useState(emptyForm);
  const [editing, setEditing] = useState<Category | null>(null);
  const [editName, setEditName] = useState("");
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    try {
      setItems(await getCategories());
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => void load(), [load]);

  async function create(event: FormEvent) {
    event.preventDefault();
    setProcessing(true);
    setError("");
    try {
      await createCategory(form.name.trim());
      setForm(emptyForm);
      setShowCreate(false);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessing(false);
    }
  }

  function openEditor(category: Category) {
    setError("");
    setEditing(category);
    setEditName(category.name);
  }

  async function persist(active: boolean) {
    if (!editing) return;
    setProcessing(true);
    setError("");
    try {
      await updateCategory(editing.id, editName.trim(), active);
      setEditing(null);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessing(false);
    }
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (editing) await persist(editing.is_active);
  }

  async function remove() {
    if (!editing || !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)) return;
    setProcessing(true);
    setError("");
    try {
      await deleteCategory(editing.id);
      setEditing(null);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessing(false);
    }
  }

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return items.filter((item) => !keyword || `${item.name} ${item.id}`.toLowerCase().includes(keyword));
  }, [items, query]);

  return (
    <>
      <section className="panel setting-collection">
        <div className="section-title">
          <div><h2>消費類別</h2><p className="muted">管理活動與交易使用的消費分類。</p></div>
          <button className="button" onClick={() => setShowCreate(true)}><Plus size={16} /> 新增類別</button>
        </div>
        {error && <div className="error preference-message">{error}</div>}
        <div className="collection-toolbar">
          <SearchField value={query} onChange={setQuery} placeholder="搜尋類別名稱" />
          <span className="collection-count">{filtered.length} 個類別</span>
        </div>
        {loading ? <p className="muted">載入消費類別中…</p> : !filtered.length ? (
          <Empty title="找不到消費類別" text={items.length ? "請調整搜尋條件。" : "建立第一個消費類別。"} />
        ) : (
          <div className="entity-grid">
            {filtered.map((category) => (
              <article className="entity-card" key={category.id}>
                <div className="entity-icon"><Shapes size={20} /></div>
                <div className="entity-copy">
                  <span className="entity-kicker">消費分類</span>
                  <h3>{category.name}</h3>
                  <StatusBadge active={category.is_active} />
                </div>
                <button className="button ghost" onClick={() => openEditor(category)}>修改</button>
              </article>
            ))}
          </div>
        )}
      </section>

      {showCreate && (
        <Dialog title="新增消費類別" onClose={() => !processing && setShowCreate(false)}>
          <form className="stack" onSubmit={create}>
            <Field label="顯示名稱">
              <input autoFocus required value={form.name} disabled={processing} placeholder="例如：影音串流" onChange={(e) => setForm({ ...form, name: e.target.value })} />
            </Field>
            <div className="dialog-actions">
              <button type="button" className="button ghost" onClick={() => setShowCreate(false)}>取消</button>
              <button className="button" disabled={processing || !form.name.trim()}>{processing ? "新增中…" : "新增類別"}</button>
            </div>
          </form>
        </Dialog>
      )}

      {editing && (
        <Dialog title="修改消費類別" onClose={() => !processing && setEditing(null)}>
          <form className="stack" onSubmit={save}>
            <Field label="顯示名稱">
              <input autoFocus required value={editName} disabled={processing} onChange={(e) => setEditName(e.target.value)} />
            </Field>
            <p className="muted">目前狀態：{editing.is_active ? "啟用" : "停用"}</p>
            <div className="dialog-actions split-actions">
              <button type="button" className="button danger" disabled={processing} onClick={() => void remove()}>刪除類別</button>
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
