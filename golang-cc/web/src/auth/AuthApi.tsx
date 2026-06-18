import { api, post } from "../api";
import type { User } from "../models";

export type AuthSession = {
  user: User;
  csrf_token: string;
};

export const getCurrentUser = () =>
  api<AuthSession>("/public/auth/me");

export const login = (email: string, password: string) =>
  post<AuthSession>("/public/auth/login", { email, password });

export const logout = () =>
  post<{ logged_out: boolean }>("/public/auth/logout");

export const acceptInvitation = (
  token: string,
  displayName: string,
  password: string,
) =>
  post<User>("/public/auth/accept-invitation", {
    token,
    display_name: displayName,
    password,
  });

export const requestPasswordReset = (email: string) =>
  post<{ accepted: boolean }>("/public/auth/request-password-reset", {
    email,
  });

export const resetPassword = (token: string, password: string) =>
  post<{ reset: boolean }>("/public/auth/reset-password", {
    token,
    password,
  });
