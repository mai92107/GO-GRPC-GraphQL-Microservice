import { describe, expect, it } from "vitest";
import type { CatalogCard } from "../models";
import { normalizeCatalogCard } from "./MemberApi";

describe("member catalog API normalization", () => {
  it("normalizes nullable catalog arrays from API payloads", () => {
    const card = normalizeCatalogCard({
      id: "card-1",
      bank_id: "bank-1",
      bank_name: "Bank",
      name: "Card",
      card_image_url: "",
      primary_color: "",
      is_active: true,
      account_tiers: null,
      qualified_type: "",
      selectable_type: "",
      networks: null,
      activities: [
        {
          id: "activity-1",
          name: "Activity",
          start_date: "2026-01-01",
          end_date: "2026-12-31",
          is_active: true,
          source_url: "",
          verified_at: null,
          networks: null,
          benefits: [
            {
              id: "benefit-1",
              reward_unit_id: "cash",
              name: "Benefit",
              display_order: 1,
              effect_type: "ADD_RATE",
              reward_value: "0.01",
              monthly_cap: null,
              layer: "base",
              stack_group: "",
              priority: 1,
              qualified_type: "",
              selectable_type: "",
              action_required: "none",
              action_message: "",
              payment_methods: null,
              category_ids: null,
              merchant_ids: null,
            },
          ],
        },
      ],
    } as unknown as CatalogCard);

    expect(card.account_tiers).toEqual([]);
    expect(card.networks).toEqual([]);
    expect(card.activities[0].networks).toEqual([]);
    expect(card.activities[0].benefits[0].payment_methods).toEqual([]);
    expect(card.activities[0].benefits[0].category_ids).toEqual([]);
    expect(card.activities[0].benefits[0].merchant_ids).toEqual([]);
  });
});
