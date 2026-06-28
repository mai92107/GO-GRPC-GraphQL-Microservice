import { Link2 } from "lucide-react";

export function TelegramGuide() {
  return (
    <aside className="panel admin-side-column guide-card">
      <div className="guide-icon"><Link2 size={22} /></div>
      <h2>如何取得 Chat ID？</h2>
      <ol>
        <li>請會員開啟 Telegram Bot</li>
        <li>輸入 <code>/recommand</code></li>
        <li>複製 Bot 回覆的 Chat ID</li>
        <li>點擊「新增綁定」完成連結</li>
      </ol>
    </aside>
  );
}
