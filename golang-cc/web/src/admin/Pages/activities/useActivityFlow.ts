import { useEffect, useMemo, useRef, useState } from "react";
import {
  createActivity,
  getActivities,
  getActivity,
  getActivityRequirementOptions,
  getCards,
  getRequirementTypes,
  getRewardUnits,
  publishActivity as publishActivityRequest,
  updateActivity,
  type ActivityRequirementOptions,
  type RequirementTypeOption,
} from "../../AdminApi";
import type { Unit } from "../../../models";
import {
  allComponents,
  emptyBenefitForm,
  emptyComponentForm,
  emptyGroupForm,
  emptyActivityFlow,
  emptyRequirementForm,
  validateActivityFlow,
  withBenefit,
  withComponent,
  withGroup,
  withRequirement,
} from "./activityFlowFactory";
import { defaultCapPeriodOptions } from "./activityFlowSettings";
import { deleteSelection } from "./activityFlowHelpers";
import {
  isUnconditionalRequirement,
  normalizeRequirementFormType,
  unconditionalRequirementOperator,
} from "./requirementHelpers";
import type {
  ActivityFlowSelection,
  ActivityFlowViewMode,
  ActivityOverviewRow,
  ActivityFlowModel,
  ActivityBenefitForm,
  ActivityComponentForm,
  ActivityGroupForm,
  ActivityRequirementForm,
} from "./activityFlowTypes";

type CatalogCardOption = {
  bank_id: string;
  bank_name: string;
  card_product_id: string;
  card_name: string;
};
const initialSelection: ActivityFlowSelection = { type: "activity", id: "" };
type ActivityEditorMode = "create" | "edit";

