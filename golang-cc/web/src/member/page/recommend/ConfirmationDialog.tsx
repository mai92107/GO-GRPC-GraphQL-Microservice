import { Dialog, Field } from "../../../components";
import { formatPercent, formatReward } from "../../../format";
import { totalLayerRate, totalLayerReward } from "../../../utils/recommendationText";
import { RewardLayerBreakdown } from "./RewardLayerBreakdown";
import type { Confirmation } from "./types";

type Props = {
  amount: string;
  confirmation: Confirmation;
  merchantName: string;
  recording: boolean;
  onChange: (confirmation: Confirmation) => void;
  onClose: () => void;
  onTransact: () => void;
};

export function ConfirmationDialog({ amount, confirmation, merchantName, recording, onChange, onClose, onTransact }: Props) {
  const option = confirmation.option;

  return (
    <Dialog title="確認刷卡" onClose={onClose}>
      <div className="stack">
        <p>將以 <strong>{confirmation.card.card_name}</strong> 建立在「{merchantName || "其他店家"}」的 NT${amount} 交易。</p>
        {option.payment_methods.length > 1 && (
          <Field label="實際支付方式">
            <select value={confirmation.paymentCode} disabled={recording} onChange={(e) => onChange({ ...confirmation, paymentCode: e.target.value })}>
              {option.payment_methods.map((method) => <option key={method.id} value={method.id}>{method.name}</option>)}
            </select>
          </Field>
        )}
        {option.reminders.map((x) => <div className="notice" key={x}>{x}</div>)}
        <div className="allocation">
          <span>Layer 回饋組成<br /><small className="muted">{option.allocations.length} 項回饋 · 合計 {formatPercent(totalLayerRate(option.layers))}</small></span>
          <strong>{option.allocations[0] ? formatReward(totalLayerReward(option.layers), option.allocations[0].reward_unit) : totalLayerReward(option.layers)}</strong>
        </div>
        {option.layers.length > 0 && <RewardLayerBreakdown layers={option.layers} compact />}
        <button className="button" disabled={recording} onClick={onTransact}>{recording ? "記錄中…" : "確認並記錄交易"}</button>
      </div>
    </Dialog>
  );
}
