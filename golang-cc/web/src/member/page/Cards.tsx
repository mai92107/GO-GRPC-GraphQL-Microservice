import { Empty } from "../../components";
import { CardDetail } from "./cards/CardDetail";
import { CardGallery } from "./cards/CardGallery";
import { CardPageHeader } from "./cards/CardPageHeader";
import { CreateMemberCardDialog } from "./cards/CreateMemberCardDialog";
import { useMemberCards } from "./cards/useMemberCards";

export default function Cards() {
  const wallet = useMemberCards();

  return (
    <>
      <CardPageHeader
        catalogCount={wallet.catalog.length}
        detail={wallet.detail}
        loading={wallet.loading}
        onBack={() => wallet.setDetail(null)}
        onCreate={() => wallet.setShowCreate(true)}
      />
      {wallet.error && <div className="error">{wallet.error}</div>}
      {wallet.loading && <p className="muted">載入卡片資料中…</p>}
      {!wallet.loading && !wallet.detail && wallet.cards.length > 0 && (
        <CardGallery
          cards={wallet.cards}
          onOpen={(card) => void wallet.openDetail(card)}
        />
      )}
      {!wallet.loading && wallet.detail && (
        <CardDetail
          card={wallet.detail}
          deletingID={wallet.deletingID}
          expanded={wallet.expanded}
          overview={wallet.overviews[wallet.detail.id]}
          overviewLoadingID={wallet.overviewLoadingID}
          qualificationSaving={wallet.qualificationSaving}
          onDelete={(id) => void wallet.remove(id)}
          onToggleGroup={(key, nextValue) =>
            wallet.setExpanded((current) => ({ ...current, [key]: nextValue }))
          }
          onToggleQualification={wallet.toggleQualification}
        />
      )}
      {!wallet.loading && !wallet.detail && wallet.cards.length === 0 && (
        <div className="panel">
          <Empty title="卡片夾是空的" text="使用上方按鈕加入目前持有的卡片。" />
        </div>
      )}
      {wallet.showCreate && (
        <CreateMemberCardDialog
          catalog={wallet.catalog}
          form={wallet.form}
          onChange={wallet.setForm}
          onClose={() => !wallet.saving && wallet.setShowCreate(false)}
          onSubmit={wallet.create}
          saving={wallet.saving}
          selectedCard={wallet.selectedCard}
        />
      )}
    </>
  );
}
