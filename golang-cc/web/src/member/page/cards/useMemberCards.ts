import { FormEvent, useCallback, useEffect, useState } from "react";
import type { Card, CatalogCard } from "../../../models";
import {
  createCard,
  deleteCard,
  getCards,
  getCatalogCards,
  getRewardOverview,
  MemberCardInput,
  RewardOverview,
  setQualificationStatus,
} from "../../MemberApi";
import { CardForm, emptyForm } from "./types";

export function useMemberCards() {
  const [cards, setCards] = useState<Card[]>([]);
  const [catalog, setCatalog] = useState<CatalogCard[]>([]);
  const [form, setForm] = useState<CardForm>(emptyForm());
  const [showCreate, setShowCreate] = useState(false);
  const [detail, setDetail] = useState<Card | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [deletingID, setDeletingID] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [overviews, setOverviews] = useState<Record<string, RewardOverview>>({});
  const [overviewLoadingID, setOverviewLoadingID] = useState("");
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [qualificationSaving, setQualificationSaving] = useState("");

  const load = useCallback(async () => {
    setError("");
    try {
      const [nextCards, nextCatalog] = await Promise.all([
        getCards(),
        getCatalogCards(),
      ]);
      setCards(nextCards);
      setCatalog(nextCatalog);
      setDetail((current) =>
        current ? nextCards.find((card) => card.id === current.id) || null : null,
      );
      setForm((current) => ({
        ...current,
        card_product_id: current.card_product_id || nextCatalog[0]?.id || "",
        card_network_id:
          current.card_network_id || nextCatalog[0]?.networks[0]?.id || "",
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

  async function openDetail(card: Card) {
    setDetail(card);
    if (overviews[card.id]) return;
    setOverviewLoadingID(card.id);
    setError("");
    try {
      const overview = await getRewardOverview(card.id);
      setOverviews((current) => ({ ...current, [card.id]: overview }));
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setOverviewLoadingID("");
    }
  }

  async function refreshOverview(cardID: string) {
    setOverviewLoadingID(cardID);
    try {
      const overview = await getRewardOverview(cardID);
      setOverviews((current) => ({ ...current, [cardID]: overview }));
    } finally {
      setOverviewLoadingID("");
    }
  }

  async function create(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      const input: MemberCardInput = {
        ...form,
        statement_day: form.statement_day ? Number(form.statement_day) : null,
        payment_due_day: form.payment_due_day ? Number(form.payment_due_day) : null,
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

  async function toggleQualification(
    cardID: string,
    planID: string,
    nextValue: boolean,
  ) {
    setQualificationSaving(`${cardID}:${planID}`);
    setError("");
    try {
      await setQualificationStatus(cardID, planID, nextValue);
      await refreshOverview(cardID);
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setQualificationSaving("");
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

  const selectedCard = catalog.find((card) => card.id === form.card_product_id);

  return {
    cards, catalog, deletingID, detail, error, expanded, form, loading,
    overviewLoadingID, overviews, qualificationSaving, saving, selectedCard,
    showCreate, create, openDetail, remove, setDetail, setExpanded, setForm,
    setShowCreate, toggleQualification,
  };
}
