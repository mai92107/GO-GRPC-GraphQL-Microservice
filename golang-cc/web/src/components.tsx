import type { ReactNode } from "react";
import { TreePine } from "lucide-react";
import type { CatalogCard, Merchant, PaymentMethod, Unit } from "./api";
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
  return <div className="dialog-backdrop" onMouseDown={onClose}><section className="dialog" role="dialog" aria-modal="true" aria-label={title} onMouseDown={e=>e.stopPropagation()}><header><h2>{title}</h2><button className="icon-button" onClick={onClose} aria-label="關閉">×</button></header>{children}</section></div>;
}

type ActivityView = CatalogCard["activities"][number] & { card_name?: string };

export function ActivityDetails({activities,units,categories,methods,merchants}:{activities:ActivityView[];units:Unit[];categories:{code:string;name:string}[];methods:PaymentMethod[];merchants:Merchant[]}){
  const name=(items:{code:string;name:string}[],code:string)=>items.find(x=>x.code===code)?.name||code;
  return <div className="stack">{activities.map(activity=><section className="activity-detail" key={activity.id}>{activity.card_name&&<span className="activity-card-name">{activity.card_name}</span>}<h3>{activity.name}</h3><p className="muted">{activity.start_date}～{activity.end_date}{activity.source_url&&<> · <a href={activity.source_url} target="_blank" rel="noreferrer">官方來源</a></>}</p>{activity.benefits.map(benefit=>{const unit=units.find(x=>x.id===benefit.reward_unit_id);return <div className="activity-benefit" key={benefit.id}><strong>{formatBenefitTitle(benefit.name,benefit.rate,benefit.payment_methods,methods)}</strong><p>回饋單位：{unit?.name||benefit.reward_unit_id}{benefit.monthly_cap&&unit&&` · 上限 ${formatReward(benefit.monthly_cap,unit)}`}</p><p>{benefit.payment_methods.length===0&&<>支付方式：不限<br/></>}消費類別：{benefit.category_codes.map(x=>name(categories,x)).join("、")}<br/>店家：{benefit.merchant_codes.length?benefit.merchant_codes.map(x=>name(merchants,x)).join("、"):"不限"}<br/>帳戶等級：{benefit.required_account_tiers.length?benefit.required_account_tiers.join("、"):"不限"}{benefit.action_message&&<><br/>提醒：{benefit.action_message}</>}</p></div>})}</section>)}</div>
}
