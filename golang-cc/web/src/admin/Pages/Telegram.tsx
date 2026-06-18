import { Trash2 } from "lucide-react";
import { useState, useEffect, FormEvent } from "react";
import { TelegramBinding, AdminUser, api, mutate, post } from "../../api";
import { Field } from "../../components";
import Head from "../Head";
import { deleteTelegramBinding, postTelegramBinding } from "../AdminApi";

export default function TelegramBindings(){
  const 
    [bindings,setBindings] = useState<TelegramBinding[]>([]),
    [users,setUsers] = useState<AdminUser[]>([]),
    [chatID,setChatID] = useState(""),
    [userID,setUserID] = useState(""),
    [notice,setNotice] = useState(""),
    [error,setError] = useState("");

    const loadMembers = ()=>Promise.all([
        api<TelegramBinding[]>("/admin/telegram-bindings"),
        api<AdminUser[]>("/admin/users")
    ]).then(
        ([b,u])=>{
            setBindings(b);
            setUsers(u);
            const available = u.filter(
                x=>
                    x.role === "member" &&
                    x.status === "active" &&
                    !b.some(binding=>binding.user_id===x.id));
            setUserID(
                current=>
                    available.some(
                        x=>x.id===current) ? current : available[0]?.id || "")
            });

    useEffect(
        ()=>{
            loadMembers().catch(
                e=>setError((e as Error).message)
            )
        },[]);

    const available = users.filter(x=>
        x.role==="member" &&
        x.status==="active" &&
        !bindings.some(binding=>binding.user_id===x.id)
    );
    function create(e:FormEvent){
        e.preventDefault();
        initBlock();
        const parsed=Number(chatID);
        if(!Number.isSafeInteger(parsed)){
            setError("Chat ID 格式無效");
            return
        }
        postTelegramBinding(parsed,userID).then(()=>{
            setChatID("");
            setNotice("Telegram Chat 綁定完成");
            loadMembers()
        }).catch(e=>{
            setError((e as Error).message)
        })
    }
  
    function remove(binding:TelegramBinding){
        initBlock();
        if(!confirm(`確定解除 ${binding.display_name} 的 Telegram 綁定？`))
            return;
        deleteTelegramBinding(binding.chat_id).then(()=>{
            setNotice("Telegram Chat 綁定已解除");
            loadMembers();
        }).catch(e=>{
            setError((e as Error).message)
        })
    }

    function initBlock(){
        setError("");
        setNotice("");
    }

  return <>
  <Head title="Telegram 綁定" text="將會員與 Telegram Chat ID 綁定後，會員即可透過 Bot 取得卡片推薦。"/>
  {error&&<div className="error preference-message">{error}</div>}
  {notice&&<div className="notice preference-message">{notice}</div>}
  <div className="settings-grid">
    <section className="panel">
        <h2>新增綁定</h2>
        <p className="muted">會員對 Bot 輸入 /recommand，首次會回覆 Chat ID。</p>
        <form className="stack" onSubmit={create}>
            <Field label="Telegram Chat ID">
                <input 
                    inputMode="numeric" 
                    pattern="-?[0-9]+" 
                    placeholder="例如：123456789" 
                    value={chatID} 
                    onChange={e=>setChatID(e.target.value.trim())} 
                    required/>
            </Field>
            <Field label="啟用會員">
                <select 
                    value={userID} 
                    onChange={e=>setUserID(e.target.value)} 
                    required 
                    disabled={!available.length}>
                    <option value="">
                        {available.length ? "請選擇會員" : "沒有可綁定的啟用會員"}
                    </option>
                    {available.map(x=>
                        <option value={x.id} key={x.id}>
                            {x.display_name} · {x.email}
                        </option>)}
                </select>
            </Field>
            <button className="button" disabled={!available.length}>
                建立綁定
            </button>
        </form>
    </section>

    <section className="panel">
        <h2>目前綁定</h2>
        <div className="list">
            {bindings.map(x=>
                <div className="list-row" key={x.chat_id}>
                    <div>
                        <h3>{x.display_name}</h3>
                        <p>{x.email}<br/>Chat ID：{x.chat_id}</p>
                    </div>
                    <button className="icon-button" aria-label={`解除 ${x.display_name} Telegram 綁定`} onClick={()=>remove(x)}><Trash2 size={16}/></button>
                </div>)}
        </div>
        {bindings.length === 0 && <p className="muted">目前沒有 Telegram 綁定。</p>}
    </section>
</div></>
}