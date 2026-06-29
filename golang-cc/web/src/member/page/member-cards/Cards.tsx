import { CardPageHeader } from "../cards/CardPageHeader";
import { CardGallery } from "./browse/CardGallery";
import { EmptyState } from "./browse/EmptyState";
import { useBrowseCards } from "./browse/useBrowseCards";
import { CreateDialog } from "./create/CreateDialog";
import { useCreateCard } from "./create/useCreateCard";
import { CardDetail } from "./detail/CardDetail";
import { useCardDetail } from "./detail/useCardDetail";

export default function Cards() {
  const browse = useBrowseCards();
  const detail = useCardDetail({ onDeleted: browse.reload });
  const create = useCreateCard({
    catalogCards: browse.catalogCards,
    onCreated: browse.reload,
  });

  return (
    <>
      <CardPageHeader
        catalogCount={browse.catalogCards.length}
        detail={detail.card}
        loading={browse.loading}
        onBack={detail.close}
        onCreate={create.open}
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
          onDelete={(id) => void detail.remove(id)}
          onToggleGroup={detail.toggleGroup}
          onToggleQualification={detail.toggleQualification}
        />
      )}
      {!browse.loading && !detail.card && browse.myCards.length === 0 && (
        <EmptyState />
      )}
      {create.showCreate && (
        <CreateDialog
          catalog={browse.catalogCards}
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
