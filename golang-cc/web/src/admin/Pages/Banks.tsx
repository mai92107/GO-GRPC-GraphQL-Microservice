import { Trash2 } from "lucide-react";
import { useState, useEffect, FormEvent } from "react";
import { api } from "../../api";
import Head from "../Head";
import { activateBank, createBank, deleteBank } from "../AdminApi";

export default function Banks() {
  const [items, setItems] = useState<any[]>([]),
    [name, setName] = useState("");

  const loadBanks = () => api<any[]>("/admin/banks").then(setItems);

  useEffect(() => {
    loadBanks();
  }, []);

  function create(e: FormEvent) {
    e.preventDefault();
    createBank(name).then(() => {
      setName("");
      loadBanks();
    });
  }
  function remove(id: string) {
    deleteBank(id)
      .then(() => {
        loadBanks();
      })
      .catch((e) => {
        alert((e as Error).message);
      });
  }
  function toggle(x: any) {
    activateBank(x.id, x.name, x.code, x.website_url, x.is_active).then(() => {
      loadBanks();
    });
  }

  return (
    <>
      <Head title="銀行管理" text="建立與維護卡片所屬銀行。" />
      <section className="panel">
        <form 
            className="toolbar" 
            onSubmit={create}
            style={{ marginBottom: "1em" }}>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="銀行名稱"
            style={{ borderRadius: "1em", flex: "1", padding: "0.1em 0.8em", fontSize: "1.2em" }}
            required
          />
          <button className="button">新增銀行</button>
        </form>
        <div className="list">
          {items.map((x) => (
            <div className="list-row" key={x.id}>
              <div>
                <h3>{x.name}</h3>
                <p>{x.is_active ? "啟用" : "停用"}</p>
              </div>
              <div className="toolbar">
                <button className="button ghost" onClick={() => toggle(x)}>
                  {x.is_active ? "停用" : "啟用"}
                </button>
                <button className="icon-button" onClick={() => remove(x.id)}>
                  <Trash2 size={16} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </section>
    </>
  );
}
