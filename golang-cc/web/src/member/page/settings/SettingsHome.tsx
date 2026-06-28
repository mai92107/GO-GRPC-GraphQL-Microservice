import { Check, ChevronRight, Info, Sparkles, WalletCards } from "lucide-react";

type Props = {
  error: string;
  highPriorityCount: number;
  notice: string;
  selectedPaymentCount: number;
  onOpenPayments: () => void;
  onOpenPreferences: () => void;
};

export function SettingsHome(props: Props) {
  return (
    <div className="settings-home">
      <div className="settings-home-head"><span className="page-eyebrow">SETTINGS</span><h1>設定</h1><p>管理個人化推薦與付款選項。</p></div>
      {props.error && <div className="error preference-message">{props.error}</div>}
      {props.notice && <div className="notice preference-message"><Check size={16} />{props.notice}</div>}
      <section className="settings-home-section">
        <span className="settings-section-label">推薦</span>
        <div className="settings-list-card">
          <button type="button" className="settings-list-row" onClick={props.onOpenPreferences}>
            <span className="settings-row-icon"><Sparkles size={19} /></span>
            <span className="settings-row-copy"><strong>回饋偏好</strong><small>調整現金、點數與里程的重視程度</small></span>
            <span className="settings-row-value">{props.highPriorityCount ? `${props.highPriorityCount} 項優先` : "均衡設定"}</span>
            <ChevronRight size={20} />
          </button>
          <button type="button" className="settings-list-row" onClick={props.onOpenPayments}>
            <span className="settings-row-icon"><WalletCards size={19} /></span>
            <span className="settings-row-copy"><strong>支付方式</strong><small>選擇你實際可以使用的付款方式</small></span>
            <span className="settings-row-value">{props.selectedPaymentCount} 項</span>
            <ChevronRight size={20} />
          </button>
        </div>
      </section>
      <section className="settings-home-section">
        <span className="settings-section-label">關於</span>
        <div className="settings-list-card">
          <div className="settings-list-row static">
            <span className="settings-row-icon"><Info size={19} /></span>
            <span className="settings-row-copy"><strong>樹卡</strong><small>依消費情境找出最適合的信用卡</small></span>
            <span className="settings-row-value">1.0</span>
          </div>
        </div>
      </section>
    </div>
  );
}
