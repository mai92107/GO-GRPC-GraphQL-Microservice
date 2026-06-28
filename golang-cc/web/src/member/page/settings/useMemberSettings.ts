import { useEffect, useMemo, useState } from "react";
import type { PaymentMethod } from "../../../models";
import type { PreferenceWrite, RewardPreference } from "../../../preferences";
import { isFixedPayment, priorityFromWeight } from "../../../utils/settingsText";
import { getPaymentMethods, getRewardPreferences, updatePaymentMethods, updateRewardPreferences } from "../../MemberApi";
import { Priority, priorityOptions, SettingsView } from "./types";

export function useMemberSettings() {
  const [view, setView] = useState<SettingsView>("overview");
  const [preferences, setPreferences] = useState<RewardPreference[]>([]);
  const [priorities, setPriorities] = useState<Record<string, Priority>>({});
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>([]);
  const [showMorePayments, setShowMorePayments] = useState(false);
  const [notice, setNotice] = useState(""), [error, setError] = useState("");
  const [loading, setLoading] = useState(true), [saving, setSaving] = useState(false);
  const [preferenceDirty, setPreferenceDirty] = useState(false);
  const [paymentDirty, setPaymentDirty] = useState(false);

  useEffect(() => {
    Promise.all([getRewardPreferences(), getPaymentMethods()]).then(([items, methods]) => {
      setPreferences(items);
      setPriorities(Object.fromEntries(items.map((item) => [item.reward_unit_id, priorityFromWeight(item.weight)])));
      setPaymentMethods(methods);
    }).catch((requestError) => setError((requestError as Error).message)).finally(() => setLoading(false));
  }, []);

  const configurablePayments = useMemo(() => paymentMethods.filter((method) => !isFixedPayment(method)), [paymentMethods]);
  const selectedPayments = useMemo(() => configurablePayments.filter((method) => method.is_available), [configurablePayments]);
  const otherPayments = useMemo(() => configurablePayments.filter((method) => !method.is_available), [configurablePayments]);
  const highPriorityCount = useMemo(() => preferences.filter((item) => priorities[item.reward_unit_id] === "high").length, [preferences, priorities]);

  function openView(nextView: SettingsView) { setNotice(""); setError(""); setView(nextView); }
  function setPriority(unitID: string, priority: Priority) {
    setPriorities((current) => ({ ...current, [unitID]: priority }));
    setPreferenceDirty(true); setNotice("");
  }
  function togglePayment(id: string) {
    setPaymentMethods((current) => current.map((method) => method.id === id ? { ...method, is_available: !method.is_available } : method));
    setPaymentDirty(true); setNotice("");
  }
  async function savePreferences() {
    setSaving(true); setError("");
    try {
      const writes: PreferenceWrite[] = preferences.map((item) => ({
        reward_unit_id: item.reward_unit_id,
        weight: priorityOptions.find((option) => option.id === (priorities[item.reward_unit_id] || "normal"))?.weight || "1.1",
      }));
      await updateRewardPreferences(writes);
      setPreferenceDirty(false); setNotice("回饋偏好已儲存。");
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setSaving(false); }
  }
  async function savePayments() {
    setSaving(true); setError("");
    try {
      await updatePaymentMethods(configurablePayments.filter((method) => method.is_available).map((method) => method.id));
      setPaymentDirty(false); setNotice("支付方式已儲存。");
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setSaving(false); }
  }

  return {
    error, highPriorityCount, loading, notice, otherPayments, paymentDirty,
    preferenceDirty, preferences, priorities, saving, selectedPayments,
    showMorePayments, view, openView, savePayments, savePreferences,
    setPriority, setShowMorePayments, togglePayment,
  };
}
