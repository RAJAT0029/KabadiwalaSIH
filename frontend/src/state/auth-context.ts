import { createContext, useContext } from "react";
import type { User } from "../types";

export type AuthContextValue = {
  user: User | null;
  loading: boolean;
  login: (
    identifier: string,
    password: string,
    rememberMe: boolean,
  ) => Promise<void>;
  register: (body: unknown) => Promise<void>;
  updateProfile: (body: unknown) => Promise<User>;
  refreshUser: () => Promise<void>;
  logout: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("useAuth must be used within AuthProvider");
  return value;
}
