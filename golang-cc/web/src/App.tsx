import { useEffect, useRef, useState, type FormEvent } from "react";
import { DndContext, DragOverlay, KeyboardSensor, PointerSensor, useDraggable, useDroppable, useSensor, useSensors, type DragEndEvent, type DragStartEvent } from "@dnd-kit/core";
import { CreditCard, GripVertical, History, Leaf, LogOut, Plus, Settings, Sparkles, Trash2 } from "lucide-react";
import { api, mutate, setCSRF, type Card, type Merchant, type PaymentOption, type Recommendation, type User } from "./api";
import { ActivityDetails, Dialog, Empty, Field, Logo } from "./components";
import { formatPercent, formatReward, formatScore } from "./format";
import { groupPreferences, movePreference, tierLabel, toPreferenceWrites, type PreferenceTier, type RewardPreference } from "./preferences";
import AdminApp from "./AdminApp";

type Page = "recommend" | "transactions" | "cards" | "settings";
const today = new Intl.DateTimeFormat("en-CA",{timeZone:"Asia/Taipei",year:"numeric",month:"2-digit",day:"2-digit"}).format(new Date());

export default function App() {
  const [user,setUser]=useState<User|null>(null),[loading,setLoading]=useState(true),[page,setPage]=useState<Page>("recommend");
  useEffect(()=>{api<{user:User;csrf_token:string}>("/public/auth/me").then(x=>{setUser(x.user);setCSRF(x.csrf_token)}).catch(()=>{}).finally(()=>setLoading(false))},[]);
  if(loading)return <div className="auth-page"><Logo/></div>;
  if(!user)return <Auth onLogin={(u,c)=>{setCSRF(c);setUser(u)}}/>;
  if(user.role==="admin")return <AdminApp user={user} onLogout={()=>mutate("/public/auth/logout","POST").finally(()=>setUser(null))}/>;
  return <div className="shell"><header className="topbar"><Logo/><div className="user"><span>{user.display_name}<br/>{user.email}</span><span className="avatar">{user.display_name.slice(0,1)}</span><button className="icon-button" aria-label="登出" onClick={()=>mutate("/public/auth/logout","POST").finally(()=>setUser(null))}><LogOut size={17}/></button></div></header>
    <main className="content">{page==="recommend"?<Recommend/>:page==="transactions"?<Transactions/>:page==="cards"?<Cards/>:<MemberSettings/>}</main>
    <nav className="bottom-nav" aria-label="主要導覽">
      <Nav active={page==="recommend"} icon={<Sparkles/>} label="推薦" onClick={()=>setPage("recommend")}/>
      <Nav active={page==="transactions"} icon={<History/>} label="交易" onClick={()=>setPage("transactions")}/>
      <Nav active={page==="cards"} icon={<CreditCard/>} label="卡片" onClick={()=>setPage("cards")}/>
      <Nav active={page==="settings"} icon={<Settings/>} label="設定" onClick={()=>setPage("settings")}/>
    </nav></div>;
}

function Nav({active,icon,label,onClick}:{active:boolean;icon:React.ReactNode;label:string;onClick:()=>void}){return <button className={active?"active":""} onClick={onClick} aria-current={active?"page":undefined}>{icon}{label}</button>}

