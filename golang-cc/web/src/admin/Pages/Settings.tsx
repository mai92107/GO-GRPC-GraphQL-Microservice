import { ComponentType, useState } from "react";
import Head from "../../tool/Head";
import Categories from "./setting/Categories";
import Merchants from "./setting/Merchant";
import PaymentMethods from "./setting/Payment";
import Units from "./setting/Unit";

type Tab = "categories" | "units" | "payments" | "merchants";

const tabs: { id: Tab; label: string; component: ComponentType }[] = [
  { id: "categories", label: "消費類別", component: Categories },
  { id: "units", label: "回饋單位", component: Units },
  { id: "payments", label: "支付方式", component: PaymentMethods },
  { id: "merchants", label: "店家", component: Merchants },
];

export default function AdminSettings() {
  const [tab, setTab] = useState<Tab>("categories");
  const Page = tabs.find((item) => item.id === tab)?.component || Categories;

  return (
    <>
      <Head title="系統設定" text="消費類別、回饋單位、支付方式與店家管理。" />
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
