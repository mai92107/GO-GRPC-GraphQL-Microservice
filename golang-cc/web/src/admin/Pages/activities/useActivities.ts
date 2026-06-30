import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import type { CatalogCard, Merchant, PaymentMethod, Unit } from "../../../models";
import { activateActivity, Activity, ActivityBenefitInput, Category, createActivity, deleteActivity, getCard, getCards, getCategories, getMerchants, getPaymentMethods, getRewardUnits, publishRewardComponentVersion, updateActivity } from "../../AdminApi";
import { benefitPayload, mapCardActivities, nextFormForMode, splitCardTypes, unavailableSelectableTypes, validateBenefitForm } from "./activityHelpers";
import { filterBankCards } from "./activityFilters";
import { ActivityForm, BenefitMode, emptyActivityForm } from "./types";
import { Network } from "lucide-react";

export function useActivities() {
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
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [editingActivity, setEditingActivity] = useState<Activity | null>(null);
  const [editBenefits, setEditBenefits] = useState<ActivityBenefitInput[]>([]);
  const [activeBank, setActiveBank] = useState("");
  const [expandedCards, setExpandedCards] = useState<string[]>([]);
  const [cardActivities, setCardActivities] = useState<Record<string, Activity[]>>({});
  const [cardLoadingID, setCardLoadingID] = useState("");
  const [form, setForm] = useState<ActivityForm>(emptyActivityForm());

  const load = useCallback(async () => {
    try {
      const [nextCards, nextUnits, nextCategories, nextMethods, nextMerchants] = await Promise.all([getCards(), getRewardUnits(), getCategories(), getPaymentMethods(), getMerchants()]);
      const cardItems = Array.isArray(nextCards) ? nextCards : [];
      setItems([]); setCardActivities({}); setCards(cardItems); setUnits(nextUnits);
      setCategories(nextCategories.filter((category) => category.is_active));
      setMethods(nextMethods.filter((method) => method.is_active && method.id !== "any_payment"));
      setMerchants(nextMerchants.filter((merchant) => merchant.is_active));
      setActiveBank((current) => current || cardItems[0]?.bank_name || "");
      setForm((current) => ({ ...current, bank_name: current.bank_name || cardItems[0]?.bank_name || "", card_product_id: current.card_product_id || cardItems[0]?.id || "", network_ids: current.network_ids.length ? current.network_ids : (cardItems[0]?.networks || []).map((network) => network), reward_unit_id: current.reward_unit_id || nextUnits[0]?.id || "", merchant_ids: current.merchant_ids || [] }));
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { void load(); }, [load]);
  const bankOptions = useMemo(() => [...new Set(cards.map((card) => card.bank_name))], [cards]);
  const availableCards = useMemo(() => cards.filter((card) => card.bank_name === form.bank_name), [cards, form.bank_name]);
  const selectedCard = cards.find((card) => card.id === form.card_product_id);
  const selectedNetworks = useMemo(() => (selectedCard?.networks || []).map((network) => network), [selectedCard]);
  const allNetworksSelected = selectedNetworks.length > 0 && selectedNetworks.every((id) => form.network_ids.includes(id));
  const qualifiedTypes = splitCardTypes(selectedCard?.qualified_type);
  const selectableTypes = splitCardTypes(selectedCard?.selectable_type);
  const unavailableTypes = useMemo(() => unavailableSelectableTypes(items, benefits, form), [benefits, form, items]);
  const selectedCards = useMemo(() => filterBankCards(cards, activeBank, query, cardActivities), [activeBank, cardActivities, cards, query]);

  useEffect(() => {
    if (!showCreate || loading) return;
    if (!form.bank_name && bankOptions[0]) { setForm((current) => ({ ...current, bank_name: bankOptions[0] })); return; }
    if (!form.bank_name || (form.card_product_id && availableCards.some((card) => card.id === form.card_product_id))) return;
    setForm((current) => ({ ...current, card_product_id: availableCards[0]?.id || "", network_ids: [], selectable_type: "" }));
  }, [availableCards, bankOptions, form.bank_name, form.card_product_id, loading, showCreate]);

  useEffect(() => {
    if (!showCreate || !selectedCard) return;
    const allowed = new Set(selectedNetworks);
    const nextIDs = form.network_ids.filter((networkID) => allowed.has(networkID));
    if (!nextIDs.length && selectedNetworks.length) setForm((current) => ({ ...current, network_ids: selectedNetworks }));
    if (nextIDs.length && nextIDs.length !== form.network_ids.length) setForm((current) => ({ ...current, network_ids: nextIDs }));
  }, [form.network_ids, selectedCard, selectedNetworks, showCreate]);

  async function refreshCardActivities(cardID: string) {
    setCardLoadingID(cardID); setError("");
    try {
      const activities = mapCardActivities(await getCard(cardID));
      setCardActivities((current) => ({ ...current, [cardID]: activities }));
      setItems((current) => [...current.filter((activity) => activity.card_product_id !== cardID), ...activities]);
      return activities;
    } catch (requestError) { setError((requestError as Error).message); return []; }
    finally { setCardLoadingID(""); }
  }

  async function toggleCard(card: CatalogCard) {
    if (expandedCards.includes(card.id)) { setExpandedCards((current) => current.filter((id) => id !== card.id)); return; }
    setExpandedCards((current) => [...current, card.id]);
    if (!cardActivities[card.id]) await refreshCardActivities(card.id);
  }

  function stageBenefit() {
    const validation = validateBenefitForm(form, unavailableTypes);
    if (validation) { setError(validation); return; }
    setError(""); setBenefits((current) => [...current, benefitPayload(form)]);
    setForm((current) => ({ ...current, benefit_name: "", monthly_cap: "", layer: "base", display_order: 0, effect_type: "ADD_RATE", reward_value: "0.01", benefit_mode: "standard", selectable_type: "", action_required: "none", action_message: "" }));
  }

  function selectBenefitMode(mode: BenefitMode) {
    setForm((current) => nextFormForMode(current, mode, selectableTypes, unavailableTypes));
  }

  async function create(event: FormEvent) {
    event.preventDefault();
    if (form.benefit_name.trim()) {
      const validation = validateBenefitForm(form, unavailableTypes);
      if (validation) { setError(validation); return; }
    }
    const allBenefits = form.benefit_name.trim() ? [...benefits, benefitPayload(form)] : benefits;
    if (!allBenefits.length) { setError("回饋方案至少需要一項優惠條件。"); return; }
    setCreating(true); setError("");
    try {
      await createActivity(form.card_product_id, form.name.trim(), form.start_date, form.end_date, form.source_url.trim(), form.network_ids, allBenefits);
      setBenefits([]); setShowCreate(false);
      setForm((current) => ({ ...current, name: "", source_url: "", network_ids: [], benefit_name: "", monthly_cap: "", layer: "base", display_order: 0, effect_type: "ADD_RATE", reward_value: "0.01", benefit_mode: "standard", selectable_type: "", action_required: "none", action_message: "" }));
      await refreshCardActivities(form.card_product_id);
      setExpandedCards((current) => current.includes(form.card_product_id) ? current : [...current, form.card_product_id]);
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setCreating(false); }
  }

  async function toggle(activity: Activity) {
    setProcessingID(activity.id); setError("");
    try { await activateActivity(activity); await refreshCardActivities(activity.card_product_id); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessingID(null); }
  }

  function openRelationshipEditor(activity: Activity) {
    setEditingActivity(activity); setEditBenefits(activity.benefits.map((benefit) => ({ ...benefit }))); setError("");
  }

  async function saveRelationships(event: FormEvent) {
    event.preventDefault();
    if (!editingActivity) return;
    setProcessingID(editingActivity.id); setError("");
    try {
      await updateActivity({ ...editingActivity, benefits: editBenefits.map((benefit) => ({ ...benefit, id: benefit.id || "", stack_group: benefit.stack_group.trim() })) });
      setEditingActivity(null); setEditBenefits([]);
      await refreshCardActivities(editingActivity.card_product_id);
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessingID(null); }
  }

  async function remove(activity: Activity) {
    if (!confirm(`確定刪除「${activity.name}」？`)) return;
    setProcessingID(activity.id); setError("");
    try { await deleteActivity(activity.id); await refreshCardActivities(activity.card_product_id); }
    catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessingID(null); }
  }

  async function scheduleVersion(benefit: Activity["benefits"][number]) {
    const rewardValue = prompt("新效果值", benefit.reward_value);
    if (!rewardValue) return;
    const effectiveDate = prompt("生效日期（YYYY-MM-DD）", "2026-07-01");
    if (!effectiveDate) return;
    const reason = prompt("變更原因", "銀行權益調整") || "";
    if (!confirm(`版本預覽\n目前：${benefit.reward_value}\n新版：${rewardValue}\n生效：${effectiveDate}\n發布後將保留舊版本，確定繼續？`)) return;
    setProcessingID(benefit.id); setError("");
    try {
      await publishRewardComponentVersion(benefit.id, { reward_unit_id: benefit.reward_unit_id, name: benefit.name, effect_type: benefit.effect_type, reward_value: rewardValue, effective_from: `${effectiveDate}T00:00:00+08:00`, effective_to: null, announced_at: new Date().toISOString(), change_reason: reason, display_change_until: `${effectiveDate}T23:59:59+08:00` });
      const owner = editingActivity || items.find((activity) => activity.benefits.some((item) => item.id === benefit.id));
      if (owner) await refreshCardActivities(owner.card_product_id);
    } catch (requestError) { setError((requestError as Error).message); }
    finally { setProcessingID(null); }
  }

  function closeRelationshipDialog() {
    if (!editingActivity || processingID === editingActivity.id) return;
    setEditingActivity(null); setEditBenefits([]);
  }

  return { activeBank, allNetworksSelected, availableCards, bankOptions, benefits, cardActivities, cardLoadingID, cards, categories, creating, editBenefits, editingActivity, error, expandedCards, form, loading, merchants, methods, processingID, qualifiedTypes, query, selectableTypes, selectedCard, selectedCards, selectedNetworks, showCreate, unavailableTypes, units, closeRelationshipDialog, create, openRelationshipEditor, remove, saveRelationships, scheduleVersion, selectBenefitMode, setActiveBank, setBenefits, setEditBenefits, setForm, setQuery, setShowCreate, stageBenefit, toggle, toggleCard };
}
