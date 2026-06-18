import { del, patch, post } from "../api";

export async function invite(email: string) {
    await post("/admin/invitations",{email});
}   
export async function toggleActivate(userId: number, status: string) {
    await patch(`/admin/users/${userId}`, {
      status: status === "active" ? "disabled" : "active",
    });
}
export async function toggleReset(userId: number) {
    await post(`/admin/users/${userId}/password-reset`);
}
export async function postTelegramBinding(chatID: number, userID: string) {
    await post("/admin/telegram-bindings",{chat_id:chatID,user_id:userID});
}
export async function deleteTelegramBinding(bindingID: number) {
    await patch(`/admin/telegram-bindings/${bindingID}`,{status:"deleted"});
}
export async function createBank(name: string) {
    await post("/admin/banks",{name,is_active:true});
}
export async function activateBank(id: string, name: string, code: string, website_url: string, is_active: boolean) {
    await patch(`/admin/banks/${id}`,{name,code,website_url,is_active:!is_active});
}
export async function deleteBank(id: string) {
    await del(`/admin/banks/${id}`);
}