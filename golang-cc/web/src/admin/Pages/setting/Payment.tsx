import { Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useState } from "react";
import type { PaymentMethod } from "../../../models";
import {
  createPaymentMethod,
  deletePaymentMethod,
  getPaymentMethods,
  updatePaymentMethod,
} from "../../AdminApi";

export default function PaymentMethods() {
  const [items, setItems] = useState<PaymentMethod[]>([]);
  const [form, setForm] = useState({ code: "", name: "" });
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [processingCode, setProcessingCode] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems(await getPaymentMethods());
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
      await createPaymentMethod(form.code.trim(), form.name.trim());
      setForm({ code: "", name: "" });
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function toggle(method: PaymentMethod) {
    setProcessingCode(method.code);
    setError("");
    try {
      await updatePaymentMethod(method.code, method.name, !method.is_active);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingCode(null);
    }
  }

  async function remove(method: PaymentMethod) {
    if (!confirm(`確定刪除「${method.name}」？`)) return;
    setProcessingCode(method.code);
    setError("");
    try {
      await deletePaymentMethod(method.code);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingCode(null);
    }
  }

  return (
    <section className="panel">
      <h2>支付方式</h2>
      {error && <div className="error">{error}</div>}
      <form className="stack" onSubmit={create}>
        <input
          className="inputBlock"
          placeholder="穩定代碼"
          value={form.code}
          disabled={creating}
          required
          onChange={(event) => setForm({ ...form, code: event.target.value })}
        />
        <input
          className="inputBlock"
          placeholder="顯示名稱"
          value={form.name}
          disabled={creating}
          required
          onChange={(event) => setForm({ ...form, name: event.target.value })}
        />
        <button className="button secondary" disabled={creating}>
          {creating ? "新增中…" : "新增支付方式"}
        </button>
      </form>
      {loading ? (
        <p className="muted">載入支付方式中…</p>
      ) : (
        <div className="list">
          {items.map((method) => {
            const processing = processingCode === method.code;
            return (
              <div className="list-row" key={method.code}>
                <div>
                  <h3>{method.name}</h3>
                  <p>
                    {method.code} · {method.is_active ? "啟用" : "停用"}
                    {method.is_system ? " · 系統預設" : ""}
                  </p>
                </div>
                <div className="toolbar">
                  <button
                    type="button"
                    className="button ghost"
                    disabled={processing}
                    onClick={() => void toggle(method)}
                  >
                    {method.is_active ? "停用" : "啟用"}
                  </button>
                  <button
                    type="button"
                    className="icon-button"
                    aria-label={`刪除 ${method.name}`}
                    disabled={processing || method.is_system}
                    onClick={() => void remove(method)}
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
