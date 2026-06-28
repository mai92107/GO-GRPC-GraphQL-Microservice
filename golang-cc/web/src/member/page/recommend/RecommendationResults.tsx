import { ArrowRight, Trophy } from "lucide-react";
import { Empty } from "../../../components";
import { formatPercent, formatScore } from "../../../format";
import type { Recommendation } from "../../../models";
import { totalLayerRate } from "../../../utils/recommendationText";
import type { Confirmation } from "./types";
import { ExcludedBenefitsDisclosure, RewardLayerBreakdown } from "./RewardLayerBreakdown";

type Props = {
  hasSearched: boolean;
  results: Recommendation[];
  onConfirm: (confirmation: Confirmation) => void;
};

export function RecommendationResults({ hasSearched, results, onConfirm }: Props) {
  if (results.length === 0) {
    return (
      <section className="results" aria-live="polite">
        <div className="panel">
          <Empty
            title={hasSearched ? "沒有符合條件的推薦" : "準備好開始推薦"}
            text={hasSearched ? "請調整消費條件，或確認卡片與回饋方案目前為啟用狀態。" : "填入左側消費資訊，我們會計算每張卡與支付方式的有效回饋。"}
          />
        </div>
      </section>
    );
  }

  return (
    <section className="results" aria-live="polite">
      {results.map((card) => (
        <article className={`recommend-card ${card.rank === 1 ? "best" : ""}`} key={`${card.card_id}-${card.total_score}`}>
          <div className="recommend-card-head">
            <span className="rank">{card.rank === 1 && <Trophy size={13} />}第 {card.rank} 名</span>
            <span className="recommend-score">偏好分數 {formatScore(card.total_score)}</span>
          </div>
          <h3>{card.card_name}</h3>
          {card.payment_options.map((option) => (
            <div className="allocations" key={option.payment_method_id}>
              <div className="payment-option-head"><div><small>建議支付方式</small><h4>{option.payment_method_name}</h4></div><strong>{formatPercent(totalLayerRate(option.layers))} 回饋</strong></div>
              {option.reminders.map((x) => <div className="notice" key={x}>{x}</div>)}
              {option.layers.length ? (
                <>
                  <RewardLayerBreakdown layers={option.layers} />
                  <ExcludedBenefitsDisclosure items={option.excluded_benefits} />
                  <button className="button secondary" onClick={() => onConfirm({ card, option, paymentCode: option.payment_methods[0]?.id || option.payment_method_id })}>選擇這個方案 <ArrowRight size={16} /></button>
                </>
              ) : <div className="notice">此推薦缺少 Layer 明細，請先在後台建立新架構回饋規則。</div>}
            </div>
          ))}
        </article>
      ))}
    </section>
  );
}
