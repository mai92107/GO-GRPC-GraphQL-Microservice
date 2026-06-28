import { CreditCard, Landmark, Layers3, ReceiptText, Store } from "lucide-react";
import { formatPercent, formatScore } from "../../../format";
import type { ExcludedBenefit, RewardLayer } from "../../../models";
import { excludeReasonLabel } from "../../../utils/recommendationText";

export function RewardLayerBreakdown({ layers, compact = false }: { layers: RewardLayer[]; compact?: boolean }) {
  return (
    <section className={`reward-layers ${compact ? "compact" : ""}`}>
      {layers.map((layer) => (
        <div className="reward-layer" key={layer.layer}>
          <div className="reward-layer-head">
            <span className="reward-layer-icon"><LayerIcon layer={layer.layer} /></span>
            <div><h5>{layer.layer}</h5></div>
            <strong>{formatPercent(layer.reward_rate)}<small>NT${formatScore(layer.reward_amount)}</small></strong>
          </div>
          <div className="reward-layer-benefits">
            {layer.benefits.map((benefit) => (
              <div className="reward-layer-benefit" key={benefit.benefit_id}>
                <span>
                  {benefit.name}
                  <small>{benefit.activity_name} · {benefit.stack_group || "未分組"}{benefit.remaining_before && ` · cap 剩餘 NT$${formatScore(benefit.remaining_before)}`}</small>
                </span>
                <strong>{benefit.effect_type === "ADD_CASH" || benefit.effect_type === "DISCOUNT" ? `NT$${formatScore(benefit.reward_amount)}` : formatPercent(benefit.reward_rate)}</strong>
              </div>
            ))}
          </div>
        </div>
      ))}
    </section>
  );
}

function LayerIcon({ layer }: { layer: number }) {
  if (layer === 1) return <CreditCard size={18} />;
  if (layer === 2) return <Landmark size={18} />;
  if (layer === 3) return <Store size={18} />;
  if (layer === 4) return <ReceiptText size={18} />;
  return <Layers3 size={18} />;
}

export function ExcludedBenefitsDisclosure({ items }: { items: ExcludedBenefit[] }) {
  if (items.length === 0) return null;
  return (
    <details className="excluded-benefits">
      <summary>未採用活動 {items.length} 項</summary>
      <div>
        {items.map((item) => (
          <p key={`${item.reason}-${item.benefit_id}`}>
            <strong>{item.name}</strong>
            <small>{excludeReasonLabel(item.reason)}{!item.layer && item.stack_group && ` · ${item.stack_group}`}</small>
          </p>
        ))}
      </div>
    </details>
  );
}
