import { AlertTriangle, CheckCircle2, CirclePlus } from "lucide-react";
import { ActivityClientPreview } from "./ActivityClientPreview";
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
          <span className="page-eyebrow">MOCK FLOW</span>
          <h2>Activity 五層建檔流程</h2>
          <p className="muted">
            先用前端假資料體驗新增、修改、刪除流程；右側即時顯示會員端會看到的回饋內容。
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
            <button
              className="button"
              type="button"
              onClick={() => activity.openActivityEditor()}
            >
              <CirclePlus size={16} /> 新增
            </button>
          )}
          {activity.viewMode === "editor" && (
            <button
              className="button ghost"
              type="button"
              onClick={activity.resetExample}
            >
              重設範例
            </button>
          )}
        </div>
      </div>

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
            <ActivityEntityTree
              flow={activity.flow}
              selection={activity.selection}
              onSelect={activity.select}
              onDelete={activity.deleteSelected}
            />
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
                  onChange={activity.setFlow}
                />
              )}
            {activity.selection.type === "benefit" &&
              activity.selectedBenefit && (
                <BenefitEditor
                  benefit={activity.selectedBenefit}
                  capPeriodOptions={activity.capPeriodOptions}
                  flow={activity.flow}
                  onChange={activity.setFlow}
                />
              )}

            <ActivityEditor flow={activity.flow} onChange={activity.setFlow} />
            <CreateGroupForm
              form={activity.groupForm}
              onAdd={activity.addGroup}
              onChange={activity.setGroupForm}
            />
            <CreateComponentForm
              flow={activity.flow}
              form={activity.componentForm}
              onAdd={activity.addComponent}
              onChange={activity.setComponentForm}
            />
            <CreateRequirementForm
              components={activity.components}
              form={activity.requirementForm}
              onAdd={activity.addRequirement}
              onChange={activity.setRequirementForm}
            />
            <CreateBenefitForm
              capPeriodOptions={activity.capPeriodOptions}
              components={activity.components}
              form={activity.benefitForm}
              onAdd={activity.addBenefit}
              onChange={activity.setBenefitForm}
            />
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
            <ActivityClientPreview flow={activity.flow} />
          </aside>
        </div>
      )}
    </section>
  );
}
