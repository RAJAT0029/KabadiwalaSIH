const BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";
let accessToken: string | null = null;
let revision = 0;
let refreshInFlight: Promise<boolean> | null = null;
export const setAccessToken = (token: string | null) => {
  accessToken = token;
  revision++;
};
export const getAccessToken = () => accessToken;

type ErrorPayload = { error?: { code?: string; message?: string } };
function expireSession() {
  setAccessToken(null);
  window.dispatchEvent(new Event("kc:session-expired"));
}
function refreshAccessToken(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;
  const startedAt = revision;
  refreshInFlight = (async () => {
    try {
      const response = await fetch(`${BASE_URL}/auth/refresh`, {
        method: "POST",
        credentials: "include",
      });
      if (startedAt !== revision) return accessToken !== null;
      if (!response.ok) {
        expireSession();
        return false;
      }
      const payload = (await response.json()) as {
        data: { accessToken: string };
      };
      if (startedAt !== revision) return accessToken !== null;
      setAccessToken(payload.data.accessToken);
      return true;
    } catch (error) {
      if (startedAt === revision) expireSession();
      throw error;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}
export async function api<T>(
  path: string,
  options: RequestInit = {},
  retry = true,
): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.body && !headers.has("Content-Type"))
    headers.set("Content-Type", "application/json");
  const requestToken = accessToken;
  if (requestToken) headers.set("Authorization", `Bearer ${requestToken}`);
  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });
  if (response.status === 401 && !path.startsWith("/auth/")) {
    if (
      retry &&
      ((accessToken && accessToken !== requestToken) ||
        (await refreshAccessToken()))
    )
      return api<T>(path, options, false);
    expireSession();
  }
  if (response.status === 204) return undefined as T;
  const payload = (await response.json().catch(() => ({}))) as ErrorPayload & {
    data?: T;
  };
  if (!response.ok)
    throw new Error(payload.error?.message || "Something went wrong");
  return payload.data as T;
}
export const authApi = {
  login: (body: {
    identifier: string;
    password: string;
    rememberMe: boolean;
  }) =>
    api<{ user: import("../types").User; accessToken: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  register: (body: unknown) =>
    api<{ user: import("../types").User; accessToken: string }>(
      "/auth/register",
      { method: "POST", body: JSON.stringify(body) },
    ),
  logout: () => api<void>("/auth/logout", { method: "POST" }),
  refresh: refreshAccessToken,
};
