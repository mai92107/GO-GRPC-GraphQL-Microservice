import { CreditCard, ChevronDown } from "lucide-react";
import type { PaymentMethod } from "../../../models";
import { PaymentPill } from "./PaymentPill";
import { SettingsDetail } from "./SettingsDetail";

type Props = {
  dirty: boolean; error: string; notice: string; otherPayments: PaymentMethod[];
  saving: boolean; selectedPayments: PaymentMethod[]; showMorePayments: boolean;
  onBack: () => void; onSave: () => void; onTogglePayment: (id: string) => void;
  onToggleMore: () => void;
};

export function PaymentsView(props: Props) {
  return (
    <SettingsDetail title="編輯支付方式" dirty={props.dirty} saving={props.saving} notice={props.notice} error={props.error} onBack={props.onBack} onSave={props.onSave}>
      <div className="settings-detail-intro"><h2>你有哪些支付方式？</h2><p>點一下即可選取或取消，推薦只會使用你已選擇的方式。</p></div>
      <section className="payment-pill-section">
        <span className="settings-section-label">常用</span>
        <div className="payment-pill-grid">
          {props.selectedPayments.map((method) => <PaymentPill key={method.id} method={method} selected onClick={() => props.onTogglePayment(method.id)} />)}
          {props.selectedPayments.length === 0 && <p className="payment-empty-copy">尚未選擇常用支付方式，可從下方加入。</p>}
        </div>
      </section>
      <section className="payment-pill-section">
        <button type="button" className="payment-more-toggle" aria-expanded={props.showMorePayments} onClick={props.onToggleMore}>
          <span>更多支付方式（{props.otherPayments.length}）</span>
          <ChevronDown className={props.showMorePayments ? "rotated" : ""} />
        </button>
        {props.showMorePayments && (
          <div className="payment-pill-grid more-payments">
            {props.otherPayments.map((method) => <PaymentPill key={method.id} method={method} selected={false} onClick={() => props.onTogglePayment(method.id)} />)}
          </div>
        )}
      </section>
      <div className="fixed-payment-note"><CreditCard size={17} />實體信用卡與線上刷卡固定可用，不需另外設定。</div>
    </SettingsDetail>
  );
}
