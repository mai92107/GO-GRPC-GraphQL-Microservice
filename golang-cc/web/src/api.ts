

let csrf = "";
export const setCSRF = (value: string) => {
  csrf = value;
};

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (options.method && options.method !== "GET" && csrf)
    headers.set("X-CSRF-Token", csrf);
  const response = await fetch(`/api${path}`, {
    ...options,
    headers,
    credentials: "include",
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.error?.message || "發生未預期錯誤");
  return payload.data as T;
}

export const post = <T>(path: string, body?: unknown) =>
  api<T>(path, {
    method: "POST",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
export const patch = <T>(path: string, body?: unknown) =>
  api<T>(path, {
    method: "PATCH",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
export const del = <T>(path: string) => 
  api<T>(path, { 
    method: "DELETE" 
  });