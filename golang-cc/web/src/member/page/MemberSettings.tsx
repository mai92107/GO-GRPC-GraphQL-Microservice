import {
  useSensors,
  useSensor,
  PointerSensor,
  KeyboardSensor,
  DragStartEvent,
  DragEndEvent,
  DndContext,
  DragOverlay,
  useDroppable,
  useDraggable,
} from "@dnd-kit/core";
import { Plus, Trash2, GripVertical } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import {
  PreferenceTier,
  RewardPreference,
  groupPreferences,
  movePreference,
  toPreferenceWrites,
  tierLabel,
} from "../../preferences";
import { getRewardPreferences, updateRewardPreferences } from "../MemberApi";

export default function MemberSettings() {
  const [tiers, setTiers] = useState<PreferenceTier[]>([
      { id: "tier-1", units: [] },
    ]),
    [activeUnit, setActiveUnit] = useState<RewardPreference | null>(null),
    [notice, setNotice] = useState(""),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(true),
    [saving, setSaving] = useState(false);
  const nextTierID = useRef(2);
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
    useSensor(KeyboardSensor),
  );
  useEffect(() => {
    getRewardPreferences()
      .then((items) => {
        const grouped = groupPreferences(items);
        setTiers(grouped);
        nextTierID.current = grouped.length + 1;
      })
      .catch((e) => setError((e as Error).message))
      .finally(() => setLoading(false));
  }, []);
  const hasEmptyTier = tiers.some((tier) => tier.units.length === 0);
  function move(unitID: string, targetTierID: string) {
    setNotice("");
    setError("");
    setTiers((current) => movePreference(current, unitID, targetTierID));
  }
  function addTier() {
    setNotice("");
    setError("");
    setTiers((current) => [
      ...current,
      { id: `tier-${nextTierID.current++}`, units: [] },
    ]);
  }
  function removeTier(id: string) {
    setNotice("");
    setError("");
    setTiers((current) =>
      current.length > 1
        ? current.filter((tier) => tier.id !== id || tier.units.length > 0)
        : current,
    );
  }
  function dragStart(event: DragStartEvent) {
    setActiveUnit(
      tiers
        .flatMap((tier) => tier.units)
        .find((unit) => unit.reward_unit_id === event.active.id) || null,
    );
  }
  function dragEnd(event: DragEndEvent) {
    if (event.over) move(String(event.active.id), String(event.over.id));
    setActiveUnit(null);
  }
  async function save() {
    if (hasEmptyTier) return;
    setError("");
    setSaving(true);
    try {
      await updateRewardPreferences(toPreferenceWrites(tiers));
      setNotice("偏好已更新，新的優先順序會套用到推薦結果。");
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setSaving(false);
    }
  }
  return (
    <>
      <div className="page-head">
        <div>
          <h1>偏好設定</h1>
          <p>把回饋單位拖到偏好層級；越上方越重視，同一層的優先程度相同。</p>
        </div>
        <button
          type="button"
          className="button secondary"
          disabled={loading || saving}
          onClick={addTier}
        >
          <Plus size={17} />
          新增層級
        </button>
      </div>
      {error && <div className="error preference-message">{error}</div>}
      {notice && <div className="notice preference-message">{notice}</div>}
      {loading && <p className="muted">載入偏好設定中…</p>}
      <DndContext
        sensors={sensors}
        onDragStart={dragStart}
        onDragCancel={() => setActiveUnit(null)}
        onDragEnd={dragEnd}
      >
        <section
          className="preference-tiers"
          aria-label="回饋偏好層級"
          aria-busy={loading}
        >
          {tiers.map((tier, index) => (
            <PreferenceTierBlock
              key={tier.id}
              tier={tier}
              index={index}
              tierCount={tiers.length}
              tiers={tiers}
              onMove={move}
              onRemove={() => removeTier(tier.id)}
            />
          ))}
        </section>
        <DragOverlay>
          {activeUnit && <PreferenceUnitPreview unit={activeUnit} />}
        </DragOverlay>
      </DndContext>
      {hasEmptyTier && (
        <div className="notice preference-message">
          空白層級不能儲存，請放入回饋單位或刪除空白層級。
        </div>
      )}
      <div className="preference-actions">
        <button
          type="button"
          className="button"
          disabled={loading || saving || hasEmptyTier}
          onClick={() => void save()}
        >
          {saving ? "儲存中…" : "儲存偏好"}
        </button>
      </div>
    </>
  );
}

