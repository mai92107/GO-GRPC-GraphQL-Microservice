import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import type { CardNetwork, CatalogCard } from "../../../models";
import {
  Bank,
  createCard,
  deleteCard,
  getBanks,
  getCard,
  getCardNetworks,
  getCards,
  updateCard,
} from "../../AdminApi";
import { filterCards } from "./cardFilters";
import { CardForm, emptyForm } from "./types";

export function useAdminCards() {
  const [items, setItems] = useState<CatalogCard[]>([]);
  const [banks, setBanks] = useState<Bank[]>([]);
  const [networks, setNetworks] = useState<CardNetwork[]>([]);
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
      const [cards, nextBanks, nextNetworks] = await Promise.all([
        getCards(),
        getBanks(),
        getCardNetworks(),
      ]);
      const bankItems = Array.isArray(nextBanks) ? nextBanks : [];
      const networkItems = Array.isArray(nextNetworks) ? nextNetworks : [];
      setItems(Array.isArray(cards) ? cards : []);
      setBanks(bankItems);
      setNetworks(networkItems);
      setForm((current) => ({
        ...current,
        bank_id: current.bank_id || bankItems[0]?.id || "",
        network_ids: current.network_ids.length
          ? current.network_ids
          : networkItems.map((network) => network.id),
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
      await createCard(
        form.bank_id,
        form.name.trim(),
        form.qualified_type,
        form.selectable_type,
        form.network_ids,
      );
      setForm(emptyForm(form.bank_id, networks.map((network) => network.id)));
      setShowCreate(false);
      await load();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function openEditor(card: CatalogCard) {
    setError("");
    setEditLoadingID(card.id);
    try {
      const loadedCard = await getCard(card.id);
      setEditing(loadedCard);
      setEditForm({
        bank_id: loadedCard.bank_id,
        name: loadedCard.name,
        qualified_type: loadedCard.qualified_type || "",
        selectable_type: loadedCard.selectable_type || "",
        network_ids: (loadedCard.networks || []).map((network) => network.id),
      });
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setEditLoadingID("");
    }
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
        editForm.qualified_type,
        editForm.selectable_type,
        editForm.network_ids,
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

  async function remove() {
    if (!editing || !confirm(`確定要刪除「${editing.name}」嗎？此操作無法復原。`)) return;
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
