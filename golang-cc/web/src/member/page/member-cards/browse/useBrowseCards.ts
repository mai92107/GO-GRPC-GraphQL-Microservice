import { useCallback, useEffect, useState } from "react";
import type { Card, CatalogCard } from "../../../../models";
import { getCards, getCatalogCards } from "../../../MemberApi";

export function useBrowseCards() {
  const [myCards, setMyCards] = useState<Card[]>([]);
  const [catalogCards, setCatalogCards] = useState<CatalogCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const reload = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [nextMyCards, nextCatalogCards] = await Promise.all([
        getCards(),
        getCatalogCards(),
      ]);
      setMyCards(nextMyCards);
      setCatalogCards(nextCatalogCards);
    } catch (requestError) {
      setError((requestError as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reload();
  }, [reload]);

  return { catalogCards, error, loading, myCards, reload };
}
