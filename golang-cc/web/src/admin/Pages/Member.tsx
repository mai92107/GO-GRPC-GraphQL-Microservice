import { useState, useEffect, FormEvent } from "react";
import { api, patch, post } from "../../api";
import { Field } from "../../components";
import Head from "../Head";
import { invite, toggleActivate, toggleReset } from "../AdminApi";

export function Members(){
    const 
    [users,setUsers]=useState<any[]>([]),
    [invites,setInvites]=useState<any[]>([]),
    [email,setEmail]=useState("");
    
    const loadMembers=()=>
        Promise.all([
            api<any[]>("/admin/users"),
            api<any[]>("/admin/invitings")
        ]).then(([u,i])=>{
            setUsers(u);
            setInvites(i);
        });

    useEffect(()=>{loadMembers()},[]);

    function inviteMember(e:FormEvent){
        e.preventDefault();
        invite(email).then(()=>{
            setEmail("");
            loadMembers();
        });
    }    
    return <>
        <Head title="會員與邀請" text="只有管理員能邀請、停用會員與寄送密碼重設信。"/>
        <div className="settings-grid">
            <section className="panel">
                <h2>邀請會員</h2>
                <form className="stack" onSubmit={inviteMember}>
                    <Field label="Email">
                        <input type="email" value={email} onChange={e=>setEmail(e.target.value)} required/>
                    </Field>
                    <button className="button">寄送啟用連結</button>
                </form>
                <div className="list">
                    {
                    invites.map(i=>
                    <div className="list-row" key={i.id}>
                        <div>
                            <h3>{i.email}</h3>
                            <p>等待接受名單</p>
                        </div>
                    </div>)}
                </div>
            </section>

            <section className="panel">
                <h2>當前會員</h2>
                <div className="list">
                    {
                    users.filter(u=>u.role==="member")
                    .map(u=>MemberCard({ user: u, onRefresh: loadMembers }))}
                </div>
            </section>
        </div>
    </>}


type MemberCardProps = {
  user: any;
  onRefresh: () => void;
};

function MemberCard({ user, onRefresh }: MemberCardProps) {
  return (
    <div className="list-row">
      <div>
        <h3>{user.display_name}</h3>
        <p>{user.email} · {user.status === "active" ? "啟用中" : "已停用"}</p>
      </div>

      <div className="toolbar">
        <button className="button ghost" onClick={() => toggleReset(user.id).then(onRefresh)}>
          重設密碼
        </button>

        <button className="button secondary" onClick={() => toggleActivate(user.id, user.status).then(onRefresh)}>
          {user.status === "active" ? "現在停用" : "現在啟用"}
        </button>
      </div>
    </div>
  );
}