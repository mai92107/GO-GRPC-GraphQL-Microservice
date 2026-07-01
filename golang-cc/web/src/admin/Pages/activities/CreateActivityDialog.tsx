import { FormEvent } from "react";
import { Dialog } from "../../../components";
import type { CatalogCard, Merchant, PaymentMethod, Unit } from "../../../models";
import type { ActivityBenefitInput, Category } from "../../AdminApi";
import { ActivityBasicFields } from "./ActivityBasicFields";
import { ActivityBenefitFields } from "./ActivityBenefitFields";
import type { ActivityForm, BenefitMode } from "./types";

type Props = {
  allNetworksSelected: boolean;
  availableCards: CatalogCard[];
  bankOptions: string[];
  benefits: ActivityBenefitInput[];
  categories: Category[];
  creating: boolean;
  error: string;
  form: ActivityForm;
  merchants: Merchant[];
  methods: PaymentMethod[];
  qualifiedTypes: string[];
  selectableTypes: string[];
  selectedCard?: CatalogCard;
  selectedNetworks: string[];
  unavailableTypes: Set<string>;
  units: Unit[];
  onAddBenefit: () => void;
  onChange: (form: ActivityForm) => void;
  onClose: () => void;
  onModeChange: (mode: BenefitMode) => void;
  onRemoveBenefit: (index: number) => void;
  onSubmit: (event: FormEvent) => void;
};

export function CreateActivityDialog(props: Props) {
  const benefitCount =
    props.benefits.length + (props.form.benefit_name.trim() ? 1 : 0);
  const disabled =
    props.creating ||
    !props.form.card_product_id ||
    !props.availableCards.length ||
    !props.form.networks.length ||
    !props.form.name ||
    !props.form.start_date ||
    !props.form.end_date ||
    (!props.benefits.length && !props.form.benefit_name.trim());

  return (
    <Dialog title="新增回饋方案" onClose={props.onClose}>
      {props.error && <div className="error preference-message">{props.error}</div>}
      <form className="stack activity-form" onSubmit={props.onSubmit}>
        <ActivityBasicFields
          allNetworksSelected={props.allNetworksSelected}
          availableCards={props.availableCards}
          bankOptions={props.bankOptions}
          creating={props.creating}
          form={props.form}
          onChange={props.onChange}
          selectedCard={props.selectedCard}
          selectedNetworks={props.selectedNetworks}
        />
        <ActivityBenefitFields
          benefits={props.benefits}
          categories={props.categories}
          creating={props.creating}
          form={props.form}
          merchants={props.merchants}
          methods={props.methods}
          onAdd={props.onAddBenefit}
          onChange={props.onChange}
          onModeChange={props.onModeChange}
          onRemove={props.onRemoveBenefit}
          qualifiedTypes={props.qualifiedTypes}
          selectableTypes={props.selectableTypes}
          unavailableTypes={props.unavailableTypes}
          units={props.units}
        />
        <div className="dialog-actions">
          <button type="button" className="button ghost" onClick={props.onClose}>
            取消
          </button>
          <button className="button" disabled={disabled}>
            {props.creating ? "建立中…" : `建立方案（${benefitCount} 項優惠）`}
          </button>
        </div>
      </form>
    </Dialog>
  );
}