function Auth({onLogin}:{onLogin:(user:User,csrf:string)=>void}){
  const params=new URLSearchParams(location.search), token=params.get("token")||"", path=location.pathname;
  const mode=path.includes("accept-invitation")?"accept":path.includes("reset-password")?"reset":"login";
  const [email,setEmail]=useState(""),[password,setPassword]=useState(""),[name,setName]=useState(""),[error,setError]=useState(""),[notice,setNotice]=useState("");
  async function submit(e:FormEvent){e.preventDefault();setError("");try{
    if(mode==="accept"){await mutate("/public/auth/accept-invitation","POST",{token,display_name:name,password});setNotice("帳號已啟用，現在可以登入");history.replaceState(null,"","/")}
    else if(mode==="reset"){await mutate("/public/auth/reset-password","POST",{token,password});setNotice("密碼已更新，現在可以登入");history.replaceState(null,"","/")}
    else {const data=await mutate<{user:User;csrf_token:string}>("/public/auth/login","POST",{email,password});onLogin(data.user,data.csrf_token)}
  }catch(e){setError((e as Error).message)}}
  async function reset(){setError("");await mutate("/public/auth/request-password-reset","POST",{email});setNotice("若帳號存在，重設連結已寄至信箱")}
  return <main className="auth-page"><section className="auth-card"><Logo/><h1>{mode==="accept"?"啟用帳號":mode==="reset"?"設定新密碼":"歡迎回來"}</h1><p>{mode==="login"?"登入後，幾秒內找到最適合這次消費的卡。":"安全連結僅能使用一次。"}</p>{error&&<div className="error">{error}</div>}{notice&&<div className="notice">{notice}</div>}<form className="stack" onSubmit={submit}>
    {mode==="accept"&&<Field label="顯示名稱"><input value={name} onChange={e=>setName(e.target.value)} required/></Field>}
    {mode==="login"&&<Field label="Email"><input type="email" value={email} onChange={e=>setEmail(e.target.value)} required autoComplete="email"/></Field>}
    <Field label={mode==="login"?"密碼":"新密碼"} hint={mode!=="login"?"至少 12 個字元":undefined}><input type="password" value={password} onChange={e=>setPassword(e.target.value)} required minLength={12} autoComplete={mode==="login"?"current-password":"new-password"}/></Field>
    <button className="button" type="submit">{mode==="login"?"登入":mode==="accept"?"啟用帳號":"更新密碼"}</button>
    {mode==="login"&&<button className="link-button" type="button" onClick={reset}>寄送密碼重設連結</button>}
  </form></section></main>
}

