import { Trash2 } from "lucide-react";
import type {
  ActivityFlowSelection,
  ActivityFlowModel,
} from "./activityFlowTypes";

export function ActivityEntityTree({
  flow,
  onDelete,
  onSelect,
  selection,
}: {
  flow: ActivityFlowModel;
  selection: ActivityFlowSelection;
  onDelete: (nextSelection: ActivityFlowSelection) => void;
  onSelect: (selection: ActivityFlowSelection) => void;
}) {
  return (
    <section className="activity-entity-tree">
      <div className="form-section-title">
        <span>1</span>
        <div>
          <h3>資料結構</h3>
          <p>選取節點後可在下方修改；刪除會同步更新右側預覽</p>
        </div>
      </div>
      <TreeButton
        active={selection.type === "activity"}
        enabled={flow.activity.is_active}
        label={`${flow.activity.bank_name} ${flow.activity.card_name} ${flow.activity.title}`}
        meta="Activity"
        onClick={() => onSelect({ type: "activity", id: flow.activity.id })}
      />
      {flow.reward_groups.map((group) => (
        <div className="entity-tree-group" key={group.id}>
          <TreeButton
            active={selection.type === "group" && selection.id === group.id}
            enabled={group.is_active}
            label={group.name}
            meta={`Group · ${group.components.length} components`}
            onClick={() => onSelect({ type: "group", id: group.id })}
            onDelete={() =>
              onDelete({ type: "activity", id: flow.activity.id })
            }
          />
          {group.components.map((component) => (
            <div className="entity-tree-component" key={component.id}>
              <TreeButton
                active={
                  selection.type === "component" &&
                  selection.id === component.id
                }
                enabled={component.is_active}
                label={component.name}
                meta={`L${component.layer} · ${component.stack_group} · ${component.stack_mode}`}
                onClick={() =>
                  onSelect({ type: "component", id: component.id })
                }
                onDelete={() =>
                  onDelete({ type: "group", id: component.reward_group_id })
                }
              />
              <div className="entity-tree-child-list">
                {component.requirements.map((requirement) => (
                  <TreeButton
                    active={
                      selection.type === "requirement" &&
                      selection.id === requirement.id
                    }
                    enabled={requirement.is_active}
                    label={
                      requirement.description || requirement.requirement_type
                    }
                    meta={`Req · ${requirement.requirement_type}`}
                    key={requirement.id}
                    onClick={() =>
                      onSelect({
                        type: "requirement",
                        id: requirement.id,
                        componentID: component.id,
                      })
                    }
                    onDelete={() =>
                      onDelete({ type: "component", id: component.id })
                    }
                  />
                ))}
                {component.benefits.map((benefit) => (
                  <TreeButton
                    active={
                      selection.type === "benefit" &&
                      selection.id === benefit.id
                    }
                    enabled={benefit.is_active}
                    label={benefit.description || benefit.benefit_type}
                    meta={`Benefit · ${benefit.value} ${benefit.reward_unit_id}`}
                    key={benefit.id}
                    onClick={() =>
                      onSelect({
                        type: "benefit",
                        id: benefit.id,
                        componentID: component.id,
                      })
                    }
                    onDelete={() =>
                      onDelete({ type: "component", id: component.id })
                    }
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      ))}
    </section>
  );
}

function TreeButton({
  active,
  label,
  meta,
  onClick,
  onDelete,
  enabled = true,
}: {
  active: boolean;
  enabled?: boolean;
  label: string;
  meta: string;
  onClick: () => void;
  onDelete?: () => void;
}) {
  return (
    <div className={`entity-tree-row ${active ? "selected" : ""} ${enabled ? "is-active" : "is-inactive"}`}>
      <button type="button" onClick={onClick}>
        <strong>{label || "未命名"}</strong>
        <small>{enabled ? meta : `${meta} · 已停用`}</small>
      </button>
      {onDelete && (
        <button
          aria-label={`刪除 ${label}`}
          className="icon-button danger-icon"
          type="button"
          onClick={onDelete}
        >
          <Trash2 size={15} />
        </button>
      )}
    </div>
  );
}

