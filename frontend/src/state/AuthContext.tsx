import { useCallback, useEffect, useState, type ReactNode } from "react";
import { api, authApi, setAccessToken } from "../services/api";
import type { User } from "../types";
import { AuthContext } from "./auth-context";
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const refreshUser = useCallback(async () => {
    setUser(await api<User>("/me"));
  }, []);
  const hydrate = useCallback(async () => {
    try {
      const ok = await authApi.refresh();
      if (ok) {
        await refreshUser();
      }
    } catch {
      setAccessToken(null);
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, [refreshUser]);
  useEffect(() => {
    void hydrate();
  }, [hydrate]);
  useEffect(() => {
    const expired = () => setUser(null);
    window.addEventListener("kc:session-expired", expired);
    return () => window.removeEventListener("kc:session-expired", expired);
  }, []);
  const login = async (
    identifier: string,
    password: string,
    rememberMe: boolean,
  ) => {
    const data = await authApi.login({ identifier, password, rememberMe });
    setAccessToken(data.accessToken);
    setUser(data.user);
  };
  const register = async (body: unknown) => {
    const data = await authApi.register(body);
    setAccessToken(data.accessToken);
    setUser(data.user);
  };
  const updateProfile = async (body: unknown) => {
    const updated = await api<User>("/me", {
      method: "PATCH",
      body: JSON.stringify(body),
    });
    setUser(updated);
    return updated;
  };
  const logout = async () => {
    setAccessToken(null);
    setUser(null);
    await authApi.logout();
  };
  const value = {
    user,
    loading,
    login,
    register,
    updateProfile,
    refreshUser,
    logout,
  };
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