function Recommend(){
  const [amount,setAmount]=useState("200"),[category,setCategory]=useState(""),[merchantCode,setMerchantCode]=useState(""),[otherMerchant,setOtherMerchant]=useState(""),[merchants,setMerchants]=useState<Merchant[]>([]),[availableCategories,setAvailableCategories]=useState<{code:string;name:string}[]>([]),[date,setDate]=useState(today),[results,setResults]=useState<Recommendation[]>([]),[error,setError]=useState(""),[confirm,setConfirm]=useState<{card:Recommendation;option:PaymentOption;paymentCode:string}|null>(null),[notice,setNotice]=useState("");
  useEffect(()=>{api<{code:string;name:string}[]>("/member/categories").then(categories=>{setAvailableCategories(categories);if(categories[0])setCategory(categories[0].code)}).catch(e=>setError((e as Error).message))},[]);
  useEffect(()=>{if(!category)return;api<Merchant[]>(`/member/merchants?category_code=${encodeURIComponent(category)}`).then(stores=>{setMerchants(stores);setMerchantCode("");setOtherMerchant("")}).catch(e=>setError((e as Error).message))},[category]);
  const merchantName=merchantCode?merchants.find(x=>x.code===merchantCode)?.name||"":otherMerchant;
  async function run(e?:FormEvent){e?.preventDefault();setError("");try{const data=await mutate<{recommendations:Recommendation[]}>("/member/recommendations","POST",{amount_minor:Math.round(Number(amount)*100),category_code:category,merchant_code:merchantCode,merchant_name:merchantName,date});setResults(data.recommendations)}catch(e){setError((e as Error).message)}}
  async function transact(card:Recommendation,option:PaymentOption,paymentCode:string){try{const data=await mutate<{recommendation_changed:boolean}>("/member/transactions","POST",{card_id:card.card_id,amount_minor:Math.round(Number(amount)*100),category_code:category,merchant_code:merchantCode,merchant_name:merchantName,payment_method_code:paymentCode,transaction_date:date,recommendation_summary:option.allocations.map(a=>({rule_id:a.rule_id,allocated_reward:a.allocated_reward}))});setNotice(data.recommendation_changed?"實際回饋因 cap 或規則更新而改變":"交易已記錄，cap 已更新");setConfirm(null);await run()}catch(e){setError((e as Error).message)}}
  return <><div className="page-head"><div><h1>這次刷哪張？</h1><p>輸入消費情境，系統會連同最佳支付方式一起推薦。</p></div></div>{notice&&<div className="notice">{notice}</div>}<div className="grid"><form className="panel recommend-form stack" onSubmit={run}><Field label="消費金額（NT$）"><input type="number" min="0.01" step="0.01" value={amount} onChange={e=>setAmount(e.target.value)} required/></Field><Field label="消費類別"><select required value={category} onChange={e=>setCategory(e.target.value)}>{availableCategories.map(c=><option key={c.code} value={c.code}>{c.name}</option>)}</select></Field><Field label="消費店家"><select value={merchantCode} onChange={e=>setMerchantCode(e.target.value)}><option value="">其他</option>{merchants.map(x=><option key={x.code} value={x.code}>{x.name}</option>)}</select></Field>{!merchantCode&&<Field label="其他店家名稱（選填）"><input value={otherMerchant} onChange={e=>setOtherMerchant(e.target.value)} placeholder="可留空"/></Field>}<Field label="消費日期"><input type="date" value={date} onChange={e=>setDate(e.target.value)} required/></Field>{error&&<div className="error">{error}</div>}<button className="button"><Sparkles size={17}/>取得推薦</button></form><section className="results" aria-live="polite">{results.length===0?<div className="panel"><Empty title="準備好開始推薦" text="填入左側消費資訊，我們會計算每張卡與支付方式的有效回饋。"/></div>:results.map(card=><article className={`recommend-card ${card.rank===1?"best":""}`} key={`${card.card_id}-${card.total_score}`}><span className="rank">第 {card.rank} 名</span><h3>{card.card_name}(得分:{formatScore(card.total_score)})</h3>{card.payment_options.map(option=><div className="allocations" key={option.payment_method_code}><h4>{option.payment_method_name} · 預估回饋({formatPercent(String(option.allocations.reduce((sum,a)=>sum+Number(a.rate),0)))})</h4>{option.reminders.map(x=><div className="notice" key={x}>{x}</div>)}{option.allocations.map(a=><div className="allocation" key={a.rule_id}><span>{a.rule_name}<br/><small className="muted">{a.activity_name} · 累計層 {a.stack_group}{a.remaining_before&&` · cap 剩餘 ${formatReward(a.remaining_before,a.reward_unit)}`}</small></span><strong>{formatReward(a.allocated_reward,a.reward_unit)}</strong></div>)}<button className="button secondary" onClick={()=>setConfirm({card,option,paymentCode:option.payment_methods[0]?.code||option.payment_method_code})}>使用以上支付方式</button></div>)}</article>)}</section></div>{confirm&&<Dialog title="確認刷卡" onClose={()=>setConfirm(null)}><div className="stack"><p>將以 <strong>{confirm.card.card_name}</strong> 建立在「{merchantName||"其他店家"}」的 NT${amount} 交易。</p>{confirm.option.payment_methods.length>1&&<Field label="實際支付方式"><select value={confirm.paymentCode} onChange={e=>setConfirm({...confirm,paymentCode:e.target.value})}>{confirm.option.payment_methods.map(method=><option key={method.code} value={method.code}>{method.name}</option>)}</select></Field>}{confirm.option.reminders.map(x=><div className="notice" key={x}>{x}</div>)}<button className="button" onClick={()=>transact(confirm.card,confirm.option,confirm.paymentCode)}>確認並記錄交易</button></div></Dialog>}</>
}

