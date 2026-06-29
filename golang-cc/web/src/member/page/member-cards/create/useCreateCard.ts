import { FormEvent, useMemo, useState } from "react";
import type { CatalogCard } from "../../../../models";
import { createCard, MemberCardInput } from "../../../MemberApi";
import { CardForm, emptyForm } from "../../cards/types";

type Options = {
  catalogCards: CatalogCard[];
  onCreated: () => Promise<void> | void;
};

export function useCreateCard({ catalogCards, onCreated }: Options) {
  const [form, setForm] = useState<CardForm>(emptyForm());
  const [showCreate, setShowCreate] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const selectedCard = useMemo(
    () => catalogCards.find((card) => card.id === form.card_id),
    [catalogCards, form.card_id],
  );

  function formForCard(cardID = catalogCards[0]?.id || "") {
    const card = catalogCards.find((item) => item.id === cardID);
    return {
      ...emptyForm(cardID),
      card_network_id: card?.networks[0] || "",
    };
  }

  function open() {
    setError("");
    setForm((current) => (current.card_id ? current : formForCard()));
    setShowCreate(true);
  }

  function close() {
    if (saving) return;
    setShowCreate(false);
  }

  async function create(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      const input: MemberCardInput = {
        ...form,
        statement_day: form.statement_day ? Number(form.statement_day) : null,
        payment_due_day: form.payment_due_day
          ? Number(form.payment_due_day)
          : null,
      };
      await createCard(input);
      setShowCreate(false);
      setForm(formForCard());
      await onCreated();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setSaving(false);
    }
  }

  return {
    close,
    create,
    error,
    form,
    open,
    saving,
    selectedCard,
    formForCard,
    setForm,
    showCreate,
  };
}
