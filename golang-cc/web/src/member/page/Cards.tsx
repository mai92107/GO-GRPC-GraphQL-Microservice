import { Leaf, Plus, Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { ActivityDetails, Dialog, Empty, Field } from "../../components";
import type {
  Card,
  CatalogCard,
  Merchant,
  PaymentMethod,
  Unit,
} from "../../models";
import {
  Category,
  createCard,
  deleteCard,
  getCards,
  getCatalogCards,
  getCategories,
  getMerchants,
  getPaymentMethods,
  getRewardUnits,
  MemberCardInput,
} from "../MemberApi";

type CardForm = {
  card_product_id: string;
  nickname: string;
  last_four: string;
  statement_day: string;
  payment_due_day: string;
  account_tier: string;
  is_active: boolean;
};

const emptyForm = (cardProductID = ""): CardForm => ({
  card_product_id: cardProductID,
  nickname: "",
  last_four: "",
  statement_day: "",
  payment_due_day: "",
  account_tier: "",
  is_active: true,
});

const todayInTaipei = () =>
  new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Taipei",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());

const activeActivities = (card: CatalogCard, today: string) =>
  card.activities.filter(
    (activity) =>
      activity.is_active &&
      activity.start_date <= today &&
      activity.end_date >= today,
  );

export default function Cards() {
  const [cards, setCards] = useState<Card[]>([]);
  const [catalog, setCatalog] = useState<CatalogCard[]>([]);
  const [units, setUnits] = useState<Unit[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [merchants, setMerchants] = useState<Merchant[]>([]);
  const [form, setForm] = useState<CardForm>(emptyForm());
  const [showCreate, setShowCreate] = useState(false);
  const [detail, setDetail] = useState<CatalogCard | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [deletingID, setDeletingID] = useState<string | null>(null);
  const [error, setError] = useState("");
  const today = useMemo(todayInTaipei, []);

  const load = useCallback(async () => {
    setError("");
    try {
      const [nextCards, nextCatalog, nextUnits, nextCategories, nextMethods] =
        await Promise.all([
          getCards(),
          getCatalogCards(),
          getRewardUnits(),
          getCategories(),
          getPaymentMethods(),
        ]);

      setCards(nextCards);
      setCatalog(nextCatalog);
      setUnits(nextUnits);
      setCategories(nextCategories);
      setMethods(nextMethods);
      setForm((current) => ({
        ...current,
        card_product_id:
          current.card_product_id || nextCatalog[0]?.id || "",
      }));

      const merchantGroups = await Promise.all(
        nextCategories.map((category) => getMerchants(category.code)),
      );
      const uniqueMerchants = new Map(
        merchantGroups.flat().map((merchant) => [merchant.code, merchant]),
      );

      setMerchants([...uniqueMerchants.values()]);
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const selectedCard = catalog.find(
    (card) => card.id === form.card_product_id,
  );
  const detailActivities = detail ? activeActivities(detail, today) : [];

  async function create(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      const input: MemberCardInput = {
        ...form,
        statement_day: form.statement_day
          ? Number(form.statement_day)
          : null,
        payment_due_day: form.payment_due_day
          ? Number(form.payment_due_day)
          : null,
      };
      await createCard(input);
      setShowCreate(false);
      setForm(emptyForm(catalog[0]?.id));
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setSaving(false);
    }
  }

  async function remove(id: string) {
    if (!confirm("確定刪除這張卡？已有交易的卡片將無法刪除。")) return;

    setDeletingID(id);
    setError("");
    try {
      await deleteCard(id);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setDeletingID(null);
    }
  }

  return (
    <>
      <div className="page-head">
        <div>
          <h1>我的卡片夾</h1>
          <p>從卡片目錄加入目前持有的卡片。</p>
        </div>
        <button
          type="button"
          className="button"
          disabled={loading || !catalog.length}
          onClick={() => setShowCreate(true)}
        >
          <Plus size={17} />
          加入卡片
        </button>
      </div>

      {error && <div className="error">{error}</div>}
      {loading && <p className="muted">載入卡片資料中…</p>}

      {!loading && cards.length > 0 && (
        <div className="card-list">
          {cards.map((card) => (
            <article
              className={`credit-card ${card.is_active ? "" : "inactive"}`}
              key={card.id}
            >
              <header>
                <span>{card.issuer}</span>
                <Leaf />
              </header>
              <div>
                <h3>{card.name}</h3>
                <small>
                  {card.is_active ? "啟用" : "停用"}
                  {card.account_tier && ` · ${card.account_tier}`} ·{" "}
                  {card.last_four
                    ? `•••• ${card.last_four}`
                    : "未設定末四碼"}
                </small>
              </div>
              <button
                type="button"
                className="button danger"
                disabled={deletingID === card.id}
                onClick={() => void remove(card.id)}
              >
                <Trash2 size={15} />
                {deletingID === card.id ? "移除中…" : "移除"}
              </button>
            </article>
          ))}
        </div>
      )}

      {!loading && cards.length === 0 && (
        <div className="panel">
          <Empty
            title="卡片夾是空的"
            text="從卡片目錄加入目前持有的卡片。"
          />
        </div>
      )}

      <section className="panel" style={{ marginTop: 18 }}>
        <div className="section-title">
          <h2>所有卡片與活動</h2>
        </div>
        <div className="list">
          {catalog.map((card) => {
            const activities = activeActivities(card, today);
            return (
              <button
                type="button"
                className="list-row clickable-row"
                key={card.id}
                onClick={() => setDetail(card)}
              >
                <div>
                  <h3>
                    {card.bank_name} · {card.name}
                  </h3>
                  <p>
                    {activities.length
                      ? `${activities.length} 個當前活動`
                      : "目前沒有有效活動"}
                  </p>
                </div>
              </button>
            );
          })}
        </div>
      </section>

      {detail && (
        <Dialog
          title={`${detail.bank_name} · ${detail.name}`}
          onClose={() => setDetail(null)}
        >
          {detailActivities.length > 0 ? (
            <ActivityDetails
              activities={detailActivities}
              units={units}
              categories={categories}
              methods={methods}
              merchants={merchants}
            />
          ) : (
            <Empty
              title="目前沒有有效活動"
              text="今天沒有啟用且在活動期間內的活動。"
            />
          )}
        </Dialog>
      )}

      {showCreate && (
        <Dialog
          title="加入卡片夾"
          onClose={() => !saving && setShowCreate(false)}
        >
          <form className="stack" onSubmit={create}>
            <Field label="銀行卡片">
              <select
                value={form.card_product_id}
                disabled={saving}
                onChange={(event) =>
                  setForm({
                    ...form,
                    card_product_id: event.target.value,
                    account_tier: "",
                  })
                }
              >
                {catalog.map((card) => (
                  <option value={card.id} key={card.id}>
                    {card.bank_name} · {card.name}
                  </option>
                ))}
              </select>
            </Field>
            {selectedCard && selectedCard.account_tiers.length > 0 && (
              <Field label="目前帳戶等級">
                <select
                  required
                  value={form.account_tier}
                  disabled={saving}
                  onChange={(event) =>
                    setForm({ ...form, account_tier: event.target.value })
                  }
                >
                  <option value="">請選擇</option>
                  {selectedCard.account_tiers.map((tier) => (
                    <option key={tier}>{tier}</option>
                  ))}
                </select>
              </Field>
            )}
            <Field label="自訂暱稱">
              <input
                value={form.nickname}
                disabled={saving}
                onChange={(event) =>
                  setForm({ ...form, nickname: event.target.value })
                }
              />
            </Field>
            <Field label="卡號末四碼">
              <input
                inputMode="numeric"
                pattern="[0-9]{4}"
                value={form.last_four}
                disabled={saving}
                onChange={(event) =>
                  setForm({ ...form, last_four: event.target.value })
                }
              />
            </Field>
            <div className="two-col">
              <Field label="結帳日">
                <input
                  type="number"
                  min="1"
                  max="31"
                  value={form.statement_day}
                  disabled={saving}
                  onChange={(event) =>
                    setForm({ ...form, statement_day: event.target.value })
                  }
                />
              </Field>
              <Field label="繳款日">
                <input
                  type="number"
                  min="1"
                  max="31"
                  value={form.payment_due_day}
                  disabled={saving}
                  onChange={(event) =>
                    setForm({ ...form, payment_due_day: event.target.value })
                  }
                />
              </Field>
            </div>
            <label>
              <input
                type="checkbox"
                checked={form.is_active}
                disabled={saving}
                onChange={(event) =>
                  setForm({ ...form, is_active: event.target.checked })
                }
              />{" "}
              啟用此卡
            </label>
            <button className="button" disabled={saving}>
              {saving ? "加入中…" : "加入卡片夾"}
            </button>
          </form>
        </Dialog>
      )}
    </>
  );
}