function Cards(){
  const [cards,setCards]=useState<Card[]>([]),[catalog,setCatalog]=useState<import("./api").CatalogCard[]>([]),[open,setOpen]=useState(false),[detail,setDetail]=useState<import("./api").CatalogCard|null>(null),[units,setUnits]=useState<import("./api").Unit[]>([]),[categories,setCategories]=useState<{code:string;name:string}[]>([]),[methods,setMethods]=useState<import("./api").PaymentMethod[]>([]),[merchants,setMerchants]=useState<Merchant[]>([]),[form,setForm]=useState({card_product_id:"",nickname:"",last_four:"",statement_day:"",payment_due_day:"",account_tier:"",is_active:true});
  const load=()=>Promise.all([api<Card[]>("/member/cards"),api<import("./api").CatalogCard[]>("/member/catalog/cards"),api<import("./api").Unit[]>("/member/reward-units"),api<{code:string;name:string}[]>("/member/categories"),api<import("./api").PaymentMethod[]>("/member/payment-methods")]).then(([c,k,u,cs,p])=>{setCards(c);setCatalog(k);setUnits(u);setCategories(cs);setMethods(p);if(!form.card_product_id&&k[0])setForm(v=>({...v,card_product_id:k[0].id}));return Promise.all(cs.map(x=>api<Merchant[]>(`/member/merchants?category_code=${encodeURIComponent(x.code)}`))).then(groups=>setMerchants(groups.flat().filter((x,i,a)=>a.findIndex(y=>y.code===x.code)===i)))});useEffect(()=>{load()},[]);
  async function create(e:FormEvent){e.preventDefault();await mutate("/member/cards","POST",{...form,statement_day:form.statement_day?Number(form.statement_day):null,payment_due_day:form.payment_due_day?Number(form.payment_due_day):null});setOpen(false);load()}
  async function remove(id:string){if(confirm("確定刪除這張卡？已有交易的卡片將無法刪除。")){try{await mutate(`/member/cards/${id}`,"DELETE");load()}catch(e){alert((e as Error).message)}}}
  const selected=catalog.find(c=>c.id===form.card_product_id);
  const currentActivities=(c:import("./api").CatalogCard)=>c.activities.filter(a=>a.is_active&&a.start_date<=today&&a.end_date>=today);
  return <>
    <div className="page-head"><div><h1>我的卡片夾</h1><p>從卡片目錄加入目前持有的卡片。</p></div><button className="button" disabled={!catalog.length} onClick={()=>setOpen(true)}><Plus size={17}/>加入卡片</button></div>
    <div className="card-list">{cards.map(c=><article className={`credit-card ${c.is_active?"":"inactive"}`} key={c.id}><header><span>{c.issuer}</span><Leaf/></header><div><h3>{c.name}</h3><small>{c.is_active?"啟用":"停用"}{c.account_tier&&` · ${c.account_tier}`} · {c.last_four?`•••• ${c.last_four}`:"未設定末四碼"}</small></div><button className="button danger" onClick={()=>remove(c.id)}><Trash2 size={15}/>移除</button></article>)}</div>
    <section className="panel" style={{marginTop:18}}><div className="section-title"><h2>所有卡片與活動</h2></div><div className="list">{catalog.map(c=><button type="button" className="list-row clickable-row" key={c.id} onClick={()=>setDetail(c)}><div><h3>{c.bank_name} · {c.name}</h3><p>{currentActivities(c).length?`${currentActivities(c).length} 個當前活動`:"目前沒有有效活動"}</p></div></button>)}</div></section>
    {detail&&<Dialog title={`${detail.bank_name} · ${detail.name}`} onClose={()=>setDetail(null)}>{currentActivities(detail).length>0?<ActivityDetails activities={currentActivities(detail)} units={units} categories={categories} methods={methods} merchants={merchants}/>:<Empty title="目前沒有有效活動" text="今天沒有啟用且在活動期間內的活動。"/>}</Dialog>}
    {open&&<Dialog title="加入卡片夾" onClose={()=>setOpen(false)}><form className="stack" onSubmit={create}><Field label="銀行卡片"><select value={form.card_product_id} onChange={e=>setForm({...form,card_product_id:e.target.value,account_tier:""})}>{catalog.map(c=><option value={c.id} key={c.id}>{c.bank_name} · {c.name}</option>)}</select></Field>{selected&&selected.account_tiers.length>0&&<Field label="目前帳戶等級"><select required value={form.account_tier} onChange={e=>setForm({...form,account_tier:e.target.value})}><option value="">請選擇</option>{selected.account_tiers.map(x=><option key={x}>{x}</option>)}</select></Field>}<Field label="自訂暱稱"><input value={form.nickname} onChange={e=>setForm({...form,nickname:e.target.value})}/></Field><Field label="卡號末四碼"><input pattern="[0-9]{4}" value={form.last_four} onChange={e=>setForm({...form,last_four:e.target.value})}/></Field><div className="two-col"><Field label="結帳日"><input type="number" min="1" max="31" value={form.statement_day} onChange={e=>setForm({...form,statement_day:e.target.value})}/></Field><Field label="繳款日"><input type="number" min="1" max="31" value={form.payment_due_day} onChange={e=>setForm({...form,payment_due_day:e.target.value})}/></Field></div><label><input type="checkbox" checked={form.is_active} onChange={e=>setForm({...form,is_active:e.target.checked})}/> 啟用此卡</label><button className="button">加入卡片夾</button></form></Dialog>}
  </>
}

