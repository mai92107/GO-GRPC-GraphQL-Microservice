import { useEffect, type ReactNode } from "react";
import { Search, TreePine, X } from "lucide-react";
import type { CatalogCard, Merchant, PaymentMethod, Unit } from "./models";
import { formatBenefitTitle, formatReward } from "./format";

export function Logo({ compact = false }: { compact?: boolean }) {
  return <div className="logo" aria-label="樹卡">
    <span className="logo-mark"><TreePine size={compact ? 20 : 28} strokeWidth={1.8}/></span>
    {!compact && <span><strong>樹卡</strong><small>每次消費，都選得更好</small></span>}
  </div>;
}

export function Field({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return <label className="field"><span>{label}</span>{children}{hint && <small>{hint}</small>}</label>;
}

export function Empty({ title, text }: { title: string; text: string }) {
  return <div className="empty"><TreePine size={34}/><strong>{title}</strong><p>{text}</p></div>;
}

export function Dialog({ title, children, onClose }: { title: string; children: ReactNode; onClose: () => void }) {
  useEffect(() => {
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, []);

  return <div className="dialog-backdrop" onMouseDown={onClose}><section className="dialog" role="dialog" aria-modal="true" aria-label={title} onMouseDown={e=>e.stopPropagation()}><header><div><span className="page-eyebrow">ADMIN ACTION</span><h2>{title}</h2></div><button className="icon-button" onClick={onClose} aria-label="關閉"><X size={18}/></button></header>{children}</section></div>;
}

export function StatusBadge({
  active,
  activeText = "啟用中",
  inactiveText = "已停用",
}: {
  active: boolean;
  activeText?: string;
  inactiveText?: string;
}) {
  return (
    <span className={`status-badge ${active ? "is-active" : "is-inactive"}`}>
      <span aria-hidden="true" />
      {active ? activeText : inactiveText}
    </span>
  );
}

export function SearchField({
  value,
  onChange,
  placeholder = "搜尋",
}: {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}) {
  return (
    <label className="search-field">
      <Search size={17} aria-hidden="true" />
      <span className="sr-only">{placeholder}</span>
      <input
        type="search"
        value={value}
        placeholder={placeholder}
        onChange={(event) => onChange(event.target.value)}
      />
    </label>
  );
}

type ActivityView = CatalogCard["activities"][number] & { card_name?: string };

export function ActivityDetails({activities,units,categories,methods,merchants}:{activities:ActivityView[];units:Unit[];categories:{id:string;name:string}[];methods:PaymentMethod[];merchants:Merchant[]}){
  const name=(items:{id:string;name:string}[],id:string)=>items.find(x=>x.id===id)?.name||id;
  const merchantName=(id:string)=>merchants.find(x=>x.id===id)?.name||id;
  return <div className="stack">{activities.map(activity=><section className="activity-detail" key={activity.id}>
    {activity.card_name&&<span className="activity-card-name">{activity.card_name}</span>}
    <h3>{activity.name}</h3>
    <p className="muted">{activity.start_date}～{activity.end_date}{activity.source_url&&<> · <a href={activity.source_url} target="_blank" rel="noreferrer">官方來源</a></>}</p>
    {activity.benefits.map(benefit=>{
      const unit=units.find(x=>x.id===benefit.reward_unit_id);
      return <div className="activity-benefit" key={benefit.id}>
        <strong>{formatBenefitTitle(benefit.name,benefit.reward_value,benefit.payment_methods,methods,benefit.effect_type)}</strong>
        <p>回饋單位：{unit?.name||benefit.reward_unit_id}{benefit.monthly_cap&&unit&&` · 上限 ${formatReward(benefit.monthly_cap,unit)}`}</p>
        <p>{benefit.payment_methods.length===0&&<>支付方式：不限<br/></>}消費類別：{benefit.category_ids.map(x=>name(categories,x)).join("、")}<br/>店家：{benefit.merchant_ids.length?benefit.merchant_ids.map(merchantName).join("、"):"不限"}<br/>帳戶等級：{benefit.qualified_type||"不限"}{benefit.action_message&&<><br/>提醒：{benefit.action_message}</>}</p>
      </div>
    })}
  </section>)}</div>
}
