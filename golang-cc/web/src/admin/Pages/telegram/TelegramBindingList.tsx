import { Send, Trash2 } from "lucide-react";
import { Empty } from "../../../components";
import type { TelegramBinding } from "../../../models";

type Props = {
  bindings: TelegramBinding[];
  deletingChatID: number | null;
  loading: boolean;
  onRemove: (binding: TelegramBinding) => void;
};

export function TelegramBindingList({ bindings, deletingChatID, loading, onRemove }: Props) {
  return (
    <section className="panel admin-main-column">
      <div className="section-title"><div><h2>目前綁定</h2><p className="muted">{bindings.length} 位會員已連結 Telegram</p></div></div>
      {loading ? <p className="muted">載入 Telegram 綁定中…</p> : <BindingGrid bindings={bindings} deletingChatID={deletingChatID} onRemove={onRemove} />}
    </section>
  );
}

function BindingGrid({ bindings, deletingChatID, onRemove }: Omit<Props, "loading">) {
  if (!bindings.length) {
    return <Empty title="尚無 Telegram 綁定" text="建立綁定後，會員即可透過 Bot 使用推薦功能。" />;
  }
  return (
    <div className="entity-grid">
      {bindings.map((binding) => (
        <article className="entity-card" key={binding.chat_id}>
          <div className="entity-icon telegram-icon"><Send size={20} /></div>
          <div className="entity-copy"><h3>{binding.display_name}</h3><p>{binding.email}</p><span className="mono-chip">Chat ID {binding.chat_id}</span></div>
          <button type="button" className="icon-button" aria-label={`解除 ${binding.display_name} Telegram 綁定`} disabled={deletingChatID === binding.chat_id} onClick={() => onRemove(binding)}>
            <Trash2 size={16} />
          </button>
        </article>
      ))}
    </div>
  );
}
