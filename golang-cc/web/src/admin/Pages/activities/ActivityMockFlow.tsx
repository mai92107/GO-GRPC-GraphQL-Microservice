import { AlertTriangle, CheckCircle2, CirclePlus, Save } from "lucide-react";
import {
  CreateBenefitForm,
  CreateComponentForm,
  CreateGroupForm,
  CreateRequirementForm,
} from "./ActivityCreateForms";
import {
  ActivityEditor,
  BenefitEditor,
  ComponentEditor,
  GroupEditor,
  RequirementEditor,
} from "./ActivityEditors";
import { ActivityEntityTree } from "./ActivityEntityTree";
import { ActivityOverviewList } from "./ActivityOverviewList";
import { useActivityMockFlow } from "./useActivityMockFlow";

export function ActivityMockFlow() {
  const activity = useActivityMockFlow();

  return (
    <section className="panel activity-mock-flow">
      <div className="section-title">
        <div>
          <span className="page-eyebrow">ACTIVITY FLOW</span>
          <h2>Activity 五層建檔流程</h2>
          <p className="muted">
            初次只載入活動清單；展開或編輯單筆活動時才載入 Group / Component / Requirement / Benefit。
          </p>
        </div>
        <div className="toolbar">
          {activity.viewMode === "editor" && (
            <button
              className="button ghost"
              type="button"
              onClick={() => activity.setViewMode("overview")}
            >
              返回總覽
            </button>
          )}
          {activity.viewMode === "overview" && (
            <button className="button" type="button" onClick={activity.reset}>
              <CirclePlus size={16} /> 新建活動
            </button>
          )}
          {activity.viewMode === "editor" && (
            <button
              className="button ghost"
              type="button"
              disabled={!activity.canSave}
              onClick={activity.saveActivity}
            >
              <Save size={16} />
              {activity.editorMode === "create" ? "新增活動" : "儲存活動"}
            </button>
          )}
        </div>
      </div>

      {activity.error && <p className="form-error">{activity.error}</p>}

      {activity.viewMode === "overview" ? (
        <ActivityOverviewList
          bankFilter={activity.bankFilter}
          bankOptions={activity.bankOptions}
          cardFilter={activity.cardFilter}
          cardOptions={activity.cardOptions}
          rows={activity.overviewRows}
          onBankFilterChange={activity.setBank}
          onCardFilterChange={activity.setCardFilter}
          onEdit={activity.openActivityEditor}
        />
      ) : (
        <div className="activity-mock-layout">
          <div className="activity-mock-builder">
            {activity.selection.type === "group" && activity.selectedGroup && (
              <GroupEditor
                flow={activity.flow}
                group={activity.selectedGroup}
                onChange={activity.setFlow}
              />
            )}
            {activity.selection.type === "component" &&
              activity.selectedComponent && (
                <ComponentEditor
                  component={activity.selectedComponent}
                  flow={activity.flow}
                  onChange={activity.setFlow}
                />
              )}
            {activity.selection.type === "requirement" &&
              activity.selectedRequirement && (
                <RequirementEditor
                  flow={activity.flow}
                  requirement={activity.selectedRequirement}
                  requirementOptions={activity.requirementOptions}
                  requirementTypes={activity.requirementTypes}
                  onChange={activity.setFlow}
                />
              )}
            {activity.selection.type === "benefit" && activity.selectedBenefit && (
              <BenefitEditor
                benefit={activity.selectedBenefit}
                capPeriodOptions={activity.capPeriodOptions}
                flow={activity.flow}
                rewardUnits={activity.rewardUnits}
                onChange={activity.setFlow}
              />
            )}

            <ActivityEditor activityCardOptions={activity.activityCardOptions} bankOptions={activity.bankOptions} flow={activity.flow} onBankChange={activity.setActivityBank} onCardChange={activity.setActivityCard} onChange={activity.setFlow} />
            <CreateGroupForm
              form={activity.groupForm}
              onAdd={activity.addGroup}
              onChange={activity.setGroupForm}
            />
            {activity.showComponentForm && (
              <CreateComponentForm
                flow={activity.flow}
                form={activity.componentForm}
                onAdd={activity.addComponent}
                onChange={activity.setComponentForm}
              />
            )}
            {activity.showRequirementForm && (
              <CreateRequirementForm
                components={activity.components}
                form={activity.requirementForm}
                requirementOptions={activity.requirementOptions}
                requirementTypes={activity.requirementTypes}
                onAdd={activity.addRequirement}
                onChange={activity.setRequirementForm}
              />
            )}
            {activity.showBenefitForm && (
              <CreateBenefitForm
                capPeriodOptions={activity.capPeriodOptions}
                components={activity.components}
                form={activity.benefitForm}
                rewardUnits={activity.rewardUnits}
                onAdd={activity.addBenefit}
                onChange={activity.setBenefitForm}
              />
            )}
          </div>

          <aside className="activity-mock-output">
            <div className="activity-mock-checks">
              {activity.checks.map((check) => (
                <span
                  className={check.pass ? "pass" : "fail"}
                  key={check.label}
                >
                  {check.pass ? (
                    <CheckCircle2 size={16} />
                  ) : (
                    <AlertTriangle size={16} />
                  )}
                  {check.label}
                </span>
              ))}
            </div>
            <ActivityEntityTree
              flow={activity.flow}
              selection={activity.selection}
              onSelect={activity.select}
              onDelete={activity.deleteSelected}
            />
          </aside>
        </div>
      )}
    </section>
  );
}
