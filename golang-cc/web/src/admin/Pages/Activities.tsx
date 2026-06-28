import { Plus } from "lucide-react";
import Head from "../../tool/Head";
import { ActivitiesPanel } from "./activities/ActivitiesPanel";
import { CreateActivityDialog } from "./activities/CreateActivityDialog";
import { RelationshipDialog } from "./activities/RelationshipDialog";
import { useActivities } from "./activities/useActivities";

export default function Activities() {
  const activities = useActivities();

  return (
    <>
      <Head
        title="回饋方案"
        text="依銀行與卡片管理回饋方案；同卡方案期間不可重疊。"
        actions={
          <button
            className="button"
            onClick={() => activities.setShowCreate(true)}
          >
            <Plus size={17} /> 新增方案
          </button>
        }
      />
      {activities.error && (
        <div className="error preference-message">{activities.error}</div>
      )}
      <ActivitiesPanel
        activeBank={activities.activeBank}
        bankOptions={activities.bankOptions}
        cardActivities={activities.cardActivities}
        cardLoadingID={activities.cardLoadingID}
        cards={activities.cards}
        expandedCards={activities.expandedCards}
        loading={activities.loading}
        onActiveBankChange={activities.setActiveBank}
        onEdit={activities.openRelationshipEditor}
        onQueryChange={activities.setQuery}
        onRemove={(activity) => void activities.remove(activity)}
        onSchedule={(benefit) => void activities.scheduleVersion(benefit)}
        onToggle={(activity) => void activities.toggle(activity)}
        onToggleCard={(card) => void activities.toggleCard(card)}
        processingID={activities.processingID}
        query={activities.query}
        selectedCards={activities.selectedCards}
      />
      {activities.showCreate && (
        <CreateActivityDialog
          allNetworksSelected={activities.allNetworksSelected}
          availableCards={activities.availableCards}
          bankOptions={activities.bankOptions}
          benefits={activities.benefits}
          categories={activities.categories}
          creating={activities.creating}
          error={activities.error}
          form={activities.form}
          merchants={activities.merchants}
          methods={activities.methods}
          onAddBenefit={activities.stageBenefit}
          onChange={activities.setForm}
          onClose={() =>
            !activities.creating && activities.setShowCreate(false)
          }
          onModeChange={activities.selectBenefitMode}
          onRemoveBenefit={(index) =>
            activities.setBenefits((current) =>
              current.filter((_, itemIndex) => itemIndex !== index),
            )
          }
          onSubmit={activities.create}
          qualifiedTypes={activities.qualifiedTypes}
          selectableTypes={activities.selectableTypes}
          selectedCard={activities.selectedCard}
          selectedNetworkIDs={activities.selectedNetworkIDs}
          unavailableTypes={activities.unavailableTypes}
          units={activities.units}
        />
      )}
      {activities.editingActivity && (
        <RelationshipDialog
          activity={activities.editingActivity}
          benefits={activities.editBenefits}
          onChange={activities.setEditBenefits}
          onClose={activities.closeRelationshipDialog}
          onSubmit={activities.saveRelationships}
          processingID={activities.processingID}
        />
      )}
    </>
  );
}