function Transactions(){
  const [items,setItems]=useState<{id:string;card_id:string;amount_minor:number;category_code:string;merchant_code:string;merchant_name:string;payment_method_code:string;payment_method_name:string;transaction_date:string;note:string}[]>([]);
  const load=()=>api<typeof items>("/member/transactions").then(setItems);useEffect(()=>{load()},[]);
  async function remove(id:string){if(confirm("確定刪除這筆交易？回饋 cap 將會恢復。")){await mutate(`/member/transactions/${id}`,"DELETE");load()}}
  return <><div className="page-head"><div><h1>交易紀錄</h1><p>刪除或修改交易後，月度 cap 會重新計算。</p></div></div><div className="list">{items.map(x=><article className="list-row" key={x.id}><div><h3>NT${(x.amount_minor/100).toLocaleString()} · {x.merchant_name||"其他店家"}</h3><p>{x.category_code} · {x.payment_method_name} · {x.transaction_date}{x.note&&` · ${x.note}`}</p></div><button className="icon-button" aria-label="刪除交易" onClick={()=>remove(x.id)}><Trash2 size={16}/></button></article>)}</div>{items.length===0&&<div className="panel"><Empty title="還沒有交易" text="在推薦結果確認使用卡片後，交易會出現在這裡。"/></div>}</>
}

function MemberSettings(){
  const [tiers,setTiers]=useState<PreferenceTier[]>([{id:"tier-1",units:[]}]),[activeUnit,setActiveUnit]=useState<RewardPreference|null>(null),[notice,setNotice]=useState(""),[error,setError]=useState("");
  const nextTierID=useRef(2);
  const sensors=useSensors(useSensor(PointerSensor,{activationConstraint:{distance:6}}),useSensor(KeyboardSensor));
  useEffect(()=>{api<RewardPreference[]>("/member/reward-preferences").then(items=>{const grouped=groupPreferences(items);setTiers(grouped);nextTierID.current=grouped.length+1}).catch(e=>setError((e as Error).message))},[]);
  const hasEmptyTier=tiers.some(tier=>tier.units.length===0);
  function move(unitID:string,targetTierID:string){setNotice("");setTiers(current=>movePreference(current,unitID,targetTierID))}
  function addTier(){setNotice("");setTiers(current=>[...current,{id:`tier-${Date.now()}-${nextTierID.current++}`,units:[]}])}
  function removeTier(id:string){setNotice("");setTiers(current=>current.length>1?current.filter(tier=>tier.id!==id||tier.units.length>0):current)}
  function dragStart(event:DragStartEvent){setActiveUnit(tiers.flatMap(tier=>tier.units).find(unit=>unit.reward_unit_id===event.active.id)||null)}
  function dragEnd(event:DragEndEvent){if(event.over)move(String(event.active.id),String(event.over.id));setActiveUnit(null)}
  async function save(){if(hasEmptyTier)return;setError("");try{await mutate("/member/reward-preferences","PATCH",{preferences:toPreferenceWrites(tiers)});setNotice("偏好已更新，新的優先順序會套用到推薦結果。")}catch(e){setError((e as Error).message)}}
  return <><div className="page-head"><div><h1>偏好設定</h1><p>把回饋單位拖到偏好層級；越上方越重視，同一層的優先程度相同。</p></div><button className="button secondary" onClick={addTier}><Plus size={17}/>新增層級</button></div>
    {error&&<div className="error preference-message">{error}</div>}{notice&&<div className="notice preference-message">{notice}</div>}
    <DndContext sensors={sensors} onDragStart={dragStart} onDragCancel={()=>setActiveUnit(null)} onDragEnd={dragEnd}>
      <section className="preference-tiers" aria-label="回饋偏好層級">
        {tiers.map((tier,index)=><PreferenceTierBlock key={tier.id} tier={tier} index={index} tierCount={tiers.length} tiers={tiers} onMove={move} onRemove={()=>removeTier(tier.id)}/>)}
      </section>
      <DragOverlay>{activeUnit&&<PreferenceUnitPreview unit={activeUnit}/>}</DragOverlay>
    </DndContext>
    {hasEmptyTier&&<div className="notice preference-message">空白層級不能儲存，請放入回饋單位或刪除空白層級。</div>}
    <div className="preference-actions"><button className="button" disabled={hasEmptyTier} onClick={save}>儲存偏好</button></div>
  </>
}

