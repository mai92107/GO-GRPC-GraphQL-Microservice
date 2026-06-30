import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import type { CatalogCard } from "../../../models";
import {
  type Bank,
  createCardFromForm,
  deleteCardByID,
  loadCardForEdit,
  loadCardPageData,
  updateCardFromForm,
} from "./cardRequests";
import {
  cardFormFromCatalogCard,
  requestMessage,
  selectedNetworksOrDefault,
} from "./cardForm";
import { filterCards } from "./cardFilters";
import { CardForm, emptyForm } from "./types";

export function useAdminCards() {
  const [items, setItems] = useState<CatalogCard[]>([]);
  const [banks, setBanks] = useState<Bank[]>([]);
  const [networks, setNetworks] = useState<string[]>([]);
  const [form, setForm] = useState<CardForm>(emptyForm());
  const [editing, setEditing] = useState<CatalogCard | null>(null);
  const [editForm, setEditForm] = useState<CardForm>(emptyForm());
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editLoadingID, setEditLoadingID] = useState("");
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [bankFilter, setBankFilter] = useState("all");
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    try {
      const data = await loadCardPageData();
      setItems(data.cards);
      setBanks(data.banks);
      setNetworks(data.networks);
      setForm((current) => ({
        ...current,
        bank_id: current.bank_id || data.banks[0]?.id || "",
        networks: selectedNetworksOrDefault(current.networks, data.networks),
      }));
    } catch (requestError) {
      setError(requestMessage(requestError));
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
      await createCardFromForm(form);
      setForm(emptyForm(form.bank_id, networks));
      setShowCreate(false);
      await load();
    } catch (requestError) {
      setError(requestMessage(requestError));
    } finally {
      setCreating(false);
    }
  }

  async function openEditor(card: CatalogCard) {
    setError("");
    setEditLoadingID(card.id);
    try {
      const loadedCard = await loadCardForEdit(card.id);
      setEditing(loadedCard);
      setEditForm(cardFormFromCatalogCard(loadedCard));
    } catch (requestError) {
      setError(requestMessage(requestError));
    } finally {
      setEditLoadingID("");
    }
  }

  async function persist(active: boolean) {
    if (!editing) return;
    setSaving(true);
    setError("");
    try {
      await updateCardFromForm(editing.id, editForm, active);
      setEditing(null);
      await load();
    } catch (requestError) {
      setError(requestMessage(requestError));
    } finally {
      setSaving(false);
    }
  }

  async function remove() {
    if (!editing || !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)) return;
    setSaving(true);
    setError("");
    try {
      await deleteCardByID(editing.id);
      setEditing(null);
      await load();
    } catch (requestError) {
      setError(requestMessage(requestError));
    } finally {
      setSaving(false);
    }
  }

  const filtered = useMemo(
    () => filterCards(items, query, bankFilter),
    [bankFilter, items, query],
  );

  return {
    banks, bankFilter, creating, editForm, editLoadingID, editing, error,
    filtered, form, items, loading, networks, query, saving, showCreate,
    create, openEditor, persist, remove, setBankFilter, setEditForm,
    setEditing, setForm, setQuery, setShowCreate,
  };
}
