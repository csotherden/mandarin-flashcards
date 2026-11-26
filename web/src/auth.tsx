import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { api, clearSession, saveSession, token } from "./api";
import type { LoginResponse, User } from "./types";

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
      .then(setUser)
      .catch(() => clearSession())
      .finally(() => setReady(true));
  }, []);

  async function login(email: string, password: string) {
    const res = await api<LoginResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    saveSession(res);
    setUser(res.user);
  }

  async function register(email: string, password: string) {
    const res = await api<LoginResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
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
