import { PaymentsView } from "./settings/PaymentsView";
import { PreferencesView } from "./settings/PreferencesView";
import { SettingsHome } from "./settings/SettingsHome";
import { useMemberSettings } from "./settings/useMemberSettings";

export default function MemberSettings() {
  const settings = useMemberSettings();

  if (settings.loading) return <div className="member-loading">正在載入你的設定…</div>;

  if (settings.view === "preferences") {
    return (
      <PreferencesView
        dirty={settings.preferenceDirty}
        error={settings.error}
        notice={settings.notice}
        onBack={() => settings.openView("overview")}
        onSave={() => void settings.savePreferences()}
        onSetPriority={settings.setPriority}
        preferences={settings.preferences}
        priorities={settings.priorities}
        saving={settings.saving}
      />
    );
  }

  if (settings.view === "payments") {
    return (
      <PaymentsView
        dirty={settings.paymentDirty}
        error={settings.error}
        notice={settings.notice}
        onBack={() => settings.openView("overview")}
        onSave={() => void settings.savePayments()}
        onToggleMore={() => settings.setShowMorePayments((current) => !current)}
        onTogglePayment={settings.togglePayment}
        otherPayments={settings.otherPayments}
        saving={settings.saving}
        selectedPayments={settings.selectedPayments}
        showMorePayments={settings.showMorePayments}
      />
    );
  }

  return (
    <SettingsHome
      error={settings.error}
      highPriorityCount={settings.highPriorityCount}
      notice={settings.notice}
      onOpenPayments={() => settings.openView("payments")}
      onOpenPreferences={() => settings.openView("preferences")}
      selectedPaymentCount={settings.selectedPayments.length}
    />
  );
}
