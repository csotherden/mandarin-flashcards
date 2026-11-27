import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { api, clearSession, saveSession, token } from "./api";
import type { LoginResponse, User } from "./types";

/** The device's IANA zone. The server keeps it so streak days are the user's own calendar days. */
export function deviceTimezone() {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone ?? "";
  } catch {
    return "";
  }
}

type AuthValue = {
  user: User | null;
  ready: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!token()) {
      setReady(true);
      return;
    }
    api<User>("/api/user/profile")
      .then(async (profile) => {
        // Keep the saved zone in step with this device (new account flow, travel, older accounts).
        const zone = deviceTimezone();
        if (zone && zone !== profile.timezone) {
          profile = await api<User>("/api/user/timezone", { method: "PUT", body: JSON.stringify({ timezone: zone }) }).catch(() => profile);
        }
        setUser(profile);
      })
      .catch(() => clearSession())
      .finally(() => setReady(true));
  }, []);

  async function login(email: string, password: string) {
    const res = await api<LoginResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password, timezone: deviceTimezone() }),
    });
    saveSession(res);
    setUser(res.user);
  }

  async function register(email: string, password: string) {
    const res = await api<LoginResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password, timezone: deviceTimezone() }),
    });
    saveSession(res);
    setUser(res.user);
  }

  function logout() {
    clearSession();
    setUser(null);
  }

  return <AuthContext.Provider value={{ user, ready, login, register, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("AuthProvider missing");
  return value;
}
