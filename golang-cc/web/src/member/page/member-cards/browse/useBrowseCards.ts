import { useCallback, useEffect, useState } from "react";
import type { Card } from "../../../../models";
import { getCards } from "../../../MemberApi";

export function useBrowseCards() {
  const [myCards, setMyCards] = useState<Card[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const reload = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextMyCards = await getCards();
      setMyCards(nextMyCards);
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reload();
  }, [reload]);

  return { error, loading, myCards, reload };
}
