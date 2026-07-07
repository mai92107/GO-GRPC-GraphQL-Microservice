import { CardPageHeader } from "./cards/CardPageHeader";
import { CardGallery } from "./member-cards/browse/CardGallery";
import { EmptyState } from "./member-cards/browse/EmptyState";
import { useBrowseCards } from "./member-cards/browse/useBrowseCards";
import { CreateDialog } from "./member-cards/create/CreateDialog";
import { useCreateCard } from "./member-cards/create/useCreateCard";
import { CardDetail } from "./member-cards/detail/CardDetail";
import { useCardDetail } from "./member-cards/detail/useCardDetail";

export default function Cards() {
  const browse = useBrowseCards();
  const detail = useCardDetail({ onDeleted: browse.reload });
  const create = useCreateCard({
    onCreated: browse.reload,
  });

  return (
    <>
      <CardPageHeader
        detail={detail.card}
        loading={browse.loading}
        onBack={detail.close}
        onCreate={() => void create.open()}
      />
      {browse.error && <div className="error">{browse.error}</div>}
      {detail.error && <div className="error">{detail.error}</div>}
      {browse.loading && <p className="muted">載入卡片資料中…</p>}
      {!browse.loading && !detail.card && browse.myCards.length > 0 && (
        <CardGallery cards={browse.myCards} onOpen={(id) => void detail.open(id)} />
      )}
      {!browse.loading && detail.loadingCardID && !detail.card && (
        <p className="muted">載入卡片回饋中…</p>
      )}
      {!browse.loading && detail.card && (
        <CardDetail
          card={detail.card}
          deletingID={detail.deletingID}
          expanded={detail.expanded}
          overview={detail.overview}
          overviewLoading={detail.loadingCardID === detail.card.member_card_id}
          qualificationSaving={detail.qualificationSaving}
          creditLimitSaving={detail.creditLimitSaving}
          onDelete={(id) => void detail.remove(id)}
          onUpdateCreditLimit={detail.updateCreditLimit}
          onToggleGroup={detail.toggleGroup}
          onToggleQualification={detail.toggleQualification}
        />
      )}
      {!browse.loading && !detail.card && browse.myCards.length === 0 && (
        <EmptyState />
      )}
      {create.showCreate && (
        <CreateDialog
          catalog={create.catalogCards}
          catalogLoading={create.catalogLoading}
          error={create.error}
          form={create.form}
          formForCard={create.formForCard}
          onChange={create.setForm}
          onClose={create.close}
          onSubmit={create.create}
          saving={create.saving}
          selectedCard={create.selectedCard}
        />
      )}
    </>
  );
}