function PreferenceTierBlock({
  tier,
  index,
  tierCount,
  tiers,
  onMove,
  onRemove,
}: {
  tier: PreferenceTier;
  index: number;
  tierCount: number;
  tiers: PreferenceTier[];
  onMove: (unitID: string, tierID: string) => void;
  onRemove: () => void;
}) {
  const { isOver, setNodeRef } = useDroppable({ id: tier.id });
  return (
    <article
      ref={setNodeRef}
      className={`preference-tier${isOver ? " over" : ""}`}
    >
      <header>
        <div>
          <span className="tier-order">{index + 1}</span>
          <div>
            <h2>{tierLabel(index, tierCount)}</h2>
            <p>{tier.units.length} 個回饋單位</p>
          </div>
        </div>
        <button
          type="button"
          className="icon-button"
          aria-label={`刪除${tierLabel(index, tierCount)}層級`}
          title={
            tierCount === 1
              ? "至少需要保留一個層級"
              : tier.units.length > 0
                ? "請先移出此層級的回饋單位"
                : "刪除空白層級"
          }
          disabled={tierCount === 1 || tier.units.length > 0}
          onClick={onRemove}
        >
          <Trash2 size={16} />
        </button>
      </header>
      <div className="preference-unit-list">
        {tier.units.map((unit) => (
          <PreferenceUnitCard
            key={unit.reward_unit_id}
            unit={unit}
            tiers={tiers}
            currentTierID={tier.id}
            onMove={onMove}
          />
        ))}
        {tier.units.length === 0 && (
          <div className="preference-drop-hint">拖放回饋單位到這裡</div>
        )}
      </div>
    </article>
  );
}

function PreferenceUnitCard({
  unit,
  tiers,
  currentTierID,
  onMove,
}: {
  unit: RewardPreference;
  tiers: PreferenceTier[];
  currentTierID: string;
  onMove: (unitID: string, tierID: string) => void;
}) {
  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({ id: unit.reward_unit_id });
  const style = transform
    ? { transform: `translate3d(${transform.x}px, ${transform.y}px, 0)` }
    : undefined;
  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`preference-unit${isDragging ? " dragging" : ""}`}
    >
      <button
        className="drag-handle"
        type="button"
        aria-label={`拖拉${unit.name}`}
        {...listeners}
        {...attributes}
      >
        <GripVertical size={18} />
      </button>
      <div className="preference-unit-name">
        <strong>{unit.name}</strong>
        <span>{unit.symbol}</span>
      </div>
      <label>
        <span className="sr-only">移動 {unit.name} 至</span>
        <select
          aria-label={`移動${unit.name}至偏好層級`}
          value={currentTierID}
          onChange={(event) => onMove(unit.reward_unit_id, event.target.value)}
        >
          {tiers.map((tier, index) => (
            <option key={tier.id} value={tier.id}>
              {tierLabel(index, tiers.length)}
            </option>
          ))}
        </select>
      </label>
    </div>
  );
}

function PreferenceUnitPreview({ unit }: { unit: RewardPreference }) {
  return (
    <div className="preference-unit overlay">
      <span className="drag-handle">
        <GripVertical size={18} />
      </span>
      <div className="preference-unit-name">
        <strong>{unit.name}</strong>
        <span>{unit.symbol}</span>
      </div>
    </div>
  );
}
