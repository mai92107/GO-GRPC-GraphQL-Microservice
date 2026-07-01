import { Plus, Store } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  Dialog,
  Empty,
  Field,
  SearchField,
  StatusBadge,
} from "../../../components";
import type { Merchant } from "../../../models";
import {
  Category,
  MerchantInput,
  createMerchant,
  deleteMerchant,
  getCategories,
  getMerchants,
  updateMerchant,
} from "../../AdminApi";

type MerchantForm = {
  name: string;
  aliases: string;
  category_ids: string[];
};

const emptyForm: MerchantForm = { name: "", aliases: "", category_ids: [] };
const splitAliases = (value: string) =>
  value
    .split(",")
    .map((alias) => alias.trim())
    .filter(Boolean);
const toForm = (merchant: Merchant): MerchantForm => ({
  name: merchant.name,
  aliases: (merchant.aliases || []).join(", "),
  category_ids: merchant.category_ids || [],
});

export default function Merchants() {
  const [items, setItems] = useState<Merchant[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<MerchantForm>(emptyForm);
  const [editing, setEditing] = useState<Merchant | null>(null);
  const [editForm, setEditForm] = useState<MerchantForm>(emptyForm);
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    try {
      const [nextMerchants, nextCategories] = await Promise.all([
        getMerchants(),
        getCategories(),
      ]);
      setItems(nextMerchants);
      setCategories(
        nextCategories.filter(
          (category) => category.is_active && category.id !== "general",
        ),
      );
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => void load(), [load]);

  const payload = (value: MerchantForm): MerchantInput => ({
    name: value.name.trim(),
    aliases: splitAliases(value.aliases),
    category_ids: value.category_ids,
  });

  async function create(event: FormEvent) {
    event.preventDefault();
    setProcessing(true);
    setError("");
    try {
      await createMerchant(payload(form));
      setForm(emptyForm);
      setShowCreate(false);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessing(false);
    }
  }

  function openEditor(merchant: Merchant) {
    setError("");
    setEditing(merchant);
    setEditForm(toForm(merchant));
  }

  async function persist(active: boolean) {
    if (!editing) return;
    setProcessing(true);
    setError("");
    try {
      await updateMerchant(
        {
          ...payload(editForm),
          id: editing.id,
          is_active: editing.is_active,
          is_system: editing.is_system,
        },
        active,
      );
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
    if (editing) await persist(Boolean(editing.is_active));
  }

  async function remove() {
    if (
      !editing ||
      editing.is_system ||
      !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)
    )
      return;
    setProcessing(true);
    setError("");
    try {
      await deleteMerchant(editing.id!);
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
    return items.filter(
      (item) =>
        !keyword ||
        `${item.name} ${(item.aliases || []).join(" ")}`
          .toLowerCase()
          .includes(keyword),
    );
  }, [items, query]);

  const merchantFields = (
    value: MerchantForm,
    update: (next: MerchantForm) => void,
  ) => (
    <>
      <Field label="顯示名稱">
        <input
          required
          value={value.name}
          disabled={processing}
          placeholder="例如：好市多"
          onChange={(e) => update({ ...value, name: e.target.value })}
        />
      </Field>
      <Field label="搜尋別名" hint="多個別名請用逗號分隔，協助交易店名辨識。">
        <input
          value={value.aliases}
          disabled={processing}
          placeholder="例如：Costco, 好市多線上購物"
          onChange={(e) => update({ ...value, aliases: e.target.value })}
        />
      </Field>
      <div className="field">
        <span>消費類別</span>
        <div className="choice-grid">
          {categories.map((category, id) => (
            <label className="choice-chip" key={id}>
              <input
                type="checkbox"
                checked={value.category_ids.includes(category.id)}
                disabled={processing}
                onChange={(e) =>
                  update({
                    ...value,
                    category_ids: e.target.checked
                      ? [...value.category_ids, category.id]
                      : value.category_ids.filter((id) => id !== category.id),
                  })
                }
              />
              <span>{category.name}</span>
            </label>
          ))}
        </div>
        <small>至少選擇一項，店家才能正確套用活動條件。</small>
      </div>
    </>
  );

  return (
    <>
      <section className="panel setting-collection">
        <div className="section-title">
          <div>
            <h2>店家</h2>
            <p className="muted">管理交易辨識名稱與適用的消費類別。</p>
          </div>
          <button
            className="button"
            onClick={() => setShowCreate(true)}
            disabled={!categories.length}
          >
            <Plus size={16} /> 新增店家
          </button>
        </div>
        {error && <div className="error preference-message">{error}</div>}
        <div className="collection-toolbar">
          <SearchField
            value={query}
            onChange={setQuery}
            placeholder="搜尋店家或別名"
          />
          <span className="collection-count">{filtered.length} 家店家</span>
        </div>
        {loading ? (
          <p className="muted">載入店家中…</p>
        ) : !filtered.length ? (
          <Empty
            title="找不到店家"
            text={items.length ? "請調整搜尋條件。" : "建立第一家店家。"}
          />
        ) : (
          <div className="entity-grid">
            {filtered.map((merchant) => (
              <article className="entity-card" key={merchant.id}>
                <div className="entity-icon">
                  <Store size={20} />
                </div>
                <div className="entity-copy">
                  <span className="entity-kicker">
                    {merchant.is_system ? "系統店家" : "店家資料"}
                  </span>
                  <h3>{merchant.name}</h3>
                  <p>
                    {(merchant.aliases || []).length
                      ? `${merchant.aliases!.length} 個別名`
                      : "無搜尋別名"}
                  </p>
                  <div className="tag-row">
                    {(merchant.category_ids || []).slice(0, 3).map((id) => (
                      <span className="tag" key={id}>
                        {categories.find((category) => category.id === id)
                          ?.name || id}
                      </span>
                    ))}
                  </div>
                  <StatusBadge active={Boolean(merchant.is_active)} />
                </div>
                <button
                  className="button ghost"
                  onClick={() => openEditor(merchant)}
                >
                  修改
                </button>
              </article>
            ))}
          </div>
        )}
      </section>

      {showCreate && (
        <Dialog
          title="新增店家"
          onClose={() => !processing && setShowCreate(false)}
        >
          <form className="stack" onSubmit={create}>
            {merchantFields(form, setForm)}
            <div className="dialog-actions">
              <button
                type="button"
                className="button ghost"
                onClick={() => setShowCreate(false)}
              >
                取消
              </button>
              <button
                className="button"
                disabled={
                  processing || !form.name.trim() || !form.category_ids.length
                }
              >
                {processing ? "新增中…" : "新增店家"}
              </button>
            </div>
          </form>
        </Dialog>
      )}

      {editing && (
        <Dialog
          title="修改店家資料"
          onClose={() => !processing && setEditing(null)}
        >
          <form className="stack" onSubmit={save}>
            {merchantFields(editForm, setEditForm)}
            {editing.is_system && (
              <div className="notice">
                這是系統預設店家，可以修改資料與狀態，但不能刪除。
              </div>
            )}
            <p className="muted">
              目前狀態：{editing.is_active ? "啟用" : "停用"}
            </p>
            <div className="dialog-actions split-actions">
              <button
                type="button"
                className="button danger"
                disabled={processing || editing.is_system}
                onClick={() => void remove()}
              >
                刪除店家
              </button>
              <span />
              <button
                className="button"
                disabled={
                  processing ||
                  !editForm.name.trim() ||
                  !editForm.category_ids.length
                }
              >
                {processing ? "處理中…" : "儲存"}
              </button>
              <button
                type="button"
                className="button secondary"
                disabled={processing}
                onClick={() => void persist(!editing.is_active)}
              >
                {editing.is_active ? "停用" : "啟用"}
              </button>
            </div>
          </form>
        </Dialog>
      )}
    </>
  );
}
