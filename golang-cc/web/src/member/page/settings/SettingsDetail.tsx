import { ArrowLeft, Check, Save } from "lucide-react";
import type React from "react";

export function SettingsDetail({
  title, dirty, saving, notice, error, onBack, onSave, children,
}: {
  title: string; dirty: boolean; saving: boolean; notice: string; error: string;
  onBack: () => void; onSave: () => void; children: React.ReactNode;
}) {
  return (
    <div className="settings-detail-page">
      <header className="settings-detail-head">
        <button type="button" className="settings-back-button" aria-label="返回設定" onClick={onBack}><ArrowLeft /></button>
        <h1>{title}</h1>
      </header>
      {error && <div className="error preference-message">{error}</div>}
      {notice && <div className="notice preference-message"><Check size={16} />{notice}</div>}
      <main className="settings-detail-content">{children}</main>
      <div className="settings-save-dock">
        <button type="button" className="button settings-save-button" disabled={!dirty || saving} onClick={onSave}>
          <Save size={17} />{saving ? "儲存中…" : "儲存"}
        </button>
      </div>
    </div>
  );
}
