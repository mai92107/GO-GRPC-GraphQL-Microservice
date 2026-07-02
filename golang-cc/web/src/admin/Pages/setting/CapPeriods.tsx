import { CalendarClock, Plus } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";
import { Dialog, Empty, Field, SearchField } from "../../../components";
import {
  defaultCapPeriodOptions,
  loadCapPeriodOptions,
  saveCapPeriodOptions,
  type MockOption,
} from "../activities/activityMockSettings";

const emptyForm = { code: "", name: "" };

export default function CapPeriods() {
  const [items, setItems] = useState<MockOption[]>(loadCapPeriodOptions);
  const [form, setForm] = useState(emptyForm);
  const [editing, setEditing] = useState<MockOption | null>(null);
  const [editForm, setEditForm] = useState(emptyForm);
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const persist = (nextItems: MockOption[]) => {
    setItems(nextItems);
    saveCapPeriodOptions(nextItems);
  };

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return items.filter((item) => !keyword || `${item.name} ${item.code}`.toLowerCase().includes(keyword));
  }, [items, query]);

  function create(event: FormEvent) {
    event.preventDefault();
    const code = form.code.trim().toUpperCase().replace(/\s+/g, "_");
    if (!code || !form.name.trim() || items.some((item) => item.code === code)) return;
    persist([...items, { code, name: form.name.trim() }]);
    setForm(emptyForm);
    setShowCreate(false);
  }

  function openEditor(item: MockOption) {
    setEditing(item);
    setEditForm({ code: item.code, name: item.name });
  }

  function save(event: FormEvent) {
    event.preventDefault();
    if (!editing || !editForm.name.trim()) return;
    persist(items.map((item) => item.code === editing.code ? { ...item, name: editForm.name.trim() } : item));
    setEditing(null);
  }

  function remove() {
    if (!editing) return;
    persist(items.filter((item) => item.code !== editing.code));
    setEditing(null);
  }

  function restoreDefaults() {
    persist(defaultCapPeriodOptions);
    setEditing(null);
    setShowCreate(false);
  }

  return (
    <>
      <section className="panel setting-collection">
        <div className="section-title">
          <div><h2>上限週期</h2><p className="muted">管理 Activity Benefit 可選擇的回饋上限週期。</p></div>
          <div className="toolbar">
            <button className="button ghost" type="button" onClick={restoreDefaults}>還原預設</button>
            <button className="button" type="button" onClick={() => setShowCreate(true)}><Plus size={16} /> 新增週期</button>
          </div>
        </div>
        <div className="collection-toolbar">
          <SearchField value={query} onChange={setQuery} placeholder="搜尋週期名稱或代碼" />
          <span className="collection-count">{filtered.length} 個週期</span>
        </div>
        {!filtered.length ? (
          <Empty title="找不到上限週期" text="新增自訂週期後會出現在 Activity Benefit 下拉選單。" />
        ) : (
          <div className="entity-grid">
            {filtered.map((item) => (
              <article className="entity-card" key={item.code}>
                <div className="entity-icon"><CalendarClock size={20} /></div>
                <div className="entity-copy">
                  <span className="entity-kicker">{item.code}</span>
                  <h3>{item.name}</h3>
                </div>
                <button className="button ghost" type="button" onClick={() => openEditor(item)}>修改</button>
              </article>
            ))}
          </div>
        )}
      </section>

      {showCreate && (
        <Dialog title="新增上限週期" onClose={() => setShowCreate(false)}>
          <form className="stack" onSubmit={create}>
            <Field label="代碼" hint="會保存為大寫代碼，例如 SEMI_ANNUAL。">
              <input autoFocus required value={form.code} placeholder="SEMI_ANNUAL" onChange={(event) => setForm({ ...form, code: event.target.value })} />
            </Field>
            <Field label="顯示名稱">
              <input required value={form.name} placeholder="例如：每半年" onChange={(event) => setForm({ ...form, name: event.target.value })} />
            </Field>
            <div className="dialog-actions">
              <button type="button" className="button ghost" onClick={() => setShowCreate(false)}>取消</button>
              <button className="button" disabled={!form.code.trim() || !form.name.trim()}>新增週期</button>
            </div>
          </form>
        </Dialog>
      )}

      {editing && (
        <Dialog title="修改上限週期" onClose={() => setEditing(null)}>
          <form className="stack" onSubmit={save}>
            <Field label="代碼">
              <input value={editForm.code} disabled />
            </Field>
            <Field label="顯示名稱">
              <input autoFocus required value={editForm.name} onChange={(event) => setEditForm({ ...editForm, name: event.target.value })} />
            </Field>
            <div className="dialog-actions split-actions">
              <button type="button" className="button danger" onClick={remove}>刪除週期</button>
              <span />
              <button className="button" disabled={!editForm.name.trim()}>儲存</button>
            </div>
          </form>
        </Dialog>
      )}
    </>
  );
}
