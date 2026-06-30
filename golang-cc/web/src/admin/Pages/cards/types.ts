export type CardForm = {
  bank_id: string;
  name: string;
  qualified_type: string;
  selectable_type: string;
  networks: string[];
};

export const emptyForm = (
  bankID = "",
  networkIDs: string[] = [],
): CardForm => ({
  bank_id: bankID,
  name: "",
  qualified_type: "",
  selectable_type: "",
  networks: networkIDs,
});

export const toggleNetwork = (
  current: string[],
  networkID: string,
  checked: boolean,
) =>
  checked
    ? [...new Set([...current, networkID])]
    : current.filter((value) => value !== networkID);
