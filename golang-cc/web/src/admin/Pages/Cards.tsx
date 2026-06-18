import { FormEvent, useCallback, useEffect, useState } from "react";
import { Dialog, Empty, Field } from "../../components";
import type { CatalogCard } from "../../models";
import Head from "../../tool/Head";
import {
  Bank,
  createCard,
  deleteCard,
  getBanks,
  getCards,
  updateCard,
} from "../AdminApi";

type CardForm = {
  bank_id: string;
  name: string;
  tiers: string;
};

const emptyForm = (bankID = ""): CardForm => ({
  bank_id: bankID,
  name: "",
  tiers: "",
});

export default function Cards() {
  const [items, setItems] = useState<CatalogCard[]>([]);
  const [banks, setBanks] = useState<Bank[]>([]);
  const [form, setForm] = useState<CardForm>(emptyForm());
  const [editing, setEditing] = useState<CatalogCard | null>(null);
  const [editForm, setEditForm] = useState<CardForm>(emptyForm());
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      const [cards, nextBanks] = await Promise.all([getCards(), getBanks()]);
      setItems(cards);
      setBanks(nextBanks);
      setForm((current) => ({
        ...current,
        bank_id: current.bank_id || nextBanks[0]?.id || "",
      }));
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
      await createCard(form.bank_id, form.name.trim(), form.tiers);
      setForm(emptyForm(form.bank_id));
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  function openEditor(card: CatalogCard) {
    setError("");
    setEditing(card);
    setEditForm({
      bank_id: card.bank_id,
      name: card.name,
      tiers: card.account_tiers.join(", "),
    });
  }

  async function persist(active: boolean) {
    if (!editing) return;
    setSaving(true);
    setError("");
    try {
      await updateCard(
        editing.id,
        editForm.bank_id,
        editForm.name.trim(),
        editForm.tiers,
        active,
      );
      setEditing(null);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setSaving(false);
    }
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (editing) await persist(editing.is_active);
  }

  async function remove() {
    if (
      !editing ||
      !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)
    ) {
      return;
    }
    setSaving(true);
    setError("");
    try {
      await deleteCard(editing.id);
      setEditing(null);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <Head title="銀行卡片目錄" text="會員只能從此目錄加入持有卡片。" />
      <section className="panel">
        {error && <div className="error">{error}</div>}
        <form className="toolbar" onSubmit={create}>
          <select
            value={form.bank_id}
            disabled={creating || !banks.length}
            onChange={(event) =>
              setForm({ ...form, bank_id: event.target.value })
            }
          >
            {banks.map((bank) => (
              <option value={bank.id} key={bank.id}>
                {bank.name}
              </option>
            ))}
          </select>
          <input
            value={form.name}
            disabled={creating}
            onChange={(event) =>
              setForm({ ...form, name: event.target.value })
            }
            placeholder="卡片名稱"
            required
          />
          <input
            value={form.tiers}
            disabled={creating}
            onChange={(event) =>
              setForm({ ...form, tiers: event.target.value })
            }
            placeholder="帳戶等級（逗號分隔）"
          />
          <button
            className="button"
            disabled={creating || !form.bank_id}
          >
            {creating ? "新增中…" : "新增卡片"}
          </button>
        </form>

        {loading ? (
          <p className="muted">載入卡片目錄中…</p>
        ) : items.length === 0 ? (
          <Empty title="尚未建立卡片" text="建立銀行後即可新增卡片。" />
        ) : (
          <div className="list" style={{ marginTop: "1em" }}>
            {items.map((card) => (
              <div className="list-row" key={card.id}>
                <div>
                  <h3>
                    {card.bank_name} · {card.name}
                  </h3>
                  <p>
                    {card.activities.length} 個活動 ·{" "}
                    {card.account_tiers.length
                      ? `等級：${card.account_tiers.join("、")}`
                      : "無帳戶等級"}{" "}
                    · {card.is_active ? "啟用" : "停用"}
                  </p>
                </div>
                <button
                  type="button"
                  className="button ghost"
                  onClick={() => openEditor(card)}
                >
                  修改
                </button>
              </div>
            ))}
          </div>
        )}
      </section>

      {editing && (
        <Dialog
          title="修改卡片資訊"
          onClose={() => !saving && setEditing(null)}
        >
          <form className="stack" onSubmit={save}>
            <Field label="發卡銀行">
              <select
                value={editForm.bank_id}
                disabled={saving}
                onChange={(event) =>
                  setEditForm({ ...editForm, bank_id: event.target.value })
                }
              >
                {banks.map((bank) => (
                  <option value={bank.id} key={bank.id}>
                    {bank.name}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="卡片名稱">
              <input
                value={editForm.name}
                disabled={saving}
                required
                onChange={(event) =>
                  setEditForm({ ...editForm, name: event.target.value })
                }
              />
            </Field>
            <Field label="帳戶等級" hint="多個種類請用逗號分隔">
              <input
                value={editForm.tiers}
                disabled={saving}
                placeholder="例如：一般戶, 財富管理戶"
                onChange={(event) =>
                  setEditForm({ ...editForm, tiers: event.target.value })
                }
              />
            </Field>
            <p className="muted">
              目前狀態：{editing.is_active ? "啟用" : "停用"}
            </p>
            <div className="toolbar">
              <button className="button" disabled={saving}>
                {saving ? "處理中…" : "儲存"}
              </button>
              <button
                type="button"
                className="button secondary"
                disabled={saving}
                onClick={() => void persist(!editing.is_active)}
              >
                {editing.is_active ? "停用" : "啟用"}
              </button>
              <button
                type="button"
                className="button danger"
                disabled={saving}
                onClick={() => void remove()}
              >
                刪除
              </button>
            </div>
          </form>
        </Dialog>
      )}
    </>
  );
}
