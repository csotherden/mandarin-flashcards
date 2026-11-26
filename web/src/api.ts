import type { LoginResponse } from "./types";

const accessKey = "mandarin.access";
const refreshKey = "mandarin.refresh";

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(message: string, status: number, code = "") {
    super(message);
    this.status = status;
    this.code = code;
  }
}

export function token() {
  return localStorage.getItem(accessKey);
}

export function saveSession(res: LoginResponse) {
  localStorage.setItem(accessKey, res.access_token);
  localStorage.setItem(refreshKey, res.refresh_token);
}

export function clearSession() {
  localStorage.removeItem(accessKey);
  localStorage.removeItem(refreshKey);
}

async function refresh(): Promise<boolean> {
  const refreshToken = localStorage.getItem(refreshKey);
  if (!refreshToken) return false;
  const res = await fetch("/api/auth/refresh", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  if (!res.ok) {
    clearSession();
    return false;
  }
  const body = (await res.json()) as { access_token: string; refresh_token: string };
  localStorage.setItem(accessKey, body.access_token);
  localStorage.setItem(refreshKey, body.refresh_token);
  return true;
}

async function send(path: string, init: RequestInit, retry: boolean): Promise<Response> {
  const headers = new Headers(init.headers);
  const access = token();
  if (access) headers.set("Authorization", `Bearer ${access}`);
  if (init.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const res = await fetch(path, { ...init, headers });
  if (res.status === 401 && retry && !path.startsWith("/api/auth/")) {
    if (await refresh()) return send(path, init, false);
  }
  return res;
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await send(path, init, true);
  if (res.status === 204) return undefined as T;
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const message = typeof body.message === "string" ? body.message : "Request failed";
    throw new ApiError(message, res.status, typeof body.code === "string" ? body.code : "");
  }
  return body as T;
}

export async function apiBlob(path: string): Promise<string> {
  const res = await send(path, {}, true);
  if (!res.ok) throw new ApiError("Could not load image", res.status);
  const blob = await res.blob();
  return URL.createObjectURL(blob);
}
