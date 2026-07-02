import Head from "../../tool/Head";
import { ActivityMockFlow } from "./activities/ActivityMockFlow";

export default function Activities() {
  return (
    <>
      <Head
        title="活動回饋"
        text="先以 Mock Flow 重構 Activity / Group / Component / Requirement / Benefit 建檔流程，確認五層資料結構後再進後端 schema 與 API。"
      />
      <ActivityMockFlow />
    </>
  );
}

