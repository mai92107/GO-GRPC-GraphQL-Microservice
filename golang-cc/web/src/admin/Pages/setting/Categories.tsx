import { Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useState } from "react";
import {
  Category,
  createCategory,
  deleteCategory,
  getCategories,
  updateCategory,
} from "../../AdminApi";

export default function Categories() {
  const [items, setItems] = useState<Category[]>([]);
  const [form, setForm] = useState({ code: "", name: "" });
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [processingCode, setProcessingCode] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems(await getCategories());
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
      await createCategory(form.code.trim(), form.name.trim());
      setForm({ code: "", name: "" });
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function toggle(category: Category) {
    setProcessingCode(category.code);
    setError("");
    try {
      await updateCategory(
        category.code,
        category.name,
        !category.is_active,
      );
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingCode(null);
    }
  }

  async function remove(category: Category) {
    if (!confirm(`確定刪除「${category.name}」？`)) return;
    setProcessingCode(category.code);
    setError("");
    try {
      await deleteCategory(category.code);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingCode(null);
    }
  }

  return (
    <section className="panel">
      <h2>消費類別</h2>
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
        <button className="button secondary" disabled={creating}>
          {creating ? "新增中…" : "新增類別"}
        </button>
      </form>
      {loading ? (
        <p className="muted">載入消費類別中…</p>
      ) : (
        <div className="list">
          {items.map((category) => {
            const processing = processingCode === category.code;
            return (
              <div className="list-row" key={category.code}>
                <div>
                  <h3>{category.name}</h3>
                  <p>
                    {category.code} ·{" "}
                    {category.is_active ? "啟用" : "停用"}
                  </p>
                </div>
                <div className="toolbar">
                  <button
                    type="button"
                    className="button ghost"
                    disabled={processing}
                    onClick={() => void toggle(category)}
                  >
                    {category.is_active ? "停用" : "啟用"}
                  </button>
                  <button
                    type="button"
                    className="icon-button"
                    aria-label={`刪除 ${category.name}`}
                    disabled={processing}
                    onClick={() => void remove(category)}
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
