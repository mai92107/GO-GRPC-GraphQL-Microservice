import { Plus } from "lucide-react";
import Head from "../../tool/Head";
import { CreateTelegramDialog } from "./telegram/CreateTelegramDialog";
import { TelegramBindingList } from "./telegram/TelegramBindingList";
import { TelegramGuide } from "./telegram/TelegramGuide";
import { useTelegramBindings } from "./telegram/useTelegramBindings";

export default function TelegramBindings() {
  const telegram = useTelegramBindings();

  return (
    <>
      <Head
        title="Telegram 綁定"
        text="管理會員與 Telegram Bot 的連結。"
        actions={
          <button className="button" onClick={() => telegram.setShowCreate(true)} disabled={!telegram.availableUsers.length}>
            <Plus size={17} /> 新增綁定
          </button>
        }
      />
      {telegram.error && <div className="error preference-message">{telegram.error}</div>}
      {telegram.notice && <div className="notice preference-message">{telegram.notice}</div>}
      <div className="admin-columns">
        <TelegramBindingList
          bindings={telegram.bindings}
          deletingChatID={telegram.deletingChatID}
          loading={telegram.loading}
          onRemove={(binding) => void telegram.remove(binding)}
        />
        <TelegramGuide />
      </div>
      {telegram.showCreate && (
        <CreateTelegramDialog
          availableUsers={telegram.availableUsers}
          chatID={telegram.chatID}
          creating={telegram.creating}
          loading={telegram.loading}
          onChatIDChange={telegram.setChatID}
          onClose={() => !telegram.creating && telegram.setShowCreate(false)}
          onSubmit={telegram.create}
          onUserIDChange={telegram.setUserID}
          userID={telegram.userID}
        />
      )}
    </>
  );
}
