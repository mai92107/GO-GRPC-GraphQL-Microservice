import {
  createRegion,
  deleteRegion,
  getRegions,
  updateRegion,
} from "../../AdminApi";
import LookupCollection from "./LookupCollection";

export default function Regions() {
  return (
    <LookupCollection
      createItem={createRegion}
      deleteItem={deleteRegion}
      emptyText="管理活動條件可選擇的地區。"
      icon="region"
      itemLabel="地區"
      loadItems={getRegions}
      searchPlaceholder="搜尋地區"
      title="地區"
      updateItem={updateRegion}
    />
  );
}
