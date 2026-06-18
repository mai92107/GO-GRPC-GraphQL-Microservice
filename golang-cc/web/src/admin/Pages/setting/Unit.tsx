import { Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { Field } from "../../../components";
import type { Unit } from "../../../models";
import {
  createRewardUnit,
  deleteRewardUnit,
  getRewardUnits,
  updateRewardUnit,
} from "../../AdminApi";

const emptyForm = {
  code: "",
  name: "",
  symbol: "",
  symbol_position: "suffix" as const,
  twd_rate: "1",
  precision: 2,
};

export default function Units() {
  const [items, setItems] = useState<Unit[]>([]);
  const [form, setForm] = useState<{
    code: string;
    name: string;
    symbol: string;
    symbol_position: "prefix" | "suffix";
    twd_rate: string;
    precision: number;
  }>(emptyForm);
  const [editingID, setEditingID] = useState<string | null>(null);
  const [editingRate, setEditingRate] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [processingID, setProcessingID] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems(await getRewardUnits());
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
      await createRewardUnit(form);
      setForm(emptyForm);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  function startEdit(unit: Unit) {
    setEditingID(unit.id);
    setEditingRate(unit.twd_rate);
  }

  async function saveRate(unit: Unit) {
    setProcessingID(unit.id);
    setError("");
    try {
      await updateRewardUnit(unit.id, {
        code: unit.code,
        name: unit.name,
        symbol: unit.symbol,
        symbol_position: unit.symbol_position,
        precision: unit.precision,
        twd_rate: editingRate,
      });
      setEditingID(null);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }

  async function remove(unit: Unit) {
    if (!confirm(`確定刪除「${unit.name}」？`)) return;
    setProcessingID(unit.id);
    setError("");
    try {
      await deleteRewardUnit(unit.id);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }

  return (
    <section className="panel">
      <h2>回饋單位</h2>
      {error && <div className="error">{error}</div>}
      <form className="stack" onSubmit={create}>
        <input
          className="inputBlock"
          placeholder="代碼"
          value={form.code}
          disabled={creating}
          required
          onChange={(event) => setForm({ ...form, code: event.target.value })}
        />
        <input
          className="inputBlock"
          placeholder="名稱"
          value={form.name}
          disabled={creating}
          required
          onChange={(event) => setForm({ ...form, name: event.target.value })}
        />
        <input
          className="inputBlock"
          placeholder="符號"
          value={form.symbol}
          disabled={creating}
          required
          onChange={(event) =>
            setForm({ ...form, symbol: event.target.value })
          }
        />
        <div className="two-col">
          <Field label="符號位置">
            <select
              value={form.symbol_position}
              disabled={creating}
              onChange={(event) =>
                setForm({
                  ...form,
                  symbol_position: event.target.value as "prefix" | "suffix",
                })
              }
            >
              <option value="prefix">置前</option>
              <option value="suffix">置後</option>
            </select>
          </Field>
          <Field label="每單位約當台幣">
            <input
              type="number"
              min="0.000001"
              step="0.000001"
              value={form.twd_rate}
              disabled={creating}
              required
              onChange={(event) =>
                setForm({ ...form, twd_rate: event.target.value })
              }
            />
          </Field>
        </div>
        <Field label="小數精度">
          <input
            type="number"
            min="0"
            max="6"
            value={form.precision}
            disabled={creating}
            onChange={(event) =>
              setForm({ ...form, precision: Number(event.target.value) })
            }
          />
        </Field>
        <button className="button secondary" disabled={creating}>
          {creating ? "新增中…" : "新增單位"}
        </button>
      </form>

      {loading ? (
        <p className="muted">載入回饋單位中…</p>
      ) : (
        <div className="list">
          {items.map((unit) => {
            const processing = processingID === unit.id;
            const editing = editingID === unit.id;
            return (
              <div className="list-row" key={unit.id}>
                <div>
                  <h3>{unit.name}</h3>
                  <p>
                    {unit.code} ·{" "}
                    {unit.symbol_position === "prefix"
                      ? `${unit.symbol}數值`
                      : `數值${unit.symbol}`}{" "}
                    · 1 單位 = NT${unit.twd_rate}
                  </p>
                  {editing && (
                    <input
                      type="number"
                      min="0.000001"
                      step="0.000001"
                      aria-label={`${unit.name} 每單位約當台幣`}
                      value={editingRate}
                      disabled={processing}
                      onChange={(event) => setEditingRate(event.target.value)}
                    />
                  )}
                </div>
                <div className="toolbar">
                  {editing ? (
                    <>
                      <button
                        type="button"
                        className="button"
                        disabled={processing || !editingRate}
                        onClick={() => void saveRate(unit)}
                      >
                        儲存換算
                      </button>
                      <button
                        type="button"
                        className="button ghost"
                        disabled={processing}
                        onClick={() => setEditingID(null)}
                      >
                        取消
                      </button>
                    </>
                  ) : (
                    <button
                      type="button"
                      className="button ghost"
                      disabled={processing}
                      onClick={() => startEdit(unit)}
                    >
                      修改換算
                    </button>
                  )}
                  <button
                    type="button"
                    className="icon-button"
                    aria-label={`刪除 ${unit.name}`}
                    disabled={processing}
                    onClick={() => void remove(unit)}
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
  );
}