export function useActivityFlow() {
  const [flow, setFlow] = useState<ActivityFlowModel>(emptyActivityFlow);
  const [overviewRows, setOverviewRows] = useState<ActivityOverviewRow[]>([]);
  const [viewMode, setViewMode] = useState<ActivityFlowViewMode>("overview");
  const [editorMode, setEditorMode] = useState<ActivityEditorMode>("edit");
  const [bankFilter, setBankFilter] = useState("all");
  const [cardFilter, setCardFilter] = useState("all");
  const [selection, setSelection] =
    useState<ActivityFlowSelection>(initialSelection);
  const [groupForm, setGroupForm] = useState<ActivityGroupForm>(emptyGroupForm);
  const [componentForm, setComponentForm] = useState<ActivityComponentForm>(() =>
    emptyComponentForm(),
  );
  const [requirementForm, setRequirementFormState] =
    useState<ActivityRequirementForm>(() => emptyRequirementForm());
  const [benefitForm, setBenefitFormState] = useState<ActivityBenefitForm>(() =>
    emptyBenefitForm(),
  );
  const [requirementTypes, setRequirementTypes] = useState<
    RequirementTypeOption[]
  >([]);
  const [requirementOptions, setRequirementOptions] =
    useState<ActivityRequirementOptions | null>(null);
  const [rewardUnits, setRewardUnits] = useState<Unit[]>([]);
  const [catalogCards, setCatalogCards] = useState<CatalogCardOption[]>([]);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState("");
  const previousActivityDates = useRef({ effective_from: "", effective_to: "" });

  const refreshOverview = async () => {
    const rows = await getActivities({
      bank_id: bankFilter === "all" ? undefined : bankFilter,
      card_product_id: cardFilter === "all" ? undefined : cardFilter,
    });
    setOverviewRows(rows);
  };

  useEffect(() => {
    refreshOverview().catch((err) => setError(err.message));
  }, [bankFilter, cardFilter]);

  useEffect(() => {
    getRewardUnits()
      .then(setRewardUnits)
      .catch((err) => setError(err.message));
    getCards()
      .then((cards) =>
        setCatalogCards(
          cards.map((card) => ({
            bank_id: card.bank_id,
            bank_name: card.bank_name,
            card_product_id: card.id,
            card_name: card.name,
          })),
        ),
      )
      .catch((err) => {
        setCatalogCards([]);
        setError(err.message);
      });
  }, []);

  const components = useMemo(() => allComponents(flow), [flow]);
  const requirements = useMemo(
    () => components.flatMap((component) => component.requirements),
    [components],
  );
  const benefits = useMemo(
    () => components.flatMap((component) => component.benefits),
    [components],
  );
  const showComponentForm = flow.reward_groups.length > 0;
  const showRequirementForm = components.length > 0;
  const showBenefitForm = requirements.length > 0;

  useEffect(() => {
    const firstGroup = flow.reward_groups[0];
    if (!firstGroup) return;
    if (
      componentForm.reward_group_ids.length > 0 &&
      componentForm.reward_group_ids.every((groupID) =>
        flow.reward_groups.some((group) => group.id === groupID),
      )
    )
      return;
    setComponentForm((current) => ({
      ...current,
      reward_group_ids: [firstGroup.id],
      effective_from: current.effective_from || flow.activity.effective_from,
      effective_to: current.effective_to || flow.activity.effective_to,
    }));
  }, [flow.reward_groups, componentForm.reward_group_ids, flow.activity.effective_from, flow.activity.effective_to]);

  useEffect(() => {
    const previous = previousActivityDates.current;
    setComponentForm((current) => ({
      ...current,
      effective_from:
        !current.effective_from || current.effective_from === previous.effective_from
          ? flow.activity.effective_from
          : current.effective_from,
      effective_to:
        !current.effective_to || current.effective_to === previous.effective_to
          ? flow.activity.effective_to
          : current.effective_to,
    }));
    previousActivityDates.current = {
      effective_from: flow.activity.effective_from,
      effective_to: flow.activity.effective_to,
    };
  }, [flow.activity.effective_from, flow.activity.effective_to]);

  useEffect(() => {
    const firstComponent = components[0];
    if (!firstComponent) return;
    if (
      requirementForm.reward_component_id &&
      components.some(
        (component) => component.id === requirementForm.reward_component_id,
      )
    )
      return;
    setRequirementFormState((current) => ({
      ...current,
      reward_component_id: firstComponent.id,
    }));
  }, [components, requirementForm.reward_component_id]);

  useEffect(() => {
    const firstComponent = components[0];
    if (!firstComponent) return;
    if (
      benefitForm.reward_component_id &&
      components.some(
        (component) => component.id === benefitForm.reward_component_id,
      )
    )
      return;
    setBenefitFormState((current) => ({
      ...current,
      reward_component_id: firstComponent.id,
    }));
  }, [components, benefitForm.reward_component_id]);

  useEffect(() => {
    if (
      viewMode !== "editor" ||
      !showRequirementForm ||
      requirementTypes.length
    )
      return;
    getRequirementTypes()
      .then((types) => {
        setRequirementTypes(types);
        const firstType = types[0]?.code as
          | ActivityRequirementForm["requirement_type"]
          | undefined;
        if (firstType)
          setRequirementFormState((current) => ({
            ...emptyRequirementForm(current.reward_component_id),
            ...normalizeRequirementFormType(firstType),
          }));
      })
      .catch((err) => setError(err.message));
  }, [viewMode, showRequirementForm, requirementTypes.length]);

  useEffect(() => {
    if (!rewardUnits.length || benefitForm.reward_unit_id) return;
    setBenefitFormState((current) => ({
      ...current,
      reward_unit_id: rewardUnits[0].id,
    }));
  }, [rewardUnits, benefitForm.reward_unit_id]);

  const bankOptions = useMemo(
    () => [
      ...new Map(catalogCards.map((card) => [card.bank_id, card])).values(),
    ],
    [catalogCards],
  );
  const cardOptions = useMemo(
    () =>
      catalogCards.filter(
        (card) => bankFilter === "all" || card.bank_id === bankFilter,
      ),
    [bankFilter, catalogCards],
  );
  const activityCardOptions = useMemo(
    () =>
      catalogCards.filter(
        (card) =>
          !flow.activity.bank_id || card.bank_id === flow.activity.bank_id,
      ),
    [catalogCards, flow.activity.bank_id],
  );
  const checks = useMemo(() => validateActivityFlow(flow), [flow]);
  const canSave = checks.every((check) => check.pass) && !isSaving;

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
      ? requirements.find((requirement) => requirement.id === selection.id)
      : undefined;
  const selectedBenefit =
    selection.type === "benefit"
      ? benefits.find((benefit) => benefit.id === selection.id)
      : undefined;
  const activeRequirementType =
    selection.type === "requirement" && selectedRequirement
      ? selectedRequirement.requirement_type
      : requirementForm.requirement_type;

  useEffect(() => {
    if (viewMode !== "editor" || !showRequirementForm || !activeRequirementType)
      return;
    let ignore = false;
    setRequirementOptions(null);
    getActivityRequirementOptions(
      activeRequirementType,
      flow.activity.card_product_id,
    )
      .then((options) => {
        if (!ignore) setRequirementOptions(options);
      })
      .catch((err) => setError(err.message));
    return () => {
      ignore = true;
    };
  }, [
    viewMode,
    showRequirementForm,
    activeRequirementType,
    flow.activity.card_product_id,
  ]);

  const resetForms = (componentID = "", groupID = "") => {
    setGroupForm(emptyGroupForm());
    setComponentForm(
      emptyComponentForm(
        groupID,
        flow.activity.effective_from,
        flow.activity.effective_to,
      ),
    );
    setRequirementFormState(emptyRequirementForm(componentID));
    setBenefitFormState(emptyBenefitForm(componentID));
  };
  const setRequirementForm = (next: ActivityRequirementForm) =>
    setRequirementFormState((current) =>
      next.requirement_type !== current.requirement_type
        ? {
            ...next,
            ...normalizeRequirementFormType(next.requirement_type),
          }
        : next,
    );
  const setBenefitForm = (next: ActivityBenefitForm) => setBenefitFormState(next);

  const select = (nextSelection: ActivityFlowSelection) => {
    setSelection(nextSelection);
    if (nextSelection.type === "component") {
      setRequirementFormState(emptyRequirementForm(nextSelection.id));
      setBenefitFormState(emptyBenefitForm(nextSelection.id));
    }
  };
  const deleteSelected = (
    target: ActivityFlowSelection,
    nextSelection: ActivityFlowSelection,
  ) => {
    setFlow((current) => deleteSelection(current, target));
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
    if (!componentForm.reward_group_ids.length || !componentForm.name.trim()) return;
    setFlow((current) => withComponent(current, componentForm));
    setComponentForm(
      emptyComponentForm(
        componentForm.reward_group_ids[0],
        flow.activity.effective_from,
        flow.activity.effective_to,
      ),
    );
  };
  const addRequirement = () => {
    if (
      !requirementForm.reward_component_id ||
      (!isUnconditionalRequirement(requirementForm.requirement_type) &&
        !requirementForm.values.trim())
    )
      return;
    setFlow((current) => withRequirement(current, requirementForm));
    setRequirementFormState(
      emptyRequirementForm(requirementForm.reward_component_id),
    );
  };
  const addBenefit = () => {
    if (
      !benefitForm.reward_component_id ||
      !benefitForm.value.trim() ||
      !benefitForm.reward_unit_id
    )
      return;
    setFlow((current) => withBenefit(current, benefitForm));
    setBenefitFormState(emptyBenefitForm(benefitForm.reward_component_id));
  };

  const openActivityEditor = async (activityID = flow.activity.id) => {
    setEditorMode("edit");
    setError("");
    const detail = await getActivity(activityID);
    setFlow(normalizeFlow(detail as unknown as ActivityFlowModel));
    setSelection({ type: "activity", id: detail.activity.id });
    setViewMode("editor");
  };
  const saveActivity = async () => {
    if (!canSave) return;
    setIsSaving(true);
    setError("");
    try {
      if (editorMode === "create") await createActivity(flow as never);
      else await updateActivity(flow as never);
      await refreshOverview();
      setEditorMode("edit");
      setViewMode("overview");
    } catch (err) {
      setError(err instanceof Error ? err.message : "儲存失敗");
    } finally {
      setIsSaving(false);
    }
  };
  const publishActivity = async (activityID = flow.activity.id) => {
    if (!activityID || editorMode === "create") return;
    setIsSaving(true);
    setError("");
    try {
      await publishActivityRequest(activityID);
      await refreshOverview();
      if (viewMode === "editor") await openActivityEditor(activityID);
    } catch (err) {
      setError(err instanceof Error ? err.message : "發布失敗");
    } finally {
      setIsSaving(false);
    }
  };
  const reset = () => {
    const next = emptyActivityFlow();
    const firstCard = catalogCards[0];
    if (firstCard)
      Object.assign(next.activity, {
        bank_id: firstCard.bank_id,
        bank_name: firstCard.bank_name,
        card_product_id: firstCard.card_product_id,
        card_name: firstCard.card_name,
      });
    setFlow(next);
    previousActivityDates.current = {
      effective_from: next.activity.effective_from,
      effective_to: next.activity.effective_to,
    };
    setEditorMode("create");
    setSelection({ type: "activity", id: next.activity.id });
    resetForms();
    setViewMode("editor");
  };
  const setActivityCard = (cardProductID: string) => {
    const card = catalogCards.find(
      (item) => item.card_product_id === cardProductID,
    );
    if (!card) return;
    setFlow((current) => ({
      ...current,
      activity: {
        ...current.activity,
        bank_id: card.bank_id,
        bank_name: card.bank_name,
        card_product_id: card.card_product_id,
        card_name: card.card_name,
      },
    }));
    setRequirementFormState((current) => ({
      ...current,
      values: "",
      description: "",
    }));
  };
  const setActivityBank = (bankID: string) => {
    const firstCard = catalogCards.find((card) => card.bank_id === bankID);
    if (firstCard) setActivityCard(firstCard.card_product_id);
  };

  return {
    addBenefit,
    addComponent,
    addGroup,
    addRequirement,
    activityCardOptions,
    bankFilter,
    bankOptions,
    benefitForm,
    canSave,
    capPeriodOptions: defaultCapPeriodOptions,
    cardFilter,
    cardOptions,
    checks,
    componentForm,
    components,
    deleteSelected,
    editorMode,
    error,
    flow,
    groupForm,
    isSaving,
    openActivityEditor,
    overviewRows,
    publishActivity,
    requirementForm,
    requirementOptions,
    requirementTypes,
    reset,
    rewardUnits,
    saveActivity,
    select,
    selectedBenefit,
    selectedComponent,
    selectedGroup,
    selectedRequirement,
    selection,
    setActivityBank,
    setActivityCard,
    setBank,
    setBenefitForm,
    setCardFilter,
    setComponentForm,
    setFlow,
    setGroupForm,
    setRequirementForm,
    setViewMode,
    showBenefitForm,
    showComponentForm,
    showRequirementForm,
    viewMode,
  };
}

function normalizeFlow(flow: ActivityFlowModel): ActivityFlowModel {
  return {
    ...flow,
    reward_groups: flow.reward_groups.map((group) => ({
      ...group,
      components: group.components.map((component) => ({
        ...component,
        is_exclusive: component.stack_mode === "EXCLUSIVE",
        is_best_only: component.stack_mode === "BEST_ONLY",
        requirements: component.requirements.map((requirement) =>
          isUnconditionalRequirement(requirement.requirement_type)
            ? {
                ...requirement,
                operator: unconditionalRequirementOperator,
                configuration_json: {},
              }
            : requirement,
        ),
      })),
    })),
  };
}
