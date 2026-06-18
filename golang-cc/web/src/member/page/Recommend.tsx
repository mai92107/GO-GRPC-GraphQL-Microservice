import { Sparkles } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { Field, Empty, Dialog } from "../../components";
import { formatScore, formatPercent, formatReward } from "../../format";
import { Merchant, Recommendation, PaymentOption } from "../../models";
import {
  Category,
  createTransaction,
  getCategories,
  getMerchants,
  getRecommendations,
} from "../MemberApi";

type Confirmation = {
  card: Recommendation;
  option: PaymentOption;
  paymentCode: string;
};

const todayInTaipei = () =>
  new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Taipei",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());

export function Recommend() {
  const [amount, setAmount] = useState("200"),
    [category, setCategory] = useState(""),
    [merchantCode, setMerchantCode] = useState(""),
    [otherMerchant, setOtherMerchant] = useState(""),
    [merchants, setMerchants] = useState<Merchant[]>([]),
    [availableCategories, setAvailableCategories] = useState<Category[]>([]),
    [date, setDate] = useState(todayInTaipei),
    [results, setResults] = useState<Recommendation[]>([]),
    [error, setError] = useState(""),
    [confirmation, setConfirmation] = useState<Confirmation | null>(null),
    [notice, setNotice] = useState(""),
    [loadingCategories, setLoadingCategories] = useState(true),
    [loadingMerchants, setLoadingMerchants] = useState(false),
    [recommending, setRecommending] = useState(false),
    [recording, setRecording] = useState(false),
    [hasSearched, setHasSearched] = useState(false);

  useEffect(() => {
    getCategories()
      .then((categories) => {
        setAvailableCategories(categories);
        if (categories[0]) setCategory(categories[0].code);
      })
      .catch((e) => setError((e as Error).message))
      .finally(() => setLoadingCategories(false));
  }, []);

  useEffect(() => {
    if (!category) return;
    setLoadingMerchants(true);
    getMerchants(category)
      .then((stores) => {
        setMerchants(stores);
        setMerchantCode("");
        setOtherMerchant("");
      })
      .catch((e) => setError((e as Error).message))
      .finally(() => setLoadingMerchants(false));
  }, [category]);

  const merchantName = useMemo(
    () =>
      merchantCode
        ? merchants.find((merchant) => merchant.code === merchantCode)?.name ||
          ""
        : otherMerchant.trim(),
    [merchantCode, merchants, otherMerchant],
  );

  async function run(e?: FormEvent) {
    if (e) {
      e.preventDefault();
      setNotice("");
    }
    setError("");
    setRecommending(true);
    try {
      const data = await getRecommendations({
        amount_minor: Math.round(Number(amount) * 100),
        category_code: category,
        merchant_code: merchantCode,
        merchant_name: merchantName,
        date,
      });
      setResults(data.recommendations);
      setHasSearched(true);
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setRecommending(false);
    }
  }

  async function transact(
    card: Recommendation,
    option: PaymentOption,
    paymentCode: string,
  ) {
    setRecording(true);
    setError("");
    try {
      const data = await createTransaction({
        card_id: card.card_id,
        amount_minor: Math.round(Number(amount) * 100),
        category_code: category,
        merchant_code: merchantCode,
        merchant_name: merchantName,
        payment_method_code: paymentCode,
        transaction_date: date,
        recommendation_summary: option.allocations.map((allocation) => ({
          rule_id: allocation.rule_id,
          allocated_reward: allocation.allocated_reward,
        })),
      });
      setNotice(
        data.recommendation_changed
          ? "實際回饋因 cap 或規則更新而改變"
          : "交易已記錄，cap 已更新",
      );
      setConfirmation(null);
      await run();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setRecording(false);
    }
  }
  return (
    <>
      <div className="page-head">
        <div>
          <h1>這次刷哪張？</h1>
          <p>輸入消費情境，系統會連同最佳支付方式一起推薦。</p>
        </div>
      </div>
      {notice && <div className="notice">{notice}</div>}
      <div className="grid">
        <form className="panel recommend-form stack" onSubmit={run}>
          <Field label="消費金額（NT$）">
            <input
              type="number"
              min="0.01"
              step="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
            />
          </Field>
          <Field label="消費類別">
            <select
              required
              value={category}
              disabled={loadingCategories || recommending}
              onChange={(e) => setCategory(e.target.value)}
            >
              {availableCategories.map((c) => (
                <option key={c.code} value={c.code}>
                  {c.name}
                </option>
              ))}
            </select>
          </Field>
          <Field label="消費店家">
            <select
              value={merchantCode}
              disabled={loadingMerchants || recommending}
              onChange={(e) => setMerchantCode(e.target.value)}
            >
              <option value="">其他</option>
              {merchants.map((x) => (
                <option key={x.code} value={x.code}>
                  {x.name}
                </option>
              ))}
            </select>
          </Field>
          {!merchantCode && (
            <Field label="其他店家名稱（選填）">
              <input
                value={otherMerchant}
                onChange={(e) => setOtherMerchant(e.target.value)}
                placeholder="可留空"
              />
            </Field>
          )}
          <Field label="消費日期">
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              required
            />
          </Field>
          {error && <div className="error">{error}</div>}
          <button
            className="button"
            disabled={!category || recommending || loadingCategories}
          >
            <Sparkles size={17} />
            {recommending ? "計算中…" : "取得推薦"}
          </button>
        </form>
        <section className="results" aria-live="polite">
          {results.length === 0 ? (
            <div className="panel">
              <Empty
                title={hasSearched ? "沒有符合條件的推薦" : "準備好開始推薦"}
                text={
                  hasSearched
                    ? "請調整消費條件，或確認卡片與活動目前為啟用狀態。"
                    : "填入左側消費資訊，我們會計算每張卡與支付方式的有效回饋。"
                }
              />
            </div>
          ) : (
            results.map((card) => (
              <article
                className={`recommend-card ${card.rank === 1 ? "best" : ""}`}
                key={`${card.card_id}-${card.total_score}`}
              >
                <span className="rank">第 {card.rank} 名</span>
                <h3>
                  {card.card_name}(得分:{formatScore(card.total_score)})
                </h3>
                {card.payment_options.map((option) => (
                  <div className="allocations" key={option.payment_method_code}>
                    <h4>
                      {option.payment_method_name} · 預估回饋(
                      {formatPercent(
                        String(
                          option.allocations.reduce(
                            (sum, a) => sum + Number(a.rate),
                            0,
                          ),
                        ),
                      )}
                      )
                    </h4>
                    {option.reminders.map((x) => (
                      <div className="notice" key={x}>
                        {x}
                      </div>
                    ))}
                    {option.allocations.map((a) => (
                      <div className="allocation" key={a.rule_id}>
                        <span>
                          {a.rule_name}
                          <br />
                          <small className="muted">
                            {a.activity_name} · 累計層 {a.stack_group}
                            {a.remaining_before &&
                              ` · cap 剩餘 ${formatReward(a.remaining_before, a.reward_unit)}`}
                          </small>
                        </span>
                        <strong>
                          {formatReward(a.allocated_reward, a.reward_unit)}
                        </strong>
                      </div>
                    ))}
                    <button
                      className="button secondary"
                      onClick={() =>
                        setConfirmation({
                          card,
                          option,
                          paymentCode:
                            option.payment_methods[0]?.code ||
                            option.payment_method_code,
                        })
                      }
                    >
                      使用以上支付方式
                    </button>
                  </div>
                ))}
              </article>
            ))
          )}
        </section>
      </div>
      {confirmation && (
        <Dialog
          title="確認刷卡"
          onClose={() => !recording && setConfirmation(null)}
        >
          <div className="stack">
            <p>
              將以 <strong>{confirmation.card.card_name}</strong> 建立在「
              {merchantName || "其他店家"}」的 NT${amount} 交易。
            </p>
            {confirmation.option.payment_methods.length > 1 && (
              <Field label="實際支付方式">
                <select
                  value={confirmation.paymentCode}
                  disabled={recording}
                  onChange={(e) =>
                    setConfirmation({
                      ...confirmation,
                      paymentCode: e.target.value,
                    })
                  }
                >
                  {confirmation.option.payment_methods.map((method) => (
                    <option key={method.code} value={method.code}>
                      {method.name}
                    </option>
                  ))}
                </select>
              </Field>
            )}
            {confirmation.option.reminders.map((x) => (
              <div className="notice" key={x}>
                {x}
              </div>
            ))}
            <button
              className="button"
              disabled={recording}
              onClick={() =>
                void transact(
                  confirmation.card,
                  confirmation.option,
                  confirmation.paymentCode,
                )
              }
            >
              {recording ? "記錄中…" : "確認並記錄交易"}
            </button>
          </div>
        </Dialog>
      )}
    </>
  );
}
