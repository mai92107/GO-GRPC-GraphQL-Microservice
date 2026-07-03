import { ComponentType, useState } from "react";
import Head from "../../tool/Head";
import CapPeriods from "./setting/CapPeriods";
import Categories from "./setting/Categories";
import Merchants from "./setting/Merchant";
import PaymentMethods from "./setting/Payment";
import Regions from "./setting/Regions";
import Units from "./setting/Unit";
import UserQualifications from "./setting/UserQualifications";

type Tab = "categories" | "units" | "cap_periods" | "payments" | "merchants" | "regions" | "user_qualifications";

const tabs: { id: Tab; label: string; component: ComponentType }[] = [
  { id: "categories", label: "消費類別", component: Categories },
  { id: "units", label: "回饋單位", component: Units },
  { id: "cap_periods", label: "上限週期", component: CapPeriods },
  { id: "payments", label: "支付方式", component: PaymentMethods },
  { id: "merchants", label: "店家", component: Merchants },
  { id: "regions", label: "地區", component: Regions },
  { id: "user_qualifications", label: "會員資格", component: UserQualifications },
];

export default function AdminSettings() {
  const [tab, setTab] = useState<Tab>("categories");
  const Page = tabs.find((item) => item.id === tab)?.component || Categories;

  return (
    <>
      <Head title="系統設定" text="消費類別、回饋單位、上限週期、支付方式、店家、地區與會員資格管理。" />
      <div className="toolbar settings-tabs" role="tablist" aria-label="系統設定">
        {tabs.map((item) => (
          <button
            type="button"
            role="tab"
            aria-selected={tab === item.id}
            className={`button ${tab === item.id ? "" : "ghost"}`}
            key={item.id}
            onClick={() => setTab(item.id)}
          >
            {item.label}
          </button>
        ))}
      </div>
      <Page />
    </>
  );
}
