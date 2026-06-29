import { Empty } from "../../../../components";

export function EmptyState() {
  return (
    <div className="panel">
      <Empty title="卡片夾是空的" text="使用上方按鈕加入目前持有的卡片。" />
    </div>
  );
}
