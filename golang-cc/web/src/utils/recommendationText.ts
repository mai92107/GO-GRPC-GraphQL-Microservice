import type { RewardLayer } from "../models";

export const todayInTaipei = () =>
  new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Taipei",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());

export function totalLayerRate(layers: RewardLayer[]) {
  return String(layers.reduce((sum, layer) => sum + Number(layer.reward_rate), 0));
}

export function totalLayerReward(layers: RewardLayer[]) {
  return String(layers.reduce((sum, layer) => sum + Number(layer.reward_amount), 0));
}

export function excludeReasonLabel(reason: string) {
  if (reason === "STACK_GROUP_LOST") return "同組活動未採用";
  if (reason === "CAP_REACHED") return "已達回饋上限";
  return reason;
}
