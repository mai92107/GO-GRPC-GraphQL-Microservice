import { Plus } from "lucide-react";
import { FormEvent } from "react";
import Head from "../../tool/Head";
import { CardCatalogList } from "./cards/CardCatalogList";
import { CreateCardDialog } from "./cards/CreateCardDialog";
import { EditCardDialog } from "./cards/EditCardDialog";
import { useAdminCards } from "./cards/useAdminCards";

export default function Cards() {
  const cards = useAdminCards();

  function save(event: FormEvent) {
    event.preventDefault();
    if (cards.editing) void cards.persist(cards.editing.is_active);
  }

  return (
    <>
      <Head
        title="銀行卡片目錄"
        text="會員只能從此目錄加入持有卡片。"
        actions={
          <button
            className="button"
            onClick={() => cards.setShowCreate(true)}
            disabled={!cards.banks.length}
          >
            <Plus size={17} /> 新增卡片
          </button>
        }
      />
      {cards.error && <div className="error">{cards.error}</div>}
      <CardCatalogList
        bankFilter={cards.bankFilter}
        banks={cards.banks}
        cards={cards.filtered}
        editLoadingID={cards.editLoadingID}
        itemsCount={cards.items.length}
        loading={cards.loading}
        onBankFilterChange={cards.setBankFilter}
        onEdit={(card) => void cards.openEditor(card)}
        onQueryChange={cards.setQuery}
        query={cards.query}
      />
      {cards.showCreate && (
        <CreateCardDialog
          banks={cards.banks}
          creating={cards.creating}
          form={cards.form}
          networks={cards.networks}
          onChange={cards.setForm}
          onClose={() => !cards.creating && cards.setShowCreate(false)}
          onSubmit={cards.create}
        />
      )}
      {cards.editing && (
        <EditCardDialog
          banks={cards.banks}
          card={cards.editing}
          form={cards.editForm}
          networks={cards.networks}
          onChange={cards.setEditForm}
          onClose={() => !cards.saving && cards.setEditing(null)}
          onDelete={() => void cards.remove()}
          onSave={save}
          onToggleActive={() =>
            cards.editing && void cards.persist(!cards.editing.is_active)
          }
          saving={cards.saving}
        />
      )}
    </>
  );
}
