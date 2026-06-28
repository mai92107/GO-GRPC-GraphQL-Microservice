import { Coins, Plus } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Dialog, Empty, Field, SearchField } from "../../../components";
import type { Unit } from "../../../models";
import {
  createRewardUnit,
  deleteRewardUnit,
  getRewardUnits,
  updateRewardUnit,
} from "../../AdminApi";

type UnitForm = {
  name: string;
  symbol: string;
  symbol_position: "prefix" | "suffix";
  twd_rate: string;
  precision: number;
};

const emptyForm: UnitForm = {
  name: "",
  symbol: "",
  symbol_position: "suffix",
  twd_rate: "1",
  precision: 2,
};

const toForm = (unit: Unit): UnitForm => ({
  name: unit.name,
  symbol: unit.symbol,
  symbol_position: unit.symbol_position,
  twd_rate: unit.twd_rate,
  precision: unit.precision,
});

export default function Units() {
  const [items, setItems] = useState<Unit[]>([]);
  const [form, setForm] = useState<UnitForm>(emptyForm);
  const [editing, setEditing] = useState<Unit | null>(null);
  const [editForm, setEditForm] = useState<UnitForm>(emptyForm);
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    try { setItems(await getRewardUnits()); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => void load(), [load]);

  async function create(event: FormEvent) {
    event.preventDefault();
    setProcessing(true); setError("");
    try {
      await createRewardUnit(form);
      setForm(emptyForm); setShowCreate(false); await load();
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  function openEditor(unit: Unit) {
    setError(""); setEditing(unit); setEditForm(toForm(unit));
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (!editing) return;
    setProcessing(true); setError("");
    try {
      await updateRewardUnit(editing.id, editForm);
      setEditing(null); await load();
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  async function remove() {
    if (!editing || !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)) return;
    setProcessing(true); setError("");
    try { await deleteRewardUnit(editing.id); setEditing(null); await load(); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessing(false); }
  }

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return items.filter((item) => !keyword || `${item.name} ${item.symbol}`.toLowerCase().includes(keyword));
  }, [items, query]);

  const unitFields = (value: UnitForm, update: (next: UnitForm) => void) => (
    <>
      <div className="two-col">
        <Field label="顯示名稱">
          <input required value={value.name} disabled={processing} placeholder="例如：航空里程" onChange={(e) => update({ ...value, name: e.target.value })} />
        </Field>
        <Field label="顯示符號">
          <input required value={value.symbol} disabled={processing} placeholder="例如：哩" onChange={(e) => update({ ...value, symbol: e.target.value })} />
        </Field>
      </div>
      <div className="two-col">
        <Field label="符號位置">
          <select value={value.symbol_position} disabled={processing} onChange={(e) => update({ ...value, symbol_position: e.target.value as UnitForm["symbol_position"] })}>
            <option value="prefix">置前，例如 $100</option>
            <option value="suffix">置後，例如 100 點</option>
          </select>
        </Field>
        <Field label="每單位約當台幣">
          <input type="number" min="0.000001" step="0.000001" required value={value.twd_rate} disabled={processing} onChange={(e) => update({ ...value, twd_rate: e.target.value })} />
        </Field>
      </div>
      <Field label="小數精度" hint="回饋結果最多顯示的小數位數。">
        <input type="number" min="0" max="6" value={value.precision} disabled={processing} onChange={(e) => update({ ...value, precision: Number(e.target.value) })} />
      </Field>
    </>
  );

  return (
    <>
      <section className="panel setting-collection">
        <div className="section-title">
          <div><h2>回饋單位</h2><p className="muted">管理現金、點數與里程的顯示及價值換算。</p></div>
          <button className="button" onClick={() => setShowCreate(true)}><Plus size={16} /> 新增單位</button>
        </div>
        {error && <div className="error preference-message">{error}</div>}
        <div className="collection-toolbar">
          <SearchField value={query} onChange={setQuery} placeholder="搜尋單位名稱或符號" />
          <span className="collection-count">{filtered.length} 種回饋單位</span>
        </div>
        {loading ? <p className="muted">載入回饋單位中…</p> : !filtered.length ? (
          <Empty title="找不到回饋單位" text={items.length ? "請調整搜尋條件。" : "建立第一種回饋單位。"} />
        ) : (
          <div className="entity-grid">
            {filtered.map((unit) => (
              <article className="entity-card" key={unit.id}>
                <div className="entity-icon"><Coins size={20} /></div>
                <div className="entity-copy">
                  <span className="entity-kicker">1 單位 ≈ NT${unit.twd_rate}</span>
                  <h3>{unit.name}</h3>
                  <p>{unit.symbol_position === "prefix" ? `${unit.symbol}100` : `100${unit.symbol}`} · 小數 {unit.precision} 位</p>
                </div>
                <button className="button ghost" onClick={() => openEditor(unit)}>修改</button>
              </article>
            ))}
          </div>
        )}
      </section>

      {showCreate && (
        <Dialog title="新增回饋單位" onClose={() => !processing && setShowCreate(false)}>
          <form className="stack" onSubmit={create}>
            {unitFields(form, setForm)}
            <div className="dialog-actions">
              <button type="button" className="button ghost" onClick={() => setShowCreate(false)}>取消</button>
              <button className="button" disabled={processing || !form.name.trim() || !form.symbol.trim()}>{processing ? "新增中…" : "新增單位"}</button>
            </div>
          </form>
        </Dialog>
      )}

      {editing && (
        <Dialog title="修改回饋單位" onClose={() => !processing && setEditing(null)}>
          <form className="stack" onSubmit={save}>
            {unitFields(editForm, setEditForm)}
            <div className="dialog-actions split-actions">
              <button type="button" className="button danger" disabled={processing} onClick={() => void remove()}>刪除單位</button>
              <span />
              <button className="button" disabled={processing || !editForm.name.trim() || !editForm.symbol.trim()}>{processing ? "處理中…" : "儲存"}</button>
            </div>
          </form>
        </Dialog>
      )}
    </>
  );
}
