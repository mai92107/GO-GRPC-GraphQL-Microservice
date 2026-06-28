import { ArrowLeft, Plus } from "lucide-react";
import type { Card } from "../../../models";

type Props = {
  catalogCount: number;
  detail: Card | null;
  loading: boolean;
  onBack: () => void;
  onCreate: () => void;
};

export function CardPageHeader({
  catalogCount,
  detail,
  loading,
  onBack,
  onCreate,
}: Props) {
  const text = detail
    ? `${detail.issuer} · 查看卡片資訊、優惠資格與目前回饋。`
    : "選擇一張卡片，查看卡片資訊與目前可用回饋。";

  return (
    <div className="page-head">
      <div>
        <span className="page-eyebrow">MY WALLET</span>
        <p>{text}</p>
      </div>
      {detail ? (
        <button
          type="button"
          className="button ghost wallet-back-button"
          onClick={onBack}
        >
          <ArrowLeft size={17} />
          返回卡片夾
        </button>
      ) : (
        <button
          type="button"
          className="button"
          disabled={loading || !catalogCount}
          onClick={onCreate}
        >
          <Plus size={17} />
          加入卡片
        </button>
      )}
    </div>
  );
}
