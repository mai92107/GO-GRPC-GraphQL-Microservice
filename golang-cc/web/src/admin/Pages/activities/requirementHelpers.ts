import type { ActivityRequirementOperator, ActivityRequirementType } from "./activityFlowTypes";

export const unconditionalRequirementType = "none";
export const unconditionalRequirementOperator: ActivityRequirementOperator = "EQ";

export function isUnconditionalRequirement(type: string) {
  return type.toLowerCase() === unconditionalRequirementType;
}

export function normalizeRequirementFormType(type: ActivityRequirementType) {
  return isUnconditionalRequirement(type)
    ? {
        requirement_type: type,
        operator: unconditionalRequirementOperator,
        values: "",
        description: "",
      }
    : {
        requirement_type: type,
        values: "",
        description: "",
      };
}
