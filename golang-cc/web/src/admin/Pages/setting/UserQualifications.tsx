import {
  createUserQualification,
  deleteUserQualification,
  getUserQualifications,
  updateUserQualification,
} from "../../AdminApi";
import LookupCollection from "./LookupCollection";

export default function UserQualifications() {
  return (
    <LookupCollection
      createItem={createUserQualification}
      deleteItem={deleteUserQualification}
      emptyText="管理活動條件可選擇的會員資格。"
      icon="qualification"
      itemLabel="會員資格"
      loadItems={getUserQualifications}
      searchPlaceholder="搜尋會員資格"
      title="會員資格"
      updateItem={updateUserQualification}
    />
  );
}
