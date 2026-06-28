import type { RewardPreference } from "../../../preferences";
import { priorityOptions, type Priority } from "./types";
import { SettingsDetail } from "./SettingsDetail";

type Props = {
  dirty: boolean;
  error: string;
  notice: string;
  preferences: RewardPreference[];
  priorities: Record<string, Priority>;
  saving: boolean;
  onBack: () => void;
  onSave: () => void;
  onSetPriority: (unitID: string, priority: Priority) => void;
};

export function PreferencesView(props: Props) {
  return (
    <SettingsDetail title="編輯回饋偏好" dirty={props.dirty} saving={props.saving} notice={props.notice} error={props.error} onBack={props.onBack} onSave={props.onSave}>
      <div className="settings-detail-intro"><h2>你重視哪種回饋？</h2><p>每種回饋選擇一個優先程度，推薦結果會依此調整。</p></div>
      <div className="preference-control-list settings-preference-list">
        {props.preferences.map((unit) => (
          <article className="preference-control-row" key={unit.reward_unit_id}>
            <div className="reward-unit-identity"><span>{unit.symbol || unit.name.slice(0, 1)}</span><div><strong>{unit.name}</strong></div></div>
            <div className="priority-segment" role="radiogroup" aria-label={`${unit.name}偏好`}>
              {priorityOptions.map((option) => {
                const selected = (props.priorities[unit.reward_unit_id] || "normal") === option.id;
                return (
                  <button type="button" role="radio" aria-checked={selected} className={selected ? "selected" : ""} key={option.id} onClick={() => props.onSetPriority(unit.reward_unit_id, option.id)}>
                    <strong>{option.label}</strong><small>{option.description}</small>
                  </button>
                );
              })}
            </div>
          </article>
        ))}
      </div>
    </SettingsDetail>
  );
}
