import { Trash2 } from "lucide-react";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { CatalogCard, Merchant, PaymentMethod, Unit } from "../../models";
import { Empty, Field } from "../../components";
import Head from "../../tool/Head";
import {
  activateActivity,
  Activity,
  ActivityBenefitInput,
  Category,
  createActivity,
  deleteActivity,
  getActivities,
  getCards,
  getCategories,
  getMerchants,
  getPaymentMethods,
  getRewardUnits,
} from "../AdminApi";

export default function Activities() {
  const [items, setItems] = useState<Activity[]>([]);
  const [cards, setCards] = useState<CatalogCard[]>([]);
  const [units, setUnits] = useState<Unit[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [merchants, setMerchants] = useState<Merchant[]>([]);
  const [benefits, setBenefits] = useState<ActivityBenefitInput[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [processingID, setProcessingID] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [form, setForm] = useState({
    card_product_id: "",
    name: "",
    start_date: "2026-07-01",
    end_date: "2026-12-31",
    source_url: "",
    reward_unit_id: "",
    benefit_name: "",
    rate: "0.01",
    monthly_cap: "",
    stack_group: "base",
    category_code: "general",
    action_required: "none" as ActivityBenefitInput["action_required"],
    action_message: "",
    payment_methods: [] as string[],
    merchant_codes: [] as string[],
  });
  const load = useCallback(async () => {
    try {
      const [activities, nextCards, nextUnits, nextCategories, nextMethods, nextMerchants] =
        await Promise.all([
          getActivities(),
          getCards(),
          getRewardUnits(),
          getCategories(),
          getPaymentMethods(),
          getMerchants(),
        ]);
      const activeCategories = nextCategories.filter(
        (category) => category.is_active,
      );
      setItems(activities);
      setCards(nextCards);
      setUnits(nextUnits);
      setCategories(activeCategories);
      setMethods(
        nextMethods.filter(
          (method) => method.is_active && method.code !== "any_payment",
        ),
      );
      setMerchants(nextMerchants.filter((merchant) => merchant.is_active));
      setForm((current) => ({
        ...current,
        card_product_id:
          current.card_product_id || nextCards[0]?.id || "",
        reward_unit_id:
          current.reward_unit_id || nextUnits[0]?.id || "",
        category_code:
          activeCategories.some(
            (category) => category.code === current.category_code,
          )
            ? current.category_code
            : activeCategories[0]?.code || "",
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

  const benefitPayload = () => ({
    reward_unit_id: form.reward_unit_id,
    name: form.benefit_name,
    rate: form.rate,
    monthly_cap: form.monthly_cap || null,
    stack_group: form.stack_group,
    priority: 100,
    required_account_tiers: [],
    action_required: form.action_required,
    action_message: form.action_message,
    payment_methods: form.payment_methods,
    category_codes: [form.category_code],
    merchant_codes: form.merchant_codes,
  });

  function stageBenefit() {
    if (!form.benefit_name.trim() || !form.reward_unit_id || !form.category_code) {
      setError("優惠名稱、回饋單位與消費類別為必填。");
      return;
    }
    setError("");
    setBenefits((current) => [...current, benefitPayload()]);
    setForm((current) => ({
      ...current,
      benefit_name: "",
      monthly_cap: "",
      action_message: "",
    }));
  }

  async function create(event: FormEvent) {
    event.preventDefault();
    const allBenefits = form.benefit_name.trim()
      ? [...benefits, benefitPayload()]
      : benefits;
    if (!allBenefits.length) {
      setError("活動至少需要一項優惠條件。");
      return;
    }
    setCreating(true);
    setError("");
    try {
      await createActivity(
        form.card_product_id,
        form.name.trim(),
        form.start_date,
        form.end_date,
        form.source_url.trim(),
        allBenefits,
      );
      setBenefits([]);
      setForm((current) => ({
        ...current,
        name: "",
        source_url: "",
        benefit_name: "",
        monthly_cap: "",
        action_message: "",
      }));
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function toggle(activity: Activity) {
    setProcessingID(activity.id);
    setError("");
    try {
      await activateActivity(activity);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }

  async function remove(activity: Activity) {
    if (!confirm(`確定刪除「${activity.name}」？`)) return;
    setProcessingID(activity.id);
    setError("");
    try {
      await deleteActivity(activity.id);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setProcessingID(null);
    }
  }
  return (
    <>
      <Head
        title="卡片活動"
        text="一期活動可包含多個優惠條件；同卡活動期間不可重疊。"
      />
      <section className="panel">
        {error && <div className="error">{error}</div>}
        <form className="stack" onSubmit={create}>
          <div className="two-col">
            <Field label="卡片">
              <select
                value={form.card_product_id}
                disabled={creating}
                onChange={(e) =>
                  setForm({ ...form, card_product_id: e.target.value })
                }
              >
                {cards.map((c) => (
                  <option value={c.id} key={c.id}>
                    {c.bank_name} · {c.name}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="活動名稱">
              <input
                required
                value={form.name}
                disabled={creating}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </Field>
          </div>
          <div className="two-col">
            <Field label="開始日期">
              <input
                type="date"
                required
                value={form.start_date}
                disabled={creating}
                onChange={(e) =>
                  setForm({ ...form, start_date: e.target.value })
                }
              />
            </Field>
            <Field label="結束日期">
              <input
                type="date"
                required
                value={form.end_date}
                disabled={creating}
                onChange={(e) => setForm({ ...form, end_date: e.target.value })}
              />
            </Field>
          </div>
          <Field label="官方來源">
            <input
              value={form.source_url}
              disabled={creating}
              onChange={(e) => setForm({ ...form, source_url: e.target.value })}
            />
          </Field>
          <h3>優惠條件（已加入 {benefits.length} 項）</h3>
          {benefits.length > 0 && (
            <div className="list">
              {benefits.map((benefit, index) => (
                <div
                  className="list-row"
                  key={`${benefit.name}-${index}`}
                >
                  <div>
                    <h3>{benefit.name}</h3>
                    <p>
                      回饋率 {benefit.rate} · 累計層 {benefit.stack_group}
                    </p>
                  </div>
                  <button
                    type="button"
                    className="icon-button"
                    aria-label={`移除 ${benefit.name}`}
                    disabled={creating}
                    onClick={() =>
                      setBenefits((current) =>
                        current.filter((_, itemIndex) => itemIndex !== index),
                      )
                    }
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              ))}
            </div>
          )}
          <div className="two-col">
            <Field label="優惠名稱">
              <input
                value={form.benefit_name}
                disabled={creating}
                onChange={(e) =>
                  setForm({ ...form, benefit_name: e.target.value })
                }
              />
            </Field>
            <Field label="回饋單位">
              <select
                value={form.reward_unit_id}
                disabled={creating}
                onChange={(e) =>
                  setForm({ ...form, reward_unit_id: e.target.value })
                }
              >
                {units.map((u) => (
                  <option value={u.id} key={u.id}>
                    {u.name}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <div className="two-col">
            <Field label="回饋率">
              <input
                value={form.rate}
                disabled={creating}
                onChange={(e) => setForm({ ...form, rate: e.target.value })}
              />
            </Field>
            <Field label="累計層">
              <input
                value={form.stack_group}
                disabled={creating}
                onChange={(e) =>
                  setForm({ ...form, stack_group: e.target.value })
                }
              />
            </Field>
          </div>
          <Field label="每月回饋上限（選填）">
            <input
              type="number"
              min="0"
              step="0.000001"
              value={form.monthly_cap}
              disabled={creating}
              onChange={(event) =>
                setForm({ ...form, monthly_cap: event.target.value })
              }
            />
          </Field>
          <div className="two-col">
            <Field label="消費類別">
              <select
                value={form.category_code}
                disabled={creating}
                onChange={(e) =>
                  setForm({ ...form, category_code: e.target.value })
                }
              >
                {categories.map((c) => (
                  <option value={c.code} key={c.code}>
                    {c.name}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="前置操作">
              <select
                value={form.action_required}
                disabled={creating}
                onChange={(e) =>
                  setForm({
                    ...form,
                    action_required: e.target
                      .value as ActivityBenefitInput["action_required"],
                  })
                }
              >
                <option value="none">無</option>
                <option value="registration">需登錄</option>
                <option value="app_switch">需 APP 切換</option>
                <option value="account_setup">需帳戶設定</option>
              </select>
            </Field>
          </div>
          <Field label="限定支付方式（未選代表不限）">
            <div className="toolbar">
              {methods.map((x) => (
                <label key={x.code}>
                  <input
                    type="checkbox"
                    checked={form.payment_methods.includes(x.code)}
                    disabled={creating}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        payment_methods: e.target.checked
                          ? [...form.payment_methods, x.code]
                          : form.payment_methods.filter(
                              (code) => code !== x.code,
                            ),
                      })
                    }
                  />{" "}
                  {x.name}
                </label>
              ))}
            </div>
          </Field>
          <Field label="限定店家（未選代表不限）">
            <div className="toolbar">
              {merchants.map((x) => (
                <label key={x.code}>
                  <input
                    type="checkbox"
                    checked={form.merchant_codes.includes(x.code)}
                    disabled={creating}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        merchant_codes: e.target.checked
                          ? [...form.merchant_codes, x.code]
                          : form.merchant_codes.filter(
                              (code) => code !== x.code,
                            ),
                      })
                    }
                  />{" "}
                  {x.name}
                </label>
              ))}
            </div>
          </Field>
          <Field label="操作提醒">
            <input
              value={form.action_message}
              disabled={creating}
              onChange={(e) =>
                setForm({ ...form, action_message: e.target.value })
              }
            />
          </Field>
          <div className="toolbar">
            <button
              type="button"
              className="button secondary"
              disabled={creating}
              onClick={stageBenefit}
            >
              暫存此優惠條件
            </button>
            <button
              className="button"
              disabled={
                creating ||
                !form.card_product_id ||
                !form.name ||
                !form.start_date ||
                !form.end_date
              }
            >
              {creating ? "建立中…" : "建立活動（包含目前條件）"}
            </button>
          </div>
        </form>
        {loading ? (
          <p className="muted">載入活動中…</p>
        ) : (
          <div className="list">
            {items.map((activity) => (
              <div className="list-row" key={activity.id}>
                <div>
                  <h3>{activity.name}</h3>
                  <p>
                    {activity.start_date}～{activity.end_date} ·{" "}
                    {activity.benefits.length} 項優惠 ·{" "}
                    {activity.is_active ? "啟用" : "停用"}
                  </p>
                </div>
                <div className="toolbar">
                  <button
                    type="button"
                    className="button ghost"
                    disabled={processingID === activity.id}
                    onClick={() => void toggle(activity)}
                  >
                    {activity.is_active ? "停用" : "啟用"}
                  </button>
                  <button
                    type="button"
                    className="icon-button"
                    aria-label={`刪除 ${activity.name}`}
                    disabled={processingID === activity.id}
                    onClick={() => void remove(activity)}
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>
            ))}
            {items.length === 0 && (
              <Empty title="尚未建立活動" text="建立第一個卡片活動。" />
            )}
          </div>
        )}
      </section>
    </>
  );
}
