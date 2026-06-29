import { useState } from "react";
import type { Card } from "../../../../models";
import {
  deleteCard,
  getCard,
  getRewardOverview,
  RewardOverview,
  setQualificationStatus,
} from "../../../MemberApi";

type Options = {
  onDeleted: () => Promise<void> | void;
};

export function useCardDetail({ onDeleted }: Options) {
  const [card, setCard] = useState<Card | null>(null);
  const [overview, setOverview] = useState<RewardOverview | undefined>();
  const [loadingCardID, setLoadingCardID] = useState("");
  const [deletingID, setDeletingID] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [qualificationSaving, setQualificationSaving] = useState("");

  async function open(memberCardID: string) {
    setLoadingCardID(memberCardID);
    setError("");
    try {
      const info = await getCard(memberCardID);
      setCard(info);
      setOverview(await getRewardOverview(memberCardID));
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoadingCardID("");
    }
  }

  function close() {
    if (deletingID) return;
    setCard(null);
    setOverview(undefined);
    setError("");
    setExpanded({});
  }

  async function refreshOverview(cardID: string) {
    setLoadingCardID(cardID);
    try {
      setOverview(await getRewardOverview(cardID));
    } finally {
      setLoadingCardID("");
    }
  }

  function toggleGroup(key: string, nextValue: boolean) {
    setExpanded((current) => ({ ...current, [key]: nextValue }));
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
      close();
      await onDeleted();
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setDeletingID(null);
    }
  }

  return {
    card,
    close,
    deletingID,
    error,
    expanded,
    loadingCardID,
    open,
    overview,
    qualificationSaving,
    remove,
    toggleGroup,
    toggleQualification,
  };
}
