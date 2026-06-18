import { Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { Field } from "../../../components";
import type { Merchant } from "../../../models";
import {
  Category,
  createMerchant,
  deleteMerchant,
  getCategories,
  getMerchants,
  updateMerchant,
} from "../../AdminApi";

type MerchantForm = {
  code: string;
  name: string;
  aliases: string;
  category_codes: string[];
};

const emptyForm: MerchantForm = {
  code: "",
  name: "",
  aliases: "",
  category_codes: [],
};

const splitAliases = (value: string) =>
  value
    .split(",")
    .map((alias) => alias.trim())
    .filter(Boolean);

export default function Merchants() {
  const [items, setItems] = useState<Merchant[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<MerchantForm>(emptyForm);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [processingCode, setProcessingCode] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      const [nextMerchants, nextCategories] = await Promise.all([
        getMerchants(),
        getCategories(),
      ]);
      setItems(nextMerchants);
      setCategories(
        nextCategories.filter(
          (category) => category.is_active && category.code !== "general",
        ),
      );
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
      await createMerchant({
        code: form.code.trim(),
        name: form.name.trim(),
        aliases: splitAliases(form.aliases),
        category_codes: form.category_codes,
      });
      setForm(emptyForm);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function toggle(merchant: Merchant) {
    setProcessingCode(merchant.code);
    setError("");
    try {
      await updateMerchant(merchant, !merchant.is_active);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingCode(null);
    }
  }

  async function remove(merchant: Merchant) {
    if (!confirm(`確定刪除「${merchant.name}」？`)) return;
    setProcessingCode(merchant.code);
    setError("");
    try {
      await deleteMerchant(merchant.code);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingCode(null);
    }
  }

  return (
    <section className="panel">
      <h2>店家</h2>
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
        <input
          className="inputBlock"
          placeholder="別名，以逗號分隔"
          value={form.aliases}
          disabled={creating}
          onChange={(event) =>
            setForm({ ...form, aliases: event.target.value })
          }
        />
        <Field label="消費類別（至少一項）">
          <div className="toolbar">
            {categories.map((category) => (
              <label key={category.code}>
                <input
                  type="checkbox"
                  checked={form.category_codes.includes(category.code)}
                  disabled={creating}
                  onChange={(event) =>
                    setForm({
                      ...form,
                      category_codes: event.target.checked
                        ? [...form.category_codes, category.code]
                        : form.category_codes.filter(
                            (code) => code !== category.code,
                          ),
                    })
                  }
                />{" "}
                {category.name}
              </label>
            ))}
          </div>
        </Field>
        <button
          className="button secondary"
          disabled={creating || !form.category_codes.length}
        >
          {creating ? "新增中…" : "新增店家"}
        </button>
      </form>

      {loading ? (
        <p className="muted">載入店家中…</p>
      ) : (
        <div className="list">
          {items.map((merchant) => {
            const processing = processingCode === merchant.code;
            return (
              <div className="list-row" key={merchant.code}>
                <div>
                  <h3>{merchant.name}</h3>
                  <p>
                    {merchant.code} · 類別：
                    {(merchant.category_codes || []).join("、")} ·{" "}
                    {(merchant.aliases || []).join("、") || "無別名"} ·{" "}
                    {merchant.is_active ? "啟用" : "停用"}
                  </p>
                </div>
                <div className="toolbar">
                  <button
                    type="button"
                    className="button ghost"
                    disabled={processing}
                    onClick={() => void toggle(merchant)}
                  >
                    {merchant.is_active ? "停用" : "啟用"}
                  </button>
                  <button
                    type="button"
                    className="icon-button"
                    aria-label={`刪除 ${merchant.name}`}
                    disabled={processing || merchant.is_system}
                    onClick={() => void remove(merchant)}
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
