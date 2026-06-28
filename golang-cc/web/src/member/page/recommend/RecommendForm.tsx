import { Sparkles } from "lucide-react";
import type React from "react";
import { Field } from "../../../components";
import type { Merchant } from "../../../models";
import type { Category } from "../../MemberApi";

type Props = {
  amount: string; categories: Category[]; category: string; date: string;
  error: string; loadingCategories: boolean; loadingMerchants: boolean;
  merchantId: string; merchants: Merchant[]; otherMerchant: string;
  recommending: boolean; onAmount: (value: string) => void;
  onCategory: (value: string) => void; onDate: (value: string) => void;
  onMerchant: (value: string) => void; onOtherMerchant: (value: string) => void;
  onSubmit: (event: React.FormEvent) => void;
};

export function RecommendForm(props: Props) {
  return (
    <form className="panel recommend-form stack" onSubmit={props.onSubmit}>
      <Field label="消費金額（NT$）"><input type="number" min="0.01" step="0.01" value={props.amount} onChange={(e) => props.onAmount(e.target.value)} required /></Field>
      <Field label="消費類別"><select required value={props.category} disabled={props.loadingCategories || props.recommending} onChange={(e) => props.onCategory(e.target.value)}>{props.categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select></Field>
      <Field label="消費店家"><select value={props.merchantId} disabled={props.loadingMerchants || props.recommending} onChange={(e) => props.onMerchant(e.target.value)}><option value="">其他</option>{props.merchants.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}</select></Field>
      {!props.merchantId && <Field label="其他店家名稱（選填）"><input value={props.otherMerchant} onChange={(e) => props.onOtherMerchant(e.target.value)} placeholder="可留空" /></Field>}
      <Field label="消費日期"><input type="date" value={props.date} onChange={(e) => props.onDate(e.target.value)} required /></Field>
      {props.error && <div className="error">{props.error}</div>}
      <button className="button" disabled={!props.category || props.recommending || props.loadingCategories}>
        <Sparkles size={17} />{props.recommending ? "計算中…" : "取得推薦"}
      </button>
    </form>
  );
}
