import { FormEvent, useEffect, useMemo, useState } from "react";
import type { Merchant, PaymentOption, Recommendation } from "../../../models";
import { todayInTaipei } from "../../../utils/recommendationText";
import { Category, createTransaction, getCategories, getMerchants, getRecommendations } from "../../MemberApi";
import type { Confirmation } from "./types";

export function useRecommend() {
  const [amount, setAmount] = useState("200"), [category, setCategory] = useState("");
  const [merchantId, setMerchantId] = useState(""), [otherMerchant, setOtherMerchant] = useState("");
  const [merchants, setMerchants] = useState<Merchant[]>([]);
  const [availableCategories, setAvailableCategories] = useState<Category[]>([]);
  const [date, setDate] = useState(todayInTaipei);
  const [results, setResults] = useState<Recommendation[]>([]);
  const [error, setError] = useState(""), [notice, setNotice] = useState("");
  const [confirmation, setConfirmation] = useState<Confirmation | null>(null);
  const [loadingCategories, setLoadingCategories] = useState(true);
  const [loadingMerchants, setLoadingMerchants] = useState(false);
  const [recommending, setRecommending] = useState(false);
  const [recording, setRecording] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  useEffect(() => {
    getCategories().then((categories) => {
      setAvailableCategories(categories);
      if (categories[0]) setCategory(categories[0].id);
    }).catch((e) => setError((e as Error).message)).finally(() => setLoadingCategories(false));
  }, []);

  useEffect(() => {
    if (!category) return;
    setLoadingMerchants(true);
    getMerchants(category).then((stores) => {
      setMerchants(stores); setMerchantId(""); setOtherMerchant("");
    }).catch((e) => setError((e as Error).message)).finally(() => setLoadingMerchants(false));
  }, [category]);

  const merchantName = useMemo(
    () => merchantId ? merchants.find((merchant) => merchant.id === merchantId)?.name || "" : otherMerchant.trim(),
    [merchantId, merchants, otherMerchant],
  );

  async function run(e?: FormEvent) {
    if (e) { e.preventDefault(); setNotice(""); }
    setError(""); setRecommending(true);
    try {
      const data = await getRecommendations({
        amount_minor: Math.round(Number(amount) * 100),
        category_id: category, merchant_id: merchantId, merchant_name: merchantName, date,
      });
      setResults(data.recommendations); setHasSearched(true);
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setRecommending(false); }
  }

  async function transact(card: Recommendation, option: PaymentOption, paymentCode: string) {
    setRecording(true); setError("");
    try {
      const data = await createTransaction({
        card_id: card.card_id, amount_minor: Math.round(Number(amount) * 100),
        category_id: category, merchant_id: merchantId, merchant_name: merchantName,
        payment_method_id: paymentCode, transaction_date: date,
        recommendation_summary: option.allocations.map((allocation) => ({
          rule_id: allocation.rule_id, allocated_reward: allocation.allocated_reward,
        })),
      });
      setNotice(data.recommendation_changed ? "實際回饋因 cap 或規則更新而改變" : "交易已記錄，cap 已更新");
      setConfirmation(null); await run();
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setRecording(false); }
  }

  return {
    amount, availableCategories, category, confirmation, date, error,
    hasSearched, loadingCategories, loadingMerchants, merchantId, merchantName,
    merchants, notice, otherMerchant, recommending, recording, results,
    run, setAmount, setCategory, setConfirmation, setDate, setMerchantId,
    setOtherMerchant, transact,
  };
}
