import { useEffect, useMemo, useState } from "react";
import {
  allComponents,
  blankMockActivity,
  emptyBenefitForm,
  emptyComponentForm,
  emptyGroupForm,
  emptyMockActivity,
  emptyRequirementForm,
  validateMockFlow,
  withBenefit,
  withComponent,
  withGroup,
  withRequirement,
} from "./mockFlowHelpers";
import {
  loadCapPeriodOptions,
  mockCatalogCards,
} from "./activityMockSettings";
import { deleteSelection, overviewItems } from "./activityFlowHelpers";
import type {
  ActivityMockSelection,
  ActivityMockViewMode,
  MockBenefitForm,
  MockComponentForm,
  MockGroupForm,
  MockRequirementForm,
} from "./mockFlowTypes";

const initialSelection: ActivityMockSelection = {
  type: "activity",
  id: "activity-sport-q3",
};

type ActivityEditorMode = "create" | "edit";

export function useActivityMockFlow() {
  const [flow, setFlow] = useState(emptyMockActivity);
  const [viewMode, setViewMode] =
    useState<ActivityMockViewMode>("overview");
  const [editorMode, setEditorMode] = useState<ActivityEditorMode>("edit");
  const [bankFilter, setBankFilter] = useState("all");
  const [cardFilter, setCardFilter] = useState("all");
  const [selection, setSelection] =
    useState<ActivityMockSelection>(initialSelection);
  const [groupForm, setGroupForm] = useState<MockGroupForm>(emptyGroupForm);
  const [componentForm, setComponentForm] = useState<MockComponentForm>(() =>
    emptyComponentForm("group-line-pay"),
  );
  const [requirementForm, setRequirementForm] = useState<MockRequirementForm>(
    () => emptyRequirementForm("component-linepay-base"),
  );
  const [benefitForm, setBenefitForm] = useState<MockBenefitForm>(() =>
    emptyBenefitForm("component-linepay-base"),
  );
  const [capPeriodOptions, setCapPeriodOptions] =
    useState(loadCapPeriodOptions);

  useEffect(() => {
    const refresh = () => setCapPeriodOptions(loadCapPeriodOptions());
    window.addEventListener("focus", refresh);
    window.addEventListener("storage", refresh);
    return () => {
      window.removeEventListener("focus", refresh);
      window.removeEventListener("storage", refresh);
    };
  }, []);

  const components = useMemo(() => allComponents(flow), [flow]);
  const bankOptions = useMemo(
    () => [
      ...new Map(mockCatalogCards.map((card) => [card.bank_id, card])).values(),
    ],
    [],
  );
  const cardOptions = useMemo(
    () =>
      mockCatalogCards.filter(
        (card) => bankFilter === "all" || card.bank_id === bankFilter,
      ),
    [bankFilter],
  );
  const overviewRows = useMemo(
    () =>
      overviewItems(flow).filter(
        (item) =>
          (bankFilter === "all" || item.activity.bank_id === bankFilter) &&
          (cardFilter === "all" ||
            item.activity.card_product_id === cardFilter),
      ),
    [bankFilter, cardFilter, flow],
  );
  const checks = useMemo(() => validateMockFlow(flow), [flow]);
  const selectedGroup =
    selection.type === "group"
      ? flow.reward_groups.find((group) => group.id === selection.id)
      : undefined;
  const selectedComponent =
    selection.type === "component"
      ? components.find((component) => component.id === selection.id)
      : undefined;
  const selectedRequirement =
    selection.type === "requirement"
      ? components
          .flatMap((component) => component.requirements)
          .find((requirement) => requirement.id === selection.id)
      : undefined;
  const selectedBenefit =
    selection.type === "benefit"
      ? components
          .flatMap((component) => component.benefits)
          .find((benefit) => benefit.id === selection.id)
      : undefined;

  const resetForms = (componentID = "", groupID = "") => {
    setGroupForm(emptyGroupForm());
    setComponentForm(emptyComponentForm(groupID));
    setRequirementForm(emptyRequirementForm(componentID));
    setBenefitForm(emptyBenefitForm(componentID));
  };

  const select = (nextSelection: ActivityMockSelection) => {
    setSelection(nextSelection);
    if (nextSelection.type === "component") {
      setRequirementForm(emptyRequirementForm(nextSelection.id));
      setBenefitForm(emptyBenefitForm(nextSelection.id));
    }
  };

  const deleteSelected = (nextSelection: ActivityMockSelection) => {
    setFlow((current) => deleteSelection(current, selection));
    setSelection(nextSelection);
  };

  const setBank = (nextBank: string) => {
    setBankFilter(nextBank);
    setCardFilter("all");
  };

  const addGroup = () => {
    if (!groupForm.name.trim()) return;
    setFlow((current) => withGroup(current, groupForm));
    setGroupForm(emptyGroupForm());
  };

  const addComponent = () => {
    if (!componentForm.reward_group_id || !componentForm.name.trim()) return;
    setFlow((current) => withComponent(current, componentForm));
    setComponentForm(emptyComponentForm(componentForm.reward_group_id));
  };

  const addRequirement = () => {
    if (!requirementForm.reward_component_id || !requirementForm.values.trim())
      return;
    setFlow((current) => withRequirement(current, requirementForm));
    setRequirementForm(
      emptyRequirementForm(requirementForm.reward_component_id),
    );
  };

  const addBenefit = () => {
    if (!benefitForm.reward_component_id || !benefitForm.value.trim()) return;
    setFlow((current) => withBenefit(current, benefitForm));
    setBenefitForm(emptyBenefitForm(benefitForm.reward_component_id));
  };

  const openActivityEditor = (activityID = flow.activity.id) => {
    setEditorMode("edit");
    setSelection({ type: "activity", id: activityID });
    setViewMode("editor");
  };

  const finishActivityEdit = () => {
    setEditorMode("edit");
    setSelection({ type: "activity", id: flow.activity.id });
    setViewMode("overview");
  };

  const reset = () => {
    const next = emptyMockActivity();
    setFlow(next);
    setEditorMode("create");
    setSelection({ type: "activity", id: next.activity.id });
    resetForms();
    setViewMode("editor");
  };

  return {
    addBenefit,
    addComponent,
    addGroup,
    addRequirement,
    bankFilter,
    bankOptions,
    benefitForm,
    capPeriodOptions,
    cardFilter,
    cardOptions,
    checks,
    componentForm,
    components,
    deleteSelected,
    editorMode,
    finishActivityEdit,
    flow,
    groupForm,
    openActivityEditor,
    overviewRows,
    requirementForm,
    reset,
    select,
    selectedBenefit,
    selectedComponent,
    selectedGroup,
    selectedRequirement,
    selection,
    setBank,
    setBenefitForm,
    setCardFilter,
    setComponentForm,
    setFlow,
    setGroupForm,
    setRequirementForm,
    setViewMode,
    viewMode,
  };
}