function PreferenceTierBlock({tier,index,tierCount,tiers,onMove,onRemove}:{tier:PreferenceTier;index:number;tierCount:number;tiers:PreferenceTier[];onMove:(unitID:string,tierID:string)=>void;onRemove:()=>void}){
  const {isOver,setNodeRef}=useDroppable({id:tier.id});
  return <article ref={setNodeRef} className={`preference-tier${isOver?" over":""}`}>
    <header><div><span className="tier-order">{index+1}</span><div><h2>{tierLabel(index,tierCount)}</h2><p>{tier.units.length} 個回饋單位</p></div></div><button className="icon-button" aria-label={`刪除${tierLabel(index,tierCount)}層級`} title={tierCount===1?"至少需要保留一個層級":tier.units.length>0?"請先移出此層級的回饋單位":"刪除空白層級"} disabled={tierCount===1||tier.units.length>0} onClick={onRemove}><Trash2 size={16}/></button></header>
    <div className="preference-unit-list">{tier.units.map(unit=><PreferenceUnitCard key={unit.reward_unit_id} unit={unit} tiers={tiers} currentTierID={tier.id} onMove={onMove}/>)}
      {tier.units.length===0&&<div className="preference-drop-hint">拖放回饋單位到這裡</div>}
    </div>
  </article>
}

function PreferenceUnitCard({unit,tiers,currentTierID,onMove}:{unit:RewardPreference;tiers:PreferenceTier[];currentTierID:string;onMove:(unitID:string,tierID:string)=>void}){
  const {attributes,listeners,setNodeRef,transform,isDragging}=useDraggable({id:unit.reward_unit_id});
  const style=transform?{transform:`translate3d(${transform.x}px, ${transform.y}px, 0)`}:undefined;
  return <div ref={setNodeRef} style={style} className={`preference-unit${isDragging?" dragging":""}`}>
    <button className="drag-handle" type="button" aria-label={`拖拉${unit.name}`} {...listeners} {...attributes}><GripVertical size={18}/></button>
    <div className="preference-unit-name"><strong>{unit.name}</strong><span>{unit.symbol}</span></div>
    <label><span className="sr-only">移動 {unit.name} 至</span><select aria-label={`移動${unit.name}至偏好層級`} value={currentTierID} onChange={event=>onMove(unit.reward_unit_id,event.target.value)}>{tiers.map((tier,index)=><option key={tier.id} value={tier.id}>{tierLabel(index,tiers.length)}</option>)}</select></label>
  </div>
}

function PreferenceUnitPreview({unit}:{unit:RewardPreference}){
  return <div className="preference-unit overlay"><span className="drag-handle"><GripVertical size={18}/></span><div className="preference-unit-name"><strong>{unit.name}</strong><span>{unit.symbol}</span></div></div>
}
